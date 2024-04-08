package policytemplate

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/openinfradev/tks-api/pkg/domain"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func GetNewParamDefs(paramdefs1 []*domain.ParameterDef, paramdefs2 []*domain.ParameterDef) (newParamdefs []*domain.ParameterDef, err error) {
	result := []*domain.ParameterDef{}

	if len(paramdefs1) > len(paramdefs2) {
		return nil, errors.New("not compatible, parameter number reduced")
	}

	for _, paramdef2 := range paramdefs2 {
		paramdef1 := findParamDefByName(paramdefs1, paramdef2.Key)

		if paramdef1 == nil {
			// Not found, it's new parameter
			result = append(result, paramdef2)
		} else if !CompareParamDef(paramdef2, paramdef1) {
			return nil, fmt.Errorf("not compatible, parameter definition of '%s' is changed", paramdef2.Key)
		}
	}

	return result, nil
}

func GetNewExtractedParamDefs(paramdefs []*domain.ParameterDef, extractedParamdefs []*domain.ParameterDef) (newParamdefs []*domain.ParameterDef, err error) {
	result := []*domain.ParameterDef{}

	if len(paramdefs) > len(extractedParamdefs) {
		return nil, errors.New("not compatible, parameter number reduced")
	}

	for _, extractedParamdef := range extractedParamdefs {
		paramdef := findParamDefByName(paramdefs, extractedParamdef.Key)

		if paramdef == nil {
			// Not found, it's new parameter
			extractedParamdef.MarkNewRecursive()
			result = append(result, extractedParamdef)
		} else if !CompareParamDefAndExtractedParamDef(paramdef, extractedParamdef) {
			return nil, fmt.Errorf("not compatible, parameter definition of '%s' is changed", extractedParamdef.Key)
		}
	}

	return result, nil
}

func findParamDefByName(paramdefs []*domain.ParameterDef, name string) *domain.ParameterDef {
	for _, paramdef := range paramdefs {
		if paramdef.Key == name {
			return paramdef
		}
	}

	return nil
}

func CompareParamDef(paramdef1 *domain.ParameterDef, paramdef2 *domain.ParameterDef) bool {
	if paramdef1 == nil || paramdef2 == nil {
		return paramdef2 == paramdef1
	}

	if paramdef1.Key != paramdef2.Key {
		return false
	}

	if paramdef1.IsArray != paramdef2.IsArray {
		return false
	}

	if paramdef1.Type != paramdef2.Type {
		return false
	}

	if paramdef1.DefaultValue != paramdef2.DefaultValue {
		return false
	}

	if len(paramdef1.Children) != len(paramdef2.Children) {
		return false
	}

	for _, child := range paramdef1.Children {
		child2 := paramdef2.GetChildrenByName(child.Key)

		equals := CompareParamDef(child, child2)

		if !equals {
			return false
		}
	}

	return true
}

func CompareParamDefAndExtractedParamDef(paramdef *domain.ParameterDef, extractedParamdef *domain.ParameterDef) bool {
	if paramdef == nil || extractedParamdef == nil {
		return extractedParamdef == paramdef
	}

	if paramdef.Key != extractedParamdef.Key {
		return false
	}

	if paramdef.IsArray != extractedParamdef.IsArray {
		return false
	}

	// object 기반이면 true, string 등 any 기반이면 false
	paramDefIsObjectBased := paramdef.Type == "object" || paramdef.Type == "object[]"

	// ovject 기반이면 true, any, anㅛ[] 등 기반이면 false
	extractedParamdefIsObjectBased := paramdef.Type == "object" || paramdef.Type == "object[]"

	if paramDefIsObjectBased != extractedParamdefIsObjectBased {
		return false
	}

	if len(paramdef.Children) != len(extractedParamdef.Children) {
		return false
	}

	for _, child := range paramdef.Children {
		child2 := extractedParamdef.GetChildrenByName(child.Key)

		equals := CompareParamDefAndExtractedParamDef(child, child2)

		if !equals {
			return false
		}
	}

	return true
}

