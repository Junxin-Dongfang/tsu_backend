// File: internal/pkg/xerrors/kratos_errors.go
package xerrors

import "strings"

// Kratos 错误 ID 类型
type KratosID int

// Kratos 错误 ID 常量（与你提供的文档保持一致）
const (
	// Info - Login (1010000+)
	InfoSelfServiceLoginRoot                 KratosID = 1010000
	InfoSelfServiceLogin                     KratosID = 1010001
	InfoSelfServiceLoginWith                 KratosID = 1010002
	InfoSelfServiceLoginReAuth               KratosID = 1010003
	InfoSelfServiceLoginMFA                  KratosID = 1010004
	InfoSelfServiceLoginVerify               KratosID = 1010005
	InfoSelfServiceLoginTOTPLabel            KratosID = 1010006
	InfoLoginLookupLabel                     KratosID = 1010007
	InfoSelfServiceLoginWebAuthn             KratosID = 1010008
	InfoLoginTOTP                            KratosID = 1010009
	InfoLoginLookup                          KratosID = 1010010
	InfoSelfServiceLoginContinueWebAuthn     KratosID = 1010011
	InfoSelfServiceLoginWebAuthnPasswordless KratosID = 1010012
	InfoSelfServiceLoginContinue             KratosID = 1010013
	InfoSelfServiceLoginCodeSent             KratosID = 1010014
	InfoSelfServiceLoginCode                 KratosID = 1010015
	InfoSelfServiceLoginLink                 KratosID = 1010016
	InfoSelfServiceLoginAndLink              KratosID = 1010017
	InfoSelfServiceLoginWithAndLink          KratosID = 1010018
	InfoSelfServiceLoginCodeMFA              KratosID = 1010019
	InfoSelfServiceLoginCodeMFAHint          KratosID = 1010020
	InfoSelfServiceLoginPasskey              KratosID = 1010021
	InfoSelfServiceLoginPassword             KratosID = 1010022
	InfoSelfServiceLoginAAL2CodeAddress      KratosID = 1010023

	// Info - Logout (1020000+)
	InfoSelfServiceLogout KratosID = 1020000

	// Info - MFA (1030000+)
	InfoSelfServiceMFA KratosID = 1030000

	// Info - Registration (1040000+)
	InfoSelfServiceRegistrationRoot              KratosID = 1040000
	InfoSelfServiceRegistration                  KratosID = 1040001
	InfoSelfServiceRegistrationWith              KratosID = 1040002
	InfoSelfServiceRegistrationContinue          KratosID = 1040003
	InfoSelfServiceRegistrationRegisterWebAuthn  KratosID = 1040004
	InfoSelfServiceRegistrationEmailWithCodeSent KratosID = 1040005
	InfoSelfServiceRegistrationRegisterCode      KratosID = 1040006
	InfoSelfServiceRegistrationRegisterPasskey   KratosID = 1040007
	InfoSelfServiceRegistrationBack              KratosID = 1040008
	InfoSelfServiceRegistrationChooseCredentials KratosID = 1040009

	// Error - Validation General (4000000+)
	ErrorValidation                               KratosID = 4000000
	ErrorValidationGeneric                        KratosID = 4000001
	ErrorValidationRequired                       KratosID = 4000002
	ErrorValidationMinLength                      KratosID = 4000003
	ErrorValidationInvalidFormat                  KratosID = 4000004
	ErrorValidationPasswordPolicyViolationGeneric KratosID = 4000005
	ErrorValidationInvalidCredentials             KratosID = 4000006
	ErrorValidationDuplicateCredentials           KratosID = 4000007
	ErrorValidationTOTPVerifierWrong              KratosID = 4000008
	ErrorValidationIdentifierMissing              KratosID = 4000009
	ErrorValidationAddressNotVerified             KratosID = 4000010
	ErrorValidationNoTOTPDevice                   KratosID = 4000011
	ErrorValidationLookupAlreadyUsed              KratosID = 4000012
	ErrorValidationNoWebAuthnDevice               KratosID = 4000013
	ErrorValidationNoLookup                       KratosID = 4000014
	ErrorValidationSuchNoWebAuthnUser             KratosID = 4000015
	ErrorValidationLookupInvalid                  KratosID = 4000016
	ErrorValidationMaxLength                      KratosID = 4000017
	ErrorValidationMinimum                        KratosID = 4000018
	ErrorValidationExclusiveMinimum               KratosID = 4000019
	ErrorValidationMaximum                        KratosID = 4000020
	ErrorValidationExclusiveMaximum               KratosID = 4000021
	ErrorValidationMultipleOf                     KratosID = 4000022
	ErrorValidationMaxItems                       KratosID = 4000023
	ErrorValidationMinItems                       KratosID = 4000024
	ErrorValidationUniqueItems                    KratosID = 4000025
	ErrorValidationWrongType                      KratosID = 4000026
	ErrorValidationDuplicateCredentialsOnOIDCLink KratosID = 4000027
	ErrorValidationDuplicateCredentialsWithHints  KratosID = 4000028
	ErrorValidationConst                          KratosID = 4000029
	ErrorValidationConstGeneric                   KratosID = 4000030
	ErrorValidationPasswordIdentifierTooSimilar   KratosID = 4000031
	ErrorValidationPasswordMinLength              KratosID = 4000032
	ErrorValidationPasswordMaxLength              KratosID = 4000033
	ErrorValidationPasswordTooManyBreaches        KratosID = 4000034
	ErrorValidationNoCodeUser                     KratosID = 4000035
	ErrorValidationTraitsMismatch                 KratosID = 4000036
	ErrorValidationAccountNotFound                KratosID = 4000037
	ErrorValidationCaptchaError                   KratosID = 4000038
	ErrorValidationPasswordNewSameAsOld           KratosID = 4000039

	// Error - Login Specific (4010000+)
	ErrorValidationLogin                            KratosID = 4010000
	ErrorValidationLoginFlowExpired                 KratosID = 4010001
	ErrorValidationLoginNoStrategyFound             KratosID = 4010002
	ErrorValidationRegistrationNoStrategyFound      KratosID = 4010003
	ErrorValidationSettingsNoStrategyFound          KratosID = 4010004
	ErrorValidationRecoveryNoStrategyFound          KratosID = 4010005
	ErrorValidationVerificationNoStrategyFound      KratosID = 4010006
	ErrorValidationLoginRetrySuccess                KratosID = 4010007
	ErrorValidationLoginCodeInvalidOrAlreadyUsed    KratosID = 4010008
	ErrorValidationLoginLinkedCredentialsDoNotMatch KratosID = 4010009
	ErrorValidationLoginAddressUnknown              KratosID = 4010010

	// Error - Registration Specific (4040000+)
	ErrorValidationRegistration                     KratosID = 4040000
	ErrorValidationRegistrationFlowExpired          KratosID = 4040001
	ErrorValidationRegistrationRetrySuccess         KratosID = 4040002
	ErrorValidationRegistrationCodeInvalidOrAlready KratosID = 4040003

	// Error - Settings (4050000+)
	ErrorValidationSettings            KratosID = 4050000
	ErrorValidationSettingsFlowExpired KratosID = 4050001

	// Error - Recovery (4060000+)
	ErrorValidationRecovery                          KratosID = 4060000
	ErrorValidationRecoveryRetrySuccess              KratosID = 4060001
	ErrorValidationRecoveryStateFailure              KratosID = 4060002
	ErrorValidationRecoveryMissingRecoveryToken      KratosID = 4060003
	ErrorValidationRecoveryTokenInvalidOrAlreadyUsed KratosID = 4060004
	ErrorValidationRecoveryFlowExpired               KratosID = 4060005
	ErrorValidationRecoveryCodeInvalidOrAlreadyUsed  KratosID = 4060006

	// Error - Verification (4070000+)
	ErrorValidationVerification                          KratosID = 4070000
	ErrorValidationVerificationTokenInvalidOrAlreadyUsed KratosID = 4070001
	ErrorValidationVerificationRetrySuccess              KratosID = 4070002
	ErrorValidationVerificationStateFailure              KratosID = 4070003
	ErrorValidationVerificationMissingVerificationToken  KratosID = 4070004
	ErrorValidationVerificationFlowExpired               KratosID = 4070005
	ErrorValidationVerificationCodeInvalidOrAlreadyUsed  KratosID = 4070006

	// Error - System (5000000+)
	ErrorSystem        KratosID = 5000000
	ErrorSystemGeneric KratosID = 5000001
)

