package handler

import (
	"go-api-starter/internal/service"
	"go-api-starter/pkg/response"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type OSSHandler struct {
	service *service.OSSService
}

func NewOSSHandler(service *service.OSSService) *OSSHandler {
	return &OSSHandler{service: service}
}

// GetUploadToken godoc
// @Summary Get upload token
// @Description Get upload token for client-side direct upload to OSS, or check if file exists by MD5
// @Tags OSS
// @Accept json
// @Produce json
// @Param md5 query string false "File MD5 hash (for checking if file exists)"
// @Param file_name query string false "File name (optional, for extension validation)"
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
		response.Error(c, http.StatusBadRequest, err.Error())
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
// @Summary OSS upload callback
// @Description Handle callback from OSS after successful upload
// @Tags OSS
// @Accept json
// @Produce json
// @Param request body CallbackRequest true "Callback request"
// @Success 200 {object} response.Response{data=model.OSSFile}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/oss/callback [post]
func (h *OSSHandler) Callback(c *gin.Context) {
	var req CallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	// Get user ID from context (if authentication is enabled)
	var userID uint
	if uid, exists := c.Get("user_id"); exists {
		userID = uid.(uint)
	}

	file, err := h.service.SaveFileRecord(req.Key, req.MD5, req.FileName, req.FileSize, req.ContentType, userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, file)
}

// ListFiles godoc
// @Summary List files
// @Description List uploaded files with pagination
// @Tags OSS
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/oss/files [get]
func (h *OSSHandler) ListFiles(c *gin.Context) {
	var pagination response.Pagination
	if err := c.ShouldBindQuery(&pagination); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid pagination parameters")
		return
	}

	// Get user ID from context (if authentication is enabled)
	var userID uint
	if uid, exists := c.Get("user_id"); exists {
		userID = uid.(uint)
	}

	files, total, err := h.service.ListFiles(userID, pagination.GetPage(), pagination.GetPageSize())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.SuccessWithPage(c, files, total, &pagination)
}

// DeleteFile godoc
// @Summary Delete file
// @Description Delete file from OSS and database
// @Tags OSS
// @Accept json
// @Produce json
// @Param id path int true "File ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/oss/files/{id} [delete]
func (h *OSSHandler) DeleteFile(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid file id")
		return
	}

	if err := h.service.DeleteFile(uint(id)); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, nil)
}