func ParamDefsToJSONSchemaProeprties(paramdefs []*domain.ParameterDef) *apiextensionsv1.JSONSchemaProps {
	if paramdefs == nil {
		return nil
	}

	result := apiextensionsv1.JSONSchemaProps{Type: "object", Properties: convert(paramdefs)}

	return &result
}

func convert(paramdefs []*domain.ParameterDef) map[string]apiextensionsv1.JSONSchemaProps {
	result := map[string]apiextensionsv1.JSONSchemaProps{}

	for _, paramdef := range paramdefs {
		isArary := paramdef.IsArray
		isObject := len(paramdef.Children) > 0

		switch {
		case isArary && isObject:
			result[paramdef.Key] = apiextensionsv1.JSONSchemaProps{
				Type: "array",
				Items: &apiextensionsv1.JSONSchemaPropsOrArray{
					Schema: ParamDefsToJSONSchemaProeprties(paramdef.Children),
				},
			}
		case isArary:
			result[paramdef.Key] = apiextensionsv1.JSONSchemaProps{
				Type: "array",
				Items: &apiextensionsv1.JSONSchemaPropsOrArray{
					Schema: &apiextensionsv1.JSONSchemaProps{Type: strings.TrimSuffix(paramdef.Type, "[]")},
				},
			}
		case isObject:
			result[paramdef.Key] = *ParamDefsToJSONSchemaProeprties(paramdef.Children)
		default:
			result[paramdef.Key] = apiextensionsv1.JSONSchemaProps{Type: paramdef.Type}
		}

	}

	return result
}

func JSONSchemaProeprtiesToParamDefs(jsschema *apiextensionsv1.JSONSchemaProps) []*domain.ParameterDef {
	return convertToParameterDef(jsschema).Children
}

func convertToParameterDef(jsschema *apiextensionsv1.JSONSchemaProps) *domain.ParameterDef {
	// result := []ParameterDef{}
	// fmt.Println(jsschema.Properties)
	switch jsschema.Type {
	case "array":
		itemDef := convertToParameterDef(jsschema.Items.Schema)
		itemDef.Type = jsschema.Items.Schema.Type + "[]"
		itemDef.IsArray = true

		return itemDef
	case "object":
		children := []*domain.ParameterDef{}
		for kc, vc := range jsschema.Properties {
			converted := convertToParameterDef(&vc)
			converted.Key = kc
			children = append(children, converted)
		}
		return &domain.ParameterDef{Key: jsschema.ID, Type: jsschema.Type, DefaultValue: "",
			Children: children}
	default:
		defaultValue := ""

		if jsschema.Default != nil {
			defaultValue = string(jsschema.Default.Raw)
		}

		return &domain.ParameterDef{Key: jsschema.ID, Type: jsschema.Type, DefaultValue: defaultValue, Children: []*domain.ParameterDef{}}
	}
}

func FillParamDefFromJsonStr(paramdefs []*domain.ParameterDef, parameters string) (err error) {
	var parametersMap map[string]interface{}

	err = json.Unmarshal([]byte(parameters), &parametersMap)

	if err != nil {
		return err
	}

	return FillParamDefFromJson(paramdefs, &parametersMap)
}

func FillParamDefFromJson(paramdefs []*domain.ParameterDef, parameters *map[string]interface{}) (err error) {
	if len(paramdefs) == 0 || parameters == nil {
		return nil
	}

	for key, value := range *parameters {
		paramdef := findParamDefByName(paramdefs, key)

		if nestedMap, ok := value.(map[string]interface{}); ok {
			return FillParamDefFromJson(paramdef.Children, &nestedMap)
		} else if nestedMapArray, ok := value.([]map[string]interface{}); ok {
			jsonByte, err := json.Marshal(nestedMapArray)

			if err != nil {
				paramdef.DefaultValue = string(jsonByte)
			}
		} else if value != nil {
			jsonByte, err := json.Marshal(value)

			if err != nil {
				paramdef.DefaultValue = string(jsonByte)
			}
		}
	}

	return nil
}
