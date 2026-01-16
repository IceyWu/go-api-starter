package service

import (
	"fmt"
	"go-api-starter/internal/config"
	"go-api-starter/internal/model"
	"go-api-starter/internal/repository"
	"go-api-starter/pkg/apperrors"
	"go-api-starter/pkg/oss"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

type OSSService struct {
	repo          repository.OSSRepositoryInterface
	multipartRepo repository.MultipartRepositoryInterface
	config        *config.OSSConfig
}

func NewOSSService(repo repository.OSSRepositoryInterface, multipartRepo repository.MultipartRepositoryInterface, cfg *config.OSSConfig) *OSSService {
	return &OSSService{
		repo:          repo,
		multipartRepo: multipartRepo,
		config:        cfg,
	}
}

// GetUploadToken generates upload token for client-side direct upload
func (s *OSSService) GetUploadToken(fileName string, userID uint) (*oss.UploadToken, error) {
	var ext string
	
	if fileName != "" {
		// Validate file extension if fileName is provided
		ext = strings.ToLower(filepath.Ext(fileName))
		if !s.isAllowedExtension(ext) {
			return nil, apperrors.FileExtensionNotAllowed(ext)
		}
	} else {
		// Use default extension if no fileName provided
		ext = ""
	}

	// Generate unique key
	key := s.generateFileKey(ext)

	// Get directory prefix (everything before the filename)
	lastSlash := strings.LastIndex(key, "/")
	var dir string
	if lastSlash > 0 {
		dir = key[:lastSlash+1] // Include the trailing slash
	} else {
		dir = ""
	}

	// Prepare endpoint for token generation
	// Remove bucket prefix if present to get the base endpoint
	endpoint := s.config.Endpoint
	if strings.HasPrefix(endpoint, s.config.Bucket+".") {
		endpoint = strings.TrimPrefix(endpoint, s.config.Bucket+".")
	}

	// Generate upload token
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

// CheckFileExists checks if file with same MD5 already exists
func (s *OSSService) CheckFileExists(md5 string, userID uint) (*model.OSSFile, bool) {
	file, err := s.repo.GetByMD5(md5, userID)
	if err != nil {
		return nil, false
	}
	// Fill URL dynamically
	file.URL = s.buildFileURL(file.Key)
	return file, true
}

// SaveFileRecord saves file record after successful upload
func (s *OSSService) SaveFileRecord(key, md5, fileName string, fileSize int64, contentType string, userID uint) (*model.OSSFile, error) {
	file := &model.OSSFile{
		Key:         key,
		MD5:         md5,
		FileName:    fileName,
		FileSize:    fileSize,
		ContentType: contentType,
		Extension:   strings.ToLower(filepath.Ext(fileName)),
		UserID:      userID,
		Status:      1,
	}

	if err := s.repo.Create(file); err != nil {
		return nil, apperrors.Internal(err, "failed to save file record")
	}

	// Fill URL dynamically before returning
	file.URL = s.buildFileURL(file.Key)

	return file, nil
}

// GetFileByKey gets file by key
func (s *OSSService) GetFileByKey(key string) (*model.OSSFile, error) {
	file, err := s.repo.GetByKey(key)
	if err != nil {
		return nil, err
	}
	// Fill URL dynamically
	file.URL = s.buildFileURL(file.Key)
	return file, nil
}

// GetFileByID gets file by ID
func (s *OSSService) GetFileByID(id uint) (*model.OSSFile, error) {
	file, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	// Fill URL dynamically
	file.URL = s.buildFileURL(file.Key)
	return file, nil
}

// ListFiles lists files with pagination
func (s *OSSService) ListFiles(userID uint, page, pageSize int) ([]model.OSSFile, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	files, total, err := s.repo.List(userID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	// Fill URL dynamically for each file
	for i := range files {
		files[i].URL = s.buildFileURL(files[i].Key)
	}
	return files, total, nil
}

// DeleteFile deletes file record and OSS file
func (s *OSSService) DeleteFile(id uint) error {
	file, err := s.repo.GetByID(id)
	if err != nil {
		return apperrors.FileNotFound(id)
	}

	// Delete from OSS first
	bucket := oss.GetBucket()
	if bucket == nil {
		return apperrors.ErrOSSNotInitialized
	}

	// Delete the file from OSS
	if err := bucket.DeleteObject(file.Key); err != nil {
		return apperrors.OSSDeleteError(err, file.Key)
	}

	// Delete record from database
	if err := s.repo.Delete(id); err != nil {
		return apperrors.Internal(err, "failed to delete file record")
	}

	return nil
}

// generateFileKey generates a unique file key with date-based directory structure
func (s *OSSService) generateFileKey(ext string) string {
	now := time.Now()
	dateDir := now.Format("2006-01-02")
	fileName := uuid.New().String() + ext
	
	// Use forward slash for OSS paths (not filepath.Join which uses backslash on Windows)
	// Example: go_oss/uploads/2026-01-15/uuid.jpg
	if s.config.UploadDir != "" {
		return fmt.Sprintf("%s/%s/%s/%s", s.config.UploadDir, s.config.BasePath, dateDir, fileName)
	}
	return fmt.Sprintf("%s/%s/%s", s.config.BasePath, dateDir, fileName)
}

// buildFileURL builds the full URL for accessing the file
func (s *OSSService) buildFileURL(key string) string {
	if s.config.Domain != "" {
		return fmt.Sprintf("%s/%s", strings.TrimRight(s.config.Domain, "/"), key)
	}
	
	// Endpoint already contains bucket name (e.g., lpalette.oss-accelerate.aliyuncs.com)
	// Just use it directly
	return fmt.Sprintf("https://%s/%s", s.config.Endpoint, key)
}

// isAllowedExtension checks if the file extension is allowed
func (s *OSSService) isAllowedExtension(ext string) bool {
	if len(s.config.AllowedExtensions) == 0 {
		return true
	}
	for _, allowed := range s.config.AllowedExtensions {
		if strings.EqualFold(ext, allowed) {
			return true
		}
	}
	return false
}


// ============ Multipart Upload (分片上传) ============

// MultipartInitResult represents the result of initializing multipart upload
type MultipartInitResult struct {
	UploadID      string `json:"upload_id"`
	Key           string `json:"key"`
	Host          string `json:"host"`
	TotalParts    int    `json:"total_parts"`
	ChunkSize     int64  `json:"chunk_size"`
	UploadedParts []int  `json:"uploaded_parts"` // Already uploaded part numbers for resuming
}

// InitMultipartUpload initializes a multipart upload or resumes an existing one
func (s *OSSService) InitMultipartUpload(fileName string, md5 string, fileSize int64, chunkSize int64, userID uint) (*MultipartInitResult, error) {
	var ext string
	if fileName != "" {
		ext = strings.ToLower(filepath.Ext(fileName))
		if !s.isAllowedExtension(ext) {
			return nil, apperrors.FileExtensionNotAllowed(ext)
		}
	}

	// Calculate total parts
	totalParts := int((fileSize + chunkSize - 1) / chunkSize)

	// Check if there's an existing upload for this file (by MD5)
	if md5 != "" && s.multipartRepo != nil {
		existingUpload, err := s.multipartRepo.GetUploadByMD5(md5, userID)
		if err == nil && existingUpload != nil {
			// Found existing upload, get uploaded parts
			uploadedParts, _ := s.multipartRepo.GetUploadedPartNumbers(existingUpload.UploadID)
			return &MultipartInitResult{
				UploadID:      existingUpload.UploadID,
				Key:           existingUpload.Key,
				Host:          fmt.Sprintf("https://%s", s.config.Endpoint),
				TotalParts:    existingUpload.TotalParts,
				ChunkSize:     existingUpload.ChunkSize,
				UploadedParts: uploadedParts,
			}, nil
		}
	}

	// Generate unique key
	key := s.generateFileKey(ext)

	// Initialize multipart upload on OSS
	result, err := oss.InitMultipartUpload(key)
	if err != nil {
		return nil, apperrors.MultipartInitError(err)
	}

	// Save upload task to database
	if s.multipartRepo != nil {
		upload := &model.MultipartUpload{
			UploadID:    result.UploadID,
			Key:         result.Key,
			MD5:         md5,
			FileName:    fileName,
			FileSize:    fileSize,
			ContentType: "",
			TotalParts:  totalParts,
			ChunkSize:   chunkSize,
			UserID:      userID,
			Status:      model.MultipartStatusInitiated,
		}
		if err := s.multipartRepo.CreateUpload(upload); err != nil {
			// Log error but don't fail - upload can still work without DB tracking
			fmt.Printf("Warning: failed to save upload task to DB: %v\n", err)
		}
	}

	return &MultipartInitResult{
		UploadID:      result.UploadID,
		Key:           result.Key,
		Host:          result.Host,
		TotalParts:    totalParts,
		ChunkSize:     chunkSize,
		UploadedParts: []int{},
	}, nil
}

// PartUploadInfo represents info for uploading a part
type PartUploadInfo struct {
	PartNumber int    `json:"part_number"`
	URL        string `json:"url"`
}

// GetPartUploadURL generates presigned URL for uploading a part
func (s *OSSService) GetPartUploadURL(key, uploadID string, partNumber int) (*PartUploadInfo, error) {
	result, err := oss.GeneratePartUploadURL(key, uploadID, partNumber, s.config.TokenExpire)
	if err != nil {
		return nil, apperrors.OSSUploadError(err, fmt.Sprintf("part %d", partNumber))
	}

	return &PartUploadInfo{
		PartNumber: result.PartNumber,
		URL:        result.URL,
	}, nil
}

// GetPartUploadURLs generates presigned URLs for multiple parts
func (s *OSSService) GetPartUploadURLs(key, uploadID string, partNumbers []int) ([]PartUploadInfo, error) {
	var results []PartUploadInfo
	for _, partNumber := range partNumbers {
		result, err := oss.GeneratePartUploadURL(key, uploadID, partNumber, s.config.TokenExpire)
		if err != nil {
			return nil, apperrors.OSSUploadError(err, fmt.Sprintf("part %d", partNumber))
		}
		results = append(results, PartUploadInfo{
			PartNumber: result.PartNumber,
			URL:        result.URL,
		})
	}
	return results, nil
}

// CompletePart represents a completed part
type CompletePart struct {
	PartNumber int    `json:"part_number"`
	ETag       string `json:"etag"`
}

// CompleteMultipartUpload completes a multipart upload and saves file record
func (s *OSSService) CompleteMultipartUpload(key, uploadID, md5, fileName string, fileSize int64, contentType string, parts []CompletePart, userID uint) (*model.OSSFile, error) {
	// Convert to oss package type
	var ossParts []oss.CompletePart
	for _, p := range parts {
		ossParts = append(ossParts, oss.CompletePart{
			PartNumber: p.PartNumber,
			ETag:       p.ETag,
		})
	}

	// Complete multipart upload
	if err := oss.CompleteMultipartUpload(key, uploadID, ossParts); err != nil {
		return nil, apperrors.MultipartCompleteError(err)
	}

	// Clean up database records
	if s.multipartRepo != nil {
		s.multipartRepo.UpdateUploadStatus(uploadID, model.MultipartStatusCompleted)
		s.multipartRepo.DeleteParts(uploadID)
	}

	// Save file record
	return s.SaveFileRecord(key, md5, fileName, fileSize, contentType, userID)
}

// AbortMultipartUpload aborts a multipart upload
func (s *OSSService) AbortMultipartUpload(key, uploadID string) error {
	err := oss.AbortMultipartUpload(key, uploadID)
	
	// Clean up database records regardless of OSS result
	if s.multipartRepo != nil {
		s.multipartRepo.UpdateUploadStatus(uploadID, model.MultipartStatusAborted)
		s.multipartRepo.DeleteParts(uploadID)
	}
	
	if err != nil {
		return apperrors.MultipartAbortError(err)
	}
	return nil
}

// ListUploadedParts lists uploaded parts for resumable upload
func (s *OSSService) ListUploadedParts(key, uploadID string) ([]CompletePart, error) {
	parts, err := oss.ListParts(key, uploadID)
	if err != nil {
		return nil, apperrors.OSSListError(err)
	}

	var result []CompletePart
	for _, p := range parts {
		result = append(result, CompletePart{
			PartNumber: p.PartNumber,
			ETag:       p.ETag,
		})
	}
	return result, nil
}

// SaveUploadedPart saves an uploaded part info to database for resumable upload
func (s *OSSService) SaveUploadedPart(uploadID string, partNumber int, etag string, size int64) error {
	if s.multipartRepo == nil {
		return nil
	}
	
	part := &model.UploadedPart{
		UploadID:   uploadID,
		PartNumber: partNumber,
		ETag:       etag,
		Size:       size,
	}
	return s.multipartRepo.SavePart(part)
}

// GetUploadedPartsFromDB gets uploaded parts from database
func (s *OSSService) GetUploadedPartsFromDB(uploadID string) ([]CompletePart, error) {
	if s.multipartRepo == nil {
		return nil, nil
	}
	
	parts, err := s.multipartRepo.GetUploadedParts(uploadID)
	if err != nil {
		return nil, err
	}
	
	var result []CompletePart
	for _, p := range parts {
		result = append(result, CompletePart{
			PartNumber: p.PartNumber,
			ETag:       p.ETag,
		})
	}
	return result, nil
}
