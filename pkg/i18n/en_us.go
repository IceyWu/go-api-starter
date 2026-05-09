package i18n

var enUS = map[string]string{
	// Auth / Token
	ErrInvalidToken:        "Invalid token",
	ErrTokenExpired:        "Token expired",
	ErrRefreshTokenExpired: "Refresh token expired",
	ErrInvalidRefreshToken: "Invalid refresh token",
	ErrUnauthenticated:     "User not authenticated",
	ErrWrongCredentials:    "Wrong phone/email or password",
	ErrAccountNotFound:     "Account not found",
	ErrAccountFrozen:       "Account has been frozen",
	ErrPasswordRequired:    "Password is required",

	// Registration / Account
	ErrEmailTaken:            "Email already registered",
	ErrMobileTaken:           "Phone number already registered",
	ErrLPIDTaken:             "LP ID already taken",
	ErrMobileOrEmailRequired: "Phone or email is required",
	ErrUserNotFound:          "User not found",

	// Verification Code
	ErrCodeRequired:          "Verification code is required",
	ErrCodeInvalid:           "Invalid verification code",
	ErrCodeExpired:           "Verification code expired",
	ErrCodeRateLimit:         "Please wait 60 seconds before resending",
	ErrCodeStoreFailed:       "Failed to store verification code",
	ErrMailSendFailed:        "Failed to send email",
	ErrBindEmailCodeRequired: "Verification code required for email binding",
	ErrProvideCode:           "Please provide verification code",
	ErrProvideMobileOrEmail:  "Please provide phone or email",

	// QR Code
	ErrQRNotFound:      "QR code not found",
	ErrQRExpired:       "QR code expired",
	ErrQRInvalidState:  "Invalid QR code state",
	ErrQRScanFirst:     "Please scan the QR code first",
	ErrQRNoPermission:  "No permission to confirm this QR code",
	ErrQRKeyMissing:    "QR code key is missing",

	// WeChat
	ErrWechatAlreadyBound:  "WeChat account already bound to another user",
	ErrMobileBoundToWechat: "Phone number already bound to another WeChat",
	ErrEmailBoundToWechat:  "Email already bound to another WeChat",

	// Validation / Common
	ErrValidationFailed:    "Validation failed",
	ErrParamInvalid:        "Invalid parameter",
	ErrInvalidUserID:       "Invalid user ID",
	ErrInvalidRoleID:       "Invalid role ID",
	ErrInvalidPermissionID: "Invalid permission ID",
	ErrInvalidLogID:        "Invalid log ID",

	// System Config
	ErrConfigNotFound:  "Configuration not found",
	ErrConfigKeyExists: "Configuration key already exists",
	ErrConfigKeyEmpty:  "Configuration key cannot be empty",

	// Resource
	ErrLogNotFound:          "Log not found",
	ErrAccountNotRegistered: "Account not registered",

	// Generic
	ErrInternalError: "Internal server error",
}

func init() {
	extra := map[string]string{
		ErrQueryUserFailed:     "Failed to query user",
		ErrCreateUserFailed:    "Failed to create user",
		ErrHashPasswordFailed:  "Failed to hash password",
		ErrVerifyPasswordFailed: "Failed to verify password",
		ErrGenerateTokenFailed: "Failed to generate token",
		ErrResetPasswordFailed: "Failed to reset password",
		ErrCreateQRFailed:      "Failed to create QR code",
		ErrWechatNotConfigured: "WeChat mini-program not configured",
		ErrWechatLoginFailed:   "WeChat login failed",
		ErrWechatBindFailed:    "Failed to bind WeChat",
	}
	for k, v := range extra {
		enUS[k] = v
	}
}
