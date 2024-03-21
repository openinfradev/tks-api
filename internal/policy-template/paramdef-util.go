package policytemplate

import (
	"strings"

	"github.com/openinfradev/tks-api/pkg/domain"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

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
		_:
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
