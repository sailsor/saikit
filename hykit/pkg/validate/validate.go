package validate

import (
	"reflect"
	"strings"

	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"
	"github.com/pkg/errors"
)

//验证服务接口
type ValidateRepo interface {
	SetTagName(name string)
	ValidateStruct(i interface{}) error
}

//Validate 验证实例
type Validate struct {
	validate *validator.Validate
	trans    ut.Translator
}

func NewValidateRepo() ValidateRepo {
	v := new(Validate)
	v.validate = validator.New()
	/*中文翻译*/
	zht := zh.New()
	uni := ut.New(zht, zht)
	v.trans, _ = uni.GetTranslator("zh")
	//语言使用中文
	_ = zh_translations.RegisterDefaultTranslations(v.validate, v.trans)
	// 设置yaml TAG name
	v.SetTagName("mapstructure")
	_ = v.validate.RegisterTranslation("required_if", v.trans,
		func(ut ut.Translator) error {
			return ut.Add("required_if", "{0}; {1}为必填字段，烦请确认;", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("required_if", fe.Param(), fe.Field())
			return t
		})
	_ = v.validate.RegisterTranslation("unique", v.trans,
		func(ut ut.Translator) error {
			return ut.Add("unique", "{0}的值不能重复，烦请确认;", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("unique", fe.Field())
			return t
		})
	return v
}

//设置校验返回信息
func (v *Validate) SetTagName(name string) {
	v.validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		tagName := strings.SplitN(fld.Tag.Get(name), ",", 2)[0]
		if tagName == "-" {
			return ""
		}
		return tagName
	})
}

func (v *Validate) ValidateStruct(i interface{}) error {
	err := v.validate.Struct(i)
	if err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			return errors.Wrapf(err, "结构体规则配置校验失败:[%s]", err)
		}
		for _, err := range err.(validator.ValidationErrors) {
			return errors.New(err.Translate(v.trans))
		}
	}
	return nil
}
