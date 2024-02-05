package http

import (
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/openinfradev/tks-api/internal/usecase"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
)

type IProjectHandler interface {
	CreateProject(w http.ResponseWriter, r *http.Request)
	GetProjectRole(w http.ResponseWriter, r *http.Request)
	GetProjectRoles(w http.ResponseWriter, r *http.Request)
	UpdateProject(w http.ResponseWriter, r *http.Request)
	DeleteProject(w http.ResponseWriter, r *http.Request)
	GetProject(w http.ResponseWriter, r *http.Request)
	GetProjects(w http.ResponseWriter, r *http.Request)

	AddProjectMember(w http.ResponseWriter, r *http.Request)
	GetProjectMember(w http.ResponseWriter, r *http.Request)
	GetProjectMembers(w http.ResponseWriter, r *http.Request)
	RemoveProjectMember(w http.ResponseWriter, r *http.Request)
	RemoveProjectMembers(w http.ResponseWriter, r *http.Request)
	UpdateProjectMemberRole(w http.ResponseWriter, r *http.Request)

	CreateProjectNamespace(w http.ResponseWriter, r *http.Request)
	IsProjectNamespaceExist(w http.ResponseWriter, r *http.Request)
	GetProjectNamespaces(w http.ResponseWriter, r *http.Request)
	GetProjectNamespace(w http.ResponseWriter, r *http.Request)
	DeleteProjectNamespace(w http.ResponseWriter, r *http.Request)

	SetFavoriteProject(w http.ResponseWriter, r *http.Request)
	SetFavoriteProjectNamespace(w http.ResponseWriter, r *http.Request)
	UnSetFavoriteProject(w http.ResponseWriter, r *http.Request)
	UnSetFavoriteProjectNamespace(w http.ResponseWriter, r *http.Request)
}

type ProjectHandler struct {
	usecase usecase.IProjectUsecase
}

func NewProjectHandler(u usecase.IProjectUsecase) IProjectHandler {
	return &ProjectHandler{
		usecase: u,
	}
}

// CreateProject godoc
// @Tags        Projects
// @Summary     Create new project
// @Description Create new project
// @Accept      json
// @Produce     json
// @Param       organizationId path     string                      true "Organization ID"
// @Param       request        body     domain.CreateProjectRequest true "Request body to create project"
// @Success     200            {object} domain.CreateProjectResponse
// @Router      /organizations/{organizationId}/projects [post]
// @Security    JWT
func (p ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	log.Debugf("organizationId = [%v]\n", organizationId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"),
			"C_INVALID_ORGANIZATION_ID", ""))
		return
	}
	var projectReq domain.CreateProjectRequest
	if err := UnmarshalRequestInput(r, &projectReq); err != nil {
		ErrorJSON(w, r, err)
		return
	}

	log.Infof("projectReq: name = %s, description = %s, projectLeaderId = %s",
		projectReq.Name, projectReq.Description, projectReq.ProjectLeaderId)

	now := time.Now()
	project := &domain.Project{
		OrganizationId: organizationId,
		Name:           projectReq.Name,
		Description:    projectReq.Description,
		CreatedAt:      now,
	}
	log.Infof("Processing CREATE request for project '%s'...", project.Name)

	projectId, err := p.usecase.CreateProject(project)
	if err != nil {
		ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
		return
	}

	log.Infof("Project Id: %s", projectId)

	project.ID = projectId
	ProjectLeaderId, err := uuid.Parse(projectReq.ProjectLeaderId)
	if err != nil {
		log.Error(err)
		ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", "Failed to parse uuid to string"))
		return
	}
	pm := &domain.ProjectMember{
		ProjectId:     projectId,
		ProjectUserId: ProjectLeaderId,
		ProjectRoleId: projectReq.ProjectRoleId,
		CreatedAt:     now,
	}

	projectMemberId, err := p.usecase.AddProjectMember(pm)
	if err != nil {
		ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
		return
	}

	pr, err := p.usecase.GetProjectRoles(usecase.ProjectLeader)
	if err != nil {
		ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
		return
	}

	pms := make([]domain.ProjectMember, 0)
	pm.ID = projectMemberId
	pm.ProjectRole = pr[0]
	pms = append(pms, *pm)

	project.ProjectMembers = pms
	projectRes := domain.CreateProjectResponse{
		Project: *project,
	}

	ResponseJSON(w, r, http.StatusOK, projectRes)

}

