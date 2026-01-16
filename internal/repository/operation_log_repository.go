package repository

import (
	"context"

	"go-api-starter/internal/model"

	"gorm.io/gorm"
)

// OperationLogRepositoryInterface 操作日志仓库接口
type OperationLogRepositoryInterface interface {
	Create(ctx context.Context, log *model.OperationLog) error
	List(ctx context.Context, offset, limit int, filters map[string]interface{}) ([]model.OperationLog, int64, error)
	GetByID(ctx context.Context, id uint) (*model.OperationLog, error)
}

// OperationLogRepository 操作日志仓库实现
type OperationLogRepository struct {
	db *gorm.DB
}

// NewOperationLogRepository 创建操作日志仓库
func NewOperationLogRepository(db *gorm.DB) *OperationLogRepository {
	return &OperationLogRepository{db: db}
}

// Create 创建操作日志
func (r *OperationLogRepository) Create(ctx context.Context, log *model.OperationLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// List 查询操作日志列表
func (r *OperationLogRepository) List(ctx context.Context, offset, limit int, filters map[string]interface{}) ([]model.OperationLog, int64, error) {
	var logs []model.OperationLog
	var total int64

	query := r.db.WithContext(ctx).Model(&model.OperationLog{})

	// 应用过滤条件
	if userID, ok := filters["user_id"]; ok {
		query = query.Where("user_id = ?", userID)
	}
	if module, ok := filters["module"]; ok {
		query = query.Where("module = ?", module)
	}
	if action, ok := filters["action"]; ok {
		query = query.Where("action = ?", action)
	}
	if startTime, ok := filters["start_time"]; ok {
		query = query.Where("created_at >= ?", startTime)
	}
	if endTime, ok := filters["end_time"]; ok {
		query = query.Where("created_at <= ?", endTime)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	if err := query.Order("id DESC").Offset(offset).Limit(limit).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// GetByID 根据ID获取操作日志
func (r *OperationLogRepository) GetByID(ctx context.Context, id uint) (*model.OperationLog, error) {
	var log model.OperationLog
	if err := r.db.WithContext(ctx).First(&log, id).Error; err != nil {
		return nil, err
	}
	return &log, nil
}