// 扩展业务错误码定义
const (
	// 扩展认证错误码 (2xxxxx)
	CodePasswordPolicyError = 200008 // 密码策略不符合
	CodePasswordTooShort    = 200009 // 密码太短
	CodePasswordTooLong     = 200010 // 密码太长
	CodePasswordTooWeak     = 200011 // 密码太弱
	CodePasswordSameAsOld   = 200012 // 新密码与旧密码相同
	CodePasswordTooSimilar  = 200013 // 密码与用户信息太相似

	// 扩展用户错误码 (4xxxxx)
	CodeAddressNotVerified = 400008 // 邮箱/地址未验证
	CodeTraitsMismatch     = 400009 // 用户特征不匹配
	CodeAccountNotFound    = 400010 // 账户不存在

	// 扩展通用错误码 (1xxxxx)
	CodeFlowExpired       = 100005 // 流程已过期
	CodeCodeInvalidOrUsed = 100006 // 验证码无效或已使用
	CodeStrategyNotFound  = 100007 // 策略未找到
	CodeCaptchaError      = 100008 // 验证码错误
	CodeTOTPError         = 100009 // TOTP 错误
	CodeWebAuthnError     = 100010 // WebAuthn 错误
)

// Kratos 错误 ID 到业务错误码的映射
var kratosErrorMap = map[KratosID]int{
	// 认证相关错误
	ErrorValidationInvalidCredentials:   CodeInvalidCredentials,
	ErrorValidationDuplicateCredentials: CodeDuplicateResource,
	ErrorValidationAccountNotFound:      CodeAccountNotFound,
	ErrorValidationIdentifierMissing:    CodeInvalidParams,

	// 密码策略错误
	ErrorValidationPasswordPolicyViolationGeneric: CodePasswordPolicyError,
	ErrorValidationPasswordMinLength:              CodePasswordTooShort,
	ErrorValidationPasswordMaxLength:              CodePasswordTooLong,
	ErrorValidationPasswordTooManyBreaches:        CodePasswordTooWeak,
	ErrorValidationPasswordNewSameAsOld:           CodePasswordSameAsOld,
	ErrorValidationPasswordIdentifierTooSimilar:   CodePasswordTooSimilar,

	// 用户相关错误
	ErrorValidationAddressNotVerified: CodeAddressNotVerified,
	ErrorValidationTraitsMismatch:     CodeTraitsMismatch,
	ErrorValidationNoCodeUser:         CodeUserNotFound,

	// 流程相关错误
	ErrorValidationLoginFlowExpired:        CodeFlowExpired,
	ErrorValidationRegistrationFlowExpired: CodeFlowExpired,
	ErrorValidationSettingsFlowExpired:     CodeFlowExpired,
	ErrorValidationRecoveryFlowExpired:     CodeFlowExpired,
	ErrorValidationVerificationFlowExpired: CodeFlowExpired,

	// 验证码相关错误
	ErrorValidationLoginCodeInvalidOrAlreadyUsed:        CodeCodeInvalidOrUsed,
	ErrorValidationRegistrationCodeInvalidOrAlready:     CodeCodeInvalidOrUsed,
	ErrorValidationRecoveryCodeInvalidOrAlreadyUsed:     CodeCodeInvalidOrUsed,
	ErrorValidationVerificationCodeInvalidOrAlreadyUsed: CodeCodeInvalidOrUsed,

	// 策略相关错误
	ErrorValidationLoginNoStrategyFound:        CodeStrategyNotFound,
	ErrorValidationRegistrationNoStrategyFound: CodeStrategyNotFound,
	ErrorValidationSettingsNoStrategyFound:     CodeStrategyNotFound,
	ErrorValidationRecoveryNoStrategyFound:     CodeStrategyNotFound,
	ErrorValidationVerificationNoStrategyFound: CodeStrategyNotFound,

	// 通用验证错误
	ErrorValidation:              CodeInvalidParams,
	ErrorValidationGeneric:       CodeInvalidParams,
	ErrorValidationRequired:      CodeInvalidParams,
	ErrorValidationMinLength:     CodeInvalidParams,
	ErrorValidationMaxLength:     CodeInvalidParams,
	ErrorValidationInvalidFormat: CodeInvalidParams,
	ErrorValidationWrongType:     CodeInvalidParams,
	ErrorValidationCaptchaError:  CodeCaptchaError,

	// TOTP/MFA 相关错误
	ErrorValidationTOTPVerifierWrong: CodeTOTPError,
	ErrorValidationNoTOTPDevice:      CodeTOTPError,
	ErrorValidationLookupAlreadyUsed: CodeTOTPError,
	ErrorValidationLookupInvalid:     CodeTOTPError,

	// WebAuthn 相关错误
	ErrorValidationNoWebAuthnDevice:   CodeWebAuthnError,
	ErrorValidationSuchNoWebAuthnUser: CodeWebAuthnError,

	// 系统错误
	ErrorSystem:        CodeInternalError,
	ErrorSystemGeneric: CodeInternalError,
}

