// File: internal/pkg/xerrors/codes.go
package xerrors

import "fmt"

// ErrorCode 错误码类型（类型安全）
type ErrorCode int

// IsValid 检查错误码是否在预定义列表中
func (c ErrorCode) IsValid() bool {
	_, exists := codeMessages[c]
	return exists
}

// String 返回错误码的字符串表示
func (c ErrorCode) String() string {
	if msg, ok := codeMessages[c]; ok {
		return fmt.Sprintf("%d (%s)", c, msg)
	}
	return fmt.Sprintf("%d (未定义的错误码)", c)
}

// Message 返回错误码对应的消息
func (c ErrorCode) Message() string {
	if msg, ok := codeMessages[c]; ok {
		return msg
	}
	return "未知错误"
}

// ToInt 转换为 int（用于 JSON 序列化等场景）
func (c ErrorCode) ToInt() int {
	return int(c)
}

// -----------------------------------------------------------------------------
// 业务错误码统一定义
// 按模块或领域对错误码进行分段，便于管理。
// -----------------------------------------------------------------------------
const (
	// 1xxxxx: 通用错误码
	CodeSuccess           ErrorCode = 100000 // 操作成功
	CodeInternalError     ErrorCode = 100001 // 内部服务错误
	CodeInvalidParams     ErrorCode = 100002 // 参数错误
	CodeInvalidRequest    ErrorCode = 100003 // 请求格式错误
	CodeResourceNotFound  ErrorCode = 100404 // 资源不存在
	CodeDuplicateResource ErrorCode = 100409 // 资源已存在
	CodeRateLimitExceeded ErrorCode = 100429 // 请求频率限制

	// 2xxxxx: 认证相关错误码
	CodeAuthenticationFailed ErrorCode = 200001 // 认证失败
	CodeInvalidToken         ErrorCode = 200002 // 无效令牌
	CodeTokenExpired         ErrorCode = 200003 // 令牌过期
	CodeInvalidCredentials   ErrorCode = 200004 // 凭据无效
	CodeAccountLocked        ErrorCode = 200005 // 账户被锁定
	CodeAccountBanned        ErrorCode = 200006 // 账户被封禁
	CodeSessionExpired       ErrorCode = 200007 // 会话过期

	// 3xxxxx: 权限相关错误码
	CodePermissionDenied       ErrorCode = 300001 // 权限不足
	CodeInsufficientPrivileges ErrorCode = 300002 // 权限级别不够
	CodeRoleNotAssigned        ErrorCode = 300003 // 角色未分配
	CodePermissionNotExists    ErrorCode = 300004 // 权限不存在

	// 4xxxxx: 用户管理错误码
	CodeUserNotFound      ErrorCode = 400001 // 用户不存在
	CodeUserAlreadyExists ErrorCode = 400002 // 用户已存在
	CodeUsernameExists    ErrorCode = 400003 // 用户名已存在
	CodeEmailExists       ErrorCode = 400004 // 邮箱已存在
	CodePhoneExists       ErrorCode = 400005 // 手机号已存在
	CodeInvalidUserStatus ErrorCode = 400006 // 用户状态无效
	CodeUserBanned        ErrorCode = 400007 // 用户被封禁

	// 5xxxxx: 角色权限管理错误码
	CodeRoleNotFound           ErrorCode = 500001 // 角色不存在
	CodeRoleAlreadyExists      ErrorCode = 500002 // 角色已存在
	CodeRoleInUse              ErrorCode = 500003 // 角色正在使用中
	CodeSystemRoleProtected    ErrorCode = 500004 // 系统角色受保护
	CodePermissionAssignFailed ErrorCode = 500005 // 权限分配失败

	// 6xxxxx: 业务逻辑错误码
	CodeBusinessLogicError  ErrorCode = 600001 // 业务逻辑错误
	CodeDataIntegrityError  ErrorCode = 600002 // 数据完整性错误
	CodeOperationNotAllowed ErrorCode = 600003 // 操作不被允许
	CodeResourceLocked      ErrorCode = 600004 // 资源被锁定
	CodeQuotaExceeded       ErrorCode = 600005 // 配额超限

	// 7xxxxx: 外部服务错误码
	CodeExternalServiceError ErrorCode = 700001 // 外部服务错误
	CodeKratosError          ErrorCode = 700002 // Kratos服务错误
	CodeDatabaseError        ErrorCode = 700003 // 数据库错误
	CodeCacheError           ErrorCode = 700004 // 缓存服务错误
	CodeMessageQueueError    ErrorCode = 700005 // 消息队列错误

	// 8xxxxx: 游戏业务错误码
	// 角色相关 (80xxxx)
	CodeHeroNotFound           ErrorCode = 800001 // 角色不存在
	CodeHeroLevelTooLow        ErrorCode = 800002 // 角色等级不足
	CodeHeroMaxCount           ErrorCode = 800003 // 角色数量已达上限
	CodeHeroNameExists         ErrorCode = 800004 // 角色名已存在
	CodeHeroStatInvalid        ErrorCode = 800005 // 角色属性无效
	CodeInsufficientResource   ErrorCode = 800006 // 资源不足
	CodeOperationExpired       ErrorCode = 800007 // 操作已过期
	CodeInsufficientExperience ErrorCode = 800008 // 经验不足
	CodeInsufficientLevel      ErrorCode = 800009 // 等级不足
	CodeInsufficientAttributes ErrorCode = 800010 // 属性不足
	CodeInsufficientSkills     ErrorCode = 800011 // 技能不足

	// 技能相关 (81xxxx)
	CodeSkillNotFound            ErrorCode = 810001 // 技能不存在
	CodeSkillNotLearned          ErrorCode = 810002 // 技能未学习
	CodeSkillCooldown            ErrorCode = 810003 // 技能冷却中
	CodeSkillManaCost            ErrorCode = 810004 // 法力值不足
	CodeSkillInvalidUse          ErrorCode = 810005 // 技能使用条件不满足
	CodeSkillPrerequisiteNotMet  ErrorCode = 810006 // 前置技能未满足
	CodeSkillInvalidPrerequisite ErrorCode = 810007 // 前置技能配置无效

	// 职业相关 (82xxxx)
	CodeClassNotFound       ErrorCode = 820001 // 职业不存在
	CodeClassNotMeetReq     ErrorCode = 820002 // 不满足职业要求
	CodeClassAlreadyAdvaced ErrorCode = 820003 // 职业已进阶
)

