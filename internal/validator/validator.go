package validator

import (
	"regexp"

	"github.com/go-playground/validator"
)

const REGEX_RFC1123 = `^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])$`

func NewValidator() *validator.Validate {
	validate := validator.New()

	_ = validate.RegisterValidation("rfc1123", validateRfc1123)
	_ = validate.RegisterValidation("name", validateName)

	return validate
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
