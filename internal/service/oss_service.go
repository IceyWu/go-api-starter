package service

import (
	"errors"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"go-api-starter/internal/config"
	"go-api-starter/internal/model"
	"go-api-starter/internal/repository"
	"go-api-starter/pkg/apperrors"
	"go-api-starter/pkg/logger"
	"go-api-starter/pkg/oss"
)

// OSSService handles OSS-related operations including simple uploads and multipart uploads.
// It deliberately does not do any media processing (EXIF / blurhash / transcoding);
// clients are expected to upload finished artifacts.
type OSSService struct {
	db            *gorm.DB
	fileRepo      repository.FileRepositoryInterface
	multipartRepo repository.MultipartRepositoryInterface
	config        *config.OSSConfig
	appEnv        string
	secUIDCache   sync.Map // userID -> secUID cache
}

// NewOSSService creates a new OSSService
func NewOSSService(db *gorm.DB, fileRepo repository.FileRepositoryInterface, multipartRepo repository.MultipartRepositoryInterface, cfg *config.OSSConfig, appEnv string) *OSSService {
	return &OSSService{
		db:            db,
		fileRepo:      fileRepo,
		multipartRepo: multipartRepo,
		config:        cfg,
		appEnv:        appEnv,
	}
}

// ======================
// Simple upload / Token
// ======================

// GetUploadToken returns an upload token with an auto-generated key.
func (s *OSSService) GetUploadToken(userID uint) (*oss.UploadToken, error) {
	return s.GetUploadTokenWithFileName(userID, "")
}

// GetUploadTokenWithFileName returns an upload token that preserves the file extension in the key.
func (s *OSSService) GetUploadTokenWithFileName(userID uint, fileName string) (*oss.UploadToken, error) {
	ext := ""
	if fileName != "" {
		ext = strings.ToLower(filepath.Ext(fileName))
	}

	userSecUID, err := s.getUserSecUID(userID)
	if err != nil {
		return nil, err
	}

	key := s.generateFileKey(ext, userSecUID)

	// Directory prefix used by the policy's starts-with condition.
	dir := ""
	if idx := strings.LastIndex(key, "/"); idx >= 0 {
		dir = key[:idx+1]
	}

	endpoint := s.config.Endpoint
	if strings.HasPrefix(endpoint, s.config.Bucket+".") {
		endpoint = strings.TrimPrefix(endpoint, s.config.Bucket+".")
	}

	token, err := oss.GenerateUploadToken(
		s.config.AccessKeyID,
		s.config.AccessKeySecret,
		s.config.Bucket,
		endpoint,
		dir,
		key,
		s.config.MaxFileSize,
		s.config.TokenExpire,
	)
	if err != nil {
		return nil, apperrors.OSSInitError(err)
	}
	return token, nil
}

// CheckFileExists reports whether a file with the given MD5 already exists
// (scoped to the user when userID > 0, used for instant upload).
func (s *OSSService) CheckFileExists(md5 string, userID uint) (*model.File, bool) {
	var file model.File
	query := s.db.Where("file_md5 = ?", md5)
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if err := query.First(&file).Error; err != nil {
		return nil, false
	}
	return &file, true
}

// SaveFileRecord creates a DB record after a successful direct client-to-OSS upload.
func (s *OSSService) SaveFileRecord(key, md5, fileName string, fileSize int64, userID uint) (*model.File, error) {
	// Dedup by MD5 (scoped to user)
	if existing, ok := s.CheckFileExists(md5, userID); ok {
		if err := s.db.Preload("User").First(existing, existing.ID).Error; err != nil {
			logger.Log.Warnf("failed to preload user for existing file: %v", err)
		}
		return existing, nil
	}

	ext := strings.ToLower(filepath.Ext(fileName))
	contentType := inferContentType(ext)

	file := &model.File{
		UserID:    userID,
		Name:      fileName,
		Type:      contentType,
		FileMd5:   md5,
		Size:      uint(fileSize),
		Key:       key,
		Extension: ext,
	}

	if err := s.db.Create(file).Error; err != nil {
		return nil, apperrors.Internal(err, "failed to save file record")
	}
	if err := s.db.Preload("User").First(file, file.ID).Error; err != nil {
		logger.Log.Warnf("failed to preload user: %v", err)
	}
	return file, nil
}