// -----------------------------------------------------------------------------
// HTTP 状态码常量定义
// -----------------------------------------------------------------------------
const (
	HTTPStatusOK        = 200 // 请求成功
	HTTPStatusCreated   = 201 // 资源已创建
	HTTPStatusAccepted  = 202 // 请求已接受但未处理
	HTTPStatusNoContent = 204 // 请求成功但无内容返回

	HTTPStatusBadRequest          = 400 // 错误请求
	HTTPStatusUnauthorized        = 401 // 未经授权
	HTTPStatusForbidden           = 403 // 禁止访问
	HTTPStatusNotFound            = 404 // 资源未找到
	HTTPStatusMethodNotAllowed    = 405 // 方法不被允许
	HTTPStatusConflict            = 409 // 资源冲突
	HTTPStatusUnprocessableEntity = 422 // 无法处理的实体
	HTTPStatusTooManyRequests     = 429 // 请求过多

	HTTPStatusInternalServerError = 500 // 内部服务器错误
	HTTPStatusNotImplemented      = 501 // 未实现
	HTTPStatusBadGateway          = 502 // 错误网关
	HTTPStatusServiceUnavailable  = 503 // 服务不可用
	HTTPStatusGatewayTimeout      = 504 // 网关超时
)

// -----------------------------------------------------------------------------
// 错误消息映射
// -----------------------------------------------------------------------------
var codeMessages = map[ErrorCode]string{
	CodeSuccess:           "操作成功",
	CodeInternalError:     "内部服务错误",
	CodeInvalidParams:     "参数错误",
	CodeInvalidRequest:    "请求格式错误",
	CodeResourceNotFound:  "资源不存在",
	CodeDuplicateResource: "资源已存在",
	CodeRateLimitExceeded: "请求频率限制",

	CodeAuthenticationFailed: "认证失败",
	CodeInvalidToken:         "无效令牌",
	CodeTokenExpired:         "令牌过期",
	CodeInvalidCredentials:   "凭据无效",
	CodeAccountLocked:        "账户被锁定",
	CodeAccountBanned:        "账户被封禁",
	CodeSessionExpired:       "会话过期",

	CodePermissionDenied:       "权限不足",
	CodeInsufficientPrivileges: "权限级别不够",
	CodeRoleNotAssigned:        "角色未分配",
	CodePermissionNotExists:    "权限不存在",

	CodeUserNotFound:      "用户不存在",
	CodeUserAlreadyExists: "用户已存在",
	CodeUsernameExists:    "用户名已存在",
	CodeEmailExists:       "邮箱已存在",
	CodePhoneExists:       "手机号已存在",
	CodeInvalidUserStatus: "用户状态无效",
	CodeUserBanned:        "用户被封禁",

	CodeRoleNotFound:           "角色不存在",
	CodeRoleAlreadyExists:      "角色已存在",
	CodeRoleInUse:              "角色正在使用中",
	CodeSystemRoleProtected:    "系统角色受保护",
	CodePermissionAssignFailed: "权限分配失败",

	CodeBusinessLogicError:  "业务逻辑错误",
	CodeDataIntegrityError:  "数据完整性错误",
	CodeOperationNotAllowed: "操作不被允许",
	CodeResourceLocked:      "资源被锁定",
	CodeQuotaExceeded:       "配额超限",

	CodeExternalServiceError: "外部服务错误",
	CodeKratosError:          "Kratos服务错误",
	CodeDatabaseError:        "数据库错误",
	CodeCacheError:           "缓存服务错误",
	CodeMessageQueueError:    "消息队列错误",

	// 游戏业务错误消息
	CodeHeroNotFound:           "角色不存在",
	CodeHeroLevelTooLow:        "角色等级不足",
	CodeHeroMaxCount:           "角色数量已达上限",
	CodeHeroNameExists:         "角色名已被使用",
	CodeHeroStatInvalid:        "角色属性无效",
	CodeInsufficientResource:   "资源不足",
	CodeOperationExpired:       "操作已过期",
	CodeInsufficientExperience: "经验不足",
	CodeInsufficientLevel:      "等级不足",
	CodeInsufficientAttributes: "属性不足",
	CodeInsufficientSkills:     "技能不足",
	CodeSkillNotFound:            "技能不存在",
	CodeSkillNotLearned:          "技能未学习",
	CodeSkillCooldown:            "技能冷却中",
	CodeSkillManaCost:            "法力值不足",
	CodeSkillInvalidUse:          "技能使用条件不满足",
	CodeSkillPrerequisiteNotMet:  "前置技能未满足",
	CodeSkillInvalidPrerequisite: "前置技能配置无效",
	CodeClassNotFound:            "职业不存在",
	CodeClassNotMeetReq:          "不满足职业要求",
	CodeClassAlreadyAdvaced:      "职业已进阶",
}

