package i18n

// ─── Auth / Token ───
const (
	ErrInvalidToken       = "AUTH_INVALID_TOKEN"
	ErrTokenExpired       = "AUTH_TOKEN_EXPIRED"
	ErrRefreshTokenExpired = "AUTH_REFRESH_TOKEN_EXPIRED"
	ErrInvalidRefreshToken = "AUTH_INVALID_REFRESH_TOKEN"
	ErrUnauthenticated    = "AUTH_UNAUTHENTICATED"
	ErrWrongCredentials   = "AUTH_WRONG_CREDENTIALS"
	ErrAccountNotFound    = "AUTH_ACCOUNT_NOT_FOUND"
	ErrAccountFrozen      = "AUTH_ACCOUNT_FROZEN"
	ErrPasswordRequired   = "AUTH_PASSWORD_REQUIRED"
)

// ─── Registration / Account ───
const (
	ErrEmailTaken         = "REG_EMAIL_TAKEN"
	ErrMobileTaken        = "REG_MOBILE_TAKEN"
	ErrLPIDTaken          = "REG_LPID_TAKEN"
	ErrMobileOrEmailRequired = "REG_MOBILE_OR_EMAIL_REQUIRED"
	ErrUserNotFound       = "USER_NOT_FOUND"
)

// ─── Verification Code ───
const (
	ErrCodeRequired       = "VERIFY_CODE_REQUIRED"
	ErrCodeInvalid        = "VERIFY_CODE_INVALID"
	ErrCodeExpired        = "VERIFY_CODE_EXPIRED"
	ErrCodeRateLimit      = "VERIFY_RATE_LIMIT"
	ErrCodeStoreFailed    = "VERIFY_STORE_FAILED"
	ErrMailSendFailed     = "VERIFY_MAIL_SEND_FAILED"
	ErrBindEmailCodeRequired = "VERIFY_BIND_EMAIL_CODE_REQUIRED"
	ErrProvideCode        = "VERIFY_PROVIDE_CODE"
	ErrProvideMobileOrEmail = "VERIFY_PROVIDE_MOBILE_OR_EMAIL"
)

// ─── QR Code ───
const (
	ErrQRNotFound         = "QR_NOT_FOUND"
	ErrQRExpired          = "QR_EXPIRED"
	ErrQRInvalidState     = "QR_INVALID_STATE"
	ErrQRScanFirst        = "QR_SCAN_FIRST"
	ErrQRNoPermission     = "QR_NO_PERMISSION"
	ErrQRKeyMissing       = "QR_KEY_MISSING"
)

// ─── WeChat ───
const (
	ErrWechatAlreadyBound   = "WECHAT_ALREADY_BOUND"
	ErrMobileBoundToWechat  = "WECHAT_MOBILE_BOUND"
	ErrEmailBoundToWechat   = "WECHAT_EMAIL_BOUND"
)

// ─── Validation / Common ───
const (
	ErrValidationFailed   = "VALIDATION_FAILED"
	ErrParamInvalid       = "PARAM_INVALID"
	ErrInvalidUserID      = "INVALID_USER_ID"
	ErrInvalidRoleID      = "INVALID_ROLE_ID"
	ErrInvalidPermissionID = "INVALID_PERMISSION_ID"
	ErrInvalidLogID       = "INVALID_LOG_ID"
)

// ─── System Config ───
const (
	ErrConfigNotFound     = "CONFIG_NOT_FOUND"
	ErrConfigKeyExists    = "CONFIG_KEY_EXISTS"
	ErrConfigKeyEmpty     = "CONFIG_KEY_EMPTY"
)

// ─── Resource ───
const (
	ErrLogNotFound        = "LOG_NOT_FOUND"
	ErrAccountNotRegistered = "ACCOUNT_NOT_REGISTERED"
)

// ─── Generic ───
const (
	ErrInternalError      = "INTERNAL_ERROR"
)

// ─── Internal / Infrastructure ───
const (
	ErrQueryUserFailed    = "INTERNAL_QUERY_USER_FAILED"
	ErrCreateUserFailed   = "INTERNAL_CREATE_USER_FAILED"
	ErrHashPasswordFailed = "INTERNAL_HASH_PASSWORD_FAILED"
	ErrVerifyPasswordFailed = "INTERNAL_VERIFY_PASSWORD_FAILED"
	ErrGenerateTokenFailed = "INTERNAL_GENERATE_TOKEN_FAILED"
	ErrResetPasswordFailed = "INTERNAL_RESET_PASSWORD_FAILED"
	ErrCreateQRFailed     = "INTERNAL_CREATE_QR_FAILED"
)

// ─── WeChat ───
const (
	ErrWechatNotConfigured = "WECHAT_NOT_CONFIGURED"
	ErrWechatLoginFailed   = "WECHAT_LOGIN_FAILED"
	ErrWechatBindFailed    = "WECHAT_BIND_FAILED"
)
