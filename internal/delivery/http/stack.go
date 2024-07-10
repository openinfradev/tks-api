package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/internal/serializer"
	"github.com/openinfradev/tks-api/internal/usecase"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
)

type StackHandler struct {
	usecase           usecase.IStackUsecase
	usecasePolicy     usecase.IPolicyUsecase
	usecaseUser       usecase.IUserUsecase
	usecasePermission usecase.IPermissionUsecase
}

func NewStackHandler(h usecase.Usecase) *StackHandler {
	return &StackHandler{
		usecase:           h.Stack,
		usecasePolicy:     h.Policy,
		usecaseUser:       h.User,
		usecasePermission: h.Permission,
	}
}

// CreateStack godoc
//
//	@Tags			Stacks
//	@Summary		Create Stack
//	@Description	Create Stack
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string						true	"organizationId"
//	@Param			body			body		domain.CreateStackRequest	true	"create cloud setting request"
//	@Success		200				{object}	domain.CreateStackResponse
//	@Router			/organizations/{organizationId}/stacks [post]
//	@Security		JWT
func (h *StackHandler) CreateStack(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid organizationId"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	input := domain.CreateStackRequest{}
	err := UnmarshalRequestInput(r, &input)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var dto model.Stack
	if err = serializer.Map(r.Context(), input, &dto); err != nil {
		log.Info(r.Context(), err)
	}
	if err = serializer.Map(r.Context(), input, &dto.Conf); err != nil {
		log.Info(r.Context(), err)
	}

	dto.Domains = make([]model.ClusterDomain, 6)
	dto.Domains[0] = model.ClusterDomain{
		DomainType: "grafana",
		Url:        input.Domain.Grafana,
	}
	dto.Domains[1] = model.ClusterDomain{
		DomainType: "loki",
		Url:        input.Domain.Loki,
	}
	dto.Domains[2] = model.ClusterDomain{
		DomainType: "minio",
		Url:        input.Domain.Minio,
	}
	dto.Domains[3] = model.ClusterDomain{
		DomainType: "thanos_sidecar",
		Url:        input.Domain.ThanosSidecar,
	}
	dto.Domains[4] = model.ClusterDomain{
		DomainType: "jaeger",
		Url:        input.Domain.Jaeger,
	}
	dto.Domains[5] = model.ClusterDomain{
		DomainType: "kiali",
		Url:        input.Domain.Kiali,
	}

	dto.OrganizationId = organizationId
	stackId, err := h.usecase.Create(r.Context(), dto)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	// Sync ClusterAdmin Permission to Keycloak
	// First get all users in the organization
	users, err := h.usecaseUser.List(r.Context(), organizationId)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}
	err = h.syncKeycloakWithClusterAdminPermission(r.Context(), organizationId, []string{stackId.String()}, *users)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	out := domain.CreateStackResponse{
		ID: stackId.String(),
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// InstallStack godoc
//
//	@Tags			Stacks
//	@Summary		Install Stack ( BYOH )
//	@Description	Install Stack ( BYOH )
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"organizationId"
//	@Param			stackId			path		string	true	"stackId"
//	@Success		200				{object}	nil
//	@Router			/organizations/{organizationId}/stacks/{stackId}/install [post]
//	@Security		JWT
func (h *StackHandler) InstallStack(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	stackId, ok := vars["stackId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid stackId"), "S_INVALID_STACK_ID", ""))
		return
	}

	err := h.usecase.Install(r.Context(), domain.StackId(stackId))
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	ResponseJSON(w, r, http.StatusOK, nil)
}