// GetHTTPStatus 根据业务错误码获取HTTP状态码
func GetHTTPStatus(code ErrorCode) int {
	switch {
	case code == CodeSuccess:
		return HTTPStatusOK
	case code >= 200000 && code < 300000:
		if code == CodeAuthenticationFailed || code == CodeInvalidToken || code == CodeTokenExpired || code == CodeInvalidCredentials {
			return HTTPStatusUnauthorized
		}
		return HTTPStatusForbidden
	case code >= 300000 && code < 400000:
		return HTTPStatusForbidden
	case code >= 400000 && code < 500000:
		if code == CodeUserNotFound {
			return HTTPStatusNotFound
		}
		if code == CodeUserAlreadyExists || code == CodeUsernameExists || code == CodeEmailExists || code == CodePhoneExists {
			return HTTPStatusConflict
		}
		return HTTPStatusBadRequest
	case code == CodeResourceNotFound:
		return HTTPStatusNotFound
	case code == CodeDuplicateResource:
		return HTTPStatusConflict
	case code == CodeInvalidParams || code == CodeInvalidRequest:
		return HTTPStatusBadRequest
	case code == CodeRateLimitExceeded:
		return HTTPStatusTooManyRequests
	case code >= 500000 && code < 600000:
		return HTTPStatusBadRequest
	case code >= 600000 && code < 700000:
		return HTTPStatusBadRequest
	case code >= 700000:
		return HTTPStatusServiceUnavailable
	default:
		return HTTPStatusInternalServerError
	}
}

// 辅助函数
// getCategoryByCode 根据错误码获取分类
func getCategoryByCode(code ErrorCode) string {
	switch {
	case code >= 100000 && code < 200000:
		return "system"
	case code >= 200000 && code < 300000:
		return "authentication"
	case code >= 300000 && code < 400000:
		return "authorization"
	case code >= 400000 && code < 500000:
		return "user"
	case code >= 500000 && code < 600000:
		return "role"
	case code >= 600000 && code < 700000:
		return "business"
	case code >= 700000 && code < 800000:
		return "external"
	case code >= 800000 && code < 900000:
		return "game"
	default:
		return "unknown"
	}
}

// getLevelByCode 根据错误码获取级别
func getLevelByCode(code ErrorCode) ErrorLevel {
	switch {
	case code == CodeSuccess:
		return LevelInfo
	case code >= 100001 && code <= 100003: // 参数错误等
		return LevelWarn
	case code >= 700001: // 外部服务错误
		return LevelCritical
	default:
		return LevelError
	}
}

// isRetryableByCode 根据错误码判断是否可重试
func isRetryableByCode(code ErrorCode) bool {
	retryableCodes := map[ErrorCode]bool{
		CodeInternalError:        true,
		CodeExternalServiceError: true,
		CodeKratosError:          true,
		CodeDatabaseError:        true,
		CodeCacheError:           true,
		CodeMessageQueueError:    true,
		CodeRateLimitExceeded:    true,
	}
	return retryableCodes[code]
}
