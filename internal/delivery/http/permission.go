package http

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/serializer"
	"github.com/openinfradev/tks-api/internal/usecase"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
)

type IPermissionHandler interface {
	GetPermissionTemplates(w http.ResponseWriter, r *http.Request)
	GetPermissionsByRoleId(w http.ResponseWriter, r *http.Request)
	UpdatePermissionsByRoleId(w http.ResponseWriter, r *http.Request)
}

type PermissionHandler struct {
	permissionUsecase usecase.IPermissionUsecase
}

func NewPermissionHandler(usecase usecase.Usecase) *PermissionHandler {
	return &PermissionHandler{
		permissionUsecase: usecase.Permission,
	}
}

// GetPermissionTemplates godoc
//
//	@Tags			Permission
//	@Summary		Get Permission Templates
//	@Description	Get Permission Templates
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	domain.PermissionSetResponse
//	@Router			/permissions/templates [get]
//	@Security		JWT
func (h PermissionHandler) GetPermissionTemplates(w http.ResponseWriter, r *http.Request) {
	permissionSet := model.NewDefaultPermissionSet()

	var premissionSetResponse domain.PermissionSetResponse
	if err := serializer.Map(permissionSet, &premissionSetResponse); err != nil {
		log.Info(r.Context(), err)
	}

	var out domain.GetPermissionTemplatesResponse
	out.Permissions = append(out.Permissions, premissionSetResponse.Dashboard)
	out.Permissions = append(out.Permissions, premissionSetResponse.Stack)
	out.Permissions = append(out.Permissions, premissionSetResponse.SecurityPolicy)
	out.Permissions = append(out.Permissions, premissionSetResponse.ProjectManagement)
	out.Permissions = append(out.Permissions, premissionSetResponse.Notification)
	out.Permissions = append(out.Permissions, premissionSetResponse.Configuration)

	ResponseJSON(w, r, http.StatusOK, out)
}

// GetPermissionsByRoleId godoc
//
//	@Tags			Permission
//	@Summary		Get Permissions By Role ID
//	@Description	Get Permissions By Role ID
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	domain.PermissionSetResponse
//	@Router			/organizations/{organizationId}/roles/{roleId}/permissions [get]
//	@Security		JWT
func (h PermissionHandler) GetPermissionsByRoleId(w http.ResponseWriter, r *http.Request) {
	// path parameter
	var roleId string

	vars := mux.Vars(r)
	if v, ok := vars["roleId"]; !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(nil, "", ""))
		return
	} else {
		roleId = v
	}

	permissionSet, err := h.permissionUsecase.GetPermissionSetByRoleId(r.Context(), roleId)
	if err != nil {
		ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
		return
	}

	var premissionSetResponse domain.PermissionSetResponse
	if err := serializer.Map(permissionSet, &premissionSetResponse); err != nil {
		log.Info(r.Context(), err)
	}

	var out domain.GetPermissionsByRoleIdResponse
	out.Permissions = append(out.Permissions, premissionSetResponse.Dashboard)
	out.Permissions = append(out.Permissions, premissionSetResponse.Stack)
	out.Permissions = append(out.Permissions, premissionSetResponse.SecurityPolicy)
	out.Permissions = append(out.Permissions, premissionSetResponse.ProjectManagement)
	out.Permissions = append(out.Permissions, premissionSetResponse.Notification)
	out.Permissions = append(out.Permissions, premissionSetResponse.Configuration)

	ResponseJSON(w, r, http.StatusOK, out)
}

// UpdatePermissionsByRoleId godoc
//
//	@Tags			Permission
//	@Summary		Update Permissions By Role ID
//	@Description	Update Permissions By Role ID
//	@Accept			json
//	@Produce		json
//	@Param			roleId	path	string									true	"Role ID"
//	@Param			body	body	domain.UpdatePermissionsByRoleIdRequest	true	"Update Permissions By Role ID Request"
//	@Success		200
//	@Router			/organizations/{organizationId}/roles/{roleId}/permissions [put]
//	@Security		JWT
func (h PermissionHandler) UpdatePermissionsByRoleId(w http.ResponseWriter, r *http.Request) {
	// path parameter
	var roleId string
	_ = roleId
	vars := mux.Vars(r)
	if v, ok := vars["roleId"]; !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(nil, "", ""))
		return
	} else {
		roleId = v
	}

	// request
	input := domain.UpdatePermissionsByRoleIdRequest{}
	err := UnmarshalRequestInput(r, &input)
	if err != nil {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
		return
	}

	for _, permissionResponse := range input.Permissions {
		var permission model.Permission
		if err := serializer.Map(permissionResponse, &permission); err != nil {
			log.Info(r.Context(), err)
		}

		if err := h.permissionUsecase.UpdatePermission(r.Context(), &permission); err != nil {
			ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
			return
		}
	}

	ResponseJSON(w, r, http.StatusOK, nil)
}
