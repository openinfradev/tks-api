package policytemplate

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/open-policy-agent/opa/ast"
	"github.com/openinfradev/tks-api/pkg/domain"
	"golang.org/x/exp/maps"
)

const (
	input_param_prefix    = "input.parameters"
	input_extract_pattern = `input(\.parameters|\[\"parameters\"\])((\[\"[\w\-]+\"\])|(\[_\])|(\.\w+))*` //(\.\w+)*` //  (\.\w+\[\"\w+\"\])|(\.\w+\[\w+\])|(\.\w+))*`
	//	input_extract_pattern = `input\.parameters((\[\".+\"\])?(\.\w+\[\"\w+\"\])|(\.\w+\[\w+\])|(\.\w+))+`
	obj_get_pattern    = `object\.get\((input|input\.parameters|input\.parameters\.[^,]+)\, \"*([^,\"]+)\"*, [^\)]+\)`
	package_name_regex = `package ([\w\.]+)[\n\r]+`
	import_regex       = `import ([\w\.]+)[\n\r]+`
)

var (
	input_extract_regex = regexp.MustCompile(input_extract_pattern)
	obj_get_regex       = regexp.MustCompile(obj_get_pattern)
)

func extractInputExprFromModule(module *ast.Module) []string {
	rules := module.Rules

	passedInputMap := []string{}

	violationRule := []*ast.Rule{}
	nonViolatonRule := []*ast.Rule{}

	for _, rule := range rules {
		if rule.Head.Name == "violation" {
			violationRule = append(violationRule, rule)
		} else {
			nonViolatonRule = append(nonViolatonRule, rule)
		}
	}

	paramRefs := map[string]string{}

	for _, rule := range violationRule {
		processRule(rule, paramRefs, passedInputMap, nonViolatonRule)
	}

	// 중복제거를 위해 사용한 맵을 소팅하기 위해 키 리스트로 변환
	paramRefsList := maps.Keys(paramRefs)

	/* 가장 depth가 깊은 변수부터 우선 처리해서 최초 trverse 시 child가 있는지 효율적으로 확인하기 위해 역순으로 정렬
	이러한 소팅을 통해서 input.parameters.labels[_] 보다는 input.parameters.labels[_]의 자식들이 리스트의 앞순에 위치하도록 함
	input.parameters.labels[_]가 먼저 오면 input.parameters.labels[_]의 자식들이 존재 여부는 이 후 리스트 목록에 따라 달라지지만
	역순으로 소팅한 상태에서는 항상 자식들이 리스트 앞에 위치

	input.parameters.labels[_].key
	input.parameters.labels[_].allowedRegex
	input.parameters.labels[_]
	*/
	sort.Sort(sort.Reverse(sort.StringSlice(paramRefsList)))

	return paramRefsList
}

func processRule(rule *ast.Rule, paramRefs map[string]string, passedParams []string, nonViolatonRule []*ast.Rule) {
	exprs := rule.Body
	localAssignMap := map[string]string{}

	for i, param := range passedParams {
		if isSubstitutionRequired(param) {
			argName := rule.Head.Args[i].String()
			localAssignMap[argName] = param
		}
	}

	for _, expr := range exprs {
		exprString := expr.String()

		exprString = substituteWithLocalAssignMap(localAssignMap, exprString)
		exprString = replaceAllObjectGet(exprString)
		exprString = substituteWithLocalAssignMap(localAssignMap, exprString)

		matches := input_extract_regex.FindAllString(exprString, -1)

		if len(matches) > 0 {
			for _, match := range matches {
				paramRefs[match] = "1"
			}
		}

		updateLocalAssignMap(expr, localAssignMap)

		if expr.IsCall() {
			call, _ := expr.Terms.([]*ast.Term)
			if len(call) > 2 {
				ruleName := call[0].String()

				args := call[1:]

				inputPassed, passingParams := processingInputArgs(args, localAssignMap)

				if inputPassed {
					for _, nvrule := range nonViolatonRule {
						if ruleName == nvrule.Head.Name.String() {

							processRule(nvrule, paramRefs, passingParams, nonViolatonRule)
						}
					}
				}
			}
		}

		ast.WalkTerms(expr, func(t *ast.Term) bool {
			switch t.Value.(type) {
			case ast.Ref: // 인자가 없는 정책을 단순히 호출하는 경우, 호출하는 정책을 따라가 봄
				{
					ruleRef := ([]*ast.Term)(t.Value.(ast.Ref))
					ruleName := ruleRef[0].Value.String()

					for _, nvrule := range nonViolatonRule {
						if ruleName == nvrule.Head.Name.String() {
							processRule(nvrule, paramRefs, []string{}, nonViolatonRule)
						}
					}
				}
			case ast.Call: // 인자가 있는 정책 호출, input 치환되는 인자가 있으면 passingParams으로 전달
				call := ([]*ast.Term)(t.Value.(ast.Call))

				// 인자가 없으므로 처리 불필요
				if len(call) < 2 {
					return false
				}

				ruleName := call[0].String()

				args := call[1:]

				inputPassed, passingParams := processingInputArgs(args, localAssignMap)

				if inputPassed {
					for _, nvrule := range nonViolatonRule {
						if ruleName == nvrule.Head.Name.String() {
							processRule(nvrule, paramRefs, passingParams, nonViolatonRule)
						}
					}
				}

				return false
			default:
				for _, nvrule := range nonViolatonRule {
					ruleName := nvrule.Head.Name.String()
					if t.Value.String() == ruleName {
						processRule(nvrule, paramRefs, []string{}, nonViolatonRule)
					}
				}

				return false
			}

			return false
		})
	}
}

