package policytemplate

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/format"
	"github.com/open-policy-agent/opa/types"
	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/pkg/domain"
	"golang.org/x/exp/maps"
)

const (
	rego_var_name_pattern = `[[a-zA-Z_][a-zA-Z0-9_]*`
	input_param_prefix    = "input.parameters"
	input_extract_pattern = `input(\.parameters|\[\"parameters\"\])((\[\"[\w\-]+\"\])|(\[_\])|(\[` + rego_var_name_pattern + `\])|(\.\w+))*` //(\.\w+)*` //  (\.\w+\[\"\w+\"\])|(\.\w+\[\w+\])|(\.\w+))*`
	//	input_extract_pattern = `input\.parameters((\[\".+\"\])?(\.\w+\[\"\w+\"\])|(\.\w+\[\w+\])|(\.\w+))+`
	obj_get_list_pattern = `object\.get\((input|input\.parameters|input\.parameters\.[^,]+), \"([^\"]+)\", \[\]\)`
	obj_get_pattern      = `object\.get\((input|input\.parameters|input\.parameters\.[^,]+), \"([^\"]+)\", [^\)]+\)`
	ref_by_key_pattern   = `\[\"([\w-]+)\"\]`
	package_name_regex   = `package ([\w\.]+)[\n\r]+`
	lib_import_regex     = `import data\.lib\.([\w\.]+)[\n\r]+`
)

var (
	input_extract_regex           = regexp.MustCompile(input_extract_pattern)
	obj_get_list_regex            = regexp.MustCompile(obj_get_list_pattern)
	obj_get_regex                 = regexp.MustCompile(obj_get_pattern)
	ref_by_key_regex              = regexp.MustCompile(ref_by_key_pattern)
	ref_by_var_regex              = regexp.MustCompile(`\[` + rego_var_name_pattern + `\]`)
	array_param_map, capabilities = buildArrayParameterMap()
)

// OPA 내장 함수 중 array 인자를 가진 함수와 array 인자의 위치를 담은 자료구조를 생성한다.
// // OPA 내장 함수 목록은 정책에 상관없으므로 처음 로딩될 때 한 번만 호출해 변수에 담아두고 사용하면 된다.
// OPA 엔진 버전에 따라 달라진다. 0.62 버전 기준으로 이 함수의 결과 값은 다음과 같다.
// map[all:[true] any:[true] array.concat:[true true] array.reverse:[true] array.slice:[true false false] concat:[false true] count:[true] glob.match:[false true false] graph.reachable:[false true] graph.reachable_paths:[false true] internal.print:[true] json.filter:[false true] json.patch:[false true] json.remove:[false true] max:[true] min:[true] net.cidr_contains_matches:[true true] net.cidr_merge:[true] object.filter:[false true] object.remove:[false true] object.subset:[true true] object.union_n:[true] product:[true] sort:[true] sprintf:[false true] strings.any_prefix_match:[true true] strings.any_suffix_match:[true true] sum:[true] time.clock:[true] time.date:[true] time.diff:[true true] time.format:[true] time.weekday:[true]]
func buildArrayParameterMap() (map[string][]bool, *ast.Capabilities) {
	compiler := ast.NewCompiler()

	// 아주 단순한 rego 코드를 컴파일해도 컴파일러의 모든 Built-In 함수 정보를 컴파일 할 수 있음
	mod, err := ast.ParseModuleWithOpts("hello", "package hello\n hello {input.message = \"world\"}", ast.ParserOptions{})
	if err != nil {
		return nil, nil
	}

	// 컴파일을 수행해야 Built-in 함수 정보 로딩할 수 있음
	modules := map[string]*ast.Module{}
	modules["hello"] = mod
	compiler.Compile(modules)

	var external_data = &ast.Builtin{
		Name: "external_data",
		Decl: types.NewFunction(
			types.Args(
				types.Named("a", types.NewObject(
					[]*types.StaticProperty{
						types.NewStaticProperty("provider", types.S),
						types.NewStaticProperty("keys", types.NewArray([]types.Type{types.S}, types.A)),
					},
					nil,
				)),
			),
			types.Named("output", types.A),
		), // TODO(sr): types.A?  ^^^^^^^ (also below)
	}

	capabilities := compiler.Capabilities()
	capabilities.Builtins = append(capabilities.Builtins, external_data)

	return getArrayParameterMap(compiler), capabilities
}

