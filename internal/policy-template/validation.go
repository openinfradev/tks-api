package policytemplate

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/xeipuuv/gojsonschema"
)

var VALID_PARAM_TYPES = []string{"string", "number", "integer", "object", "boolean", "null"}

func ValidateParamDef(paramdef *domain.ParameterDef) error {
	paramType := paramdef.Type

	baseType := strings.TrimSuffix(paramType, "[]")

	// 타입과 []를 제거한 타입이 다르면 array인데 이것과 isArray 속성이 다르면 에러임
	if (baseType != paramType) != paramdef.IsArray {
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

func ValidateJSONusingParamdefs(paramdefs []*domain.ParameterDef, jsonStr string) error {
	jsonSchema := ParamDefsToJSONSchemaProeprties(paramdefs)

	if jsonSchema == nil {
		// 파라미터가 없는데 "{}" 이나 ""이면 에러가 아님
		if isEmptyValue(jsonStr) {
			return nil
		} else {
			return fmt.Errorf("schema has no field but value '%v' specified", jsonStr)
		}
	}

	// Load JSON Schema
	schemaLoader := gojsonschema.NewGoLoader(jsonSchema)

	// Load JSON data
	documentLoader := gojsonschema.NewStringLoader(jsonStr)

	// Validate JSON against JSON Schema
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return err
	}

	// Check if the result is valid
	if result.Valid() {
		return nil
	}

	schemaBytes, _ := json.Marshal(paramdefs)
	jsonSchemaBytes, _ := json.Marshal(jsonSchema)

	return fmt.Errorf("value '%s' is not valid against schemas:\njsonschema='%s',\nparamdefs='%s'",
		jsonStr, string(jsonSchemaBytes), string(schemaBytes))
}

func isEmptyValue(value string) bool {
	val := strings.TrimSpace(value)
	val = strings.TrimPrefix(val, "{")
	val = strings.TrimSuffix(val, "}")
	val = strings.TrimSpace(val)

	return len(val) == 0
}