// GetStacks godoc
//
//	@Tags			Stacks
//	@Summary		Get Stacks
//	@Description	Get Stacks
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"organizationId"
//	@Param			limit			query		string	false	"pageSize"
//	@Param			page			query		string	false	"pageNumber"
//	@Param			soertColumn		query		string	false	"sortColumn"
//	@Param			sortOrder		query		string	false	"sortOrder"
//	@Success		200				{object}	domain.GetStacksResponse
//	@Router			/organizations/{organizationId}/stacks [get]
//	@Security		JWT
func (h *StackHandler) GetStacks(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid organizationId"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	urlParams := r.URL.Query()
	pg := pagination.NewPagination(&urlParams)
	stacks, err := h.usecase.Fetch(r.Context(), organizationId, pg)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var out domain.GetStacksResponse
	out.Stacks = make([]domain.StackResponse, len(stacks))
	for i, stack := range stacks {
		if err := serializer.Map(r.Context(), stack, &out.Stacks[i]); err != nil {
			log.Info(r.Context(), err)
		}

		if err := serializer.Map(r.Context(), stack.CreatedAt, &out.Stacks[i].CreatedAt); err != nil {
			log.Info(r.Context(), err)
		}

		err = json.Unmarshal(stack.StackTemplate.Services, &out.Stacks[i].StackTemplate.Services)
		if err != nil {
			log.Info(r.Context(), err)
		}
	}

	if out.Pagination, err = pg.Response(r.Context()); err != nil {
		log.Info(r.Context(), err)
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// GetStack godoc
//
//	@Tags			Stacks
//	@Summary		Get Stack
//	@Description	Get Stack
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"organizationId"
//	@Param			stackId			path		string	true	"stackId"
//	@Success		200				{object}	domain.GetStackResponse
//	@Router			/organizations/{organizationId}/stacks/{stackId} [get]
//	@Security		JWT
func (h *StackHandler) GetStack(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	strId, ok := vars["stackId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid stackId"), "C_INVALID_STACK_ID", ""))
		return
	}

	stack, err := h.usecase.Get(r.Context(), domain.StackId(strId))
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var out domain.GetStackResponse
	if err := serializer.Map(r.Context(), stack, &out.Stack); err != nil {
		log.Info(r.Context(), err)
	}
	for _, domain := range stack.Domains {
		switch domain.DomainType {
		case "grafana":
			out.Stack.Domain.Grafana = domain.Url
		case "loki":
			out.Stack.Domain.Loki = domain.Url
		case "minio":
			out.Stack.Domain.Minio = domain.Url
		case "thanos_sidecar":
			out.Stack.Domain.ThanosSidecar = domain.Url
		case "jaeger":
			out.Stack.Domain.Jaeger = domain.Url
		case "kiali":
			out.Stack.Domain.Kiali = domain.Url
		}
	}

	err = json.Unmarshal(stack.StackTemplate.Services, &out.Stack.StackTemplate.Services)
	if err != nil {
		log.Info(r.Context(), err)
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// GetStackStatus godoc
//
//	@Tags			Stacks
//	@Summary		Get Stack Status
//	@Description	Get Stack Status
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"organizationId"
//	@Param			stackId			path		string	true	"stackId"
//	@Success		200				{object}	domain.GetStackStatusResponse
//	@Router			/organizations/{organizationId}/stacks/{stackId}/status [get]
//	@Security		JWT
func (h *StackHandler) GetStackStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	strId, ok := vars["stackId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid stackId"), "C_INVALID_STACK_ID", ""))
		return
	}

	steps, status, err := h.usecase.GetStepStatus(r.Context(), domain.StackId(strId))
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var out domain.GetStackStatusResponse
	out.StepStatus = make([]domain.StackStepStatus, len(steps))
	for i, step := range steps {
		if err := serializer.Map(r.Context(), step, &out.StepStatus[i]); err != nil {
			log.Info(r.Context(), err)
		}
	}
	out.StackStatus = status

	ResponseJSON(w, r, http.StatusOK, out)
}

