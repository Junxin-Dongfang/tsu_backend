package xerrors

import "log/slog"

// kratosErrMapping 是 Kratos 错误 ID 到我们自定义错误码和消息的映射。
// 该映射用于将 Kratos 的 UI 消息 ID 翻译为统一的业务错误码与文案。
var kratosErrMapping = map[int]struct {
	Code    int
	Message string
}{
	// --- Kratos 注册流程中的常见错误 ID ---
	// 可根据需要持续补充
	4000001: {Code: Validation, Message: "输入信息格式不正确"},
	4000002: {Code: PasswordPolicyError, Message: "密码不符合安全策略，请设置更复杂的密码"},
	4000003: {Code: IdentityAlreadyExists, Message: "该电子邮箱地址已被注册"},
	4000004: {Code: InvalidCredentials, Message: "提供的凭证格式无效"},
	4000005: {Code: PasswordTooShort, Message: "密码长度不足，请设置更长的密码"},
	4000007: {Code: PasswordTooShort, Message: "密码太短，至少需要8个字符"},
	4000008: {Code: EmailFormatError, Message: "电子邮箱地址格式不正确"},
	4000009: {Code: IdentityAlreadyExists, Message: "该用户名已被使用"},
	4000010: {Code: EmailFormatError, Message: "请输入有效的电子邮箱地址"},

	// --- Kratos 登录流程中的常见错误 ID ---
	4000006: {Code: SessionAlreadyAvailable, Message: "您已经登录，无需重复操作"},
	4010001: {Code: InvalidCredentials, Message: "用户名或密码不正确"},
	4010002: {Code: InvalidCredentials, Message: "账户不存在或密码错误"},
}

// TranslateKratosError 根据 Kratos 的 UI message ID 返回我们自定义的错误码和消息。
func TranslateKratosError(kratosID int) (int, string) {
	if errInfo, ok := kratosErrMapping[kratosID]; ok {
		return errInfo.Code, errInfo.Message
	}

	// 找不到映射时返回通用的校验错误，并记录日志便于后续补充映射表
	slog.Warn("未知的 Kratos 错误 ID", "kratos_id", kratosID)
	return Validation, "提交的信息有误，请检查后重试"
}

// TranslateKratosErrorText 根据 Kratos 返回的消息文本进行兜底翻译（当 ID 不可用或未映射时）
func TranslateKratosErrorText(text string) (int, string) {
	// 身份已存在（邮箱/用户名）
	if containsAnyFold(text, []string{
		"exists already", "already exists", "already taken",
		"identifier exists", "account exists",
		"已被注册", "已存在", "已被使用",
	}) {
		return IdentityAlreadyExists, "该电子邮箱地址已被注册"
	}
	// 密码过短/策略问题
	if containsAnyFold(text, []string{
		"too short", "at least 8", "minimum length", "not long enough",
		"密码太短", "至少", "长度不足", "不满足密码策略",
	}) {
		return PasswordTooShort, "密码太短，至少需要8个字符"
	}
	// 邮箱格式
	if containsAnyFold(text, []string{"email", "邮箱", "电子邮箱"}) && containsAnyFold(text, []string{"invalid", "not valid", "格式", "不正确"}) {
		return EmailFormatError, "电子邮箱地址格式不正确"
	}
	// 凭证无效（登录）
	if containsAnyFold(text, []string{"invalid credentials", "wrong password", "unknown email", "user not found", "no such user"}) {
		return InvalidCredentials, "用户名或密码不正确"
	}
	return Validation, "提交的信息有误，请检查后重试"
}

// containsAnyFold 做大小写不敏感包含匹配
func containsAnyFold(haystack string, needles []string) bool {
	lower := toLower(haystack)
	for _, n := range needles {
		if n == "" {
			continue
		}
		if indexOf(lower, toLower(n)) >= 0 {
			return true
		}
	}
	return false
}

// 以下三个小工具避免引入strings，保持与现有导入一致
func toLower(s string) string {
	b := []byte(s)
	for i := range b {
		if b[i] >= 'A' && b[i] <= 'Z' {
			b[i] = b[i] + 32
		}
	}
	return string(b)
}

func indexOf(s, sub string) int {
	n, m := len(s), len(sub)
	if m == 0 {
		return 0
	}
	if m > n {
		return -1
	}
	for i := 0; i <= n-m; i++ {
		if s[i:i+m] == sub {
			return i
		}
	}
	return -1
}
