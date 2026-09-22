package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"go-api-starter/internal/config"
	"go-api-starter/internal/model"
	"go-api-starter/internal/platform/apperrors"
	"go-api-starter/internal/platform/geo"
	"go-api-starter/internal/platform/logger"
	"go-api-starter/internal/platform/storage"
	"go-api-starter/internal/repository"
)

// StorageService handles provider-neutral object storage and file records.
// It deliberately does not do any media processing (EXIF / blurhash / transcoding);
// clients are expected to upload finished artifacts.
type StorageService struct {
	db            *sqlx.DB
	fileRepo      repository.FileRepositoryInterface
	multipartRepo repository.MultipartRepositoryInterface
	provider      storage.ObjectStorage
	config        *config.StorageConfig
	appEnv        string
	geoService    geo.GeocodingService
	uidCache      sync.Map // userID -> uid cache
}

type UploadMode string

const (
	UploadModeDirect    UploadMode = "direct"
	UploadModeMultipart UploadMode = "multipart"
)

type UploadSession struct {
	ID            string                   `json:"id"`
	Mode          UploadMode               `json:"mode"`
	Key           string                   `json:"key"`
	FileName      string                   `json:"file_name"`
	ContentType   string                   `json:"content_type"`
	Size          int64                    `json:"size"`
	PartSize      int64                    `json:"part_size,omitempty"`
	TotalParts    int                      `json:"total_parts"`
	UploadedParts []CompletePart           `json:"uploaded_parts,omitempty"`
	Upload        *storage.PresignedUpload `json:"upload,omitempty"`
	Parts         []PartUploadInfo         `json:"parts,omitempty"`
	ExistingFile  *model.File              `json:"existing_file,omitempty"`
}

// NewStorageService creates a new StorageService.
func NewStorageService(db *sqlx.DB, fileRepo repository.FileRepositoryInterface, multipartRepo repository.MultipartRepositoryInterface, provider storage.ObjectStorage, cfg *config.StorageConfig, appEnv string, amapAPIKey string) *StorageService {
	var geoService geo.GeocodingService
	if amapAPIKey != "" {
		geoService = geo.NewAMapGeocodingService(amapAPIKey)
	}
	return &StorageService{
		db:            db,
		fileRepo:      fileRepo,
		multipartRepo: multipartRepo,
		provider:      provider,
		config:        cfg,
		appEnv:        appEnv,
		geoService:    geoService,
	}
}

// UploadPublic stores an object through the same provider abstraction as authenticated uploads.
func (s *StorageService) UploadPublic(ctx context.Context, key string, body io.Reader, contentType string) (*storage.ObjectInfo, error) {
	if s.provider == nil {
		return nil, apperrors.StorageInitError(errors.New("object storage is not configured"))
	}
	result, err := s.provider.PutObject(ctx, key, body, contentType)
	if err != nil {
		return nil, apperrors.Internal(err, "failed to upload object")
	}
	return result, nil
}