// 扩展错误消息映射
func init() {
	// 新增错误消息
	additionalMessages := map[int]string{
		CodePasswordPolicyError: "密码不符合安全策略",
		CodePasswordTooShort:    "密码长度不够",
		CodePasswordTooLong:     "密码长度过长",
		CodePasswordTooWeak:     "密码强度不够",
		CodePasswordSameAsOld:   "新密码不能与旧密码相同",
		CodePasswordTooSimilar:  "密码不能与用户信息太相似",
		CodeAddressNotVerified:  "邮箱地址未验证",
		CodeTraitsMismatch:      "用户信息不匹配",
		CodeAccountNotFound:     "账户不存在",
		CodeFlowExpired:         "操作流程已过期，请重新开始",
		CodeCodeInvalidOrUsed:   "验证码无效或已使用",
		CodeStrategyNotFound:    "认证策略不可用",
		CodeCaptchaError:        "验证码错误",
		CodeTOTPError:           "动态验证码错误",
		CodeWebAuthnError:       "生物识别验证失败",
	}

	// 合并到全局消息映射
	for code, msg := range additionalMessages {
		codeMessages[code] = msg
	}
}

// TranslateKratosError 将 Kratos 错误 ID 转换为业务错误码和消息
func TranslateKratosError(kratosID int) (int, string) {
	kratosIDType := KratosID(kratosID)

	if appCode, exists := kratosErrorMap[kratosIDType]; exists {
		if msg, exists := codeMessages[appCode]; exists {
			return appCode, msg
		}
	}

	// 兜底返回通用验证错误
	return CodeInvalidParams, "输入信息有误，请检查后重试"
}

