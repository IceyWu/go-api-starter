package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"go-api-starter/internal/config"
	"go-api-starter/internal/model"
	"go-api-starter/internal/platform/apperrors"
	"go-api-starter/internal/platform/auth"
	"go-api-starter/internal/platform/i18n"
	"go-api-starter/internal/platform/logger"
	"go-api-starter/internal/repository"
)

const (
	wxJsCode2SessionURL = "https://api.weixin.qq.com/sns/jscode2session"
	wxGetAccessTokenURL = "https://api.weixin.qq.com/cgi-bin/token"
	wxGetPhoneNumberURL = "https://api.weixin.qq.com/wxa/business/getuserphonenumber"
	maxWechatResponse   = 1 << 20
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
	logger.Log.Infof("[WxLogin] === 开始处理 === open_id_present=%t, phone_code_present=%t, mobile_present=%t",
		req.OpenID != nil && *req.OpenID != "", req.PhoneCode != "", req.Mobile != nil && *req.Mobile != "")

	if s.cfg.AppID == "" || s.cfg.Secret == "" {
		logger.Log.Info("[WxLogin] 错误: AppID 或 Secret 未配置")
		return nil, apperrors.InternalCode(nil, i18n.ErrWechatNotConfigured)
	}

	// 分步绑定阶段（兼容）: open_id + phone_code/mobile
	if req.OpenID != nil && *req.OpenID != "" {
		logger.Log.Info("[WxLogin] 进入绑定阶段")
		return s.bindAndLogin(ctx, req)
	}

	// 必须有 js_code
	if req.JsCode == "" {
		logger.Log.Info("[WxLogin] 错误: js_code 为空")
		return nil, apperrors.BadRequestCode(i18n.ErrParamInvalid)
	}

	// code2session
	logger.Log.Info("[WxLogin] 调用 code2Session")
	session, err := s.code2Session(ctx, req.JsCode)
	if err != nil {
		logger.Log.Warnf("[WxLogin] code2Session 失败: %v", err)
		return nil, apperrors.WrapCode(err, i18n.ErrWechatLoginFailed)
	}

	logger.Log.Infof("[WxLogin] code2Session 结果: errcode=%d, errmsg=%s", session.ErrCode, session.ErrMsg)

	if session.OpenID == "" {
		logger.Log.Warnf("[WxLogin] 错误: openid 为空, errcode=%d, errmsg=%s", session.ErrCode, session.ErrMsg)
		return nil, apperrors.BadRequestCode(i18n.ErrWechatLoginFailed)
	}

	// Check if user exists
	user, err := s.userRepo.FindByOpenID(ctx, session.OpenID)
	if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
		logger.Log.Warnf("[WxLogin] 查询用户失败: %v", err)
		return nil, apperrors.WrapCode(err, i18n.ErrQueryUserFailed)
	}

	if user != nil {
		logger.Log.Infof("[WxLogin] ✅ 用户已存在, uid=%s", user.UID)
		return s.buildLoginResponse(user)
	}

	// 新用户：如果同时带了 phone_code，直接一步完成注册
	if req.PhoneCode != "" {
		logger.Log.Info("[WxLogin] 新用户 + phone_code, 一步完成注册绑定")
		openID := session.OpenID
		req.OpenID = &openID
		return s.bindAndLogin(ctx, req)
	}

	// 新用户，没带 phone_code，需要绑定
	logger.Log.Info("[WxLogin] 新用户，需要绑定手机号")
	return &model.WxLoginResponse{
		NeedBind: true,
		OpenID:   &session.OpenID,
	}, nil
}