// CreateUpload creates a provider-neutral upload session. The client never chooses
// the object key or talks to a provider-specific policy API.
func (s *StorageService) CreateUpload(fileName, contentType, checksum string, fileSize, partSize int64, userID uint) (*UploadSession, error) {
	if s.provider == nil {
		return nil, apperrors.StorageInitError(errors.New("object storage is not configured"))
	}
	if fileName == "" || fileSize <= 0 || checksum == "" {
		return nil, apperrors.BadRequest("file_name, size, and checksum are required")
	}
	if s.config.MaxFileSize > 0 && fileSize > s.config.MaxFileSize {
		maxMiB := (s.config.MaxFileSize + (1 << 20) - 1) / (1 << 20)
		return nil, apperrors.BadRequestWithDetails(
			fmt.Sprintf("file size exceeds the maximum allowed size of %d MiB", maxMiB),
			map[string]int64{"max_file_size": s.config.MaxFileSize},
		)
	}
	ext := strings.ToLower(filepath.Ext(fileName))
	if len(s.config.AllowedExtensions) > 0 {
		allowed := false
		for _, candidate := range s.config.AllowedExtensions {
			candidate = strings.ToLower(strings.TrimSpace(candidate))
			if candidate != "" && !strings.HasPrefix(candidate, ".") {
				candidate = "." + candidate
			}
			if ext == candidate {
				allowed = true
				break
			}
		}
		if !allowed {
			return nil, apperrors.FileExtensionNotAllowed(ext)
		}
	}
	if existing, ok := s.CheckFileExists(checksum, userID); ok {
		return &UploadSession{ID: "", Mode: UploadModeDirect, Key: existing.Key, FileName: existing.Name, ContentType: existing.Type, Size: int64(existing.Size), ExistingFile: existing}, nil
	}
	if contentType == "" {
		contentType = inferContentType(strings.ToLower(filepath.Ext(fileName)))
	}
	if existing, existingErr := s.multipartRepo.GetUploadByMD5(checksum, userID); existingErr == nil && existing != nil {
		return s.resumeUploadSession(existing)
	}
	userUID, err := s.getUserUID(userID)
	if err != nil {
		return nil, err
	}
	key := s.generateFileKey(ext, userUID)
	if partSize <= 0 {
		partSize = 5 * 1024 * 1024
	}
	const multipartThreshold = 5 * 1024 * 1024
	mode := UploadModeDirect
	totalParts := 1
	sessionID := uuid.NewString()
	var upload *storage.MultipartUpload
	if fileSize >= multipartThreshold {
		mode = UploadModeMultipart
		totalParts = int((fileSize + partSize - 1) / partSize)
		upload, err = s.provider.CreateMultipartUpload(context.Background(), key, contentType)
		if err != nil {
			return nil, apperrors.Internal(err, "failed to create upload session")
		}
		sessionID = upload.UploadID
	}
	session := &model.MultipartUpload{UploadID: sessionID, Key: key, MD5: checksum, FileName: fileName, FileSize: fileSize, ContentType: contentType, TotalParts: totalParts, ChunkSize: partSize, UserID: userID, Status: model.MultipartStatusInitiated}
	if err := s.multipartRepo.CreateUpload(session); err != nil {
		if upload != nil {
			_ = s.provider.AbortMultipartUpload(context.Background(), key, upload.UploadID)
		}
		return nil, apperrors.Internal(err, "failed to persist upload session")
	}
	result := &UploadSession{ID: sessionID, Mode: mode, Key: key, FileName: fileName, ContentType: contentType, Size: fileSize, PartSize: partSize, TotalParts: totalParts}
	if mode == UploadModeDirect {
		result.Upload, err = s.provider.PresignPutObject(context.Background(), key, contentType, s.uploadURLTTL())
		if err != nil {
			_ = s.multipartRepo.UpdateUploadStatus(sessionID, model.MultipartStatusAborted)
			return nil, apperrors.Internal(err, "failed to sign upload URL")
		}
	} else {
		result.Parts, err = s.presignParts(key, sessionID, totalParts)
		if err != nil {
			_ = s.provider.AbortMultipartUpload(context.Background(), key, sessionID)
			_ = s.multipartRepo.UpdateUploadStatus(sessionID, model.MultipartStatusAborted)
			return nil, err
		}
	}
	return result, nil
}

func (s *StorageService) resumeUploadSession(session *model.MultipartUpload) (*UploadSession, error) {
	result := &UploadSession{ID: session.UploadID, Key: session.Key, FileName: session.FileName, ContentType: session.ContentType, Size: session.FileSize, PartSize: session.ChunkSize, TotalParts: session.TotalParts}
	if session.TotalParts <= 1 {
		result.Mode = UploadModeDirect
		var err error
		result.Upload, err = s.provider.PresignPutObject(context.Background(), session.Key, session.ContentType, s.uploadURLTTL())
		if err != nil {
			return nil, apperrors.Internal(err, "failed to sign upload URL")
		}
		return result, nil
	}
	result.Mode = UploadModeMultipart
	parts, err := s.provider.ListParts(context.Background(), session.Key, session.UploadID)
	if err != nil {
		return nil, apperrors.Internal(err, "failed to list uploaded parts")
	}
	result.UploadedParts = make([]CompletePart, len(parts))
	for i, part := range parts {
		result.UploadedParts[i] = CompletePart{PartNumber: part.PartNumber, ETag: part.ETag}
	}
	result.Parts, err = s.presignParts(session.Key, session.UploadID, session.TotalParts)
	return result, err
}

