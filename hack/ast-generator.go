package main

//func main() {
//	fset := token.NewFileSet()
//
//	node, err := parser.ParseFile(fset, "./internal/delivery/api/endpoint.go", nil, parser.ParseComments)
//	if err != nil {
//		log.Fatalf("파싱 오류: %v", err)
//	}
//
//	ast.Print(fset, node)
//
//	// write ast to file
//	f, err := os.Create("ast.txt")
//	if err != nil {
//		log.Fatalf("파일 생성 오류: %v", err)
//	}
//	defer f.Close()
//	ast.Fprint(f, fset, node, nil)
//
//}
