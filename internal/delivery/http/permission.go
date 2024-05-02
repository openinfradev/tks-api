package http

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"net/http"

	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/usecase"
	"github.com/openinfradev/tks-api/pkg/domain"
)

type IPermissionHandler interface {
	GetPermissionTemplates(w http.ResponseWriter, r *http.Request)
	GetEndpoints(w http.ResponseWriter, r *http.Request)
}

type PermissionHandler struct {
	permissionUsecase usecase.IPermissionUsecase
	userUsecase       usecase.IUserUsecase
}

func NewPermissionHandler(usecase usecase.Usecase) IPermissionHandler {
	return &PermissionHandler{
		permissionUsecase: usecase.Permission,
		userUsecase:       usecase.User,
	}
}

// GetPermissionTemplates godoc
//
//	@Tags			Permission
//	@Summary		Get Permission Templates
//	@Description	Get Permission Templates
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	domain.GetPermissionTemplatesResponse
//	@Router			/permissions/templates [get]
//	@Security		JWT
func (h PermissionHandler) GetPermissionTemplates(w http.ResponseWriter, r *http.Request) {
	permissionSet := model.NewDefaultPermissionSet()

	var out domain.GetPermissionTemplatesResponse
	out.Permissions = make([]*domain.TemplateResponse, 0)

	out.Permissions = append(out.Permissions, convertModelToPermissionTemplateResponse(r.Context(), permissionSet.Dashboard))
	out.Permissions = append(out.Permissions, convertModelToPermissionTemplateResponse(r.Context(), permissionSet.Stack))
	out.Permissions = append(out.Permissions, convertModelToPermissionTemplateResponse(r.Context(), permissionSet.Policy))
	out.Permissions = append(out.Permissions, convertModelToPermissionTemplateResponse(r.Context(), permissionSet.ProjectManagement))
	out.Permissions = append(out.Permissions, convertModelToPermissionTemplateResponse(r.Context(), permissionSet.Notification))
	out.Permissions = append(out.Permissions, convertModelToPermissionTemplateResponse(r.Context(), permissionSet.Configuration))

	ResponseJSON(w, r, http.StatusOK, out)
}

func convertModelToPermissionTemplateResponse(ctx context.Context, permission *model.Permission) *domain.TemplateResponse {
	var permissionResponse domain.TemplateResponse

	permissionResponse.Key = permission.Key
	permissionResponse.Name = permission.Name
	permissionResponse.EdgeKey = permission.EdgeKey

	for _, child := range permission.Children {
		permissionResponse.Children = append(permissionResponse.Children, convertModelToPermissionTemplateResponse(ctx, child))
	}

	return &permissionResponse
}

// GetEndpoints godoc
//
//	@Tags			Permission
//	@Summary		Get Endpoints
//	@Description	Get Endpoints
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	domain.GetEndpointsResponse
//	@Router			/permissions/{permissionId}/endpoints [get]
//	@Security		JWT
func (h PermissionHandler) GetEndpoints(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	permissionId, ok := vars["permissionId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("permissionId not found"), "PE_INVALID_PERMISSIONID", "permissionId not found"))
		return
	}

	permissionUuid, err := uuid.Parse(permissionId)
	if err != nil {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("permissionId is invalid"), "PE_INVALID_PERMISSIONID", "permissionId is invalid"))
		return
	}

	endpoints, err := h.permissionUsecase.GetEndpointsByPermissionId(r.Context(), permissionUuid)
	if err != nil {
		ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "PE_GET_ENDPOINTS_FAILED", "Failed to get endpoints"))
		return
	}

	var out domain.GetEndpointsResponse
	out.Endpoints = make([]domain.EndpointResponse, 0)
	for _, endpoint := range endpoints {
		out.Endpoints = append(out.Endpoints, convertEndpointToDomain(endpoint))
	}

	ResponseJSON(w, r, http.StatusOK, out)
}