func (s *StorageService) uploadURLTTL() time.Duration {
	expires := time.Duration(s.config.TokenExpire) * time.Second
	if expires <= 0 {
		return 30 * time.Minute
	}
	return expires
}

func (s *StorageService) presignParts(key, uploadID string, totalParts int) ([]PartUploadInfo, error) {
	result := make([]PartUploadInfo, 0, totalParts)
	for number := 1; number <= totalParts; number++ {
		part, err := s.provider.PresignUploadPart(context.Background(), key, uploadID, number, s.uploadURLTTL())
		if err != nil {
			return nil, apperrors.Internal(err, "failed to sign part URL")
		}
		result = append(result, PartUploadInfo{PartNumber: part.PartNumber, URL: part.URL, Expire: part.ExpiresAt, Headers: part.Headers})
	}
	return result, nil
}

func (s *StorageService) CompleteUpload(uploadID string, userID uint, parts []CompletePart, metadata *ClientMediaMetadata) (*model.File, error) {
	session, err := s.ownedUpload(uploadID, userID)
	if err != nil {
		return nil, err
	}
	if session.TotalParts > 1 {
		providerParts := make([]storage.CompletedPart, 0, len(parts))
		for _, part := range parts {
			providerParts = append(providerParts, storage.CompletedPart{PartNumber: part.PartNumber, ETag: part.ETag})
		}
		if len(providerParts) != session.TotalParts {
			return nil, apperrors.BadRequest("all upload parts are required")
		}
		seen := make(map[int]struct{}, len(providerParts))
		for _, part := range providerParts {
			if part.PartNumber < 1 || part.PartNumber > session.TotalParts || part.ETag == "" {
				return nil, apperrors.BadRequest("invalid upload part")
			}
			if _, exists := seen[part.PartNumber]; exists {
				return nil, apperrors.BadRequest("duplicate upload part")
			}
			seen[part.PartNumber] = struct{}{}
		}
		if len(seen) != session.TotalParts {
			return nil, apperrors.BadRequest("all upload parts are required")
		}
		if err := s.provider.CompleteMultipartUpload(context.Background(), session.Key, session.UploadID, providerParts); err != nil {
			return nil, apperrors.Internal(err, "failed to complete upload")
		}
	}
	object, err := s.provider.HeadObject(context.Background(), session.Key)
	if err != nil {
		return nil, apperrors.Internal(err, "uploaded object was not found")
	}
	if object.Size != session.FileSize {
		return nil, apperrors.BadRequest("uploaded object size does not match the session")
	}
	if err := s.multipartRepo.UpdateUploadStatus(uploadID, model.MultipartStatusCompleted); err != nil {
		return nil, apperrors.Internal(err, "failed to complete upload session")
	}
	_ = s.multipartRepo.DeleteParts(uploadID)
	return s.SaveFileRecord(session.Key, session.MD5, session.FileName, object.Size, userID, metadata)
}

func (s *StorageService) AbortUpload(uploadID string, userID uint) error {
	session, err := s.ownedUpload(uploadID, userID)
	if err != nil {
		return err
	}
	if session.TotalParts > 1 && s.provider != nil {
		if err := s.provider.AbortMultipartUpload(context.Background(), session.Key, session.UploadID); err != nil {
			return apperrors.Internal(err, "failed to abort upload")
		}
	}
	if err := s.multipartRepo.UpdateUploadStatus(uploadID, model.MultipartStatusAborted); err != nil {
		return apperrors.Internal(err, "failed to abort upload session")
	}
	_ = s.multipartRepo.DeleteParts(uploadID)
	return nil
}

