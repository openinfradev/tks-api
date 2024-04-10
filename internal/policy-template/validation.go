package policytemplate

import (
	"fmt"
	"slices"
	"strings"

	"github.com/openinfradev/tks-api/pkg/domain"
)

var VALID_PARAM_TYPES = []string{"string", "number", "integer", "object", "boolean", "null"}

func ValidateParamDef(paramdef *domain.ParameterDef) error {
	paramType := paramdef.Type

	baseType := strings.TrimSuffix(paramType, "[]")

	// 타입과 []를 제거한 타입이 같은데 배열이면 에러임
	if (baseType == paramType) != paramdef.IsArray {
		return fmt.Errorf("type is '%s', but IsArray=%v", paramType, paramdef.IsArray)
	}

	if slices.Contains(VALID_PARAM_TYPES, baseType) {
		return nil
	}

	return fmt.Errorf("%s is not valid type", paramType)
}

func ValidateParamDefs(paramdefs []*domain.ParameterDef) error {
	for _, paramdef := range paramdefs {
		err := ValidateParamDef(paramdef)
		if err != nil {
			return err
		}

		err = ValidateParamDefs(paramdef.Children)

		if err != nil {
			return err
		}
	}

	return nil
}