// UpdateStack godoc
//
//	@Tags			Stacks
//	@Summary		Update Stack
//	@Description	Update Stack
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string						true	"organizationId"
//	@Param			stackId			path		string						true	"stackId"
//	@Param			body			body		domain.UpdateStackRequest	true	"Update cloud setting request"
//	@Success		200				{object}	nil
//	@Router			/organizations/{organizationId}/stacks/{stackId} [put]
//	@Security		JWT
func (h *StackHandler) UpdateStack(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	strId, ok := vars["stackId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid stackId"), "C_INVALID_STACK_ID", ""))
		return
	}
	stackId := domain.StackId(strId)
	if !stackId.Validate() {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid stackId"), "C_INVALID_STACK_ID", ""))
		return
	}

	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid organizationId"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	input := domain.UpdateStackRequest{}
	err := UnmarshalRequestInput(r, &input)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var dto model.Stack
	if err = serializer.Map(r.Context(), input, &dto); err != nil {
		log.Info(r.Context(), err)
	}
	dto.ID = stackId
	dto.OrganizationId = organizationId

	err = h.usecase.Update(r.Context(), dto)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	ResponseJSON(w, r, http.StatusOK, nil)

}

// DeleteStack godoc
//
//	@Tags			Stacks
//	@Summary		Delete Stack
//	@Description	Delete Stack
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"organizationId"
//	@Param			stackId			path		string	true	"stackId"
//	@Success		200				{object}	nil
//	@Router			/organizations/{organizationId}/stacks/{stackId} [delete]
//	@Security		JWT
func (h *StackHandler) DeleteStack(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	strId, ok := vars["stackId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid stackId"), "C_INVALID_STACK_ID", ""))
		return
	}
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid organizationId"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	var dto model.Stack
	dto.ID = domain.StackId(strId)
	dto.OrganizationId = organizationId

	// Delete Policies
	policyIds, err := h.usecasePolicy.GetPolicyIDsByClusterID(r.Context(), domain.ClusterId(dto.ID))
	if err != nil {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "S_FAILED_DELETE_POLICIES", ""))
		return
	}

	if policyIds != nil && len(*policyIds) > 0 {
		err = h.usecasePolicy.DeletePoliciesForClusterID(r.Context(), organizationId, domain.ClusterId(dto.ID), *policyIds)
		if err != nil {
			ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "S_FAILED_DELETE_POLICIES", ""))
			return
		}
	}

	err = h.usecase.Delete(r.Context(), dto)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	ResponseJSON(w, r, http.StatusOK, nil)
}

// CheckStackName godoc
//
//	@Tags			Stacks
//	@Summary		Check name for stack
//	@Description	Check name for stack
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"organizationId"
//	@Param			stackId			path		string	true	"stackId"
//	@Param			name			path		string	true	"name"
//	@Success		200				{object}	nil
//	@Router			/organizations/{organizationId}/stacks/name/{name}/existence [GET]
//	@Security		JWT
func (h *StackHandler) CheckStackName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid organizationId"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}
	name, ok := vars["name"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid name"), "S_INVALID_STACK_NAME", ""))
		return
	}

	exist := true
	_, err := h.usecase.GetByName(r.Context(), organizationId, name)
	if err != nil {
		if _, code := httpErrors.ErrorResponse(err); code == http.StatusNotFound {
			exist = false
		} else {
			ErrorJSON(w, r, err)
			return
		}
	}

	var out domain.CheckStackNameResponse
	out.Existed = exist

	ResponseJSON(w, r, http.StatusOK, out)
}

// GetStackKubeconfig godoc
//
//	@Tags			Stacks
//	@Summary		Get Kubeconfig by stack
//	@Description	Get Kubeconfig by stack
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"organizationId"
//	@Param			stackId			path		string	true	"organizationId"
//	@Success		200				{object}	domain.GetStackKubeconfigResponse
//	@Router			/organizations/{organizationId}/stacks/{stackId}/kube-config [get]
//	@Security		JWT
func (h *StackHandler) GetStackKubeconfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	_, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid organizationId"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	strId, ok := vars["stackId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid organizationId"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}
	stackId := domain.StackId(strId)
	if !stackId.Validate() {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid stackId"), "C_INVALID_STACK_ID", ""))
		return
	}

	kubeconfig, err := h.usecase.GetKubeconfig(r.Context(), domain.StackId(strId))
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var out = domain.GetStackKubeconfigResponse{
		Kubeconfig: kubeconfig,
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// SetFavorite godoc
//
//	@Tags			Stacks
//	@Summary		Set favorite stack
//	@Description	Set favorite stack
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"organizationId"
//	@Param			stackId			path		string	true	"stackId"
//	@Success		200				{object}	nil
//	@Router			/organizations/{organizationId}/stacks/{stackId}/favorite [post]
//	@Security		JWT
func (h *StackHandler) SetFavorite(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	strId, ok := vars["stackId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid stackId"), "C_INVALID_STACK_ID", ""))
		return
	}

	err := h.usecase.SetFavorite(r.Context(), domain.StackId(strId))
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}
	ResponseJSON(w, r, http.StatusOK, nil)
}

