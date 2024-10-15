package http

import (
	"context"
	"net/http"

	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/usecase"
	"github.com/openinfradev/tks-api/pkg/domain"
)

type IPermissionHandler interface {
	GetPermissionTemplates(w http.ResponseWriter, r *http.Request)
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
	if permission.IsAllowed != nil {
		permissionResponse.IsAllowed = permission.IsAllowed
	}

	for _, child := range permission.Children {
		permissionResponse.Children = append(permissionResponse.Children, convertModelToPermissionTemplateResponse(ctx, child))
	}

	return &permissionResponse
}