// bindAndLogin handles the second-step binding (open_id + phone_code or mobile).
func (s *WechatService) bindAndLogin(ctx context.Context, req *model.WxLoginRequest) (*model.WxLoginResponse, error) {
	openID := *req.OpenID
	logger.Log.Infof("[WxBind] === 绑定阶段 === phone_code_present=%t, mobile_present=%t", req.PhoneCode != "", req.Mobile != nil && *req.Mobile != "")

	// Resolve mobile: prefer phone_code (微信授权获取), fallback to manual mobile
	var mobile string
	if req.PhoneCode != "" {
		logger.Log.Info("[WxBind] 使用 phone_code 获取手机号...")
		phoneNumber, err := s.getPhoneNumber(ctx, req.PhoneCode)
		if err != nil {
			logger.Log.Warnf("[WxBind] ❌ getPhoneNumber 失败: %v", err)
			return nil, apperrors.WrapCode(err, i18n.ErrWechatLoginFailed)
		}
		mobile = phoneNumber
		logger.Log.Info("[WxBind] ✅ 解析到手机号")
	} else if req.Mobile != nil && *req.Mobile != "" {
		mobile = *req.Mobile
		logger.Log.Info("[WxBind] 使用手动输入的手机号")
	} else {
		logger.Log.Info("[WxBind] 错误: phone_code 和 mobile 都为空")
		return nil, apperrors.BadRequestCode(i18n.ErrProvideMobileOrEmail)
	}

	// Check if openid already bound
	existingByOpenID, err := s.userRepo.FindByOpenID(ctx, openID)
	if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
		logger.Log.Warnf("[WxBind] 查询 openid 用户失败: %v", err)
		return nil, apperrors.WrapCode(err, i18n.ErrQueryUserFailed)
	}
	if existingByOpenID != nil {
		logger.Log.Infof("[WxBind] openid 已绑定用户 uid=%s, 直接登录", existingByOpenID.UID)
		return s.buildLoginResponse(existingByOpenID)
	}

	// Check if mobile already has an account
	existing, err := s.userRepo.FindByMobile(ctx, mobile)
	if err == nil && existing != nil {
		logger.Log.Infof("[WxBind] 手机号已有账号 uid=%s", existing.UID)
		// Mobile exists — bind openid to it
		if existing.OpenID != nil && *existing.OpenID != "" {
			logger.Log.Warn("[WxBind] ❌ 手机号已绑定其他微信")
			return nil, apperrors.ConflictCode(i18n.ErrMobileBoundToWechat)
		}
		existing.OpenID = &openID
		if err := s.userRepo.Update(ctx, existing); err != nil {
			logger.Log.Warnf("[WxBind] ❌ 绑定 openid 到已有用户失败: %v", err)
			return nil, apperrors.WrapCode(err, i18n.ErrWechatBindFailed)
		}
		logger.Log.Infof("[WxBind] ✅ 已绑定 openid 到已有用户 uid=%s", existing.UID)
		return s.buildLoginResponse(existing)
	}

	// Create new user with openid + mobile + default password
	logger.Log.Info("[WxBind] 创建新用户")
	var hashedPwd *string
	if s.defaultPassword != "" {
		hasher := auth.NewPasswordHasher()
		h, err := hasher.HashPassword(s.defaultPassword)
		if err != nil {
			logger.Log.Warnf("[WxBind] ⚠️ 默认密码哈希失败: %v, 跳过设置密码", err)
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
		logger.Log.Warnf("[WxBind] ❌ 创建用户失败: %v", err)
		return nil, apperrors.WrapCode(err, i18n.ErrCreateUserFailed)
	}

	// Reload to get generated fields (UID, LPID, etc.)
	newUser, err = s.userRepo.FindByID(ctx, newUser.ID)
	if err != nil {
		logger.Log.Warnf("[WxBind] ❌ 重新查询用户失败: %v", err)
		return nil, apperrors.WrapCode(err, i18n.ErrQueryUserFailed)
	}

	logger.Log.Info("[WxBind] ✅ 新用户创建成功")
	return s.buildLoginResponse(newUser)
}

