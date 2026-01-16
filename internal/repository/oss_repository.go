package repository

import (
	"go-api-starter/internal/model"

	"gorm.io/gorm"
)

// Compile-time interface check
var _ OSSRepositoryInterface = (*OSSRepository)(nil)

type OSSRepository struct {
	db *gorm.DB
}

func NewOSSRepository(db *gorm.DB) *OSSRepository {
	return &OSSRepository{db: db}
}

// Create creates a new OSS file record or updates if key/md5 exists
func (r *OSSRepository) Create(file *model.OSSFile) error {
	// Try to find existing record by MD5 and userID
	var existing model.OSSFile
	err := r.db.Unscoped().Where("md5 = ? AND user_id = ?", file.MD5, file.UserID).First(&existing).Error
	
	if err == nil {
		// Record exists, update it (keep the same ID and created_at)
		file.ID = existing.ID
		file.CreatedAt = existing.CreatedAt
		file.DeletedAt = gorm.DeletedAt{} // Clear soft delete
		file.Status = 1 // Ensure status is active
		return r.db.Unscoped().Save(file).Error
	}
	
	// Record doesn't exist, create new
	return r.db.Create(file).Error
}

// GetByKey gets an OSS file by key
func (r *OSSRepository) GetByKey(key string) (*model.OSSFile, error) {
	var file model.OSSFile
	err := r.db.Where("`key` = ? AND status = 1", key).First(&file).Error
	return &file, err
}

// GetByMD5 gets an OSS file by MD5 hash
func (r *OSSRepository) GetByMD5(md5 string, userID uint) (*model.OSSFile, error) {
	var file model.OSSFile
	query := r.db.Where("md5 = ? AND status = 1", md5)
	// Only filter by userID if it's provided (not 0)
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	err := query.First(&file).Error
	return &file, err
}

// GetByID gets an OSS file by ID
func (r *OSSRepository) GetByID(id uint) (*model.OSSFile, error) {
	var file model.OSSFile
	err := r.db.Where("id = ? AND status = 1", id).First(&file).Error
	return &file, err
}

// List lists OSS files with pagination
func (r *OSSRepository) List(userID uint, page, pageSize int) ([]model.OSSFile, int64, error) {
	var files []model.OSSFile
	var total int64

	query := r.db.Model(&model.OSSFile{}).Where("status = 1")
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&files).Error
	return files, total, err
}

// Delete soft deletes an OSS file
func (r *OSSRepository) Delete(id uint) error {
	return r.db.Model(&model.OSSFile{}).Where("id = ?", id).Update("status", 0).Error
}

// DeleteByKey soft deletes an OSS file by key
func (r *OSSRepository) DeleteByKey(key string) error {
	return r.db.Model(&model.OSSFile{}).Where("`key` = ?", key).Update("status", 0).Error
}
