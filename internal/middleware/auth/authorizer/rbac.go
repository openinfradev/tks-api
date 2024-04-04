package authorizer

import (
	"github.com/openinfradev/tks-api/internal/repository"
	"net/http"
)

func RBACFilterWithEndpoint(handler http.Handler, repo repository.Repository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//requestEndpointInfo, ok := request.EndpointFrom(r.Context())
		//if !ok {
		//	internalHttp.ErrorJSON(w, r, httpErrors.NewInternalServerError(fmt.Errorf("endpoint not found"), "", ""))
		//	return
		//}
		//
		//requestUserInfo, ok := request.UserFrom(r.Context())
		//if !ok {
		//	internalHttp.ErrorJSON(w, r, httpErrors.NewInternalServerError(fmt.Errorf("user not found"), "", ""))
		//	return
		//}
		//
		//// if requestEndpointInfo.String() is one of ProjectEndpoints, print true
		//if internalApi.ApiMap[requestEndpointInfo].Group == "Project" && requestEndpointInfo != internalApi.CreateProject {
		//	log.Infof("RBACFilterWithEndpoint: %s is ProjectEndpoint", requestEndpointInfo.String())
		//	vars := mux.Vars(r)
		//	var projectRole string
		//	if projectId, ok := vars["projectId"]; ok {
		//		projectRole = requestUserInfo.GetRoleProjectMapping()[projectId]
		//	} else {
		//		log.Warn("RBACFilterWithEndpoint: projectId not found. Passing through unsafely.")
		//	}
		//	if !internalRole.IsRoleAllowed(requestEndpointInfo, internalRole.StrToRole(projectRole)) {
		//		internalHttp.ErrorJSON(w, r, httpErrors.NewForbiddenError(fmt.Errorf("permission denied"), "", ""))
		//		return
		//
		//	}
		//}

		handler.ServeHTTP(w, r)
	})
}