// GetProjects godoc
// @Tags        Projects
// @Summary     Get projects
// @Description Get projects
// @Accept      json
// @Produce     json
// @Param       organizationId     path     string true "Organization ID"
// @Success     200                {object} domain.GetProjectsResponse
// @Router      /organizations/{organizationId}/projects [get]
// @Security    JWT
func (p ProjectHandler) GetProjects(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"),
			"C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	ps, err := p.usecase.GetProjects(organizationId)
	if err != nil {
		log.ErrorWithContext(r.Context(), "Failed to retrieve projects ", err)
		ErrorJSON(w, r, err)
		return
	}

	var out domain.GetProjectsResponse
	out.Projects = ps

	if ps == nil {
		ResponseJSON(w, r, http.StatusNotFound, out)
	} else {
		ResponseJSON(w, r, http.StatusOK, out)
	}
}

func (p ProjectHandler) GetProject(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
}

func (p ProjectHandler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
}

func (p ProjectHandler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
}

// GetProjectRole godoc
// @Tags        Projects
// @Summary     Get project role
// @Description Get project role by id
// @Accept      json
// @Produce     json
// @Param       organizationId path     string true  "Organization ID"
// @Param       projectId      path     string true "Project ID"
// @Success     200            {object} domain.GetProjectRoleResponse
// @Router      /organizations/{organizationId}/projects/project-roles/{projectRoleId} [get]
// @Security    JWT
func (p ProjectHandler) GetProjectRole(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	log.Debugf("organizationId = [%v]\n", organizationId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"),
			"C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	projectRoleId, ok := vars["projectRoleId"]
	log.Debugf("projectRoleId = [%v]\n", projectRoleId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid projectRoleId"),
			"C_INVALID_PROJECT_ROLE_ID", ""))
		return
	}

	pr, err := p.usecase.GetProjectRole(projectRoleId)
	if err != nil {
		log.ErrorWithContext(r.Context(), "Failed to get project roles ", err)
		ErrorJSON(w, r, err)
		return
	}

	var out domain.GetProjectRoleResponse
	out.ProjectRole = *pr

	ResponseJSON(w, r, http.StatusOK, out)
}

// GetProjectRoles godoc
// @Tags        Projects
// @Summary     Get project roles
// @Description Get project roles by giving params
// @Accept      json
// @Produce     json
// @Param       organizationId     path     string true "Organization ID"
// @Param       query          query    string false "project role search by query (query=all), (query=leader), (query=member), (query=viewer)"
// @Success     200            {object} domain.GetProjectRolesResponse
// @Router      /organizations/{organizationId}/projects/project-roles [get]
// @Security    JWT
func (p ProjectHandler) GetProjectRoles(w http.ResponseWriter, r *http.Request) {
	urlParams := r.URL.Query()

	queryParam := urlParams.Get("query")
	query := usecase.ProjectAll
	if queryParam == "" || strings.EqualFold(queryParam, "all") {
		query = usecase.ProjectAll
	} else if strings.EqualFold(queryParam, "leader") {
		query = usecase.ProjectLeader
	} else if strings.EqualFold(queryParam, "member") {
		query = usecase.ProjectMember
	} else if strings.EqualFold(queryParam, "viewer") {
		query = usecase.ProjectViewer
	} else {
		log.ErrorWithContext(r.Context(), "Invalid query params. Err: ")
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid query params"),
			"C_INVALID_QUERY_PARAM", ""))
		return
	}

	prs, err := p.usecase.GetProjectRoles(query)
	if err != nil {
		log.ErrorWithContext(r.Context(), "Failed to get project roles ", err)
		ErrorJSON(w, r, err)
		return
	}

	var out domain.GetProjectRolesResponse
	out.ProjectRoles = prs

	ResponseJSON(w, r, http.StatusOK, out)
}

