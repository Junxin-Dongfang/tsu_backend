// File: internal/pkg/i18n/i18n.go
package i18n

import (
	"context"
	"strings"

	"tsu-self/internal/pkg/ctxkey"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// 支持的语言
var (
	// 默认语言为中文
	DefaultLanguage = language.Chinese
	// 支持的语言列表
	SupportedLanguages = []language.Tag{
		language.Chinese, // zh
		language.English, // en
	}
	// 语言匹配器
	matcher = language.NewMatcher(SupportedLanguages)
)

// WithLanguage 在 context 中设置语言偏好
func WithLanguage(ctx context.Context, lang language.Tag) context.Context {
	return context.WithValue(ctx, ctxkey.Language, lang)
}

// GetLanguage 从 context 中获取语言偏好
func GetLanguage(ctx context.Context) language.Tag {
	if lang, ok := ctx.Value(ctxkey.Language).(language.Tag); ok {
		return lang
	}
	return DefaultLanguage
}

// ParseAcceptLanguage 解析 Accept-Language 头部
// 例如: "zh-CN,zh;q=0.9,en;q=0.8,en-US;q=0.7"
func ParseAcceptLanguage(acceptLanguage string) language.Tag {
	if acceptLanguage == "" {
		return DefaultLanguage
	}

	// 解析并匹配最佳语言
	tags, _, err := language.ParseAcceptLanguage(acceptLanguage)
	if err != nil || len(tags) == 0 {
		return DefaultLanguage
	}

	// 使用 matcher 找到最佳匹配
	tag, _, _ := matcher.Match(tags...)
	return tag
}

// ParseLanguageCode 从语言代码解析 Tag
// 支持: "zh", "zh-CN", "en", "en-US" 等
func ParseLanguageCode(code string) language.Tag {
	code = strings.ToLower(strings.TrimSpace(code))
	if code == "" {
		return DefaultLanguage
	}

	tag, err := language.Parse(code)
	if err != nil {
		return DefaultLanguage
	}

	// 匹配到支持的语言
	matched, _, _ := matcher.Match(tag)
	return matched
}

// Printer 返回指定语言的打印器
func Printer(lang language.Tag) *message.Printer {
	return message.NewPrinter(lang)
}

// T 翻译函数 - 从 context 中获取语言并翻译
func T(ctx context.Context, key message.Reference, args ...interface{}) string {
	lang := GetLanguage(ctx)
	p := message.NewPrinter(lang)
	return p.Sprintf(key, args...)
}

// Translate 直接翻译（不依赖 context）
func Translate(lang language.Tag, key message.Reference, args ...interface{}) string {
	p := message.NewPrinter(lang)
	return p.Sprintf(key, args...)
}

// IsSupported 检查语言是否被支持
func IsSupported(lang language.Tag) bool {
	for _, supported := range SupportedLanguages {
		if lang == supported {
			return true
		}
	}
	return false
}

// GetLanguageCode 获取语言代码 (zh, en)
func GetLanguageCode(lang language.Tag) string {
	base, _ := lang.Base()
	return base.String()
}
