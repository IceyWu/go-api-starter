package i18n

var zhCN = map[string]string{
	// Auth / Token
	ErrInvalidToken:        "无效的令牌",
	ErrTokenExpired:        "令牌已过期",
	ErrRefreshTokenExpired: "刷新令牌已过期",
	ErrInvalidRefreshToken: "无效的刷新令牌",
	ErrUnauthenticated:     "用户未认证",
	ErrWrongCredentials:    "手机号/邮箱或密码错误",
	ErrAccountNotFound:     "账号不存在",
	ErrAccountFrozen:       "账号已被冻结",
	ErrPasswordRequired:    "密码不能为空",

	// Registration / Account
	ErrEmailTaken:            "邮箱已被注册",
	ErrMobileTaken:           "手机号已被注册",
	ErrLPIDTaken:             "该LP号已被占用",
	ErrMobileOrEmailRequired: "手机号或邮箱至少提供一个",
	ErrUserNotFound:          "用户不存在",

	// Verification Code
	ErrCodeRequired:          "验证码不能为空",
	ErrCodeInvalid:           "验证码错误",
	ErrCodeExpired:           "验证码已过期或不存在",
	ErrCodeRateLimit:         "请等待60秒后再发送验证码",
	ErrCodeStoreFailed:       "存储验证码失败",
	ErrMailSendFailed:        "发送邮件失败",
	ErrBindEmailCodeRequired: "绑定邮箱需要提供验证码",
	ErrProvideCode:           "请提供验证码",
	ErrProvideMobileOrEmail:  "请提供手机号或邮箱",

	// QR Code
	ErrQRNotFound:      "二维码不存在",
	ErrQRExpired:       "二维码已过期",
	ErrQRInvalidState:  "二维码状态无效",
	ErrQRScanFirst:     "请先扫描二维码",
	ErrQRNoPermission:  "无权确认此二维码",
	ErrQRKeyMissing:    "缺少二维码 key",

	// WeChat
	ErrWechatAlreadyBound:  "该微信号已绑定其他账号",
	ErrMobileBoundToWechat: "该手机号已绑定其他微信号",
	ErrEmailBoundToWechat:  "该邮箱已绑定其他微信号",

	// Validation / Common
	ErrValidationFailed:   "参数验证失败",
	ErrParamInvalid:       "参数错误",
	ErrInvalidUserID:      "无效的用户ID",
	ErrInvalidRoleID:      "无效的角色ID",
	ErrInvalidPermissionID: "无效的权限ID",
	ErrInvalidLogID:       "无效的日志ID",

	// System Config
	ErrConfigNotFound:  "配置不存在",
	ErrConfigKeyExists: "配置键已存在",
	ErrConfigKeyEmpty:  "配置键不能为空",

	// Resource
	ErrLogNotFound:          "日志不存在",
	ErrAccountNotRegistered: "该账号未注册",

	// Generic
	ErrInternalError: "服务器内部错误",
}

func init() {
	extra := map[string]string{
		ErrQueryUserFailed:     "查询用户失败",
		ErrCreateUserFailed:    "创建用户失败",
		ErrHashPasswordFailed:  "密码加密失败",
		ErrVerifyPasswordFailed: "密码验证失败",
		ErrGenerateTokenFailed: "生成令牌失败",
		ErrResetPasswordFailed: "密码重置失败",
		ErrCreateQRFailed:      "创建二维码失败",
		ErrWechatNotConfigured: "微信小程序未配置",
		ErrWechatLoginFailed:   "微信登录失败",
		ErrWechatBindFailed:    "绑定微信失败",
	}
	for k, v := range extra {
		zhCN[k] = v
	}
}