// AddProjectMember godoc
// @Tags        Projects
// @Summary     Add project member to project
// @Description Add project member to project
// @Accept      json
// @Produce     json
// @Param       organizationId path     string                         true "Organization ID"
// @Param       projectId      path     string                         true "Project ID"
// @Param       request        body     domain.AddProjectMemberRequest true "Request body to add project member"
// @Success     200            {object} domain.AddProjectMemberResponse
// @Router      /organizations/{organizationId}/projects/{projectId}/members [post]
// @Security    JWT
func (p ProjectHandler) AddProjectMember(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	log.Debugf("organizationId = [%v]\n", organizationId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"),
			"C_INVALID_ORGANIZATION_ID", ""))
		return
	}
	projectId, ok := vars["projectId"]
	log.Debugf("projectId = [%v]\n", projectId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid projectId"),
			"C_INVALID_PROJECT_ID", ""))
		return
	}

	var projectMemberReq domain.AddProjectMemberRequest
	if err := UnmarshalRequestInput(r, &projectMemberReq); err != nil {
		ErrorJSON(w, r, err)
		return
	}

	projectMemberResponse := domain.AddProjectMemberResponse{
		ProjectMembers: make([]domain.ProjectMember, 0),
	}
	now := time.Now()
	for _, pmr := range projectMemberReq.ProjectMemberRequests {
		ProjectUserId, err := uuid.Parse(pmr.ProjectUserId)
		if err != nil {
			log.Error(err)
			ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", "Failed to parse uuid to string"))
			return
		}

		pm := &domain.ProjectMember{
			ProjectId:     projectId,
			ProjectUserId: ProjectUserId,
			ProjectRoleId: pmr.ProjectRoleId,
			CreatedAt:     now,
		}
		pmId, err := p.usecase.AddProjectMember(pm)
		if err != nil {
			ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
			return
		}

		pr, err := p.usecase.GetProjectRole(pm.ProjectRoleId)
		if err != nil {
			ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
			return
		}
		pm.ID = pmId
		pm.ProjectRole = *pr
		projectMemberResponse.ProjectMembers = append(projectMemberResponse.ProjectMembers, *pm)
	}

	ResponseJSON(w, r, http.StatusOK, projectMemberResponse)
}

// GetProjectMember godoc
// @Tags        Projects
// @Summary     Get project member
// @Description Get project member
// @Accept      json
// @Produce     json
// @Param       organizationId path     string true "Organization ID"
// @Param       projectId          path     string true "Project ID"
// @Param       projectMemberId path     string true "Project Member ID"
// @Success     200             {object} domain.GetProjectMemberResponse
// @Router      /organizations/{organizationId}/projects/{projectId}/members/{projectMemberId} [get]
// @Security    JWT
func (p ProjectHandler) GetProjectMember(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	log.Debugf("organizationId = [%v]\n", organizationId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"),
			"C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	projectId, ok := vars["projectId"]
	log.Debugf("projectId = [%v]\n", projectId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid projectId"),
			"C_INVALID_PROJECT_ID", ""))
		return
	}

	projectMemberId, ok := vars["projectMemberId"]
	log.Debugf("projectMemberId = [%v]\n", projectMemberId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid projectMemberId"),
			"C_INVALID_PROJECT_MEMBER_ID", ""))
		return
	}

	pm, err := p.usecase.GetProjectMemberById(projectMemberId)
	if err != nil {
		log.ErrorWithContext(r.Context(), "Failed to get project member ", err)
		ErrorJSON(w, r, err)
		return
	}

	out := domain.GetProjectMemberResponse{ProjectMember: pm}
	if pm == nil {
		ResponseJSON(w, r, http.StatusNotFound, out)
		return
	}
	ResponseJSON(w, r, http.StatusOK, out)
}

// GetProjectMembers godoc
// @Tags        Projects
// @Summary     Get project members
// @Description Get project members
// @Accept      json
// @Produce     json
// @Param       organizationId path     string true "Organization ID"
// @Param       projectId      path     string true "Project ID"
// @Success     200            {object} domain.GetProjectMembersResponse
// @Router      /organizations/{organizationId}/projects/{projectId}/members [get]
// @Security    JWT
func (p ProjectHandler) GetProjectMembers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	log.Debugf("organizationId = [%v]\n", organizationId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"),
			"C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	projectId, ok := vars["projectId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid projectId"),
			"C_INVALID_PROJECT_ID", ""))
		return
	}

	pms, err := p.usecase.GetProjectMembersByProjectId(projectId)
	if err != nil {
		log.ErrorWithContext(r.Context(), "Failed to get project roles ", err)
		ErrorJSON(w, r, err)
		return
	}

	out := domain.GetProjectMembersResponse{ProjectMembers: pms}

	ResponseJSON(w, r, http.StatusOK, out)
}