func (s *StorageService) ownedUpload(uploadID string, userID uint) (*model.MultipartUpload, error) {
	if uploadID == "" || userID == 0 {
		return nil, apperrors.BadRequest("upload session is invalid")
	}
	session, err := s.multipartRepo.GetUploadByID(uploadID)
	if err != nil || session == nil || session.UserID != userID {
		return nil, apperrors.NotFound("upload session not found")
	}
	return session, nil
}

// CheckFileExists reports whether a file with the given MD5 already exists
// (scoped to the user when userID > 0, used for instant upload).
func (s *StorageService) CheckFileExists(md5 string, userID uint) (*model.File, bool) {
	var file model.File
	query := `SELECT id,uid,user_id,name,type,file_md5,size,` + "`key`" + `,extension,width,height,blurhash,arthash,arthash_codec,lng,lat,country,country_code,province,city,district,address,altitude,taken_at,device_make,device_model,lens_model,f_number,exposure_time,iso,focal_length,exif_raw,duration,codec,bitrate,frame_rate,video_metadata,transcoding_task_id,is_private,created_at,updated_at FROM files WHERE file_md5 = ?`
	args := []any{md5}
	if userID > 0 {
		query += " AND user_id = ?"
		args = append(args, userID)
	}
	query += " LIMIT 1"
	if err := s.db.Get(&file, query, args...); err != nil {
		return nil, false
	}
	file.PrepareForResponse()
	return &file, true
}

// SaveFileRecord creates a DB record after a successful direct object-storage upload.
func (s *StorageService) SaveFileRecord(key, md5, fileName string, fileSize int64, userID uint, metadata *ClientMediaMetadata) (*model.File, error) {
	ext := strings.ToLower(filepath.Ext(fileName))
	if metadata == nil {
		metadata = &ClientMediaMetadata{}
	}
	if metadata.Basic.MD5 != "" && !strings.EqualFold(metadata.Basic.MD5, md5) {
		return nil, apperrors.BadRequest("metadata md5 does not match upload md5")
	}
	if metadata.Basic.Size != 0 && metadata.Basic.Size != fileSize {
		return nil, apperrors.BadRequest("metadata size does not match upload size")
	}
	metadata.Basic.MD5 = md5
	metadata.Basic.Size = fileSize
	if metadata.Basic.Type == "" {
		metadata.Basic.Type = inferContentType(ext)
	}
	// Dedup by MD5 (scoped to user)
	if md5 != "" {
		if existing, ok := s.CheckFileExists(md5, userID); ok {
			return existing, nil
		}
	}

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
	if metadata != nil {
		applyClientMetadata(file, metadata)
		applyClientGeocoding(file, s.geoService)
	}

	if err := s.fileRepo.Create(context.Background(), file); err != nil {
		return nil, apperrors.Internal(err, "failed to save file record")
	}
	if metadata != nil && metadata.Image != nil {
		for _, item := range metadata.Image.Colors {
			color := &model.Color{Hex: item.Hex, R: item.R, G: item.G, B: item.B}
			if err := s.db.Get(color, `SELECT id,hex,r,g,b,created_at FROM colors WHERE hex = ? LIMIT 1`, item.Hex); err != nil {
				if !errors.Is(err, sql.ErrNoRows) {
					continue
				}
				result, createErr := s.db.Exec(`INSERT INTO colors(hex,r,g,b,created_at) VALUES(?,?,?,?,?)`, item.Hex, item.R, item.G, item.B, time.Now())
				if createErr != nil {
					continue
				}
				id, idErr := result.LastInsertId()
				if idErr != nil {
					continue
				}
				color.ID = uint(id)
			}
			percentage := item.Percentage
			_, _ = s.db.Exec(`INSERT INTO file_colors(file_id,color_id,is_primary,`+"`rank`"+`,percentage,created_at) VALUES(?,?,?,?,?,?)`, file.ID, color.ID, item.IsPrimary, item.Rank, &percentage, time.Now())
		}
	}
	file.PrepareForResponse()
	return file, nil
}

// ======================
// File queries / updates
// ======================

