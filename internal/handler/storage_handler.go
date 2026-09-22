package handler

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"go-api-starter/internal/transport"

	"go-api-starter/internal/config"
	"go-api-starter/internal/model"
	"go-api-starter/internal/platform/apperrors"
	"go-api-starter/internal/platform/logger"
	"go-api-starter/internal/platform/response"
	"go-api-starter/internal/service"
)

// Imports referenced for swagger auto-generation.
var _ = model.File{}

type StorageHandler struct {
	service     service.StorageServiceInterface
	userService service.UserServiceInterface
	taskManager *service.TaskManager
}

// NewStorageHandler creates a new StorageHandler.
func NewStorageHandler(svc service.StorageServiceInterface, userSvc service.UserServiceInterface) *StorageHandler {
	return &StorageHandler{service: svc, userService: userSvc}
}

func (h *StorageHandler) SetTaskManager(tm *service.TaskManager) { h.taskManager = tm }

// ============ 统一上传接口 ============

// UploadInitRequest 统一的上传初始化请求
type UploadInitRequest struct {
	FileName string `json:"file_name" binding:"required"`
	Size     int64  `json:"size" binding:"required"`
	Checksum string `json:"checksum" binding:"required"`
}

// UploadInit godoc
// @Summary 初始化上传
// @Description 统一的上传初始化接口，自动检查秒传，根据文件大小决定普通/分片上传
// @Tags 文件管理
// @Accept json
// @Produce json
// @Param request body UploadInitRequest true "初始化请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /api/v1/uploads [post]
func (h *StorageHandler) UploadInit(c *transport.Context) {
	var req UploadInitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.BadRequest("invalid request: " + err.Error()))
		return
	}

	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	result, err := h.service.CreateUpload(req.FileName, "", req.Checksum, req.Size, 0, userID)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, result)
}

// UploadCompleteRequest 统一的上传完成请求
type UploadCompleteRequest struct {
	Parts     []service.CompletePart       `json:"parts,omitempty"`
	IsPrivate bool                         `json:"is_private"`
	Metadata  *service.ClientMediaMetadata `json:"metadata,omitempty"`
}

// UploadComplete godoc
// @Summary 完成上传
// @Description 统一的上传完成接口，支持普通上传和分片上传
// @Tags 文件管理
// @Accept json
// @Produce json
// @Param request body UploadCompleteRequest true "完成请求"
// @Success 200 {object} response.Response{data=model.File}
// @Failure 400 {object} response.Response
// @Router /api/v1/uploads/{upload_id}/complete [post]
func (h *StorageHandler) UploadComplete(c *transport.Context) {
	var req UploadCompleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.BadRequest("invalid request: " + err.Error()))
		return
	}

	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	file, err := h.service.CompleteUpload(c.Param("upload_id"), userID, req.Parts, req.Metadata)
	if err != nil {
		c.Error(err)
		return
	}

	if req.IsPrivate {
		isPrivate := true
		_ = h.service.UpdateFile(file.UID, &model.UpdateFileRequest{IsPrivate: &isPrivate})
		file.IsPrivate = true
	}
	if h.taskManager != nil && h.taskManager.Enabled() && strings.HasPrefix(file.Type, "video/") && file.TranscodingTaskID == nil {
		if taskID, taskErr := h.taskManager.CreateTask(model.CreateTaskRequest{FileID: &file.ID, SourceURL: file.URL, Resolutions: []string{"original", "1080p", "720p", "480p"}}); taskErr == nil {
			file.TranscodingTaskID = &taskID
			_ = h.service.UpdateFileTranscodingTask(file.UID, taskID)
		}
	}

	response.Success(c, file)
}

// GetFile godoc
// @Summary 获取文件详情
// @Description 根据 UID 获取文件信息
// @Tags 文件管理
// @Accept json
// @Produce json
// @Param uid path string true "文件 UID"
// @Success 200 {object} response.Response{data=model.File}
// @Failure 404 {object} response.Response
// @Router /api/v1/files/{uid} [get]
func (h *StorageHandler) GetFile(c *transport.Context) {
	uid, ok := GetUID(c)
	if !ok {
		return
	}
	file, err := h.service.GetFileByUID(uid)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, file.ToSimpleResponse())
}

