package validator

import (
	"context"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	validator "github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/log"
)

const (
	REGEX_RFC1123           = `^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])$`
	REGEX_SIMPLE_SEMVER     = `^v\d+\.\d+\.\d+$`
	REGEX_PASCAL_CASE       = `^([A-Z][a-z\d]+)+$` // 대문자로 시작하는 camel case(pascal case or upper camel case)를 표현한 정규식
	REGEX_RFC1123_DNS_LABEL = "[a-z0-9]([-a-z0-9]*[a-z0-9])?"
	REGEX_RESOURCE_NAME     = `^` + REGEX_RFC1123_DNS_LABEL + "$"
	REGEX_RFC1123_SUBDOMAIN = `^` + REGEX_RFC1123_DNS_LABEL + `(\.` + REGEX_RFC1123_DNS_LABEL + `)*$`
	REGEX_TEMPLATE_KIND     = `^[A-Z][a-zA-Z0-9]+$`
)

func NewValidator() (*validator.Validate, *ut.UniversalTranslator) {
	en := en.New()
	uni := ut.New(en, en)
	trans, _ := uni.GetTranslator("en")

	v := validator.New()
	err := en_translations.RegisterDefaultTranslations(v, trans)
	if err != nil {
		log.Error(context.TODO(), err)
	}

	// register custom validator
	_ = v.RegisterValidation("rfc1123", validateRfc1123)
	_ = v.RegisterValidation("name", validateName)
	_ = v.RegisterValidation("version", validateVersion)
	_ = v.RegisterValidation("pascalcase", validatePascalCase)
	_ = v.RegisterValidation("resourcename", validateResourceName)
	_ = v.RegisterValidation("matchnamespace", validateMatchNamespace)
	_ = v.RegisterValidation("matchkinds", validateMatchKinds)
	_ = v.RegisterValidation("templatekind", validateTemplateKind)

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
	if utf8.RuneCountInString(fl.Field().String()) > 30 {
		return false
	}
	r, _ := regexp.Compile(REGEX_RFC1123)
	return r.MatchString(fl.Field().String())

}

func validateName(fl validator.FieldLevel) bool {
	if fl.Field().String() == "" {
		return false
	}

	return utf8.RuneCountInString(fl.Field().String()) <= 30
}

func validateVersion(fl validator.FieldLevel) bool {
	if fl.Field().String() == "" {
		return false
	}

	r, _ := regexp.Compile(REGEX_SIMPLE_SEMVER)
	return r.MatchString(fl.Field().String())
}

func validatePascalCase(fl validator.FieldLevel) bool {
	if fl.Field().String() == "" {
		return false
	}

	r, _ := regexp.Compile(REGEX_PASCAL_CASE)
	return r.MatchString(fl.Field().String())
}

func validateResourceName(fl validator.FieldLevel) bool {
	// 정책 리소스 이름을 지정하지 않으면 지정하기 때문에 유효함
	// 다른 리소스 이름을 처리할 때에는 분리 필요
	if fl.Field().String() == "" {
		return true
	}

	r, _ := regexp.Compile(REGEX_RESOURCE_NAME)
	return r.MatchString(fl.Field().String())
}

func validateMatchKinds(fl validator.FieldLevel) bool {
	kinds, ok := fl.Field().Interface().([]domain.Kinds)
	if !ok {
		return false
	}

	for _, kind := range kinds {
		if ok := validateMatchKindAPIGroup(kind.APIGroups) && validateMatchKindKind(kind.Kinds); !ok {
			return false
		}
	}

	return true
}

func validateTemplateKind(fl validator.FieldLevel) bool {
	if fl.Field().String() == "" {
		return false
	}

	r, _ := regexp.Compile(REGEX_TEMPLATE_KIND)
	return r.MatchString(fl.Field().String())
}

func validateMatchKindAPIGroup(apigroups []string) bool {
	if len(apigroups) == 0 {
		return true
	}

	containsWildcard := false

	r, _ := regexp.Compile(REGEX_RFC1123_SUBDOMAIN)

	for _, apigroup := range apigroups {
		if apigroup == "*" || apigroup == "" {
			containsWildcard = true
		} else {
			if !r.MatchString(apigroup) {
				return false
			}
		}
	}

	if containsWildcard && len(apigroups) != 1 {
		return false
	}

	return true
}

func validateMatchKindKind(kinds []string) bool {
	if len(kinds) == 0 {
		return true
	}

	containsWildcard := false

	r, _ := regexp.Compile(REGEX_PASCAL_CASE)

	for _, kind := range kinds {
		if kind == "*" || kind == "" {
			containsWildcard = true
		} else {
			if !r.MatchString(kind) {
				return false
			}
		}
	}

	if containsWildcard && len(kinds) != 1 {
		return false
	}

	return true
}

func validateMatchNamespace(fl validator.FieldLevel) bool {
	namespaces, ok := fl.Field().Interface().([]string)
	if !ok {
		return false
	}

	if len(namespaces) == 0 {
		return true
	}

	containsWildcard := false

	r, _ := regexp.Compile(REGEX_RESOURCE_NAME)

	for _, namespace := range namespaces {
		if namespace == "*" || namespace == "" {
			containsWildcard = true
		} else {
			trimmed := strings.TrimSuffix(strings.TrimPrefix(namespace, "*"), "*")
			if !r.MatchString(trimmed) {
				return false
			}
		}
	}

	if containsWildcard && len(namespaces) != 1 {
		return false
	}

	return true
}
