package policytemplate

import (
	"fmt"
	"slices"
	"strings"
)

var VALID_PARAM_TYPES = []string{"string", "number", "integer", "object", "boolean", "null"}

func ValidateParamDefType(paramType string) error {
	baseType := strings.TrimSuffix(paramType, "[]")

	if slices.Contains(VALID_PARAM_TYPES, baseType) {
		return nil
	}

	return fmt.Errorf("%s is not valid type", paramType)
}