// ======================
// File queries / updates
// ======================

// GetFileBySecUID returns a file by its SecUID.
func (s *OSSService) GetFileBySecUID(secUID string) (*model.File, error) {
	var file model.File
	if err := s.db.Preload("User").Where("sec_uid = ?", secUID).First(&file).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("file not found")
		}
		return nil, apperrors.Internal(err, "failed to query file")
	}
	return &file, nil
}

// UpdateFile updates mutable fields (name / visibility) on a file.
func (s *OSSService) UpdateFile(secUID string, req *model.UpdateFileRequest) error {
	file, err := s.GetFileBySecUID(secUID)
	if err != nil {
		return err
	}
	if req.Name != nil {
		file.Name = *req.Name
	}
	if req.IsPrivate != nil {
		file.IsPrivate = *req.IsPrivate
	}
	if err := s.db.Save(file).Error; err != nil {
		return apperrors.Internal(err, "failed to update file")
	}
	return nil
}

// ListFiles returns a paginated list of files, optionally filtered by owner and privacy.
func (s *OSSService) ListFiles(userID uint, isPrivate *bool, offset, limit int, sort string) ([]model.File, int64, error) {
	query := s.db.Model(&model.File{})
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if isPrivate != nil {
		query = query.Where("is_private = ?", *isPrivate)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, apperrors.Internal(err, "failed to count files")
	}

	if sort == "" {
		sort = "created_at DESC"
	}

	var files []model.File
	err := query.Preload("User").Offset(offset).Limit(limit).Order(sort).Find(&files).Error
	if err != nil {
		return nil, 0, apperrors.Internal(err, "failed to list files")
	}
	return files, total, nil
}

// DeleteFile removes the file from OSS and deletes the DB record.
func (s *OSSService) DeleteFile(secUID string) error {
	file, err := s.GetFileBySecUID(secUID)
	if err != nil {
		return err
	}

	// Best-effort OSS delete; don't fail the request if cleanup fails.
	if file.Key != "" {
		if err := oss.DeleteFile(file.Key); err != nil {
			logger.Log.Warnf("failed to delete %s from OSS: %v", file.Key, err)
		}
	}

	if err := s.db.Delete(file).Error; err != nil {
		return apperrors.Internal(err, "failed to delete file")
	}
	return nil
}

// ======================
// Multipart upload
// ======================

// MultipartInitResult is returned by InitMultipartUpload.
type MultipartInitResult struct {
	UploadID      string         `json:"upload_id"`
	Key           string         `json:"key"`
	Host          string         `json:"host"`
	TotalParts    int            `json:"total_parts"`
	ChunkSize     int64          `json:"chunk_size"`
	UploadedParts []CompletePart `json:"uploaded_parts"`
}

// PartUploadInfo represents a signed upload URL for a single part.
type PartUploadInfo struct {
	PartNumber int    `json:"part_number"`
	URL        string `json:"url"`
	Expire     int64  `json:"expire"`
}

// CompletePart represents a completed part.
type CompletePart struct {
	PartNumber int    `json:"part_number"`
	ETag       string `json:"etag"`
}