// getPhoneNumber calls WeChat API to get phone number from the phone_code.
func (s *WechatService) getPhoneNumber(ctx context.Context, phoneCode string) (string, error) {
	logger.Log.Info("[WxPhone] 开始获取手机号")

	token, err := s.getAccessToken(ctx)
	if err != nil {
		logger.Log.Warnf("[WxPhone] ❌ 获取 access_token 失败: %v", err)
		return "", fmt.Errorf("get access_token failed: %w", err)
	}
	logger.Log.Info("[WxPhone] access_token 获取成功")

	requestURL := wxGetPhoneNumberURL + "?" + url.Values{"access_token": []string{token}}.Encode()

	body, _ := json.Marshal(map[string]string{"code": phoneCode})
	logger.Log.Info("[WxPhone] 请求微信手机号接口")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.httpClient.Do(req)
	if err != nil {
		logger.Log.Warnf("[WxPhone] ❌ HTTP 请求失败: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxWechatResponse+1))
	if err != nil {
		logger.Log.Warnf("[WxPhone] ❌ 读取响应失败: %v", err)
		return "", err
	}
	if len(respBody) > maxWechatResponse {
		return "", fmt.Errorf("wechat response exceeds %d bytes", maxWechatResponse)
	}

	logger.Log.Infof("[WxPhone] 微信响应: status=%d", resp.StatusCode)

	var result wxPhoneNumberResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		logger.Log.Warnf("[WxPhone] ❌ 解析 JSON 失败: %v", err)
		return "", err
	}

	if result.ErrCode != 0 {
		logger.Log.Warnf("[WxPhone] ❌ 微信返回错误: errcode=%d, errmsg=%s", result.ErrCode, result.ErrMsg)
		return "", fmt.Errorf("wx getPhoneNumber errcode=%d, errmsg=%s", result.ErrCode, result.ErrMsg)
	}

	phoneNumber := result.PhoneInfo.PurePhoneNumber
	if phoneNumber == "" {
		phoneNumber = result.PhoneInfo.PhoneNumber
	}
	logger.Log.Info("[WxPhone] ✅ 手机号获取成功")

	return phoneNumber, nil
}

// getAccessToken gets a cached or fresh access_token for the mini-program.
func (s *WechatService) getAccessToken(ctx context.Context) (string, error) {
	s.tokenMu.Lock()
	defer s.tokenMu.Unlock()

	// Return cached token if still valid (with 5 min buffer)
	if s.accessToken != "" && time.Now().Before(s.tokenExpires.Add(-5*time.Minute)) {
		logger.Log.Infof("[WxToken] 使用缓存的 access_token, 过期时间: %s", s.tokenExpires.Format("15:04:05"))
		return s.accessToken, nil
	}

	logger.Log.Info("[WxToken] 缓存过期或为空, 重新获取 access_token...")
	requestURL := wxGetAccessTokenURL + "?" + url.Values{
		"grant_type": []string{"client_credential"},
		"appid":      []string{s.cfg.AppID},
		"secret":     []string{s.cfg.Secret},
	}.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return "", err
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		logger.Log.Warnf("[WxToken] ❌ HTTP 请求失败: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxWechatResponse+1))
	if err != nil {
		logger.Log.Warnf("[WxToken] ❌ 读取响应失败: %v", err)
		return "", err
	}
	if len(body) > maxWechatResponse {
		return "", fmt.Errorf("wechat response exceeds %d bytes", maxWechatResponse)
	}

	logger.Log.Info("[WxToken] 微信响应已收到")

	var result wxAccessTokenResponse
	if err := json.Unmarshal(body, &result); err != nil {
		logger.Log.Warnf("[WxToken] ❌ 解析 JSON 失败: %v", err)
		return "", err
	}

	if result.ErrCode != 0 {
		logger.Log.Warnf("[WxToken] ❌ 微信返回错误: errcode=%d, errmsg=%s", result.ErrCode, result.ErrMsg)
		return "", fmt.Errorf("wx getAccessToken errcode=%d, errmsg=%s", result.ErrCode, result.ErrMsg)
	}

	s.accessToken = result.AccessToken
	s.tokenExpires = time.Now().Add(time.Duration(result.ExpiresIn) * time.Second)

	logger.Log.Infof("[WxToken] ✅ access_token 刷新成功, expires_in=%d秒, 过期时间: %s",
		result.ExpiresIn, s.tokenExpires.Format("15:04:05"))
	return s.accessToken, nil
}

// code2Session calls the WeChat jscode2session API.
func (s *WechatService) code2Session(ctx context.Context, jsCode string) (*wxSessionResponse, error) {
	requestURL := wxJsCode2SessionURL + "?" + url.Values{
		"appid":      []string{s.cfg.AppID},
		"secret":     []string{s.cfg.Secret},
		"js_code":    []string{jsCode},
		"grant_type": []string{"authorization_code"},
	}.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxWechatResponse+1))
	if err != nil {
		return nil, err
	}
	if len(body) > maxWechatResponse {
		return nil, fmt.Errorf("wechat response exceeds %d bytes", maxWechatResponse)
	}

	logger.Log.Info("[WxLogin] wx API 响应已收到")

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
