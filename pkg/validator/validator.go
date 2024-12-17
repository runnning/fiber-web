package validator

import (
	"errors"
	"reflect"
	"strings"

	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	entranslations "github.com/go-playground/validator/v10/translations/en"
	zhtranslations "github.com/go-playground/validator/v10/translations/zh"
)

// Validator 封装验证器
type Validator struct {
	Validate *validator.Validate
	Trans    ut.Translator
}

// Config 验证器配置
type Config struct {
	Language string // 语言设置: "en", "zh"
}

// New 创建新的验证器实例
func New(cfg *Config) *Validator {
	v := &Validator{
		Validate: validator.New(),
	}

	// 注册标签处理函数
	v.Validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	// 设置翻译器
	enLocale := en.New()
	zhLocale := zh.New()
	uni := ut.New(enLocale, zhLocale)

	// 获取指定语言的翻译器
	var found bool
	v.Trans, found = uni.GetTranslator(cfg.Language)
	if !found {
		v.Trans, _ = uni.GetTranslator("en")
	}

	// 注册翻译器
	switch cfg.Language {
	case "zh":
		_ = zhtranslations.RegisterDefaultTranslations(v.Validate, v.Trans)
	default:
		_ = entranslations.RegisterDefaultTranslations(v.Validate, v.Trans)
	}

	return v
}

// ValidateStruct 验证结构体 (改名以避免与字段冲突)
func (v *Validator) ValidateStruct(s interface{}) error {
	return v.Validate.Struct(s)
}

// ValidateVar 验证单个变量
func (v *Validator) ValidateVar(field interface{}, tag string) error {
	return v.Validate.Var(field, tag)
}

// TranslateError 转换验证错误为易读的消息
func (v *Validator) TranslateError(err error) map[string]string {
	if err == nil {
		return nil
	}
	errorMap := make(map[string]string)
	var validatorErrs validator.ValidationErrors
	if !errors.As(err, &validatorErrs) {
		return nil
	}
	for _, e := range validatorErrs {
		errorMap[e.Field()] = e.Translate(v.Trans)
	}

	return errorMap
}
