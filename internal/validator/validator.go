package validator

import (
	"regexp"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	validator "github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
)

const REGEX_RFC1123 = `^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])$`

func NewValidator() (*validator.Validate, *ut.UniversalTranslator) {
	en := en.New()
	uni := ut.New(en, en)
	trans, _ := uni.GetTranslator("en")

	v := validator.New()
	en_translations.RegisterDefaultTranslations(v, trans)

	// register custom validator
	_ = v.RegisterValidation("rfc1123", validateRfc1123)
	_ = v.RegisterValidation("name", validateName)

	// register custom error
	_ = v.RegisterTranslation("required", trans, func(ut ut.Translator) error {
		return ut.Add("required", "[{0}] 값이 입력되지 않았습니다.", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("required", fe.Field())
		return t
	})

	_ = v.RegisterTranslation("name", trans, func(ut ut.Translator) error {
		return ut.Add("name", "[{0}] 값은 최대 30자내로 입력하세요.", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("name", fe.Field())
		return t
	})

	return v, uni
}

func validateRfc1123(fl validator.FieldLevel) bool {
	if fl.Field().String() == "" {
		return false
	}
	if len(fl.Field().String()) > 30 {
		return false
	}
	r, _ := regexp.Compile(REGEX_RFC1123)
	return r.MatchString(fl.Field().String())

}

func validateName(fl validator.FieldLevel) bool {
	if fl.Field().String() == "" {
		return false
	}
	if len(fl.Field().String()) > 30 {
		return false
	}
	return true
}
