package usecase

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/open-policy-agent/opa/ast"
	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-api/pkg/domain"
)

type IUtilityUsecase interface {
	RegoCompile(request *domain.RegoCompileRequest, parseParameter bool) (response *domain.RegoCompileResponse, err error)
}

type UtilityUsecase struct {
}

func NewUtilityUsecase(r repository.Repository) IUtilityUsecase {
	return &UtilityUsecase{}
}

func (u *UtilityUsecase) RegoCompile(request *domain.RegoCompileRequest, parseParameter bool) (response *domain.RegoCompileResponse, err error) {
	modules := map[string]*ast.Module{}

	response = &domain.RegoCompileResponse{}
	response.Errors = []domain.RegoCompieError{}

	mod, err := ast.ParseModuleWithOpts("rego", request.Rego, ast.ParserOptions{})
	if err != nil {
		return nil, err
	}
	modules["rego"] = mod

	compiler := ast.NewCompiler()
	compiler.Compile(modules)

	if compiler.Failed() {
		for _, compileError := range compiler.Errors {
			response.Errors = append(response.Errors, domain.RegoCompieError{
				Status:  400,
				Code:    "P_INVALID_REGO_SYNTAX",
				Message: "Invalid rego syntax",
				Text: fmt.Sprintf("[%d:%d] %s",
					compileError.Location.Row, compileError.Location.Col,
					compileError.Message),
			})
		}
	}

	if parseParameter {
		response.ParametersSchema = extractParameter(request.Rego)
	}

	return response, nil
}

func extractParameter(rego string) []*domain.ParameterDef {
	pattern := `input\.parameters\.[\w\.\[\]]+`

	prefix := "input.parameters."

	// Compile the regex pattern
	regex := regexp.MustCompile(pattern)

	matches := regex.FindAllString(rego, -1)

	defStore := NewParamDefStore()

	for _, match := range matches {
		remainder := match[len(prefix):]

		// 문법 변환: aa["a"]["B"][_]->aa.a.B[_]
		regex := regexp.MustCompile(`\[\"(\w+)\"\]`)
		remainder = regex.ReplaceAllString(remainder, ".$1")

		params := strings.Split(remainder, ".")

		if len(params) == 0 {
			continue
		}

		defStore.AddDefinition(params)
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
		isLast := i == len(params)-1

		key := findKey(s.store, param)

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
	finalType := "any"

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
	}

	return newDef
}