// InitMultipartUpload starts (or resumes) a multipart upload for the given file.
// If a matching upload already exists for this user, it is reused.
func (s *OSSService) InitMultipartUpload(fileName, md5 string, fileSize, chunkSize int64, userID uint) (*MultipartInitResult, error) {
	// Try to resume an existing upload for the same MD5 + user.
	if existing, err := s.multipartRepo.GetUploadByMD5(md5, userID); err == nil && existing != nil {
		uploaded, _ := s.GetUploadedPartsFromDB(existing.UploadID)
		return &MultipartInitResult{
			UploadID:      existing.UploadID,
			Key:           existing.Key,
			Host:          s.buildHost(),
			TotalParts:    existing.TotalParts,
			ChunkSize:     existing.ChunkSize,
			UploadedParts: uploaded,
		}, nil
	}

	userSecUID, err := s.getUserSecUID(userID)
	if err != nil {
		return nil, err
	}
	ext := strings.ToLower(filepath.Ext(fileName))
	key := s.generateFileKey(ext, userSecUID)
	contentType := inferContentType(ext)

	initResp, err := oss.InitMultipartUpload(key, contentType)
	if err != nil {
		return nil, apperrors.Internal(err, "failed to init multipart upload")
	}

	totalParts := int((fileSize + chunkSize - 1) / chunkSize)
	upload := &model.MultipartUpload{
		UploadID:    initResp.UploadID,
		Key:         initResp.Key,
		MD5:         md5,
		FileName:    fileName,
		FileSize:    fileSize,
		ContentType: contentType,
		TotalParts:  totalParts,
		ChunkSize:   chunkSize,
		UserID:      userID,
		Status:      model.MultipartStatusInitiated,
	}
	if err := s.multipartRepo.CreateUpload(upload); err != nil {
		// Upload already created in OSS — abort to avoid dangling
		_ = oss.AbortMultipartUpload(initResp.Key, initResp.UploadID)
		return nil, apperrors.Internal(err, "failed to record multipart upload")
	}

	return &MultipartInitResult{
		UploadID:      initResp.UploadID,
		Key:           initResp.Key,
		Host:          initResp.Host,
		TotalParts:    totalParts,
		ChunkSize:     chunkSize,
		UploadedParts: nil,
	}, nil
}

// GetPartUploadURL returns a presigned URL for uploading a single part.
func (s *OSSService) GetPartUploadURL(key, uploadID string, partNumber int) (*PartUploadInfo, error) {
	expire := s.config.TokenExpire
	if expire <= 0 {
		expire = 1800
	}
	u, err := oss.GeneratePartUploadURL(key, uploadID, partNumber, expire)
	if err != nil {
		return nil, apperrors.Internal(err, "failed to sign part URL")
	}
	return &PartUploadInfo{PartNumber: u.PartNumber, URL: u.URL, Expire: u.Expire}, nil
}

// GetPartUploadURLs returns presigned URLs for a batch of parts.
func (s *OSSService) GetPartUploadURLs(key, uploadID string, partNumbers []int) ([]PartUploadInfo, error) {
	results := make([]PartUploadInfo, 0, len(partNumbers))
	for _, n := range partNumbers {
		info, err := s.GetPartUploadURL(key, uploadID, n)
		if err != nil {
			return nil, err
		}
		results = append(results, *info)
	}
	return results, nil
}

// CompleteMultipartUpload completes a multipart upload and persists a file record.
func (s *OSSService) CompleteMultipartUpload(key, uploadID, md5, fileName string, fileSize int64, parts []CompletePart, userID uint) (*model.File, error) {
	ossParts := make([]oss.CompletePart, 0, len(parts))
	for _, p := range parts {
		ossParts = append(ossParts, oss.CompletePart{PartNumber: p.PartNumber, ETag: p.ETag})
	}
	if err := oss.CompleteMultipartUpload(key, uploadID, ossParts); err != nil {
		return nil, apperrors.Internal(err, "failed to complete multipart upload")
	}

	// Mark the record completed and clean up parts.
	if err := s.multipartRepo.UpdateUploadStatus(uploadID, model.MultipartStatusCompleted); err != nil {
		logger.Log.Warnf("failed to update multipart status: %v", err)
	}
	if err := s.multipartRepo.DeleteParts(uploadID); err != nil {
		logger.Log.Warnf("failed to delete part records: %v", err)
	}

	return s.SaveFileRecord(key, md5, fileName, fileSize, userID)
}

// AbortMultipartUpload aborts an in-progress multipart upload.
func (s *OSSService) AbortMultipartUpload(key, uploadID string) error {
	if err := oss.AbortMultipartUpload(key, uploadID); err != nil {
		logger.Log.Warnf("failed to abort OSS multipart upload: %v", err)
	}
	_ = s.multipartRepo.UpdateUploadStatus(uploadID, model.MultipartStatusAborted)
	_ = s.multipartRepo.DeleteParts(uploadID)
	return nil
}