// TranslateKratosErrorText 根据 Kratos 错误文本进行模糊匹配翻译
func TranslateKratosErrorText(errorText string) (int, string) {
	if errorText == "" {
		return CodeInvalidParams, "输入信息有误，请检查后重试"
	}

	// 转为小写以便匹配
	errorTextLower := strings.ToLower(errorText)

	// 按优先级顺序进行模式匹配
	patterns := []struct {
		keywords []string
		code     int
	}{
		// 凭据相关（最高优先级）
		{[]string{"credentials are invalid", "invalid credentials", "wrong credentials"}, CodeInvalidCredentials},
		{[]string{"account not found", "user not found", "no such user"}, CodeAccountNotFound},

		// 重复资源（高优先级）
		{[]string{"already taken", "already exists", "duplicate", "identifier is already taken"}, CodeDuplicateResource},
		{[]string{"email is already taken", "email already exists"}, CodeEmailExists},
		{[]string{"username is already taken", "username already exists"}, CodeUsernameExists},

		// 密码策略（中等优先级）
		{[]string{"password is too short", "password too short", "minimum length"}, CodePasswordTooShort},
		{[]string{"password is too long", "password too long", "maximum length"}, CodePasswordTooLong},
		{[]string{"password policy", "password security", "password strength", "too many breaches"}, CodePasswordPolicyError},
		{[]string{"password is too similar", "too similar to identifier"}, CodePasswordTooSimilar},
		{[]string{"same as old password", "new password same as old"}, CodePasswordSameAsOld},

		// 验证和流程相关
		{[]string{"address is not verified", "email not verified", "not verified"}, CodeAddressNotVerified},
		{[]string{"flow expired", "session expired", "expired"}, CodeFlowExpired},
		{[]string{"code is invalid", "code already used", "invalid code", "code expired"}, CodeCodeInvalidOrUsed},
		{[]string{"traits do not match", "traits mismatch"}, CodeTraitsMismatch},
		{[]string{"captcha", "verification failed"}, CodeCaptchaError},

		// TOTP/MFA 相关
		{[]string{"totp", "authenticator", "recovery code", "lookup"}, CodeTOTPError},
		{[]string{"webauthn", "security key", "biometric"}, CodeWebAuthnError},

		// 通用验证错误（最低优先级）
		{[]string{"required", "missing", "cannot be empty"}, CodeInvalidParams},
		{[]string{"invalid format", "format is invalid", "malformed"}, CodeInvalidParams},
		{[]string{"validation failed", "invalid input", "bad request"}, CodeInvalidParams},
	}

	// 按顺序匹配模式
	for _, pattern := range patterns {
		for _, keyword := range pattern.keywords {
			if strings.Contains(errorTextLower, keyword) {
				if msg, exists := codeMessages[pattern.code]; exists {
					return pattern.code, msg
				}
			}
		}
	}

	// 兜底返回通用验证错误
	return CodeInvalidParams, "输入信息有误，请检查后重试"
}

