package handler

import (
	"go-api-starter/internal/model"
	"go-api-starter/internal/service"
	"go-api-starter/pkg/apperrors"
	"go-api-starter/pkg/oss"
	"go-api-starter/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Ensure packages are imported for swagger
var _ = oss.UploadToken{}
var _ = model.OSSFile{}

type OSSHandler struct {
	service service.OSSServiceInterface
}

func NewOSSHandler(service service.OSSServiceInterface) *OSSHandler {
	return &OSSHandler{service: service}
}

// GetUploadToken godoc
// @Summary 获取上传令牌
// @Description 获取客户端直传 OSS 的上传令牌，或通过 MD5 检查文件是否已存在
// @Tags OSS文件管理
// @Accept json
// @Produce json
// @Param md5 query string false "文件 MD5 哈希值（用于检查文件是否存在）"
// @Param file_name query string false "文件名（可选，用于扩展名验证）"
// @Success 200 {object} response.Response{data=oss.UploadToken}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/oss/token [get]
func (h *OSSHandler) GetUploadToken(c *gin.Context) {
	md5 := c.Query("md5")
	fileName := c.Query("file_name")

	// Get user ID from context (if authentication is enabled)
	var userID uint
	if uid, exists := c.Get("user_id"); exists {
		userID = uid.(uint)
	}

	// If MD5 is provided, check if file already exists
	if md5 != "" {
		file, exists := h.service.CheckFileExists(md5, userID)
		if exists {
			response.Success(c, gin.H{
				"exists": true,
				"file":   file,
			})
			return
		}
	}

	// File doesn't exist, generate token
	token, err := h.service.GetUploadToken(fileName, userID)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, gin.H{
		"exists": false,
		"token":  token,
	})
}

// CallbackRequest represents the callback request from OSS
type CallbackRequest struct {
	Key         string `json:"key" binding:"required"`
	MD5         string `json:"md5" binding:"required"`
	FileName    string `json:"file_name" binding:"required"`
	FileSize    int64  `json:"file_size" binding:"required"`
	ContentType string `json:"content_type"`
}

// Callback godoc
// @Summary OSS 上传回调
// @Description 处理 OSS 上传成功后的回调请求
// @Tags OSS文件管理
// @Accept json
// @Produce json
// @Param request body CallbackRequest true "回调请求数据"
// @Success 200 {object} response.Response{data=model.OSSFile}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/oss/callback [post]
func (h *OSSHandler) Callback(c *gin.Context) {
	var req CallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.BadRequest("invalid request: " + err.Error()))
		return
	}

	// Get user ID from context (if authentication is enabled)
	var userID uint
	if uid, exists := c.Get("user_id"); exists {
		userID = uid.(uint)
	}

	file, err := h.service.SaveFileRecord(req.Key, req.MD5, req.FileName, req.FileSize, req.ContentType, userID)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, file)
}

// ListFiles godoc
// @Summary 获取文件列表
// @Description 获取已上传文件的分页列表
// @Tags OSS文件管理
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/oss/files [get]
func (h *OSSHandler) ListFiles(c *gin.Context) {
	var pagination response.Pagination
	if err := c.ShouldBindQuery(&pagination); err != nil {
		c.Error(apperrors.BadRequest("invalid pagination parameters"))
		return
	}

	// Get user ID from context (if authentication is enabled)
	var userID uint
	if uid, exists := c.Get("user_id"); exists {
		userID = uid.(uint)
	}

	files, total, err := h.service.ListFiles(userID, pagination.GetPage(), pagination.GetPageSize())
	if err != nil {
		c.Error(err)
		return
	}

	response.SuccessWithPage(c, files, total, &pagination)
}

// DeleteFile godoc
// @Summary 删除文件
// @Description 从 OSS 和数据库中删除文件
// @Tags OSS文件管理
// @Accept json
// @Produce json
// @Param id path int true "文件ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/oss/files/{id} [delete]
func (h *OSSHandler) DeleteFile(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.Error(apperrors.BadRequest("invalid file id"))
		return
	}

	if err := h.service.DeleteFile(uint(id)); err != nil {
		c.Error(err)
		return
	}

	response.Success(c, nil)
}


// ============ Multipart Upload (分片上传) ============

// InitMultipartRequest represents the request to initialize multipart upload
type InitMultipartRequest struct {
	FileName  string `json:"file_name" binding:"required"`
	MD5       string `json:"md5"`
	FileSize  int64  `json:"file_size" binding:"required"`
	ChunkSize int64  `json:"chunk_size" binding:"required"`
}

// InitMultipart godoc
// @Summary 初始化分片上传
// @Description 初始化分片上传，返回 uploadId 和 key。如果存在相同MD5的未完成上传，会返回已上传的分片列表用于断点续传
// @Tags OSS文件管理
// @Accept json
// @Produce json
// @Param request body InitMultipartRequest true "初始化请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /api/v1/oss/multipart/init [post]
func (h *OSSHandler) InitMultipart(c *gin.Context) {
	var req InitMultipartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.BadRequest("invalid request: " + err.Error()))
		return
	}

	var userID uint
	if uid, exists := c.Get("user_id"); exists {
		userID = uid.(uint)
	}

	// Check if file already exists by MD5
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

	result, err := h.service.InitMultipartUpload(req.FileName, req.MD5, req.FileSize, req.ChunkSize, userID)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, gin.H{
		"exists":         false,
		"upload_id":      result.UploadID,
		"key":            result.Key,
		"host":           result.Host,
		"total_parts":    result.TotalParts,
		"chunk_size":     result.ChunkSize,
		"uploaded_parts": result.UploadedParts,
	})
}

