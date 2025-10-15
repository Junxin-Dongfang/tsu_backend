// File: internal/pkg/i18n/error_messages.go
package i18n

import (
	"tsu-self/internal/pkg/xerrors"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// ErrorMessages 错误消息的多语言映射
var ErrorMessages = map[xerrors.ErrorCode]map[language.Tag]string{
	// 1xxxxx: 通用错误码
	xerrors.CodeSuccess:           {language.Chinese: "操作成功", language.English: "Operation successful"},
	xerrors.CodeInternalError:     {language.Chinese: "内部服务错误", language.English: "Internal server error"},
	xerrors.CodeInvalidParams:     {language.Chinese: "参数错误", language.English: "Invalid parameters"},
	xerrors.CodeInvalidRequest:    {language.Chinese: "请求格式错误", language.English: "Invalid request format"},
	xerrors.CodeResourceNotFound:  {language.Chinese: "资源不存在", language.English: "Resource not found"},
	xerrors.CodeDuplicateResource: {language.Chinese: "资源已存在", language.English: "Resource already exists"},
	xerrors.CodeRateLimitExceeded: {language.Chinese: "请求频率限制", language.English: "Rate limit exceeded"},

	// 2xxxxx: 认证相关错误码
	xerrors.CodeAuthenticationFailed: {language.Chinese: "认证失败", language.English: "Authentication failed"},
	xerrors.CodeInvalidToken:         {language.Chinese: "无效令牌", language.English: "Invalid token"},
	xerrors.CodeTokenExpired:         {language.Chinese: "令牌过期", language.English: "Token expired"},
	xerrors.CodeInvalidCredentials:   {language.Chinese: "凭据无效", language.English: "Invalid credentials"},
	xerrors.CodeAccountLocked:        {language.Chinese: "账户被锁定", language.English: "Account locked"},
	xerrors.CodeAccountBanned:        {language.Chinese: "账户被封禁", language.English: "Account banned"},
	xerrors.CodeSessionExpired:       {language.Chinese: "会话过期", language.English: "Session expired"},

	// 3xxxxx: 权限相关错误码
	xerrors.CodePermissionDenied:       {language.Chinese: "权限不足", language.English: "Permission denied"},
	xerrors.CodeInsufficientPrivileges: {language.Chinese: "权限级别不够", language.English: "Insufficient privileges"},
	xerrors.CodeRoleNotAssigned:        {language.Chinese: "角色未分配", language.English: "Role not assigned"},
	xerrors.CodePermissionNotExists:    {language.Chinese: "权限不存在", language.English: "Permission does not exist"},

	// 4xxxxx: 用户管理错误码
	xerrors.CodeUserNotFound:      {language.Chinese: "用户不存在", language.English: "User not found"},
	xerrors.CodeUserAlreadyExists: {language.Chinese: "用户已存在", language.English: "User already exists"},
	xerrors.CodeUsernameExists:    {language.Chinese: "用户名已存在", language.English: "Username already exists"},
	xerrors.CodeEmailExists:       {language.Chinese: "邮箱已存在", language.English: "Email already exists"},
	xerrors.CodePhoneExists:       {language.Chinese: "手机号已存在", language.English: "Phone number already exists"},
	xerrors.CodeInvalidUserStatus: {language.Chinese: "用户状态无效", language.English: "Invalid user status"},
	xerrors.CodeUserBanned:        {language.Chinese: "用户被封禁", language.English: "User banned"},

	// 5xxxxx: 角色权限管理错误码
	xerrors.CodeRoleNotFound:           {language.Chinese: "角色不存在", language.English: "Role not found"},
	xerrors.CodeRoleAlreadyExists:      {language.Chinese: "角色已存在", language.English: "Role already exists"},
	xerrors.CodeRoleInUse:              {language.Chinese: "角色正在使用中", language.English: "Role is in use"},
	xerrors.CodeSystemRoleProtected:    {language.Chinese: "系统角色受保护", language.English: "System role is protected"},
	xerrors.CodePermissionAssignFailed: {language.Chinese: "权限分配失败", language.English: "Permission assignment failed"},

	// 6xxxxx: 业务逻辑错误码
	xerrors.CodeBusinessLogicError:  {language.Chinese: "业务逻辑错误", language.English: "Business logic error"},
	xerrors.CodeDataIntegrityError:  {language.Chinese: "数据完整性错误", language.English: "Data integrity error"},
	xerrors.CodeOperationNotAllowed: {language.Chinese: "操作不被允许", language.English: "Operation not allowed"},
	xerrors.CodeResourceLocked:      {language.Chinese: "资源被锁定", language.English: "Resource locked"},
	xerrors.CodeQuotaExceeded:       {language.Chinese: "配额超限", language.English: "Quota exceeded"},

	// 7xxxxx: 外部服务错误码
	xerrors.CodeExternalServiceError: {language.Chinese: "外部服务错误", language.English: "External service error"},
	xerrors.CodeKratosError:          {language.Chinese: "Kratos服务错误", language.English: "Kratos service error"},
	xerrors.CodeDatabaseError:        {language.Chinese: "数据库错误", language.English: "Database error"},
	xerrors.CodeCacheError:           {language.Chinese: "缓存服务错误", language.English: "Cache service error"},
	xerrors.CodeMessageQueueError:    {language.Chinese: "消息队列错误", language.English: "Message queue error"},

	// 8xxxxx: 游戏业务错误码
	// 角色相关 (80xxxx)
	xerrors.CodeHeroNotFound:    {language.Chinese: "角色不存在", language.English: "Hero not found"},
	xerrors.CodeHeroLevelTooLow: {language.Chinese: "角色等级不足", language.English: "Hero level too low"},
	xerrors.CodeHeroMaxCount:    {language.Chinese: "角色数量已达上限", language.English: "Hero count limit reached"},
	xerrors.CodeHeroNameExists:  {language.Chinese: "角色名已被使用", language.English: "Hero name already taken"},
	xerrors.CodeHeroStatInvalid: {language.Chinese: "角色属性无效", language.English: "Invalid hero stat"},

	// 技能相关 (81xxxx)
	xerrors.CodeSkillNotFound:   {language.Chinese: "技能不存在", language.English: "Skill not found"},
	xerrors.CodeSkillNotLearned: {language.Chinese: "技能未学习", language.English: "Skill not learned"},
	xerrors.CodeSkillCooldown:   {language.Chinese: "技能冷却中", language.English: "Skill on cooldown"},
	xerrors.CodeSkillManaCost:   {language.Chinese: "法力值不足", language.English: "Insufficient mana"},
	xerrors.CodeSkillInvalidUse: {language.Chinese: "技能使用条件不满足", language.English: "Skill requirements not met"},

	// 职业相关 (82xxxx)
	xerrors.CodeClassNotFound:       {language.Chinese: "职业不存在", language.English: "Class not found"},
	xerrors.CodeClassNotMeetReq:     {language.Chinese: "不满足职业要求", language.English: "Class requirements not met"},
	xerrors.CodeClassAlreadyAdvaced: {language.Chinese: "职业已进阶", language.English: "Class already advanced"},
}

// GetErrorMessage 获取错误码对应语言的消息
func GetErrorMessage(code xerrors.ErrorCode, lang language.Tag) string {
	if messages, ok := ErrorMessages[code]; ok {
		if msg, ok := messages[lang]; ok {
			return msg
		}
		// 如果指定语言没有翻译，返回中文（默认）
		if msg, ok := messages[language.Chinese]; ok {
			return msg
		}
	}
	// 如果完全没有定义，返回通用错误消息
	if lang == language.English {
		return "Unknown error"
	}
	return "未知错误"
}

// init 初始化消息目录
func init() {
	// 为每个错误码注册翻译
	for code, messages := range ErrorMessages {
		codeInt := code.ToInt()
		for lang, msg := range messages {
			message.SetString(lang, string(rune(codeInt)), msg)
		}
	}
}
