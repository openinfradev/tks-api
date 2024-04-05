//go:build ignore

package main

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"html/template"
	"io/ioutil"
	"log"
	"strings"
)

const endpointFilePath = "./internal/delivery/api/endpoint.go"

type endpointDecl struct {
	Name  string
	Group string
}

const indexTemplateStr = ` // This is generated code. DO NOT EDIT.

package api

`

//const endpointTemplateStr = `// Comment below is special purpose for code generation.
//// Do not edit this comment.
//// Endpoint for Code Generation
//const (
//{{- range .}}
//	{{.Name}} Endpoint = iota
//{{- end}}
//)
//`

const apiMapTemplateStr = `var MapWithEndpoint = map[Endpoint]EndpointInfo{
{{- range .}}
    {{.Name}}: {
		Name: "{{.Name}}", 
		Group: "{{.Group}}",
	},
{{- end}}
}
`

const restCodeTemplateStr = `var MapWithName = reverseApiMap()

func reverseApiMap() map[string]Endpoint {
	m := make(map[string]Endpoint)
	for k, v := range MapWithEndpoint {
		m[v.Name] = k
	}
	return m
}

func (e Endpoint) String() string {
	return MapWithEndpoint[e].Name
}

func GetEndpoint(name string) Endpoint {
	return MapWithName[name]
}

`

//
//const stringFunctionTemplateStr = `func (e Endpoint) String() string {
//	switch e {
//{{- range .}}
//	case {{.Name}}:
//		return "{{.Name}}"
//{{- end}}
//	default:
//		return ""
//	}
//}
//`
//
//const getEndpointFunctionTemplateStr = `func GetEndpoint(name string) Endpoint {
//	switch name {
//{{- range .}}
//	case "{{.Name}}":
//		return {{.Name}}
//{{- end}}
//	default:
//		return -1
//	}
//}
//`

func main() {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, endpointFilePath, nil, parser.ParseComments)
	if err != nil {
		log.Fatalf("failed to parse file: %v", err)
	}

	var endpoints []endpointDecl
	var currentGroup string

	// AST를 탐색합니다.
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.GenDecl:
			if x.Tok == token.CONST {
				if x.Doc != nil {
					for _, comment := range x.Doc.List {
						if strings.Contains(comment.Text, "Endpoint for Code Generation") {
							continue
						}
						if strings.HasPrefix(comment.Text, "//") {
							currentGroup = strings.TrimSpace(strings.TrimPrefix(comment.Text, "//"))
						}
					}
				}
				for _, spec := range x.Specs {
					vs, ok := spec.(*ast.ValueSpec)
					if !ok {
						continue
					}
					if vs.Doc != nil {
						for _, comment := range vs.Doc.List {
							if strings.HasPrefix(comment.Text, "//") {
								currentGroup = strings.TrimSpace(strings.TrimPrefix(comment.Text, "//"))
							}
						}
					}

					for _, name := range vs.Names {
						endpoints = append(endpoints, endpointDecl{
							Name:  name.Name,
							Group: currentGroup,
						})
					}
				}
			}
		}
		return true
	})

	for _, ep := range endpoints {
		log.Printf("Endpoint: %s, Group: %s\n", ep.Name, ep.Group)
	}

	// contents for index
	indexTemplate := template.New("index")
	indexTemplate, err = indexTemplate.Parse(indexTemplateStr)
	if err != nil {
		log.Fatalf("failed to parse template: %v", err)
	}
	var indexCode bytes.Buffer
	if err := indexTemplate.Execute(&indexCode, endpoints); err != nil {
		log.Fatalf("failed to execute template: %v", err)
	}

	//// contents for endpoint
	//endpointTemplate := template.New("endpoint")
	//endpointTemplate, err = endpointTemplate.Parse(endpointTemplateStr)
	//if err != nil {
	//	log.Fatalf("failed to parse template: %v", err)
	//}
	//var endpointCode bytes.Buffer
	//if err := endpointTemplate.Execute(&endpointCode, endpoints); err != nil {
	//	log.Fatalf("failed to execute template: %v", err)
	//}

	// contents for apiMap
	apiMapTemplate := template.New("apiMap")
	apiMapTemplate, err = apiMapTemplate.Parse(apiMapTemplateStr)
	if err != nil {
		log.Fatalf("failed to parse template: %v", err)
	}
	var apiMapCode bytes.Buffer
	if err := apiMapTemplate.Execute(&apiMapCode, endpoints); err != nil {
		log.Fatalf("failed to execute template: %v", err)
	}

	restCodeTemplate := template.New("restCode")
	restCodeTemplate, err = restCodeTemplate.Parse(restCodeTemplateStr)
	if err != nil {
		log.Fatalf("failed to parse template: %v", err)
	}
	var restCode bytes.Buffer
	if err := restCodeTemplate.Execute(&restCode, nil); err != nil {
		log.Fatalf("failed to execute template: %v", err)
	}
	//
	//// contents for stringFunction
	//stringFunctionTemplate := template.New("stringFunction")
	//stringFunctionTemplate, err = stringFunctionTemplate.Parse(stringFunctionTemplateStr)
	//if err != nil {
	//	log.Fatalf("failed to parse template: %v", err)
	//}
	//var stringFunctionCode bytes.Buffer
	//if err := stringFunctionTemplate.Execute(&stringFunctionCode, endpoints); err != nil {
	//	log.Fatalf("failed to execute template: %v", err)
	//}
	//
	//// contents for getEndpointFunction
	//getEndpointFunctionTemplate := template.New("getEndpointFunction")
	//getEndpointFunctionTemplate, err = getEndpointFunctionTemplate.Parse(getEndpointFunctionTemplateStr)
	//if err != nil {
	//	log.Fatalf("failed to parse template: %v", err)
	//}
	//var getEndpointFunctionCode bytes.Buffer
	//if err := getEndpointFunctionTemplate.Execute(&getEndpointFunctionCode, endpoints); err != nil {
	//	log.Fatalf("failed to execute template: %v", err)
	//}

	// replace original file(endpointFilePath) with new contents
	//contents := indexCode.String() + endpointCode.String() + apiMapCode.String() + stringFunctionCode.String() + getEndpointFunctionCode.String()
	//contents := indexCode.String() + apiMapCode.String() + stringFunctionCode.String() + getEndpointFunctionCode.String()
	contents := indexCode.String() + apiMapCode.String() + restCode.String()
	newFilePath := strings.Replace(endpointFilePath, "endpoint", "generated_endpoints.go", 1)

	if err := ioutil.WriteFile(newFilePath, []byte(contents), 0644); err != nil {
		log.Fatalf("failed to write file: %v", err)
	}

	log.Println("Code generation is done.")
}