// ListUploadedParts lists the parts OSS currently knows about for the given upload.
func (s *OSSService) ListUploadedParts(key, uploadID string) ([]CompletePart, error) {
	parts, err := oss.ListParts(key, uploadID)
	if err != nil {
		return nil, apperrors.Internal(err, "failed to list uploaded parts")
	}
	result := make([]CompletePart, len(parts))
	for i, p := range parts {
		result[i] = CompletePart{PartNumber: p.PartNumber, ETag: p.ETag}
	}
	return result, nil
}

// SaveUploadedPart persists a part record (used by clients that want resumable uploads
// to reconstruct progress without asking OSS).
func (s *OSSService) SaveUploadedPart(uploadID string, partNumber int, etag string, size int64) error {
	part := &model.UploadedPart{
		UploadID:   uploadID,
		PartNumber: partNumber,
		ETag:       etag,
		Size:       size,
	}
	if err := s.multipartRepo.SavePart(part); err != nil {
		return apperrors.Internal(err, "failed to save part record")
	}
	return nil
}

// GetUploadedPartsFromDB returns parts recorded in the local DB.
func (s *OSSService) GetUploadedPartsFromDB(uploadID string) ([]CompletePart, error) {
	parts, err := s.multipartRepo.GetUploadedParts(uploadID)
	if err != nil {
		return nil, apperrors.Internal(err, "failed to get uploaded parts")
	}
	result := make([]CompletePart, len(parts))
	for i, p := range parts {
		result[i] = CompletePart{PartNumber: p.PartNumber, ETag: p.ETag}
	}
	return result, nil
}

// ======================
// Internal helpers
// ======================

// getUserSecUID returns the secUID of the given user, caching results in-memory.
func (s *OSSService) getUserSecUID(userID uint) (string, error) {
	if userID == 0 {
		return "anonymous", nil
	}
	if v, ok := s.secUIDCache.Load(userID); ok {
		return v.(string), nil
	}
	var user model.User
	if err := s.db.Select("sec_uid").First(&user, userID).Error; err != nil {
		return "", apperrors.Internal(err, "failed to load user")
	}
	s.secUIDCache.Store(userID, user.SecUID)
	return user.SecUID, nil
}

// generateFileKey builds an OSS object key like `<upload_dir>/<userSecUID>/<date>/<uuid>.<ext>`.
func (s *OSSService) generateFileKey(ext, userSecUID string) string {
	dir := s.config.UploadDir
	if dir == "" {
		dir = "uploads"
	}
	date := time.Now().Format("2006-01-02")
	name := uuid.New().String()
	if ext != "" {
		name += ext
	}
	if userSecUID != "" {
		return dir + "/" + userSecUID + "/" + date + "/" + name
	}
	return dir + "/" + date + "/" + name
}

// buildHost returns the OSS host used in multipart init responses.
func (s *OSSService) buildHost() string {
	endpoint := s.config.Endpoint
	endpoint = strings.TrimPrefix(endpoint, "https://")
	endpoint = strings.TrimPrefix(endpoint, "http://")
	if strings.HasPrefix(endpoint, s.config.Bucket+".") {
		return "https://" + endpoint
	}
	return "https://" + s.config.Bucket + "." + endpoint
}

// inferContentType returns a best-guess MIME type for an extension.
func inferContentType(ext string) string {
	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".heic", ".heif":
		return "image/heic"
	case ".bmp":
		return "image/bmp"
	case ".svg":
		return "image/svg+xml"
	case ".mp4":
		return "video/mp4"
	case ".mov":
		return "video/quicktime"
	case ".avi":
		return "video/x-msvideo"
	case ".webm":
		return "video/webm"
	case ".mkv":
		return "video/x-matroska"
	case ".pdf":
		return "application/pdf"
	case ".doc":
		return "application/msword"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".xls":
		return "application/vnd.ms-excel"
	case ".xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case ".zip":
		return "application/zip"
	case ".mp3":
		return "audio/mpeg"
	case ".wav":
		return "audio/wav"
	case ".txt":
		return "text/plain"
	case ".json":
		return "application/json"
	default:
		return "application/octet-stream"
	}
}