// RemoveProjectMember godoc
// @Tags        Projects
// @Summary     Remove project members to project
// @Description Remove project members to project
// @Accept      json
// @Produce     json
// @Param       organizationId  path     string true "Organization ID"
// @Param       projectId       path     string true "Project ID"
// @Param       projectMemberId path     string true "Project Member ID"
// @Success     200             {object} domain.CommonProjectResponse
// @Router      /organizations/{organizationId}/projects/{projectId}/members/{projectMemberId} [delete]
// @Security    JWT
func (p ProjectHandler) RemoveProjectMember(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	log.Debugf("organizationId = [%v]\n", organizationId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"),
			"C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	projectId, ok := vars["projectId"]
	log.Debugf("projectId = [%v]\n", projectId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid projectId"),
			"C_INVALID_PROJECT_ID", ""))
		return
	}

	projectMemberId, ok := vars["projectMemberId"]
	log.Debugf("projectMemberId = [%v]\n", projectMemberId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid projectMemberId"),
			"C_INVALID_PROJECT_MEMBER_ID", ""))
		return
	}
	if err := p.usecase.RemoveProjectMember(projectMemberId); err != nil {
		ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
		return

	}
	ResponseJSON(w, r, http.StatusOK, domain.CommonProjectResponse{Result: "OK"})
}

// RemoveProjectMembers godoc
// @Tags        Projects
// @Summary     Remove project members to project
// @Description Remove project members to project
// @Accept      json
// @Produce     json
// @Param       organizationId path     string                            true "Organization ID"
// @Param       projectId      path     string                            true "Project ID"
// @Param       request        body     domain.RemoveProjectMemberRequest true "Request body to remove project member"
// @Success     200            {object} domain.CommonProjectResponse
// @Router      /organizations/{organizationId}/projects/{projectId}/members [delete]
// @Security    JWT
func (p ProjectHandler) RemoveProjectMembers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	log.Debugf("organizationId = [%v]\n", organizationId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"),
			"C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	projectId, ok := vars["projectId"]
	log.Debugf("projectId = [%v]\n", projectId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid projectId"),
			"C_INVALID_PROJECT_ID", ""))
		return
	}

	var projectMemberReq domain.RemoveProjectMemberRequest
	if err := UnmarshalRequestInput(r, &projectMemberReq); err != nil {
		ErrorJSON(w, r, err)
		return
	}

	for _, pm := range projectMemberReq.ProjectMember {
		if err := p.usecase.RemoveProjectMember(pm.ProjectMemberId); err != nil {
			ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
			return
		}
	}
	ResponseJSON(w, r, http.StatusOK, domain.CommonProjectResponse{Result: "OK"})
}

// UpdateProjectMemberRole godoc
// @Tags        Projects
// @Summary     Update project  member   Role
// @Description Update project  member   Role
// @Accept      json
// @Produce     json
// @Param       organizationId  path     string                                true "Organization ID"
// @Param       projectId       path     string                                true "Project ID"
// @Param       projectMemberId path     string                                true "Project Member ID"
// @Param       request         body     domain.UpdateProjectMemberRoleRequest true "Request body to update project member role"
// @Success     200             {object} domain.CommonProjectResponse
// @Router      /organizations/{organizationId}/projects/{projectId}/members/{projectMemberId}/role [put]
// @Security    JWT
func (p ProjectHandler) UpdateProjectMemberRole(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	log.Debugf("organizationId = [%v]\n", organizationId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"),
			"C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	projectId, ok := vars["projectId"]
	log.Debugf("projectId = [%v]\n", projectId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid projectId"),
			"C_INVALID_PROJECT_ID", ""))
		return
	}

	projectMemberId, ok := vars["projectMemberId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid projectMemberId"),
			"C_INVALID_PROJECT_MEMBER_ID", ""))
		return
	}

	var pmrReq domain.UpdateProjectMemberRoleRequest
	if err := UnmarshalRequestInput(r, &pmrReq); err != nil {
		ErrorJSON(w, r, err)
		return
	}

	if err := p.usecase.UpdateProjectMemberRole(projectMemberId, pmrReq.ProjectRoleId); err != nil {
		ErrorJSON(w, r, err)
		return
	}

	ResponseJSON(w, r, http.StatusOK, domain.CommonProjectResponse{Result: "OK"})
}