// GetPartURLRequest represents the request to get part upload URLs
type GetPartURLRequest struct {
	Key         string `json:"key" binding:"required"`
	UploadID    string `json:"upload_id" binding:"required"`
	PartNumbers []int  `json:"part_numbers" binding:"required"`
}

// GetPartUploadURLs godoc
// @Summary 获取分片上传URL
// @Description 获取分片上传的预签名URL
// @Tags OSS文件管理
// @Accept json
// @Produce json
// @Param request body GetPartURLRequest true "请求参数"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /api/v1/oss/multipart/urls [post]
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

	response.Success(c, gin.H{
		"urls": urls,
	})
}

// CompleteMultipartRequest represents the request to complete multipart upload
type CompleteMultipartRequest struct {
	Key         string                `json:"key" binding:"required"`
	UploadID    string                `json:"upload_id" binding:"required"`
	MD5         string                `json:"md5" binding:"required"`
	FileName    string                `json:"file_name" binding:"required"`
	FileSize    int64                 `json:"file_size" binding:"required"`
	ContentType string                `json:"content_type"`
	Parts       []service.CompletePart `json:"parts" binding:"required"`
}

// CompleteMultipart godoc
// @Summary 完成分片上传
// @Description 完成分片上传，合并所有分片
// @Tags OSS文件管理
// @Accept json
// @Produce json
// @Param request body CompleteMultipartRequest true "完成请求"
// @Success 200 {object} response.Response{data=model.OSSFile}
// @Failure 400 {object} response.Response
// @Router /api/v1/oss/multipart/complete [post]
func (h *OSSHandler) CompleteMultipart(c *gin.Context) {
	var req CompleteMultipartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.BadRequest("invalid request: " + err.Error()))
		return
	}

	var userID uint
	if uid, exists := c.Get("user_id"); exists {
		userID = uid.(uint)
	}

	file, err := h.service.CompleteMultipartUpload(
		req.Key, req.UploadID, req.MD5, req.FileName,
		req.FileSize, req.ContentType, req.Parts, userID,
	)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, file)
}

// AbortMultipartRequest represents the request to abort multipart upload
type AbortMultipartRequest struct {
	Key      string `json:"key" binding:"required"`
	UploadID string `json:"upload_id" binding:"required"`
}

// AbortMultipart godoc
// @Summary 取消分片上传
// @Description 取消分片上传，清理已上传的分片
// @Tags OSS文件管理
// @Accept json
// @Produce json
// @Param request body AbortMultipartRequest true "取消请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /api/v1/oss/multipart/abort [post]
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

// ListPartsRequest represents the request to list uploaded parts
type ListPartsRequest struct {
	Key      string `form:"key" binding:"required"`
	UploadID string `form:"upload_id" binding:"required"`
}

// ListParts godoc
// @Summary 获取已上传分片列表
// @Description 获取已上传的分片列表，用于断点续传
// @Tags OSS文件管理
// @Accept json
// @Produce json
// @Param key query string true "文件key"
// @Param upload_id query string true "上传ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /api/v1/oss/multipart/parts [get]
func (h *OSSHandler) ListParts(c *gin.Context) {
	var req ListPartsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.Error(apperrors.BadRequest("invalid request: " + err.Error()))
		return
	}

	parts, err := h.service.ListUploadedParts(req.Key, req.UploadID)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, gin.H{
		"parts": parts,
	})
}

// SavePartRequest represents the request to save uploaded part info
type SavePartRequest struct {
	UploadID   string `json:"upload_id" binding:"required"`
	PartNumber int    `json:"part_number" binding:"required"`
	ETag       string `json:"etag" binding:"required"`
	Size       int64  `json:"size" binding:"required"`
}

// SavePart godoc
// @Summary 保存已上传分片信息
// @Description 保存已上传分片信息到数据库，用于断点续传
// @Tags OSS文件管理
// @Accept json
// @Produce json
// @Param request body SavePartRequest true "分片信息"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /api/v1/oss/multipart/part [post]
func (h *OSSHandler) SavePart(c *gin.Context) {
	var req SavePartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.BadRequest("invalid request: " + err.Error()))
		return
	}

	if err := h.service.SaveUploadedPart(req.UploadID, req.PartNumber, req.ETag, req.Size); err != nil {
		c.Error(err)
		return
	}

	response.Success(c, nil)
}

// GetUploadedPartsFromDB godoc
// @Summary 从数据库获取已上传分片列表
// @Description 从数据库获取已上传的分片列表，用于断点续传时获取完整的分片信息
// @Tags OSS文件管理
// @Accept json
// @Produce json
// @Param upload_id query string true "上传ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /api/v1/oss/multipart/db-parts [get]
func (h *OSSHandler) GetUploadedPartsFromDB(c *gin.Context) {
	uploadID := c.Query("upload_id")
	if uploadID == "" {
		c.Error(apperrors.BadRequest("upload_id is required"))
		return
	}

	parts, err := h.service.GetUploadedPartsFromDB(uploadID)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, gin.H{
		"parts": parts,
	})
}
