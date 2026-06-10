package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"go-api-starter/internal/config"
	"go-api-starter/internal/model"
	"go-api-starter/internal/repository"
	"go-api-starter/pkg/apperrors"
	"go-api-starter/pkg/auth"
	"go-api-starter/pkg/i18n"
)

const (
	wxJsCode2SessionURL = "https://api.weixin.qq.com/sns/jscode2session"
	wxGetAccessTokenURL = "https://api.weixin.qq.com/cgi-bin/token"
	wxGetPhoneNumberURL = "https://api.weixin.qq.com/wxa/business/getuserphonenumber"
)

type wxSessionResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

type wxAccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
}

type wxPhoneNumberResponse struct {
	ErrCode   int    `json:"errcode"`
	ErrMsg    string `json:"errmsg"`
	PhoneInfo struct {
		PhoneNumber     string `json:"phoneNumber"`
		PurePhoneNumber string `json:"purePhoneNumber"`
		CountryCode     string `json:"countryCode"`
	} `json:"phone_info"`
}

// WechatService handles WeChat mini-program login logic.
type WechatService struct {
	cfg             *config.WechatConfig
	defaultPassword string
	userRepo        repository.UserRepositoryInterface
	jwtManager      *auth.JWTManager
	httpClient      *http.Client

	// access_token 缓存
	tokenMu      sync.Mutex
	accessToken  string
	tokenExpires time.Time
}

// NewWechatService creates a new WechatService.
func NewWechatService(cfg *config.WechatConfig, defaultPassword string, userRepo repository.UserRepositoryInterface, jwtManager *auth.JWTManager) *WechatService {
	return &WechatService{
		cfg:             cfg,
		defaultPassword: defaultPassword,
		userRepo:        userRepo,
		jwtManager:      jwtManager,
		httpClient:      &http.Client{Timeout: 10 * time.Second},
	}
}

// Login handles the unified wx-login flow:
//  1. js_code + phone_code 同时传入 → code2session + 获取手机号 → 一步完成登录/注册
//  2. 仅 js_code → code2session → 返回 JWT(已有用户) 或 need_bind=true(新用户)
//  3. open_id + phone_code/mobile → 绑定阶段（兼容分步流程）
func (s *WechatService) Login(ctx context.Context, req *model.WxLoginRequest) (*model.WxLoginResponse, error) {
	log.Printf("[WxLogin] === 开始处理 === js_code=%q, open_id=%v, phone_code=%q, mobile=%v",
		req.JsCode, req.OpenID, req.PhoneCode, req.Mobile)

	if s.cfg.AppID == "" || s.cfg.Secret == "" {
		log.Printf("[WxLogin] 错误: AppID 或 Secret 未配置")
		return nil, apperrors.InternalCode(nil, i18n.ErrWechatNotConfigured)
	}

	// 分步绑定阶段（兼容）: open_id + phone_code/mobile
	if req.OpenID != nil && *req.OpenID != "" {
		log.Printf("[WxLogin] 进入绑定阶段, open_id=%s", *req.OpenID)
		return s.bindAndLogin(ctx, req)
	}

	// 必须有 js_code
	if req.JsCode == "" {
		log.Printf("[WxLogin] 错误: js_code 为空")
		return nil, apperrors.BadRequestCode(i18n.ErrParamInvalid)
	}

	// code2session
	log.Printf("[WxLogin] 调用 code2Session, js_code=%s", req.JsCode)
	session, err := s.code2Session(req.JsCode)
	if err != nil {
		log.Printf("[WxLogin] code2Session 失败: %v", err)
		return nil, apperrors.WrapCode(err, i18n.ErrWechatLoginFailed)
	}

	log.Printf("[WxLogin] code2Session 结果: openid=%s, errcode=%d, errmsg=%s",
		session.OpenID, session.ErrCode, session.ErrMsg)

	if session.OpenID == "" {
		log.Printf("[WxLogin] 错误: openid 为空, errcode=%d, errmsg=%s", session.ErrCode, session.ErrMsg)
		return nil, apperrors.BadRequestCode(i18n.ErrWechatLoginFailed)
	}

	// Check if user exists
	user, err := s.userRepo.FindByOpenID(ctx, session.OpenID)
	if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
		log.Printf("[WxLogin] 查询用户失败: %v", err)
		return nil, apperrors.WrapCode(err, i18n.ErrQueryUserFailed)
	}

	if user != nil {
		log.Printf("[WxLogin] ✅ 用户已存在, uid=%s, mobile=%v", user.UID, user.Mobile)
		return s.buildLoginResponse(user)
	}

	// 新用户：如果同时带了 phone_code，直接一步完成注册
	if req.PhoneCode != "" {
		log.Printf("[WxLogin] 新用户 + phone_code, 一步完成注册绑定")
		openID := session.OpenID
		req.OpenID = &openID
		return s.bindAndLogin(ctx, req)
	}

	// 新用户，没带 phone_code，需要绑定
	log.Printf("[WxLogin] 新用户, openid=%s, 需要绑定手机号", session.OpenID)
	return &model.WxLoginResponse{
		NeedBind: true,
		OpenID:   &session.OpenID,
	}, nil
}

