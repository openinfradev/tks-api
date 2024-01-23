package api

//
//import (
//	"github.com/gorilla/mux"
//	"net/http"
//)
//
//type IApi interface {
//	//SetPath()
//	API() Api
//	//SetMethod()
//	//GetMethod() string
//	//SetHandler()
//	//GetHandler() http.Handler
//	//
//	//RegisterApi(router *mux.Router)
//}
//type Api struct {
//	Path    string
//	Method  string
//	Handler http.Handler
//}
//
//func (a Api) GetPath() string {
//	return a.Path
//}
//
//func (a Api) GetMethod() string {
//	return a.Method
//}
//
//func (a Api) GetHandler() http.Handler {
//	return a.Handler
//}
//
//func (a Api) SetBasePath(path string) {
//	a.Path = path
//}
//
//func (a Api) SetMethod(method string) {
//	a.Method = method
//}
//
//func (a Api) SetHandler(handler http.Handler) {
//	a.Handler = handler
//}
//
//func (a Api) RegisterApi(router *mux.Router) {
//	router.Handle(a.GetPath(), a.GetHandler())
//}