// object.get(object.get(input, "parameters", {}), "exemptImages", [])) -> input.parameters.exemptImages와 같은 패턴 변환
func replaceAllObjectGet(expr string) string {
	if !strings.Contains(expr, "object.get") {
		return expr
	}

	result := obj_get_regex.ReplaceAllString(expr, "$1.$2")

	if result == expr {
		return expr
	}

	return replaceAllObjectGet(result)
}

func processingInputArgs(args []*ast.Term, localAssignMap map[string]string) (bool, []string) {
	inputPassed := false
	passingParams := []string{}

	for i := 0; i < len(args); i++ {
		if args[i] != nil {
			arg := args[i].String()
			arg = substituteWithLocalAssignMap(localAssignMap, arg)
			arg = replaceAllObjectGet(arg)
			arg = substituteWithLocalAssignMap(localAssignMap, arg)

			if isSubstitutionRequired(arg) {
				passingParams = append(passingParams, arg)
				inputPassed = true
			}
			passingParams = append(passingParams, "")
		}
	}
	return inputPassed, passingParams
}

func updateLocalAssignMap(expr *ast.Expr, localAssignMap map[string]string) {
	if expr.IsAssignment() {
		vars := expr.Operand(0).Vars()
		assigned := replaceAllObjectGet(expr.Operand(1).String())

		if len(vars) == 1 && isSubstitutionRequired(assigned) {
			localAssignMap[expr.Operand(0).String()] = assigned
		}
	}
}

func substituteWithLocalAssignMap(localAssignMap map[string]string, exprString string) string {
	for k, v := range localAssignMap {
		//pattern := `([^\w\"])` + "parameters" + `([^\w\"])`
		// pattern := `([ \[\]\(\)])` + k + `([ \[\]\(\)\.])`

		if strings.Contains(exprString, v) {
			continue
		} else if exprString == k {
			exprString = v
		} else {
			pattern := `(\W)` + k + `(\W)`

			rx := regexp.MustCompile(pattern)

			exprString = rx.ReplaceAllString(exprString, `${1}`+v+`${2}`)
		}

	}
	return exprString
}

func isSubstitutionRequired(expr string) bool {
	trimmed := strings.TrimSpace(expr)

	// input.review 등 input.parameters가 아닌 input은 해석할 필요 없음
	// input이 assign 되었으면 input.review가 될지 parameter가 될지 모르므로 일단 input으로 대체 필요
	// input.parameters 자체를 변수로 assign하는 패턴도 처리 필요
	// input.parameters 하위의 속성도 처리 필요
	return trimmed == "input" ||
		trimmed == "input.parameters" ||
		strings.HasPrefix(trimmed, "input.parameters.")
}

func ExtractParameter(modules map[string]*ast.Module) []*domain.ParameterDef {
	defStore := NewParamDefStore()

	for _, module := range modules {
		inputExprs := extractInputExprFromModule(module)

		for _, inputExpr := range inputExprs {
			remainder := inputExpr[len(input_param_prefix):]

			// 문법 변환: aa["a"]["B"][_]->aa.a.B[_]
			regex := regexp.MustCompile(`\[\"([\w-]+)\"\]`)
			remainder = regex.ReplaceAllString(remainder, ".${1}")

			params := strings.Split(remainder, ".")

			if len(params) == 0 {
				continue
			}

			defStore.AddDefinition(params)
		}

	}

	return defStore.store
}