// CreateProjectNamespace godoc
// @Tags        Projects
// @Summary     Create project namespace
// @Description Create project namespace
// @Accept      json
// @Produce     json
// @Param       organizationId path     string                               true "Organization ID"
// @Param       projectId      path     string                               true "Project ID"
// @Param       stackId        path     string                               true "Stack ID"
// @Param       request        body     domain.CreateProjectNamespaceRequest true "Request body to create project namespace"
// @Success     200            {object} domain.CreateProjectNamespaceResponse
// @Router      /organizations/{organizationId}/projects/{projectId}/stacks/{stackId}/namespaces [post]
// @Security    JWT
func (p ProjectHandler) CreateProjectNamespace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	log.Debugf("organizationId = [%v]\n", organizationId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"),
			"C_INVALID_ORGANIZATION_ID", ""))
		return
	}
	projectId, ok := vars["projectId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid projectId"),
			"C_INVALID_PROJECT_ID", ""))
		return
	}
	stackId, ok := vars["stackId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid stackId"),
			"C_INVALID_STACK_ID", ""))
		return
	}

	var projectNamespaceReq domain.CreateProjectNamespaceRequest
	if err := UnmarshalRequestInput(r, &projectNamespaceReq); err != nil {
		ErrorJSON(w, r, err)
		return
	}

	now := time.Now()
	pn := &domain.ProjectNamespace{
		ProjectId:   projectId,
		StackId:     stackId,
		Namespace:   projectNamespaceReq.Namespace,
		Description: projectNamespaceReq.Description,
		Status:      "CREATING",
		CreatedAt:   now,
	}

	pnId, err := p.usecase.CreateProjectNamespace(pn)
	if err != nil {
		ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
		return
	}
	out := domain.CreateProjectNamespaceResponse{ProjectNamesapceId: pnId}
	ResponseJSON(w, r, http.StatusOK, out)
}

// IsProjectNamespaceExist godoc
// @Tags        Projects
// @Summary     Check project namespace exist
// @Description Check project namespace exist
// @Accept      json
// @Produce     json
// @Param       organizationId   path     string true "Organization ID"
// @Param       projectId        path     string true "Project ID"
// @Param       stackId          path     string true "Project Stack ID"
// @Param       projectNamespace path     string true "Project Namespace"
// @Success     200              {object} domain.CheckExistedResponse
// @Router      /organizations/{organizationId}/projects/{projectId}/stacks/{stackId}/namespaces/{projectNamespace}/existence [get]
// @Security    JWT
func (p ProjectHandler) IsProjectNamespaceExist(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("organizationId not found in path"),
			"C_INVALID_ORGANIZATION_ID", ""))
		return
	}
	projectId, ok := vars["projectId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid projectId"),
			"C_INVALID_PROJECT_ID", ""))
		return
	}
	stackId, ok := vars["stackId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid stackId"),
			"C_INVALID_STACK_ID", ""))
		return
	}
	projectNamespace, ok := vars["projectNamespace"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("projectNamespace not found in path"),
			"C_INVALID_PROJECT_NAMESPACE", ""))
		return
	}

	exist, err := p.usecase.IsProjectNamespaceExist(organizationId, projectId, stackId, projectNamespace)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var out domain.CheckExistedResponse
	out.Existed = exist

	ResponseJSON(w, r, http.StatusOK, out)
}

