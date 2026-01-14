package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"go-api-starter/internal/model"
	"go-api-starter/pkg/response"
)

type AuthHandler struct {
	db        *gorm.DB
	jwtSecret string
}

func NewAuthHandler(db *gorm.DB, jwtSecret string) *AuthHandler {
	return &AuthHandler{
		db:        db,
		jwtSecret: jwtSecret,
	}
}

// Register godoc
// @Summary 注册新用户
// @Description 注册一个新的用户账号
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body model.RegisterRequest true "注册请求数据"
// @Success 201 {object} response.Response{data=model.User}
// @Failure 400 {object} response.Response
// @Failure 409 {object} response.Response
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Check if email already exists
	var existingUser model.User
	if err := h.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		response.Conflict(c, "邮箱已被注册")
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		response.InternalError(c, "密码加密失败")
		return
	}

	// Create user
	user := &model.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
		Age:      req.Age,
	}

	if err := h.db.Create(user).Error; err != nil {
		response.InternalError(c, "创建用户失败")
		return
	}

	response.Created(c, user)
}

// Login godoc
// @Summary 用户登录
// @Description 使用邮箱和密码登录
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body model.LoginRequest true "登录请求数据"
// @Success 200 {object} response.Response{data=model.LoginResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Find user by email
	var user model.User
	if err := h.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		response.Unauthorized(c, "邮箱或密码错误")
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		response.Unauthorized(c, "邮箱或密码错误")
		return
	}

	// Generate JWT token
	token, err := h.generateToken(user.ID)
	if err != nil {
		response.InternalError(c, "生成令牌失败")
		return
	}

	loginResp := &model.LoginResponse{
		Token: token,
		User:  &user,
	}

	response.Success(c, loginResp)
}

// GetCurrentUser godoc
// @Summary 获取当前用户信息
// @Description 获取当前已认证用户的详细信息
// @Tags 认证
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=model.User}
// @Failure 401 {object} response.Response
// @Router /api/v1/auth/me [get]
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "用户未认证")
		return
	}

	userID, ok := userIDVal.(uint)
	if !ok {
		response.Unauthorized(c, "用户ID无效")
		return
	}

	var user model.User
	if err := h.db.First(&user, userID).Error; err != nil {
		response.NotFound(c, "用户不存在")
		return
	}

	response.Success(c, &user)
}

func (h *AuthHandler) generateToken(userID uint) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(), // 7 days
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.jwtSecret))
}

// ResetPassword godoc
// @Summary 重置用户密码
// @Description 重置指定用户的密码（仅管理员）
// @Tags 认证
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Param request body model.ResetPasswordRequest true "重置密码请求数据"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/auth/reset-password/{id} [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	userID := c.Param("id")

	var req model.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Find user
	var user model.User
	if err := h.db.First(&user, userID).Error; err != nil {
		response.NotFound(c, "用户不存在")
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		response.InternalError(c, "密码加密失败")
		return
	}

	// Update password
	if err := h.db.Model(&user).Update("password", string(hashedPassword)).Error; err != nil {
		response.InternalError(c, "密码重置失败")
		return
	}

	response.Success(c, gin.H{"message": "密码重置成功"})
}