// DeleteFavorite godoc
//
//	@Tags			Stacks
//	@Summary		Delete favorite stack
//	@Description	Delete favorite stack
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"organizationId"
//	@Param			stackId			path		string	true	"stackId"
//	@Success		200				{object}	nil
//	@Router			/organizations/{organizationId}/stacks/{stackId}/favorite [delete]
//	@Security		JWT
func (h *StackHandler) DeleteFavorite(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	strId, ok := vars["stackId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid stackId"), "C_INVALID_STACK_ID", ""))
		return
	}

	err := h.usecase.DeleteFavorite(r.Context(), domain.StackId(strId))
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}
	ResponseJSON(w, r, http.StatusOK, nil)
}

// syncKeycloakWithClusterAdminPermission sync the permissions with Keycloak
// 1. Get all roles assigned to the user and merge the permissions
// 2. Then get the cluster admin permissions for the stack
// 3. Finally, sync the permissions with Keycloak
func (h StackHandler) syncKeycloakWithClusterAdminPermission(ctx context.Context, organizationId string, clusterIds []string, users []model.User) error {
	for _, user := range users {
		// 1-step
		// Merge the permissions of the userUuid
		var permissionSets []*model.PermissionSet
		for _, role := range user.Roles {
			permissionSet, err := h.usecasePermission.GetPermissionSetByRoleId(ctx, role.ID)
			if err != nil {
				return err
			}
			permissionSets = append(permissionSets, permissionSet)
		}
		mergedPermissionSet := h.usecasePermission.MergePermissionWithOrOperator(ctx, permissionSets...)

		// 2-step
		// Then get the cluster admin permissions for the stack
		var targetEdgePermissions []*model.Permission

		var targetPermission *model.Permission
		for _, permission := range mergedPermissionSet.Stack.Children {
			if permission.Key == model.MiddleClusterAccessControlKey {
				targetPermission = permission
			}
		}
		edgePermissions := model.GetEdgePermission(targetPermission, targetEdgePermissions, nil)

		// 3-step
		//  sync the permissions with Keycloak
		for _, clusterId := range clusterIds {
			if len(edgePermissions) > 0 {
				var err error
				for _, edgePermission := range edgePermissions {
					switch edgePermission.Key {
					case model.OperationCreate:
						err = h.usecasePermission.SyncKeycloakWithClusterAdminPermission(ctx, organizationId,
							clusterId+"-k8s-api", user.ID.String(), "cluster-admin-create", *edgePermission.IsAllowed)
					case model.OperationRead:
						err = h.usecasePermission.SyncKeycloakWithClusterAdminPermission(ctx, organizationId,
							clusterId+"-k8s-api", user.ID.String(), "cluster-admin-read", *edgePermission.IsAllowed)
					case model.OperationUpdate:
						err = h.usecasePermission.SyncKeycloakWithClusterAdminPermission(ctx, organizationId,
							clusterId+"-k8s-api", user.ID.String(), "cluster-admin-update", *edgePermission.IsAllowed)
					case model.OperationDelete:
						err = h.usecasePermission.SyncKeycloakWithClusterAdminPermission(ctx, organizationId,
							clusterId+"-k8s-api", user.ID.String(), "cluster-admin-delete", *edgePermission.IsAllowed)
					}
					if err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}
