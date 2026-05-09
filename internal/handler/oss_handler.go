package handler

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"go-api-starter/internal/config"
	"go-api-starter/internal/model"
	"go-api-starter/internal/service"
	"go-api-starter/pkg/apperrors"
	"go-api-starter/pkg/logger"
	"go-api-starter/pkg/oss"
	"go-api-starter/pkg/response"
)

// Imports referenced for swagger auto-generation
var (
	_ = oss.UploadToken{}
	_ = model.File{}
)

type OSSHandler struct {
	service     service.OSSServiceInterface
	userService service.UserServiceInterface
}

// NewOSSHandler creates a new OSSHandler.
func NewOSSHandler(svc service.OSSServiceInterface, userSvc service.UserServiceInterface) *OSSHandler {
	return &OSSHandler{service: svc, userService: userSvc}
}

// ============ 统一上传接口 ============

// UploadInitRequest 统一的上传初始化请求
type UploadInitRequest struct {
	FileName  string `json:"file_name" binding:"required"`
	FileSize  int64  `json:"file_size" binding:"required"`
	MD5       string `json:"md5" binding:"required"`
	ChunkSize int64  `json:"chunk_size"` // 可选，不传则使用默认值或普通上传
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
// @Router /api/v1/file/upload/init [post]
func (h *OSSHandler) UploadInit(c *gin.Context) {
	var req UploadInitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.BadRequest("invalid request: " + err.Error()))
		return
	}

	userID := GetOptionalUserID(c)

	// 1. 检查秒传
	if req.MD5 != "" {
		file, exists := h.service.CheckFileExists(req.MD5, userID)
		if exists {
			response.Success(c, gin.H{
				"exists": true,
				"file":   file,
			})
			return
		}
	}

	// 2. 根据文件大小决定上传方式：< 5MB 普通上传，>= 5MB 分片上传
	const multipartThreshold = 5 * 1024 * 1024

	if req.FileSize < multipartThreshold {
		token, err := h.service.GetUploadTokenWithFileName(userID, req.FileName)
		if err != nil {
			c.Error(err)
			return
		}
		response.Success(c, gin.H{
			"exists": false,
			"mode":   "simple",
			"token":  token,
		})
		return
	}

	chunkSize := req.ChunkSize
	if chunkSize <= 0 {
		chunkSize = 5 * 1024 * 1024 // 默认 5MB
	}

	result, err := h.service.InitMultipartUpload(req.FileName, req.MD5, req.FileSize, chunkSize, userID)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, gin.H{
		"exists":         false,
		"mode":           "multipart",
		"upload_id":      result.UploadID,
		"key":            result.Key,
		"host":           result.Host,
		"total_parts":    result.TotalParts,
		"chunk_size":     result.ChunkSize,
		"uploaded_parts": result.UploadedParts,
	})
}

// UploadCompleteRequest 统一的上传完成请求
type UploadCompleteRequest struct {
	Key       string                 `json:"key" binding:"required"`
	MD5       string                 `json:"md5" binding:"required"`
	FileName  string                 `json:"file_name" binding:"required"`
	FileSize  int64                  `json:"file_size" binding:"required"`
	IsPrivate *bool                  `json:"is_private"`
	UploadID  string                 `json:"upload_id"` // 分片上传专用
	Parts     []service.CompletePart `json:"parts"`     // 分片上传专用
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
// @Router /api/v1/file/upload/complete [post]
func (h *OSSHandler) UploadComplete(c *gin.Context) {
	var req UploadCompleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.BadRequest("invalid request: " + err.Error()))
		return
	}

	userID := GetOptionalUserID(c)

	var (
		file *model.File
		err  error
	)

	if req.UploadID != "" && len(req.Parts) > 0 {
		file, err = h.service.CompleteMultipartUpload(
			req.Key, req.UploadID, req.MD5, req.FileName,
			req.FileSize, req.Parts, userID,
		)
	} else {
		file, err = h.service.SaveFileRecord(req.Key, req.MD5, req.FileName, req.FileSize, userID)
	}
	if err != nil {
		c.Error(err)
		return
	}

	if req.IsPrivate != nil && *req.IsPrivate {
		isPrivate := true
		_ = h.service.UpdateFile(file.SecUID, &model.UpdateFileRequest{IsPrivate: &isPrivate})
		file.IsPrivate = true
	}

	response.Success(c, file)
}

