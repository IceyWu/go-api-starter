package repository

import (
	"go-api-starter/internal/model"

	"gorm.io/gorm"
)

type MultipartRepository struct {
	db *gorm.DB
}

func NewMultipartRepository(db *gorm.DB) *MultipartRepository {
	return &MultipartRepository{db: db}
}

// CreateUpload creates a new multipart upload record
func (r *MultipartRepository) CreateUpload(upload *model.MultipartUpload) error {
	return r.db.Create(upload).Error
}

// GetUploadByMD5 gets an ongoing upload by MD5 and user
func (r *MultipartRepository) GetUploadByMD5(md5 string, userID uint) (*model.MultipartUpload, error) {
	var upload model.MultipartUpload
	err := r.db.Where("md5 = ? AND user_id = ? AND status = 1", md5, userID).First(&upload).Error
	return &upload, err
}

// GetUploadByID gets upload by upload_id
func (r *MultipartRepository) GetUploadByID(uploadID string) (*model.MultipartUpload, error) {
	var upload model.MultipartUpload
	err := r.db.Where("upload_id = ? AND status = 1", uploadID).First(&upload).Error
	return &upload, err
}

// UpdateUploadStatus updates upload status
func (r *MultipartRepository) UpdateUploadStatus(uploadID string, status int) error {
	return r.db.Model(&model.MultipartUpload{}).Where("upload_id = ?", uploadID).Update("status", status).Error
}

// DeleteUpload deletes upload record
func (r *MultipartRepository) DeleteUpload(uploadID string) error {
	return r.db.Where("upload_id = ?", uploadID).Delete(&model.MultipartUpload{}).Error
}

// SavePart saves an uploaded part record
func (r *MultipartRepository) SavePart(part *model.UploadedPart) error {
	return r.db.Save(part).Error
}

// GetUploadedParts gets all uploaded parts for an upload
func (r *MultipartRepository) GetUploadedParts(uploadID string) ([]model.UploadedPart, error) {
	var parts []model.UploadedPart
	err := r.db.Where("upload_id = ?", uploadID).Order("part_number ASC").Find(&parts).Error
	return parts, err
}

// DeleteParts deletes all parts for an upload
func (r *MultipartRepository) DeleteParts(uploadID string) error {
	return r.db.Where("upload_id = ?", uploadID).Delete(&model.UploadedPart{}).Error
}

// GetUploadedPartNumbers gets list of uploaded part numbers
func (r *MultipartRepository) GetUploadedPartNumbers(uploadID string) ([]int, error) {
	var partNumbers []int
	err := r.db.Model(&model.UploadedPart{}).Where("upload_id = ?", uploadID).Pluck("part_number", &partNumbers).Error
	return partNumbers, err
}