// bindAndLogin handles the second-step binding (open_id + phone_code or mobile).
func (s *WechatService) bindAndLogin(ctx context.Context, req *model.WxLoginRequest) (*model.WxLoginResponse, error) {
	openID := *req.OpenID
	log.Printf("[WxBind] === 绑定阶段 === open_id=%s, phone_code=%q, mobile=%v", openID, req.PhoneCode, req.Mobile)

	// Resolve mobile: prefer phone_code (微信授权获取), fallback to manual mobile
	var mobile string
	if req.PhoneCode != "" {
		log.Printf("[WxBind] 使用 phone_code 获取手机号...")
		phoneNumber, err := s.getPhoneNumber(req.PhoneCode)
		if err != nil {
			log.Printf("[WxBind] ❌ getPhoneNumber 失败: %v", err)
			return nil, apperrors.WrapCode(err, i18n.ErrWechatLoginFailed)
		}
		mobile = phoneNumber
		log.Printf("[WxBind] ✅ 解析到手机号: %s", mobile)
	} else if req.Mobile != nil && *req.Mobile != "" {
		mobile = *req.Mobile
		log.Printf("[WxBind] 使用手动输入的手机号: %s", mobile)
	} else {
		log.Printf("[WxBind] 错误: phone_code 和 mobile 都为空")
		return nil, apperrors.BadRequestCode(i18n.ErrProvideMobileOrEmail)
	}

	// Check if openid already bound
	existingByOpenID, err := s.userRepo.FindByOpenID(ctx, openID)
	if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
		log.Printf("[WxBind] 查询 openid 用户失败: %v", err)
		return nil, apperrors.WrapCode(err, i18n.ErrQueryUserFailed)
	}
	if existingByOpenID != nil {
		log.Printf("[WxBind] openid 已绑定用户 uid=%s, 直接登录", existingByOpenID.UID)
		return s.buildLoginResponse(existingByOpenID)
	}

	// Check if mobile already has an account
	existing, err := s.userRepo.FindByMobile(ctx, mobile)
	if err == nil && existing != nil {
		log.Printf("[WxBind] 手机号 %s 已有账号 uid=%s", mobile, existing.UID)
		// Mobile exists — bind openid to it
		if existing.OpenID != nil && *existing.OpenID != "" {
			log.Printf("[WxBind] ❌ 手机号已绑定其他微信: open_id=%s", *existing.OpenID)
			return nil, apperrors.ConflictCode(i18n.ErrMobileBoundToWechat)
		}
		existing.OpenID = &openID
		if err := s.userRepo.Update(ctx, existing); err != nil {
			log.Printf("[WxBind] ❌ 绑定 openid 到已有用户失败: %v", err)
			return nil, apperrors.WrapCode(err, i18n.ErrWechatBindFailed)
		}
		log.Printf("[WxBind] ✅ 已绑定 openid 到已有用户 uid=%s", existing.UID)
		return s.buildLoginResponse(existing)
	}

	// Create new user with openid + mobile + default password
	log.Printf("[WxBind] 创建新用户: openid=%s, mobile=%s", openID, mobile)
	var hashedPwd *string
	if s.defaultPassword != "" {
		hasher := auth.NewPasswordHasher()
		h, err := hasher.HashPassword(s.defaultPassword)
		if err != nil {
			log.Printf("[WxBind] ⚠️ 默认密码哈希失败: %v, 跳过设置密码", err)
		} else {
			hashedPwd = &h
		}
	}
	newUser := &model.User{
		OpenID:   &openID,
		Mobile:   &mobile,
		Password: hashedPwd,
		Freezed:  false,
	}
	// 微信用户默认用户名
	wxUsername := "微信用户"
	newUser.Username = &wxUsername
	if err := s.userRepo.Create(ctx, newUser); err != nil {
		log.Printf("[WxBind] ❌ 创建用户失败: %v", err)
		return nil, apperrors.WrapCode(err, i18n.ErrCreateUserFailed)
	}

	// Reload to get generated fields (UID, LPID, etc.)
	newUser, err = s.userRepo.FindByID(ctx, newUser.ID)
	if err != nil {
		log.Printf("[WxBind] ❌ 重新查询用户失败: %v", err)
		return nil, apperrors.WrapCode(err, i18n.ErrQueryUserFailed)
	}

	log.Printf("[WxBind] ✅ 新用户创建成功: uid=%s, lp_id=%s, mobile=%s", newUser.UID, newUser.LPID, mobile)
	return s.buildLoginResponse(newUser)
}

