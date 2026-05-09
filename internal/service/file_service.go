package service

import (
	"context"
	"errors"
	"mime/multipart"

	"go-api-starter/internal/model"
	"go-api-starter/internal/repository"
	"go-api-starter/pkg/apperrors"
	"go-api-starter/pkg/oss"
	"go-api-starter/pkg/utils"
)

// FileService handles file business logic for simple uploads via multipart form.
// For large files / chunked uploads, use OSSService instead.
type FileService struct {
	repo repository.FileRepositoryInterface
}

// NewFileService creates a new FileService
func NewFileService(repo repository.FileRepositoryInterface) *FileService {
	return &FileService{repo: repo}
}

// Upload uploads a file via multipart form and creates a file record.
// Intended for small files uploaded through the API server directly.
func (s *FileService) Upload(ctx context.Context, userID uint, file multipart.File, header *multipart.FileHeader) (*model.File, error) {
	// Calculate MD5
	md5Hash, err := oss.CalculateMD5(file)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to calculate MD5")
	}

	// Instant upload: if a file with the same MD5 exists, return it
	existingFile, err := s.repo.FindByMD5(ctx, md5Hash)
	if err == nil && existingFile != nil {
		return existingFile, nil
	}

	// Upload to OSS
	objectKey := oss.GenerateObjectKey(header.Filename, "")
	uploadResult, err := oss.UploadFile(file, header, objectKey)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to upload file")
	}

	// Best-effort dimension extraction for images
	var width, height *uint
	if isImage(header.Header.Get("Content-Type")) {
		if w, h, err := utils.GetImageDimensions(file); err == nil {
			uw := uint(w)
			uh := uint(h)
			width = &uw
			height = &uh
		}
	}

	fileRecord := &model.File{
		UserID:    userID,
		Name:      header.Filename,
		Type:      header.Header.Get("Content-Type"),
		FileMd5:   md5Hash,
		Size:      uint(uploadResult.Size),
		Key:       uploadResult.Key,
		Width:     width,
		Height:    height,
		IsPrivate: false,
	}

	if err := s.repo.Create(ctx, fileRecord); err != nil {
		return nil, apperrors.Wrap(err, "failed to create file record")
	}
	return fileRecord, nil
}

// GetByID returns a file by ID
func (s *FileService) GetByID(ctx context.Context, id uint, userID *uint) (*model.File, error) {
	file, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrFileNotFound) {
			return nil, apperrors.NotFound("file not found")
		}
		return nil, apperrors.Wrap(err, "failed to get file")
	}

	if file.IsPrivate && (userID == nil || *userID != file.UserID) {
		return nil, apperrors.Forbidden("you don't have permission to access this file")
	}

	return file, nil
}

// List returns files with pagination, filtering, and sorting
func (s *FileService) List(ctx context.Context, filter model.FileFilter, offset, limit int, sort string, requestUserID *uint) ([]model.File, int64, error) {
	// If not the owner, only show public files
	if filter.UserID != nil && (requestUserID == nil || *requestUserID != *filter.UserID) {
		isPrivate := false
		filter.IsPrivate = &isPrivate
	}

	if sort == "" {
		sort = "created_at DESC"
	}

	files, total, err := s.repo.List(ctx, filter, offset, limit, sort)
	if err != nil {
		return nil, 0, apperrors.Wrap(err, "failed to list files")
	}
	return files, total, nil
}

// Update updates a file
func (s *FileService) Update(ctx context.Context, id uint, userID uint, req *model.UpdateFileRequest) (*model.File, error) {
	file, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrFileNotFound) {
			return nil, apperrors.NotFound("file not found")
		}
		return nil, apperrors.Wrap(err, "failed to find file")
	}

	if file.UserID != userID {
		return nil, apperrors.Forbidden("you don't have permission to update this file")
	}

	if req.Name != nil {
		file.Name = *req.Name
	}
	if req.IsPrivate != nil {
		file.IsPrivate = *req.IsPrivate
	}

	if err := s.repo.Update(ctx, file); err != nil {
		return nil, apperrors.Wrap(err, "failed to update file")
	}
	return file, nil
}

// Delete deletes a file
func (s *FileService) Delete(ctx context.Context, id uint, userID uint) error {
	file, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrFileNotFound) {
			return apperrors.NotFound("file not found")
		}
		return apperrors.Wrap(err, "failed to find file")
	}

	if file.UserID != userID {
		return apperrors.Forbidden("you don't have permission to delete this file")
	}

	// Best-effort OSS delete
	if file.Key != "" {
		_ = oss.DeleteFile(file.Key)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return apperrors.Wrap(err, "failed to delete file")
	}
	return nil
}

// isImage checks if content type is an image
func isImage(contentType string) bool {
	return contentType == "image/jpeg" ||
		contentType == "image/jpg" ||
		contentType == "image/png" ||
		contentType == "image/gif" ||
		contentType == "image/webp"
}
