package service

import (
	"fmt"
	"go-api-starter/internal/config"
	"go-api-starter/internal/model"
	"go-api-starter/internal/repository"
	"go-api-starter/pkg/oss"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

type OSSService struct {
	repo   *repository.OSSRepository
	config *config.OSSConfig
}

func NewOSSService(repo *repository.OSSRepository, cfg *config.OSSConfig) *OSSService {
	return &OSSService{
		repo:   repo,
		config: cfg,
	}
}

// GetUploadToken generates upload token for client-side direct upload
func (s *OSSService) GetUploadToken(fileName string, userID uint) (*oss.UploadToken, error) {
	var ext string
	
	if fileName != "" {
		// Validate file extension if fileName is provided
		ext = strings.ToLower(filepath.Ext(fileName))
		if !s.isAllowedExtension(ext) {
			return nil, fmt.Errorf("file extension %s is not allowed", ext)
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
		return nil, fmt.Errorf("failed to generate upload token: %w", err)
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
		return nil, fmt.Errorf("failed to save file record: %w", err)
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
		return fmt.Errorf("file not found: %w", err)
	}

	// Delete from OSS first
	bucket := oss.GetBucket()
	if bucket == nil {
		return fmt.Errorf("OSS bucket not initialized")
	}

	// Delete the file from OSS
	if err := bucket.DeleteObject(file.Key); err != nil {
		return fmt.Errorf("failed to delete OSS file %s: %w", file.Key, err)
	}

	// Delete record from database
	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete file record: %w", err)
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