// GetFileByUID returns a file by its UID.
func (s *StorageService) GetFileByUID(uid string) (*model.File, error) {
	file, err := s.fileRepo.FindByUID(context.Background(), uid)
	if err != nil {
		if errors.Is(err, repository.ErrFileNotFound) {
			return nil, apperrors.NotFound("file not found")
		}
		return nil, apperrors.Internal(err, "failed to query file")
	}
	return file, nil
}

// UpdateFile updates mutable fields (name / visibility) on a file.
func (s *StorageService) UpdateFile(uid string, req *model.UpdateFileRequest) error {
	file, err := s.GetFileByUID(uid)
	if err != nil {
		return err
	}
	if req.Name != nil {
		file.Name = *req.Name
	}
	if req.IsPrivate != nil {
		file.IsPrivate = *req.IsPrivate
	}
	if err := s.fileRepo.Update(context.Background(), file); err != nil {
		return apperrors.Internal(err, "failed to update file")
	}
	return nil
}

// ListFiles returns a paginated list of files, optionally filtered by owner and privacy.
func (s *StorageService) ListFiles(userID uint, isPrivate *bool, offset, limit int, sort string) ([]model.File, int64, error) {
	filter := model.FileFilter{}
	if userID > 0 {
		filter.UserID = &userID
	}
	filter.IsPrivate = isPrivate
	files, total, err := s.fileRepo.List(context.Background(), filter, offset, limit, sort)
	if err != nil {
		return nil, 0, apperrors.Internal(err, "failed to list files")
	}
	return files, total, nil
}

// DeleteFile removes the object and its DB record.
func (s *StorageService) DeleteFile(uid string) error {
	file, err := s.GetFileByUID(uid)
	if err != nil {
		return err
	}

	// Best-effort object deletion; don't fail the request if cleanup fails.
	if file.Key != "" && s.provider != nil {
		if err := s.provider.DeleteObject(context.Background(), file.Key); err != nil {
			logger.Log.Warnf("failed to delete object %s: %v", file.Key, err)
		}
	}

	if err := s.fileRepo.Delete(context.Background(), file.ID); err != nil {
		return apperrors.Internal(err, "failed to delete file")
	}
	return nil
}

func (s *StorageService) UpdateFileTranscodingTask(uid, taskID string) error {
	file, err := s.fileRepo.FindByUID(context.Background(), uid)
	if err != nil {
		return err
	}
	file.TranscodingTaskID = &taskID
	return s.fileRepo.Update(context.Background(), file)
}

// PartUploadInfo represents a signed upload URL for a single part.
type PartUploadInfo struct {
	PartNumber int               `json:"part_number"`
	URL        string            `json:"url"`
	Expire     int64             `json:"expires_at"`
	Headers    map[string]string `json:"headers,omitempty"`
}

// CompletePart represents a completed part.
type CompletePart struct {
	PartNumber int    `json:"part_number"`
	ETag       string `json:"etag"`
}

// ======================
// Internal helpers
// ======================

// getUserUID returns the UID of the given user, caching results in-memory.
func (s *StorageService) getUserUID(userID uint) (string, error) {
	if userID == 0 {
		return "anonymous", nil
	}
	if v, ok := s.uidCache.Load(userID); ok {
		return v.(string), nil
	}
	var user model.User
	if err := s.db.Get(&user, `SELECT uid FROM users WHERE id = ?`, userID); err != nil {
		return "", apperrors.Internal(err, "failed to load user")
	}
	s.uidCache.Store(userID, user.UID)
	return user.UID, nil
}

// generateFileKey builds an object key like `<upload_dir>/<userUID>/<date>/<uuid>.<ext>`.
func (s *StorageService) generateFileKey(ext, userUID string) string {
	dir := s.config.UploadDir
	if dir == "" {
		dir = "uploads"
	}
	date := time.Now().Format("2006-01-02")
	name := uuid.New().String()
	if ext != "" {
		name += ext
	}
	if userUID != "" {
		return dir + "/" + userUID + "/" + date + "/" + name
	}
	return dir + "/" + date + "/" + name
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