// GetProjectNamespaces godoc
// @Tags        Projects
// @Summary     Get project namespaces
// @Description Get project namespaces
// @Accept      json
// @Produce     json
// @Param       organizationId path     string true "Organization ID"
// @Param       projectId      path     string true "Project ID"
// @Param       stackId        path     string true "Project Stack ID"
// @Success     200            {object} domain.GetProjectNamespacesResponse
// @Router      /organizations/{organizationId}/projects/{projectId}/stacks/{stackId}/namespaces [get]
// @Security    JWT
func (p ProjectHandler) GetProjectNamespaces(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("organizationId not found in path"),
			"C_INVALID_ORGANIZATION_ID", ""))
		return
	}
	projectId, ok := vars["projectId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid projectId"),
			"C_INVALID_PROJECT_ID", ""))
		return
	}
	stackId, ok := vars["stackId"]
	log.Debugf("stackId = [%v]\n", stackId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid stackId"),
			"C_INVALID_STACK_ID", ""))
		return
	}

	pns, err := p.usecase.GetProjectNamespaces(organizationId, projectId, stackId)
	if err != nil {
		log.ErrorWithContext(r.Context(), "Failed to get project namespaces.", err)
		ErrorJSON(w, r, err)
		return
	}

	var out domain.GetProjectNamespacesResponse
	out.ProjectNamespaces = pns

	ResponseJSON(w, r, http.StatusOK, out)

}

// GetProjectNamespace godoc
// @Tags        Projects
// @Summary     Get project namespace
// @Description Get project namespace
// @Accept      json
// @Produce     json
// @Param       organizationId     path     string true "Organization ID"
// @Param       projectId          path     string true "Project ID"
// @Param       stackId            path     string true "Project Stack ID"
// @Param       projectNamespaceId path     string true "Project Namespace ID"
// @Success     200                {object} domain.GetProjectNamespaceResponse
// @Router      /organizations/{organizationId}/projects/{projectId}/stacks/{stackId}/namespaces/{projectNamespaceId} [get]
// @Security    JWT
func (p ProjectHandler) GetProjectNamespace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("organizationId not found in path"),
			"C_INVALID_ORGANIZATION_ID", ""))
		return
	}
	projectId, ok := vars["projectId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid projectId"),
			"C_INVALID_PROJECT_ID", ""))
		return
	}
	stackId, ok := vars["stackId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid stackId"),
			"C_INVALID_STACK_ID", ""))
		return
	}
	projectNamespaceId, ok := vars["projectNamespaceId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid projectNamespaceId"),
			"C_INVALID_PROJECT_NAMESPACE_ID", ""))
		return
	}

	pn, err := p.usecase.GetProjectNamespace(organizationId, projectId, stackId, projectNamespaceId)
	if err != nil {
		log.ErrorWithContext(r.Context(), "Failed to get project namespace.", err)
		ErrorJSON(w, r, err)
		return
	}

	var out domain.GetProjectNamespaceResponse
	out.ProjectNamespace = pn
	if pn == nil {
		ResponseJSON(w, r, http.StatusNotFound, out)
	} else {
		ResponseJSON(w, r, http.StatusOK, out)
	}
}

// DeleteProjectNamespace godoc
// @Tags        Projects
// @Summary     Delete project namespace
// @Description Delete project namespace
// @Accept      json
// @Produce     json
// @Param       organizationId     path     string true "Organization ID"
// @Param       projectId          path     string true "Project ID"
// @Param       stackId            path     string true "Stack ID"
// @Param       projectNamespaceId path     string true "Project Namespace ID"
// @Success     200                {object} domain.CommonProjectResponse
// @Router      /organizations/{organizationId}/projects/{projectId}/stacks/{stackId}/namespaces/{projectNamespaceId} [delete]
// @Security    JWT
func (p ProjectHandler) DeleteProjectNamespace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("organizationId not found in path"),
			"C_INVALID_ORGANIZATION_ID", ""))
		return
	}
	projectId, ok := vars["projectId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid projectId"),
			"C_INVALID_PROJECT_ID", ""))
		return
	}
	stackId, ok := vars["stackId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid stackId"),
			"C_INVALID_STACK_ID", ""))
		return
	}
	projectNamespaceId, ok := vars["projectNamespaceId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid projectNamespaceId"),
			"C_INVALID_PROJECT_NAMESPACE_ID", ""))
		return
	}

	if err := p.usecase.DeleteProjectNamespace(organizationId, projectId, stackId, projectNamespaceId); err != nil {
		ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
		return

	}
	ResponseJSON(w, r, http.StatusOK, domain.CommonProjectResponse{Result: "OK"})
}

func (p ProjectHandler) SetFavoriteProject(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
}

func (p ProjectHandler) SetFavoriteProjectNamespace(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
}

func (p ProjectHandler) UnSetFavoriteProject(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
}

func (p ProjectHandler) UnSetFavoriteProjectNamespace(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
}