// ListFiles godoc
// @Summary 获取文件列表（可选认证）
// @Description 获取文件分页列表，未登录只返回公开文件
// @Tags 文件管理
// @Produce json
// @Param page query int false "页码（默认 1）"
// @Param page_size query int false "每页数量（默认 10）"
// @Param sort query string false "排序（如 created_at,desc）"
// @Param user_uid query string false "按用户 UID 筛选"
// @Param is_private query bool false "是否仅返回私密文件（需认证）"
// @Success 200 {object} response.Response
// @Router /api/v1/files [get]
func (h *StorageHandler) ListFiles(c *transport.Context) {
	p, ok := BindPagination(c)
	if !ok {
		return
	}

	var userID uint
	if uid := c.Query("user_uid"); uid != "" {
		user, err := h.userService.GetByUID(c.Request.Context(), uid)
		if err != nil {
			c.Error(err)
			return
		}
		userID = user.ID
	}

	// 未登录时强制仅返回公开文件
	currentUserID := GetOptionalUserID(c)
	isAuthenticated := currentUserID > 0

	var isPrivate *bool
	if isAuthenticated {
		if v := c.Query("is_private"); v != "" {
			b := v == "true" || v == "1"
			isPrivate = &b
		}
	} else {
		f := false
		isPrivate = &f
	}

	files, total, err := h.service.ListFiles(userID, isPrivate, p.GetOffset(), p.GetPageSize(), p.GetSort())
	if err != nil {
		c.Error(err)
		return
	}

	result := make([]*model.FileSimpleResponse, len(files))
	for i := range files {
		result[i] = files[i].ToSimpleResponse()
	}
	response.SuccessWithPage(c, result, total, p)
}

// DeleteFile godoc
// @Summary 删除文件
// @Description 从对象存储和数据库中删除文件
// @Tags 文件管理
// @Produce json
// @Security BearerAuth
// @Param uid path string true "文件 UID"
// @Success 200 {object} response.Response
// @Router /api/v1/files/{uid} [delete]
func (h *StorageHandler) DeleteFile(c *transport.Context) {
	uid, ok := GetUID(c)
	if !ok {
		return
	}
	if err := h.service.DeleteFile(uid); err != nil {
		c.Error(err)
		return
	}
	response.Success(c, nil)
}

// UpdateFile godoc
// @Summary 更新文件信息
// @Description 更新文件的名称或隐私设置
// @Tags 文件管理
// @Accept json
// @Produce json
// @Param uid path string true "文件 UID"
// @Param request body model.UpdateFileRequest true "更新请求"
// @Success 200 {object} response.Response{data=model.File}
// @Router /api/v1/files/{uid} [put]
func (h *StorageHandler) UpdateFile(c *transport.Context) {
	uid, ok := GetUID(c)
	if !ok {
		return
	}

	var req model.UpdateFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.BadRequest("invalid request: " + err.Error()))
		return
	}

	userID := GetOptionalUserID(c)

	file, err := h.service.GetFileByUID(uid)
	if err != nil {
		c.Error(err)
		return
	}
	if file.UserID != userID {
		c.Error(apperrors.Forbidden("you don't have permission to update this file"))
		return
	}
	if err := h.service.UpdateFile(uid, &req); err != nil {
		c.Error(err)
		return
	}

	file, err = h.service.GetFileByUID(uid)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, file)
}

// AbortUpload cancels an upload session owned by the current user.
func (h *StorageHandler) AbortUpload(c *transport.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	if err := h.service.AbortUpload(c.Param("upload_id"), userID); err != nil {
		c.Error(err)
		return
	}
	response.Success(c, nil)
}

// ============ 公开上传（无需鉴权）============

// PublicUpload godoc
// @Summary 公开文件上传（无需鉴权）
// @Description 直接上传文件到对象存储的 public 目录，不落库。适合头像等小文件场景。
// @Tags 文件管理
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "要上传的文件"
// @Success 200 {object} response.Response
// @Router /api/v1/files/public/upload [post]
func (h *StorageHandler) PublicUpload(c *transport.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.Error(apperrors.BadRequest("file is required: " + err.Error()))
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.Error(apperrors.Internal(err, "failed to open uploaded file"))
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	dateDir := time.Now().Format("2006-01-02")
	fileName := uuid.New().String() + ext

	cfg := config.GetConfig()
	var objectKey string
	if cfg != nil && cfg.Storage.UploadDir != "" {
		objectKey = fmt.Sprintf("%s/public/%s/%s", cfg.Storage.UploadDir, dateDir, fileName)
	} else {
		objectKey = fmt.Sprintf("public/%s/%s", dateDir, fileName)
	}

	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" || contentType == "application/octet-stream" {
		contentType = inferContentTypeFromExt(ext)
	}

	result, err := h.service.UploadPublic(c.Request.Context(), objectKey, file, contentType)
	if err != nil {
		logger.Log.Errorf("public upload failed: %v", err)
		c.Error(apperrors.Internal(err, "failed to upload file to object storage"))
		return
	}

	response.Success(c, transport.H{
		"key":  result.Key,
		"url":  result.URL,
		"name": fileHeader.Filename,
		"size": fileHeader.Size,
		"type": contentType,
	})
}

// inferContentTypeFromExt infers MIME type from file extension (handler-local fallback).
func inferContentTypeFromExt(ext string) string {
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".pdf":
		return "application/pdf"
	default:
		return "application/octet-stream"
	}
}