// GetFile godoc
// @Summary 获取文件详情
// @Description 根据 SecUID 获取文件信息
// @Tags 文件管理
// @Accept json
// @Produce json
// @Param sec_uid path string true "文件 SecUID"
// @Success 200 {object} response.Response{data=model.File}
// @Failure 404 {object} response.Response
// @Router /api/v1/file/{sec_uid} [get]
func (h *OSSHandler) GetFile(c *gin.Context) {
	secUID, ok := GetSecUID(c)
	if !ok {
		return
	}
	file, err := h.service.GetFileBySecUID(secUID)
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
// @Param user_sec_uid query string false "按用户 SecUID 筛选"
// @Param is_private query bool false "是否仅返回私密文件（需认证）"
// @Success 200 {object} response.Response
// @Router /api/v1/file [get]
func (h *OSSHandler) ListFiles(c *gin.Context) {
	p, ok := BindPagination(c)
	if !ok {
		return
	}

	var userID uint
	if secUID := c.Query("user_sec_uid"); secUID != "" {
		user, err := h.userService.GetBySecUID(c.Request.Context(), secUID)
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
// @Description 从 OSS 和数据库中删除文件
// @Tags 文件管理
// @Produce json
// @Security BearerAuth
// @Param sec_uid path string true "文件 SecUID"
// @Success 200 {object} response.Response
// @Router /api/v1/file/{sec_uid} [delete]
func (h *OSSHandler) DeleteFile(c *gin.Context) {
	secUID, ok := GetSecUID(c)
	if !ok {
		return
	}
	if err := h.service.DeleteFile(secUID); err != nil {
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
// @Param sec_uid path string true "文件 SecUID"
// @Param request body model.UpdateFileRequest true "更新请求"
// @Success 200 {object} response.Response{data=model.File}
// @Router /api/v1/file/{sec_uid} [put]
func (h *OSSHandler) UpdateFile(c *gin.Context) {
	secUID, ok := GetSecUID(c)
	if !ok {
		return
	}

	var req model.UpdateFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.BadRequest("invalid request: " + err.Error()))
		return
	}

	userID := GetOptionalUserID(c)

	file, err := h.service.GetFileBySecUID(secUID)
	if err != nil {
		c.Error(err)
		return
	}
	if file.UserID != userID {
		c.Error(apperrors.Forbidden("you don't have permission to update this file"))
		return
	}
	if err := h.service.UpdateFile(secUID, &req); err != nil {
		c.Error(err)
		return
	}

	file, err = h.service.GetFileBySecUID(secUID)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, file)
}

// ============ 分片上传辅助接口 ============

type GetPartURLRequest struct {
	Key         string `json:"key" binding:"required"`
	UploadID    string `json:"upload_id" binding:"required"`
	PartNumbers []int  `json:"part_numbers" binding:"required"`
}

// GetPartUploadURLs godoc
// @Summary 获取分片上传URL
// @Description 获取分片上传的预签名 URL
// @Tags 文件管理
// @Accept json
// @Produce json
// @Param request body GetPartURLRequest true "请求参数"
// @Success 200 {object} response.Response
// @Router /api/v1/file/upload/urls [post]
func (h *OSSHandler) GetPartUploadURLs(c *gin.Context) {
	var req GetPartURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.BadRequest("invalid request: " + err.Error()))
		return
	}
	urls, err := h.service.GetPartUploadURLs(req.Key, req.UploadID, req.PartNumbers)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, gin.H{"urls": urls})
}

// AbortMultipartRequest represents the request to abort a multipart upload
type AbortMultipartRequest struct {
	Key      string `json:"key" binding:"required"`
	UploadID string `json:"upload_id" binding:"required"`
}

// AbortMultipart godoc
// @Summary 取消分片上传
// @Description 取消分片上传并清理已上传的分片
// @Tags 文件管理
// @Accept json
// @Produce json
// @Param request body AbortMultipartRequest true "取消请求"
// @Success 200 {object} response.Response
// @Router /api/v1/file/upload/abort [post]
func (h *OSSHandler) AbortMultipart(c *gin.Context) {
	var req AbortMultipartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.BadRequest("invalid request: " + err.Error()))
		return
	}
	if err := h.service.AbortMultipartUpload(req.Key, req.UploadID); err != nil {
		c.Error(err)
		return
	}
	response.Success(c, nil)
}

// ============ 公开上传（无需鉴权）============

// PublicUpload godoc
// @Summary 公开文件上传（无需鉴权）
// @Description 直接上传文件到 OSS 的 public 目录，不落库。适合头像等小文件场景。
// @Tags 文件管理
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "要上传的文件"
// @Success 200 {object} response.Response
// @Router /api/v1/file/public/upload [post]
func (h *OSSHandler) PublicUpload(c *gin.Context) {
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
	if cfg != nil && cfg.OSS.UploadDir != "" {
		objectKey = fmt.Sprintf("%s/public/%s/%s", cfg.OSS.UploadDir, dateDir, fileName)
	} else {
		objectKey = fmt.Sprintf("public/%s/%s", dateDir, fileName)
	}

	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" || contentType == "application/octet-stream" {
		contentType = inferContentTypeFromExt(ext)
	}

	result, err := oss.UploadFile(file, fileHeader, objectKey)
	if err != nil {
		logger.Log.Errorf("public upload failed: %v", err)
		c.Error(apperrors.Internal(err, "failed to upload file to OSS"))
		return
	}

	response.Success(c, gin.H{
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
