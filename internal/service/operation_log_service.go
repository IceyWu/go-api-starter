package service

import (
	"context"

	"go-api-starter/internal/model"
	"go-api-starter/internal/repository"
)

// OperationLogServiceInterface 操作日志服务接口
type OperationLogServiceInterface interface {
	Create(ctx context.Context, log *model.OperationLog) error
	List(ctx context.Context, offset, limit int, filters map[string]interface{}) ([]model.OperationLog, int64, error)
	GetByID(ctx context.Context, id uint) (*model.OperationLog, error)
}

// OperationLogService 操作日志服务实现
type OperationLogService struct {
	repo repository.OperationLogRepositoryInterface
}

// NewOperationLogService 创建操作日志服务
func NewOperationLogService(repo repository.OperationLogRepositoryInterface) *OperationLogService {
	return &OperationLogService{repo: repo}
}

// Create 创建操作日志
func (s *OperationLogService) Create(ctx context.Context, log *model.OperationLog) error {
	return s.repo.Create(ctx, log)
}

// List 查询操作日志列表
func (s *OperationLogService) List(ctx context.Context, offset, limit int, filters map[string]interface{}) ([]model.OperationLog, int64, error) {
	return s.repo.List(ctx, offset, limit, filters)
}

// GetByID 根据ID获取操作日志
func (s *OperationLogService) GetByID(ctx context.Context, id uint) (*model.OperationLog, error) {
	return s.repo.GetByID(ctx, id)
}
