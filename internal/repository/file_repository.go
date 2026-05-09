package repository

import (
	"context"
	"errors"

	"go-api-starter/internal/model"

	"gorm.io/gorm"
)

var ErrFileNotFound = errors.New("file not found")

// Compile-time interface check
var _ FileRepositoryInterface = (*FileRepository)(nil)

// FileRepository handles file data operations
type FileRepository struct {
	db *gorm.DB
}

// NewFileRepository creates a new FileRepository
func NewFileRepository(db *gorm.DB) *FileRepository {
	return &FileRepository{db: db}
}

// Create creates a new file
func (r *FileRepository) Create(ctx context.Context, file *model.File) error {
	return r.db.WithContext(ctx).Create(file).Error
}

// FindByID finds a file by ID with preloaded relations
func (r *FileRepository) FindByID(ctx context.Context, id uint) (*model.File, error) {
	var file model.File
	err := r.db.WithContext(ctx).
		Preload("User").
		First(&file, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrFileNotFound
	}
	return &file, err
}

// FindByMD5 finds a file by MD5 hash
func (r *FileRepository) FindByMD5(ctx context.Context, md5 string) (*model.File, error) {
	var file model.File
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("file_md5 = ?", md5).
		First(&file).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrFileNotFound
	}
	return &file, err
}

// FindBySecUID finds a file by SecUID
func (r *FileRepository) FindBySecUID(ctx context.Context, secUID string) (*model.File, error) {
	var file model.File
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("sec_uid = ?", secUID).
		First(&file).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrFileNotFound
	}
	return &file, err
}

// Update updates a file
func (r *FileRepository) Update(ctx context.Context, file *model.File) error {
	return r.db.WithContext(ctx).Save(file).Error
}

// Delete soft deletes a file by ID
func (r *FileRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&model.File{}, id)
	if result.RowsAffected == 0 {
		return ErrFileNotFound
	}
	return result.Error
}

// List returns files with pagination, filtering, and sorting
func (r *FileRepository) List(ctx context.Context, filter model.FileFilter, offset, limit int, sort string) ([]model.File, int64, error) {
	var files []model.File
	var total int64

	query := r.db.WithContext(ctx).Model(&model.File{})

	if filter.UserID != nil {
		query = query.Where("user_id = ?", *filter.UserID)
	}
	if filter.Type != nil {
		query = query.Where("type = ?", *filter.Type)
	}
	if filter.IsPrivate != nil {
		query = query.Where("is_private = ?", *filter.IsPrivate)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("User").
		Offset(offset).
		Limit(limit).
		Order(sort).
		Find(&files).Error

	return files, total, err
}