// getPhoneNumber calls WeChat API to get phone number from the phone_code.
func (s *WechatService) getPhoneNumber(phoneCode string) (string, error) {
	log.Printf("[WxPhone] 开始获取手机号, phone_code=%s", phoneCode)

	token, err := s.getAccessToken()
	if err != nil {
		log.Printf("[WxPhone] ❌ 获取 access_token 失败: %v", err)
		return "", fmt.Errorf("get access_token failed: %w", err)
	}
	log.Printf("[WxPhone] access_token 获取成功 (前20字符): %s...", token[:min(20, len(token))])

	url := fmt.Sprintf("%s?access_token=%s", wxGetPhoneNumberURL, token)

	body, _ := json.Marshal(map[string]string{"code": phoneCode})
	log.Printf("[WxPhone] 请求 URL: %s, body: %s", wxGetPhoneNumberURL, string(body))

	resp, err := s.httpClient.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		log.Printf("[WxPhone] ❌ HTTP 请求失败: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[WxPhone] ❌ 读取响应失败: %v", err)
		return "", err
	}

	log.Printf("[WxPhone] 微信响应: status=%d, body=%s", resp.StatusCode, string(respBody))

	var result wxPhoneNumberResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		log.Printf("[WxPhone] ❌ 解析 JSON 失败: %v", err)
		return "", err
	}

	if result.ErrCode != 0 {
		log.Printf("[WxPhone] ❌ 微信返回错误: errcode=%d, errmsg=%s", result.ErrCode, result.ErrMsg)
		return "", fmt.Errorf("wx getPhoneNumber errcode=%d, errmsg=%s", result.ErrCode, result.ErrMsg)
	}

	phoneNumber := result.PhoneInfo.PurePhoneNumber
	if phoneNumber == "" {
		phoneNumber = result.PhoneInfo.PhoneNumber
	}
	log.Printf("[WxPhone] ✅ 手机号获取成功: %s (countryCode=%s)", phoneNumber, result.PhoneInfo.CountryCode)

	return phoneNumber, nil
}

// getAccessToken gets a cached or fresh access_token for the mini-program.
func (s *WechatService) getAccessToken() (string, error) {
	s.tokenMu.Lock()
	defer s.tokenMu.Unlock()

	// Return cached token if still valid (with 5 min buffer)
	if s.accessToken != "" && time.Now().Before(s.tokenExpires.Add(-5*time.Minute)) {
		log.Printf("[WxToken] 使用缓存的 access_token, 过期时间: %s", s.tokenExpires.Format("15:04:05"))
		return s.accessToken, nil
	}

	log.Printf("[WxToken] 缓存过期或为空, 重新获取 access_token...")
	url := fmt.Sprintf("%s?grant_type=client_credential&appid=%s&secret=%s",
		wxGetAccessTokenURL, s.cfg.AppID, s.cfg.Secret)

	resp, err := s.httpClient.Get(url)
	if err != nil {
		log.Printf("[WxToken] ❌ HTTP 请求失败: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[WxToken] ❌ 读取响应失败: %v", err)
		return "", err
	}

	log.Printf("[WxToken] 微信响应: %s", string(body))

	var result wxAccessTokenResponse
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("[WxToken] ❌ 解析 JSON 失败: %v", err)
		return "", err
	}

	if result.ErrCode != 0 {
		log.Printf("[WxToken] ❌ 微信返回错误: errcode=%d, errmsg=%s", result.ErrCode, result.ErrMsg)
		return "", fmt.Errorf("wx getAccessToken errcode=%d, errmsg=%s", result.ErrCode, result.ErrMsg)
	}

	s.accessToken = result.AccessToken
	s.tokenExpires = time.Now().Add(time.Duration(result.ExpiresIn) * time.Second)

	log.Printf("[WxToken] ✅ access_token 刷新成功, expires_in=%d秒, 过期时间: %s",
		result.ExpiresIn, s.tokenExpires.Format("15:04:05"))
	return s.accessToken, nil
}

// code2Session calls the WeChat jscode2session API.
func (s *WechatService) code2Session(jsCode string) (*wxSessionResponse, error) {
	url := fmt.Sprintf("%s?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		wxJsCode2SessionURL, s.cfg.AppID, s.cfg.Secret, jsCode)

	resp, err := s.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	log.Printf("[WxLogin] wx API response: %s", string(body))

	var session wxSessionResponse
	if err := json.Unmarshal(body, &session); err != nil {
		return nil, err
	}

	return &session, nil
}

// buildLoginResponse generates JWT tokens for the user.
func (s *WechatService) buildLoginResponse(user *model.User) (*model.WxLoginResponse, error) {
	accessToken, refreshToken, err := s.jwtManager.GenerateTokenPair(user.ID)
	if err != nil {
		return nil, apperrors.InternalCode(err, i18n.ErrGenerateTokenFailed)
	}

	return &model.WxLoginResponse{
		NeedBind:     false,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    s.jwtManager.AccessTokenExpiresIn(),
		User:         user.ToResponse(),
	}, nil
}