func getArrayParameterMap(compiler *ast.Compiler) map[string][]bool {
	capabilities := compiler.Capabilities()

	var result = map[string][]bool{}

	if capabilities != nil {
		for _, builtin := range capabilities.Builtins {
			args := builtin.Decl.FuncArgs().Args
			isArrayParam := make([]bool, len(args))

			arrayCount := 0
			for i, typeVal := range args {
				isArrayParam[i] = IsArray(typeVal)

				if isArrayParam[i] {
					arrayCount += 1
				}
			}

			if arrayCount > 0 {
				result[builtin.Name] = isArrayParam
			}
		}
	}

	return result
}

func IsArray(t types.Type) bool {
	switch specific_type := t.(type) {
	case *types.Array:
		return true
	case types.Any:
		{
			for _, anyType := range specific_type {
				if IsArray(anyType) {
					return true
				}
			}
		}
	}

	return false
}

func extractInputExprFromModule(module *ast.Module) []string {
	rules := module.Rules

	passedInputMap := []string{}
	globalAssignMap := map[string]string{}

	violationRule := []*ast.Rule{}
	nonViolatonRule := []*ast.Rule{}

	for _, rule := range rules {
		if rule.Head.Name == "violation" {
			violationRule = append(violationRule, rule)
		} else {
			nonViolatonRule = append(nonViolatonRule, rule)

			if rule.Head.Assign && rule.Head.Value != nil {
				if _, ok := rule.Head.Value.Value.(ast.Call); !ok {
					globalAssignMap[string(rule.Head.Name)] = rule.Head.Value.String()
				}
			}
		}
	}

	paramRefs := map[string]string{}

	for _, rule := range violationRule {
		processRule(rule, globalAssignMap, paramRefs, passedInputMap, nonViolatonRule)
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

func processRule(rule *ast.Rule, globalAssignMap map[string]string, paramRefs map[string]string, passedParams []string, nonViolatonRule []*ast.Rule) map[string]string {
	localAssignMap := map[string]string{}

	// 규칙이 단순 assign이면 value의 정책 호출을 따라가 봐야 함
	if rule.Head.Assign {
		if call, ok := rule.Head.Value.Value.(ast.Call); ok {
			ruleName := call[0].String()

			args := call[1:]

			argStrs := make([]string, len(args))

			for i, arg := range args {
				argStrs[i] = arg.String()
			}

			for _, nvrule := range nonViolatonRule {
				if ruleName == nvrule.Head.Name.String() {
					return processRule(nvrule, globalAssignMap, paramRefs, argStrs, nonViolatonRule)
				}
			}
		} else {
			value := rule.Head.Value.String()

			if isSubstitutionRequired(value) {
				paramRefs[value] = "1"

				localAssignMap[string(rule.Head.Name)] = value

				// fmt.Println("1818181818", rule.Head.Value)
				return localAssignMap
			}
		}

		// 더 처리할 건 없음
		return nil
	}

	exprs := rule.Body

	for i, param := range passedParams {
		if isSubstitutionRequired(param) {
			argName := rule.Head.Args[i].String()
			localAssignMap[argName] = param
		}
	}

	for _, expr := range exprs {
		exprString := expr.String()

		if len(localAssignMap) > 0 {
			exprString = substituteWithAssignMap(localAssignMap, exprString)
			exprString = replaceAllObjectGet(exprString)
			exprString = substituteWithAssignMap(localAssignMap, exprString)
		}

		if len(globalAssignMap) > 0 {
			exprString = substituteWithAssignMap(globalAssignMap, exprString)
			exprString = replaceAllObjectGet(exprString)
			exprString = substituteWithAssignMap(globalAssignMap, exprString)
		}

		matches := input_extract_regex.FindAllString(exprString, -1)

		if len(matches) > 0 {
			for _, match := range matches {
				paramRefs[match] = "1"
			}
		}

		updateLocalAssignMap(expr, localAssignMap)

		if expr.IsCall() {
			call, _ := expr.Terms.([]*ast.Term)
			if len(call) > 1 {
				ruleName := call[0].String()
				args := call[1:]

				inputPassed, passingParams := processingInputArgs(args, localAssignMap)

				if inputPassed {
					if is_arrays, ok := array_param_map[ruleName]; ok {
						for i, passingParam := range passingParams {
							is_array := is_arrays[i]

							if is_array && strings.HasPrefix(passingParam, "input.parameters.") &&
								!strings.HasSuffix(passingParam, "[_]") {

								paramRefs[passingParam+"[_]"] = "1"
							}
						}
					}

					for _, nvrule := range nonViolatonRule {
						if ruleName == nvrule.Head.Name.String() {
							updateLocals := processRule(nvrule, globalAssignMap, paramRefs, passingParams, nonViolatonRule)
							for k, v := range updateLocals {
								if _, ok := localAssignMap[k]; !ok {
									localAssignMap[k] = v
								}
							}
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
							updateLocals := processRule(nvrule, globalAssignMap, paramRefs, []string{}, nonViolatonRule)

							for k, v := range updateLocals {
								if _, ok := localAssignMap[k]; !ok {
									localAssignMap[k] = v
								}
							}
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
							updateLocals := processRule(nvrule, globalAssignMap, paramRefs, passingParams, nonViolatonRule)

							for k, v := range updateLocals {
								if _, ok := localAssignMap[k]; !ok {
									localAssignMap[k] = v
								}
							}
						}
					}
				}

				return false
			default:
				for _, nvrule := range nonViolatonRule {
					ruleName := nvrule.Head.Name.String()
					if t.Value.String() == ruleName {
						updateLocals := processRule(nvrule, globalAssignMap, paramRefs, []string{}, nonViolatonRule)

						for k, v := range updateLocals {
							if _, ok := localAssignMap[k]; !ok {
								localAssignMap[k] = v
							}
						}
					}
				}

				return false
			}

			return false
		})
	}

	headKey := rule.Head.Key

	if headKey != nil {
		for _, nvrule := range nonViolatonRule {
			ruleName := nvrule.Head.Name.String()
			if strings.Contains(headKey.String(), ruleName) {
				updateLocals := processRule(nvrule, globalAssignMap, paramRefs, []string{}, nonViolatonRule)

				for k, v := range updateLocals {
					if _, ok := localAssignMap[k]; !ok {
						localAssignMap[k] = v
					}
				}
			}
		}
	}

	return nil
}

// object.get(object.get(input, "parameters", {}), "exemptImages", [])) -> input.parameters.exemptImages와 같은 패턴 변환
func replaceAllObjectGet(expr string) string {
	if !strings.Contains(expr, "object.get") {
		return expr
	}

	result := obj_get_list_regex.ReplaceAllString(expr, "$1.$2"+`[_]`)
	result = obj_get_regex.ReplaceAllString(result, "$1.$2")

	// 정규식의 영향 없음 그냥 리턴
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

			if len(localAssignMap) > 0 {
				arg = substituteWithAssignMap(localAssignMap, arg)
				arg = replaceAllObjectGet(arg)
				arg = substituteWithAssignMap(localAssignMap, arg)
			}

			if isSubstitutionRequired(arg) {
				passingParams = append(passingParams, arg)
				inputPassed = true
			} else {
				passingParams = append(passingParams, "")
			}
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

func substituteWithAssignMap(assignMap map[string]string, exprString string) string {
	for k, v := range assignMap {
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
			remainder = ref_by_key_regex.ReplaceAllString(remainder, ".${1}")

			// 문법 변환2: input.parameters.aa[varname]에서 aa는 객체가 아닌 배열이라고 확신할 수 있으므로
			// input.parameters.aa[_]로 변환해도 무방함
			remainder = ref_by_var_regex.ReplaceAllString(remainder, "[_]")

			params := strings.Split(remainder, ".")

			if len(params) == 0 {
				continue
			}

			defStore.AddDefinition(params)
		}

	}

	return defStore.store
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

func cutTrailingArrayNote(val string) string {
	cut, found := strings.CutSuffix(val, "[_]")

	if found {
		return cutTrailingArrayNote(cut)
	}

	return val
}

func createKey(key string, isLast bool) *domain.ParameterDef {
	var finalType string

	pKey := key
	isArray := false

	if strings.HasSuffix(pKey, "[_]") {
		pKey = cutTrailingArrayNote(pKey)
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

func processLibs(libs []string) []string {
	// libs 에 --- 로 코딩되어 여러 개가 한 번에 들어온 경우 분할
	newLibs := []string{}
	for _, lib := range libs {
		newLibs = append(newLibs, strings.Split(stripCarriageReturn(lib), model.FILE_DELIMETER)...)
	}

	return newLibs
}

func CompileRegoWithLibs(rego string, libs []string) (compiler *ast.Compiler, err error) {
	modules := map[string]*ast.Module{}

	regoPackage := GetPackageFromRegoCode(rego)

	regoModule, err := ast.ParseModuleWithOpts(regoPackage, rego, ast.ParserOptions{Capabilities: capabilities})
	if err != nil {
		return nil, err
	}

	modules[regoPackage] = regoModule

	for i, lib := range processLibs(libs) {
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

		libModule, err := ast.ParseModuleWithOpts(libPackage, lib, ast.ParserOptions{Capabilities: capabilities})
		if err != nil {
			return nil, err
		}

		modules[libPackage] = libModule
	}

	compiler = ast.NewCompiler().WithCapabilities(capabilities)

	compiler.Compile(modules)

	return compiler, nil
}

func MergeRegoAndLibs(rego string, libs []string) string {
	if len(libs) == 0 {
		return rego
	}

	var re = regexp.MustCompile(lib_import_regex)
	var re2 = regexp.MustCompile(package_name_regex)

	// data.lib import 모두 제거
	result := re.ReplaceAllString(rego, "")

	lib_imports := re.FindAllStringSubmatch(rego, -1)

	// data.lib import 시 data.lib.<라이브러리 이름>인 경우 소스에서 <라이브러리 이름>.<정책 이름>으로 참조됨
	// 이 경우 임프로를 제거했으므로 <라이브러리 이름>.<정책 이름>을 <정책 이름>으로 바꾸기 위해 제거
	// <라이브러리 이름>.<정책 이름>.<규칙 이름> 등의 형식은 <규칙 이름>으로 참조되기 때문에 처리할 필요 없음
	for _, lib_import := range lib_imports {
		lib_and_rule := lib_import[1]

		if !strings.Contains(lib_and_rule, ".") {
			remove_lib_prefix := lib_and_rule + "."
			fmt.Printf("'%s'\n", remove_lib_prefix)
			result = strings.ReplaceAll(result, remove_lib_prefix, "")
		}
	}
	// "<라이브러리 이름>." 제거 로직 끝

	for _, lib := range processLibs(libs) {
		result += re2.ReplaceAllString(lib, "")
	}

	return result
}

func MergeAndCompileRegoWithLibs(rego string, libs []string) (modules map[string]*ast.Module, err error) {
	modules = map[string]*ast.Module{}

	regoPackage := GetPackageFromRegoCode(rego)

	merged := MergeRegoAndLibs(rego, libs)

	module, err := ast.ParseModuleWithOpts(regoPackage, merged, ast.ParserOptions{Capabilities: capabilities})
	if err != nil {
		return modules, err
	}

	modules[regoPackage] = module

	compiler := ast.NewCompiler().WithCapabilities(capabilities)

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

func FormatRegoCode(rego string) string {
	packageName := GetPackageFromRegoCode(rego)

	// 패키지 명을 파싱할 수 없으면 포맷팅할 수 있는 코드가 아닐 것이므로 그냥 리턴
	if packageName == "" {
		return rego
	}

	bytes, err := format.Source("rego", []byte(rego))

	if err != nil {
		return rego
	}

	return strings.Replace(string(bytes), "\t", "  ", -1)
}

func FormatLibCode(libs []string) []string {
	processedLibs := processLibs(libs)

	result := make([]string, len(processedLibs))

	for i, lib := range processedLibs {
		result[i] = FormatRegoCode(lib)
	}

	return result
}