// GetKratosErrorPriority 获取错误的优先级（数字越小优先级越高）
func GetKratosErrorPriority(code int) int {
	priorityMap := map[int]int{
		CodeInvalidCredentials:  1,  // 最高优先级：认证失败
		CodeAccountNotFound:     2,  // 账户不存在
		CodeDuplicateResource:   3,  // 资源重复
		CodeEmailExists:         4,  // 邮箱重复
		CodeUsernameExists:      5,  // 用户名重复
		CodePasswordTooShort:    6,  // 密码太短
		CodePasswordTooLong:     7,  // 密码太长
		CodePasswordPolicyError: 8,  // 密码策略
		CodePasswordTooSimilar:  9,  // 密码太相似
		CodeAddressNotVerified:  10, // 地址未验证
		CodeFlowExpired:         11, // 流程过期
		CodeCodeInvalidOrUsed:   12, // 验证码错误
		CodeInvalidParams:       99, // 最低优先级：通用参数错误
	}

	if priority, exists := priorityMap[code]; exists {
		return priority
	}
	return 50 // 默认中等优先级
}

// IsRetryableKratosError 判断 Kratos 错误是否可重试
func IsRetryableKratosError(kratosID int) bool {
	retryableErrors := map[KratosID]bool{
		ErrorValidationLoginFlowExpired:        true,
		ErrorValidationRegistrationFlowExpired: true,
		ErrorValidationSettingsFlowExpired:     true,
		ErrorValidationRecoveryFlowExpired:     true,
		ErrorValidationVerificationFlowExpired: true,
		ErrorSystem:                            true,
		ErrorSystemGeneric:                     true,
	}

	return retryableErrors[KratosID(kratosID)]
}