func MergeRegoAndLibs(rego string, libs []string) string {
	if len(libs) == 0 {
		return rego
	}

	var re = regexp.MustCompile(import_regex)
	var re2 = regexp.MustCompile(package_name_regex)

	result := re.ReplaceAllString(rego, "")

	for _, lib := range libs {
		result += re2.ReplaceAllString(lib, "")
	}

	return result
}

type ParamDefStore struct {
	store []*domain.ParameterDef
}

func NewParamDefStore() *ParamDefStore {
	return &ParamDefStore{store: []*domain.ParameterDef{}}
}

func (s *ParamDefStore) GetStore() []*domain.ParameterDef {
	return s.store
}

func (s *ParamDefStore) AddDefinition(params []string) {
	init := &s.store

	for i, param := range params {
		if param == "" {
			continue
		}

		isLast := i == len(params)-1

		key := findKey(*init, param)

		if key == nil {
			key = createKey(param, isLast)

			*init = append(*init, key)
		}

		init = &key.Children
	}
}

func findKey(defs []*domain.ParameterDef, key string) *domain.ParameterDef {
	for _, def := range defs {
		if def.Key == key || def.Key+"[_]" == key {
			return def
		}
	}

	return nil
}

func createKey(key string, isLast bool) *domain.ParameterDef {
	var finalType string

	pKey := key
	isArray := false

	if strings.HasSuffix(pKey, "[_]") {
		pKey, _ = strings.CutSuffix(pKey, "[_]")
		isArray = true
	}

	if isLast {
		if isArray {
			finalType = "any[]"
		} else {
			finalType = "any"
		}
	} else {
		if isArray {
			finalType = "object[]"
		} else {
			finalType = "object"
		}
	}

	newDef := &domain.ParameterDef{
		Key:      pKey,
		Type:     finalType,
		Children: []*domain.ParameterDef{},
		IsArray:  isArray,
	}

	return newDef
}

func CompileRegoWithLibs(rego string, libs []string) (compiler *ast.Compiler, err error) {
	modules := map[string]*ast.Module{}

	regoPackage := GetPackageFromRegoCode(rego)

	regoModule, err := ast.ParseModuleWithOpts(regoPackage, rego, ast.ParserOptions{})
	if err != nil {
		return nil, err
	}

	modules[regoPackage] = regoModule

	for i, lib := range libs {
		// Lib이 공백이면 무시
		if len(strings.TrimSpace(lib)) == 0 {
			continue
		}

		libPackage := GetPackageFromRegoCode(lib)

		// Lib의 패키지 명이 공백이면 rego에서 import 될 수 없기 때문에 에러 처리
		// 패키지 명이 Parse할 때 비어있으면 에러가 나지만, rego인지 lib인지 정확히 알기 어려울 수 있으므로 알려 줌
		if len(strings.TrimSpace(libPackage)) == 0 {
			return nil, fmt.Errorf("lib[%d] is not valid, empty package name", i)
		}

		libModule, err := ast.ParseModuleWithOpts(libPackage, lib, ast.ParserOptions{})
		if err != nil {
			return nil, err
		}

		modules[libPackage] = libModule
	}

	compiler = ast.NewCompiler()
	compiler.Compile(modules)

	return compiler, nil
}

func MergeAndCompileRegoWithLibs(rego string, libs []string) (modules map[string]*ast.Module, err error) {
	modules = map[string]*ast.Module{}

	regoPackage := GetPackageFromRegoCode(rego)

	merged := MergeRegoAndLibs(rego, libs)

	module, err := ast.ParseModuleWithOpts(regoPackage, merged, ast.ParserOptions{})
	if err != nil {
		return modules, err
	}

	modules[regoPackage] = module

	compiler := ast.NewCompiler()
	compiler.Compile(modules)

	return modules, nil
}

func GetPackageFromRegoCode(regoCode string) string {
	packageRegex := regexp.MustCompile(package_name_regex)

	match := packageRegex.FindStringSubmatch(regoCode)

	if len(match) > 1 {
		return match[1]
	}

	return ""
}
