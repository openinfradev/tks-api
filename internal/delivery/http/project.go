package http

import (
	"fmt"
	"github.com/openinfradev/tks-api/internal/keycloak"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/middleware/auth/request"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/internal/serializer"

	"github.com/gorilla/mux"
	"github.com/openinfradev/tks-api/internal/model"
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
	IsProjectNameExist(w http.ResponseWriter, r *http.Request)
	GetProjects(w http.ResponseWriter, r *http.Request)
	Admin_GetProjects(w http.ResponseWriter, r *http.Request)

	AddProjectMember(w http.ResponseWriter, r *http.Request)
	GetProjectMember(w http.ResponseWriter, r *http.Request)
	GetProjectMembers(w http.ResponseWriter, r *http.Request)
	GetProjectMemberCount(w http.ResponseWriter, r *http.Request)
	RemoveProjectMember(w http.ResponseWriter, r *http.Request)
	RemoveProjectMembers(w http.ResponseWriter, r *http.Request)
	UpdateProjectMemberRole(w http.ResponseWriter, r *http.Request)
	UpdateProjectMembersRole(w http.ResponseWriter, r *http.Request)

	CreateProjectNamespace(w http.ResponseWriter, r *http.Request)
	IsProjectNamespaceExist(w http.ResponseWriter, r *http.Request)
	GetProjectNamespaces(w http.ResponseWriter, r *http.Request)
	GetProjectNamespace(w http.ResponseWriter, r *http.Request)
	UpdateProjectNamespace(w http.ResponseWriter, r *http.Request)
	DeleteProjectNamespace(w http.ResponseWriter, r *http.Request)

	SetFavoriteProject(w http.ResponseWriter, r *http.Request)
	SetFavoriteProjectNamespace(w http.ResponseWriter, r *http.Request)
	UnSetFavoriteProject(w http.ResponseWriter, r *http.Request)
	UnSetFavoriteProjectNamespace(w http.ResponseWriter, r *http.Request)

	GetProjectKubeconfig(w http.ResponseWriter, r *http.Request)
	GetProjectNamespaceK8sResources(w http.ResponseWriter, r *http.Request)
	GetProjectNamespaceResourcesUsage(w http.ResponseWriter, r *http.Request)
	GetProjectNamespaceKubeconfig(w http.ResponseWriter, r *http.Request)
}

type ProjectHandler struct {
	usecase     usecase.IProjectUsecase
	authUsecase usecase.IAuthUsecase
}

func NewProjectHandler(u usecase.Usecase) IProjectHandler {
	return &ProjectHandler{
		usecase:     u.Project,
		authUsecase: u.Auth,
	}
}

// CreateProject godoc
//
//	@Tags			Projects
//	@Summary		Create new project
//	@Description	Create new project
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string						true	"Organization ID"
//	@Param			request			body		domain.CreateProjectRequest	true	"Request body to create project"
//	@Success		200				{object}	domain.CreateProjectResponse
//	@Router			/organizations/{organizationId}/projects [post]
//	@Security		JWT
func (p ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	log.Debugf(r.Context(), "organizationId = [%v]\n", organizationId)
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

	now := time.Now()
	project := &model.Project{
		OrganizationId: organizationId,
		Name:           projectReq.Name,
		Description:    projectReq.Description,
		CreatedAt:      now,
	}
	log.Infof(r.Context(), "Processing CREATE request for project '%s'...", project.Name)

	projectId, err := p.usecase.CreateProject(r.Context(), project)
	if err != nil {
		ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
		return
	}

	project.ID = projectId
	ProjectLeaderId, err := uuid.Parse(projectReq.ProjectLeaderId)
	if err != nil {
		log.Error(r.Context(), err)
		ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", "Failed to parse uuid to string"))
		return
	}

	prs, err := p.usecase.GetProjectRoles(r.Context(), usecase.ProjectLeader)
	if err != nil {
		log.Error(r.Context(), err)
		ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", "Failed to retrieve project-leader id"))
		return
	}

	//Don't add ProjectUser Object because of Cascading
	pm := &model.ProjectMember{
		ProjectId: projectId,
		//ProjectUser: &model.ProjectUser{ID: ProjectLeaderId},
		//ProjectRole: &model.ProjectRole{ID: projectReq.ProjectRoleId},
		ProjectUserId:   ProjectLeaderId,
		ProjectRoleId:   prs[0].ID,
		IsProjectLeader: true,
		CreatedAt:       now,
	}

	projectMemberId, err := p.usecase.AddProjectMember(r.Context(), organizationId, pm)
	if err != nil {
		log.Errorf(r.Context(), "projectMemberId: %v", projectMemberId)
		ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
		return
	}

	err = p.usecase.AssignKeycloakClientRoleToMember(r.Context(), organizationId, projectId, keycloak.DefaultClientID, projectMemberId)
	if err != nil {
		log.Error(r.Context(), err)
		ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
		return
	}

	out := domain.CreateProjectResponse{ProjectId: projectId}
	ResponseJSON(w, r, http.StatusOK, out)

}

// GetProjects godoc
//
//	@Tags			Projects
//	@Summary		Get projects
//	@Description	Get projects
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"Organization ID"
//	@Param			query			query		string	false	"(all | only)"
//	@Param			projectName		query		string	false	"Project Name"
//	@Success		200				{object}	domain.GetProjectsResponse
//	@Router			/organizations/{organizationId}/projects [get]
//	@Security		JWT
func (p ProjectHandler) GetProjects(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"),
			"C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	urlParams := r.URL.Query()
	queryName := urlParams.Get("query")
	onlyMyProject := false
	if queryName == "only" {
		onlyMyProject = true
	}
	projectName := urlParams.Get("projectName")

	pg := pagination.NewPagination(&urlParams)

	// get myUserId from login component
	requestUserInfo, ok := request.UserFrom(r.Context())
	if !ok {
		log.Error(r.Context(), "Failed to retrieve user info from request")
		ErrorJSON(w, r, fmt.Errorf("failed to retrieve user info from request"))
	}
	myUserId := requestUserInfo.GetUserId().String()
	pr, err := p.usecase.GetProjects(r.Context(), organizationId, myUserId, onlyMyProject, projectName, pg)
	if err != nil {
		log.Error(r.Context(), "Failed to retrieve projects ", err)
		ErrorJSON(w, r, err)
		return
	}

	var out domain.GetProjectsResponse
	out.Projects = pr
	if out.Pagination, err = pg.Response(r.Context()); err != nil {
		log.Info(r.Context(), err)
	}

	if pr == nil {
		out.Projects = make([]domain.ProjectResponse, 0)
	}
	ResponseJSON(w, r, http.StatusOK, out)
}

// Admin_GetProjects godoc
//
//	@Tags			Projects
//	@Summary		Get projects as admin
//	@Description	Get projects as admin
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"Organization ID"
//	@Param			projectName		query		string	false	"Project Name"
//	@Success		200				{object}	domain.GetProjectsResponse
//	@Router			/admin/organizations/{organizationId}/projects [get]
//	@Security		JWT
func (p ProjectHandler) Admin_GetProjects(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"),
			"C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	urlParams := r.URL.Query()
	projectName := urlParams.Get("projectName")
	pg := pagination.NewPagination(&urlParams)
	pr, err := p.usecase.GetProjects(r.Context(), organizationId, "", false, projectName, pg)
	if err != nil {
		log.Error(r.Context(), "Failed to retrieve projects ", err)
		ErrorJSON(w, r, err)
		return
	}

	var out domain.GetProjectsResponse
	out.Projects = pr
	if out.Pagination, err = pg.Response(r.Context()); err != nil {
		log.Info(r.Context(), err)
	}

	if pr == nil {
		ResponseJSON(w, r, http.StatusNotFound, domain.GetProjectsResponse{})
	} else {
		ResponseJSON(w, r, http.StatusOK, out)
	}
}

// GetProject   godoc
//
//	@Tags			Projects
//	@Summary		Get projects
//	@Description	Get projects
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"Organization ID"
//	@Param			projectId		path		string	true	"Project ID"
//	@Success		200				{object}	domain.GetProjectResponse
//	@Router			/organizations/{organizationId}/projects/{projectId} [get]
//	@Security		JWT
func (p ProjectHandler) GetProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
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

	project, err := p.usecase.GetProjectWithLeader(r.Context(), organizationId, projectId)
	if err != nil {
		log.Error(r.Context(), "Failed to retrieve project", err)
		ErrorJSON(w, r, err)
		return
	}
	if project == nil {
		project, err := p.usecase.GetProject(r.Context(), organizationId, projectId)
		if err != nil {
			log.Error(r.Context(), "Failed to retrieve project", err)
			ErrorJSON(w, r, err)
			return
		}
		if project == nil {
			ResponseJSON(w, r, http.StatusOK, domain.GetProjectResponse{})
		}
	}

	requestUserInfo, ok := request.UserFrom(r.Context())
	if !ok {
		ErrorJSON(w, r, fmt.Errorf("failed to retrieve user info from request"))
	}
	myUserId := requestUserInfo.GetUserId().String()

	var projectRoleId, projectRoleName string
	//userProjectRole := requestUserInfo.GetRoleProjectMapping()
	//if userProjectRole != nil {
	//	projectRoleName = userProjectRole[project.ID]
	//}
	//projectRoles, err := p.usecase.GetProjectRoles(r.Context(), usecase.ProjectAll)
	//if err != nil {
	//	ErrorJSON(w, r, fmt.Errorf("failed to retrieve project role"))
	//}
	//for _, projectRole := range projectRoles {
	//	if projectRoleName == projectRole.Name {
	//		projectRoleId = projectRole.ID
	//		break
	//	}
	//}

	pm, err := p.usecase.GetProjectMemberByUserId(r.Context(), project.ID, myUserId)
	if err != nil {
		log.Warnf(r.Context(), "failed to retrieve project member %+v", err)
	}
	if pm != nil {
		projectRoleId = pm.ProjectRole.ID
		projectRoleName = pm.ProjectRole.Name
	}

	//appCount, err := p.usecase.GetAppCount(organizationId, projectId)
	//if err != nil {
	//	log.Error(r.Context(), "Failed to retrieve app count", err)
	//	ErrorJSON(w, r, err)
	//	return
	//}

	var out domain.GetProjectResponse
	var projectLeaderId, projectLeaderName, projectLeaderAccountId, projectLeaderDepartment string
	for _, pu := range project.ProjectMembers {
		projectLeaderId = pu.ProjectUser.ID.String()
		projectLeaderName = pu.ProjectUser.Name
		projectLeaderAccountId = pu.ProjectUser.AccountId
		projectLeaderDepartment = pu.ProjectUser.Department
	}

	var pdr domain.ProjectDetailResponse
	if err = serializer.Map(r.Context(), *project, &pdr); err != nil {
		log.Error(r.Context(), err)
		ErrorJSON(w, r, err)
		return
	}
	pdr.ProjectLeaderId = projectLeaderId
	pdr.ProjectLeaderName = projectLeaderName
	pdr.ProjectLeaderAccountId = projectLeaderAccountId
	pdr.ProjectLeaderDepartment = projectLeaderDepartment
	pdr.ProjectRoleId = projectRoleId
	pdr.ProjectRoleName = projectRoleName
	//pdr.AppCount = appCount

	out.Project = &pdr
	ResponseJSON(w, r, http.StatusOK, out)
}

// IsProjectNameExist godoc
//
//	@Tags			Projects
//	@Summary		Check project name exist
//	@Description	Check project name exist
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"Organization ID"
//	@Param			type			query		string	false	"type (name)"
//	@Param			value			query		string	true	"value (project name)"
//	@Success		200				{object}	domain.CheckExistedResponse
//	@Router			/organizations/{organizationId}/projects/existence [get]
//	@Security		JWT
func (p ProjectHandler) IsProjectNameExist(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("organizationId not found in path"),
			"C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	urlParams := r.URL.Query()
	projectName := urlParams.Get("value")

	exist, err := p.usecase.IsProjectNameExist(r.Context(), organizationId, projectName)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var out domain.CheckExistedResponse
	out.Existed = exist

	ResponseJSON(w, r, http.StatusOK, out)
}

// UpdateProject godoc
//
//	@Tags			Projects
//	@Summary		Update project
//	@Description	Update project
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string						true	"Organization ID"
//	@Param			projectId		path		string						true	"Project ID"
//	@Param			request			body		domain.UpdateProjectRequest	true	"Request body to update project"
//	@Success		200				{object}	domain.CommonProjectResponse
//	@Router			/organizations/{organizationId}/projects/{projectId} [put]
//	@Security		JWT
func (p ProjectHandler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	log.Debugf(r.Context(), "organizationId = [%v]\n", organizationId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"),
			"C_INVALID_ORGANIZATION_ID", ""))
		return
	}
	projectId, ok := vars["projectId"]
	log.Debugf(r.Context(), "projectId = [%v]\n", projectId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid projectId"),
			"C_INVALID_PROJECT_ID", ""))
		return
	}

	var projectReq domain.UpdateProjectRequest
	if err := UnmarshalRequestInput(r, &projectReq); err != nil {
		ErrorJSON(w, r, err)
		return
	}

	now := time.Now()
	project, err := p.usecase.GetProjectWithLeader(r.Context(), organizationId, projectId)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}
	if project == nil {
		project, err = p.usecase.GetProject(r.Context(), organizationId, projectId)
		if err != nil {
			ErrorJSON(w, r, err)
			return
		}
	}

	project.Name = projectReq.Name
	project.Description = projectReq.Description
	project.UpdatedAt = &now

	if err := p.usecase.UpdateProject(r.Context(), project, projectReq.ProjectLeaderId); err != nil {
		ErrorJSON(w, r, err)
		return
	}

	ResponseJSON(w, r, http.StatusOK, domain.CommonProjectResponse{Result: "OK"})
}

func (p ProjectHandler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	//ToDo: to donggyu. implement cleanup logic for k8s & keycloak

	//TODO implement me
}

// GetProjectRole godoc
//
//	@Tags			Projects
//	@Summary		Get project role
//	@Description	Get project role by id
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"Organization ID"
//	@Param			projectRoleId	path		string	true	"Project Role ID"
//	@Success		200				{object}	domain.GetProjectRoleResponse
//	@Router			/organizations/{organizationId}/projects/project-roles/{projectRoleId} [get]
//	@Security		JWT
func (p ProjectHandler) GetProjectRole(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	log.Debugf(r.Context(), "organizationId = [%v]\n", organizationId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"),
			"C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	projectRoleId, ok := vars["projectRoleId"]
	log.Debugf(r.Context(), "projectRoleId = [%v]\n", projectRoleId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid projectRoleId"),
			"C_INVALID_PROJECT_ROLE_ID", ""))
		return
	}

	pr, err := p.usecase.GetProjectRole(r.Context(), projectRoleId)
	if err != nil {
		log.Error(r.Context(), "Failed to get project roles ", err)
		ErrorJSON(w, r, err)
		return
	}

	var out domain.GetProjectRoleResponse
	out.ProjectRole = domain.ProjectRoleResponse(*pr)

	ResponseJSON(w, r, http.StatusOK, out)
}

// GetProjectRoles godoc
//
//	@Tags			Projects
//	@Summary		Get project roles
//	@Description	Get project roles by giving params
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"Organization ID"
//	@Param			query			query		string	false	"project role search by query (query=all), (query=leader), (query=member), (query=viewer)"
//	@Success		200				{object}	domain.GetProjectRolesResponse
//	@Router			/organizations/{organizationId}/projects/project-roles [get]
//	@Security		JWT
func (p ProjectHandler) GetProjectRoles(w http.ResponseWriter, r *http.Request) {
	urlParams := r.URL.Query()

	queryParam := urlParams.Get("query")
	var query int
	if queryParam == "" || strings.EqualFold(queryParam, "all") {
		query = usecase.ProjectAll
	} else if strings.EqualFold(queryParam, "leader") {
		query = usecase.ProjectLeader
	} else if strings.EqualFold(queryParam, "member") {
		query = usecase.ProjectMember
	} else if strings.EqualFold(queryParam, "viewer") {
		query = usecase.ProjectViewer
	} else {
		log.Error(r.Context(), "Invalid query params. Err: ")
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid query params"),
			"C_INVALID_QUERY_PARAM", ""))
		return
	}

	prs, err := p.usecase.GetProjectRoles(r.Context(), query)
	if err != nil {
		log.Error(r.Context(), "Failed to get project roles ", err)
		ErrorJSON(w, r, err)
		return
	}

	var out domain.GetProjectRolesResponse
	for _, pr := range prs {
		out.ProjectRoles = append(out.ProjectRoles, domain.ProjectRoleResponse(pr))
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// AddProjectMember godoc
//
//	@Tags			Projects
//	@Summary		Add project member to project
//	@Description	Add project member to project
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string							true	"Organization ID"
//	@Param			projectId		path		string							true	"Project ID"
//	@Param			request			body		domain.AddProjectMemberRequest	true	"Request body to add project member"
//	@Success		200				{object}	domain.CommonProjectResponse
//	@Router			/organizations/{organizationId}/projects/{projectId}/members [post]
//	@Security		JWT
func (p ProjectHandler) AddProjectMember(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	log.Debugf(r.Context(), "organizationId = [%v]\n", organizationId)
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

	var projectMemberReq domain.AddProjectMemberRequest
	if err := UnmarshalRequestInput(r, &projectMemberReq); err != nil {
		ErrorJSON(w, r, err)
		return
	}

	urlParams := r.URL.Query()
	pg := pagination.NewPagination(&urlParams)
	pns, err := p.usecase.GetProjectNamespaces(r.Context(), organizationId, projectId, pg)
	if err != nil {
		log.Error(r.Context(), err)
		ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
		return
	}
	stackIds := make(map[string]struct{})
	for _, pn := range pns {
		stackIds[pn.StackId] = struct{}{}
	}

	now := time.Now()
	for _, pmr := range projectMemberReq.ProjectMemberRequests {
		pu, err := p.usecase.GetProjectUser(r.Context(), pmr.ProjectUserId)
		if err != nil {
			log.Error(r.Context(), err)
			ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid projectUserId"),
				"C_INVALID_PROJECT_USER_ID", ""))
			return
		}

		pr, err := p.usecase.GetProjectRole(r.Context(), pmr.ProjectRoleId)
		if err != nil {
			log.Error(r.Context(), err)
			ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
			return
		}
		if pr == nil {
			ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid projectRoleId"),
				"C_INVALID_PROJECT_ROLE_ID", ""))
			return
		}

		pm := &model.ProjectMember{
			ProjectId:     projectId,
			ProjectUserId: pu.ID,
			ProjectUser:   nil,
			ProjectRoleId: pr.ID,
			ProjectRole:   nil,
			CreatedAt:     now,
		}
		pmId, err := p.usecase.AddProjectMember(r.Context(), organizationId, pm)
		if err != nil {
			log.Errorf(r.Context(), "projectMemberId: %s", pmId)
			ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
			return
		}

		err = p.usecase.AssignKeycloakClientRoleToMember(r.Context(), organizationId, projectId, keycloak.DefaultClientID, pmId)
		if err != nil {
			log.Error(r.Context(), err)
			ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
			return
		}

		// tasks for keycloak & k8s
		for stackId := range stackIds {
			if err := p.usecase.AssignKeycloakClientRoleToMember(r.Context(), organizationId, projectId, stackId+"-k8s-api", pmId); err != nil {
				log.Error(r.Context(), err)
				ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
				return
			}
		}

		err = p.authUsecase.UpdateExpiredTimeOnToken(r.Context(), organizationId, pmr.ProjectUserId)
		if err != nil {
			log.Error(r.Context(), err)
			ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
			return
		}
	}

	out := domain.CommonProjectResponse{Result: "OK"}
	ResponseJSON(w, r, http.StatusOK, out)
}

// GetProjectMember godoc
//
//	@Tags			Projects
//	@Summary		Get project member
//	@Description	Get project member
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"Organization ID"
//	@Param			projectId		path		string	true	"Project ID"
//	@Param			projectMemberId	path		string	true	"Project Member ID"
//	@Success		200				{object}	domain.GetProjectMemberResponse
//	@Router			/organizations/{organizationId}/projects/{projectId}/members/{projectMemberId} [get]
//	@Security		JWT
func (p ProjectHandler) GetProjectMember(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	log.Debugf(r.Context(), "organizationId = [%v]\n", organizationId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"),
			"C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	projectId, ok := vars["projectId"]
	log.Debugf(r.Context(), "projectId = [%v]\n", projectId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid projectId"),
			"C_INVALID_PROJECT_ID", ""))
		return
	}

	projectMemberId, ok := vars["projectMemberId"]
	log.Debugf(r.Context(), "projectMemberId = [%v]\n", projectMemberId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid projectMemberId"),
			"C_INVALID_PROJECT_MEMBER_ID", ""))
		return
	}

	pm, err := p.usecase.GetProjectMember(r.Context(), projectMemberId)
	if err != nil {
		log.Error(r.Context(), "Failed to get project member ", err)
		ErrorJSON(w, r, err)
		return
	}

	var out domain.GetProjectMemberResponse
	if pm == nil {
		ResponseJSON(w, r, http.StatusNotFound, out)
		return
	}

	pmr := &domain.ProjectMemberResponse{
		ID:                    pm.ID,
		ProjectId:             pm.ProjectId,
		ProjectUserId:         pm.ProjectUser.ID.String(),
		ProjectUserName:       pm.ProjectUser.Name,
		ProjectUserAccountId:  pm.ProjectUser.AccountId,
		ProjectUserEmail:      pm.ProjectUser.Email,
		ProjectUserDepartment: pm.ProjectUser.Department,
		ProjectRoleId:         pm.ProjectRole.ID,
		ProjectRoleName:       pm.ProjectRole.Name,
		CreatedAt:             pm.CreatedAt,
		UpdatedAt:             pm.UpdatedAt,
	}

	out.ProjectMember = pmr
	ResponseJSON(w, r, http.StatusOK, out)
}

// GetProjectMembers godoc
//
//	@Tags			Projects
//	@Summary		Get project members
//	@Description	Get project members
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"Organization ID"
//	@Param			projectId		path		string	true	"Project ID"
//	@Param			query			query		string	false	"project member search by query (query=all), (query=leader), (query=member), (query=viewer)"
//	@Success		200				{object}	domain.GetProjectMembersResponse
//	@Router			/organizations/{organizationId}/projects/{projectId}/members [get]
//	@Security		JWT
func (p ProjectHandler) GetProjectMembers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	log.Debugf(r.Context(), "organizationId = [%v]\n", organizationId)
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

	urlParams := r.URL.Query()

	queryParam := urlParams.Get("query")
	var query int
	if queryParam == "" || strings.EqualFold(queryParam, "all") {
		query = usecase.ProjectAll
	} else if strings.EqualFold(queryParam, "leader") {
		query = usecase.ProjectLeader
	} else if strings.EqualFold(queryParam, "member") {
		query = usecase.ProjectMember
	} else if strings.EqualFold(queryParam, "viewer") {
		query = usecase.ProjectViewer
	} else {
		log.Error(r.Context(), "Invalid query params. Err: ")
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid query params"),
			"C_INVALID_QUERY_PARAM", ""))
		return
	}

	pg := pagination.NewPagination(&urlParams)
	pms, err := p.usecase.GetProjectMembers(r.Context(), projectId, query, pg)
	if err != nil {
		log.Error(r.Context(), "Failed to get project members ", err)
		ErrorJSON(w, r, err)
		return
	}

	var out domain.GetProjectMembersResponse
	if pms == nil {
		ResponseJSON(w, r, http.StatusNotFound, out)
		return
	}

	pmrs := make([]domain.ProjectMemberResponse, 0)
	for _, pm := range pms {
		pmr := domain.ProjectMemberResponse{
			ID:                    pm.ID,
			ProjectId:             pm.ProjectId,
			ProjectUserId:         pm.ProjectUser.ID.String(),
			ProjectUserName:       pm.ProjectUser.Name,
			ProjectUserAccountId:  pm.ProjectUser.AccountId,
			ProjectUserEmail:      pm.ProjectUser.Email,
			ProjectUserDepartment: pm.ProjectUser.Department,
			ProjectRoleId:         pm.ProjectRole.ID,
			ProjectRoleName:       pm.ProjectRole.Name,
			CreatedAt:             pm.CreatedAt,
			UpdatedAt:             pm.UpdatedAt,
		}
		pmrs = append(pmrs, pmr)
	}

	out = domain.GetProjectMembersResponse{ProjectMembers: pmrs}
	if out.Pagination, err = pg.Response(r.Context()); err != nil {
		log.Info(r.Context(), err)
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// GetProjectMemberCount godoc
//
//	@Tags			Projects
//	@Summary		Get project member count group by project role
//	@Description	Get project member count group by project role
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"Organization ID"
//	@Param			projectId		path		string	true	"Project ID"
//	@Success		200				{object}	domain.GetProjectMemberCountResponse
//	@Router			/organizations/{organizationId}/projects/{projectId}/members/count [get]
//	@Security		JWT
func (p ProjectHandler) GetProjectMemberCount(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	log.Debugf(r.Context(), "organizationId = [%v]\n", organizationId)
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

	pmcr, err := p.usecase.GetProjectMemberCount(r.Context(), projectId)
	if err != nil {
		log.Error(r.Context(), "Failed to get project member count", err)
		ErrorJSON(w, r, err)
		return
	}

	if pmcr == nil {
		ResponseJSON(w, r, http.StatusNotFound, domain.GetProjectMembersResponse{})
		return
	}
	ResponseJSON(w, r, http.StatusOK, pmcr)
}

// RemoveProjectMember godoc
//
//	@Tags			Projects
//	@Summary		Remove project members to project
//	@Description	Remove project members to project
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"Organization ID"
//	@Param			projectId		path		string	true	"Project ID"
//	@Param			projectMemberId	path		string	true	"Project Member ID"
//	@Success		200				{object}	domain.CommonProjectResponse
//	@Router			/organizations/{organizationId}/projects/{projectId}/members/{projectMemberId} [delete]
//	@Security		JWT
func (p ProjectHandler) RemoveProjectMember(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	log.Debugf(r.Context(), "organizationId = [%v]\n", organizationId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"),
			"C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	projectId, ok := vars["projectId"]
	log.Debugf(r.Context(), "projectId = [%v]\n", projectId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid projectId"),
			"C_INVALID_PROJECT_ID", ""))
		return
	}

	projectMemberId, ok := vars["projectMemberId"]
	log.Debugf(r.Context(), "projectMemberId = [%v]\n", projectMemberId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid projectMemberId"),
			"C_INVALID_PROJECT_MEMBER_ID", ""))
		return
	}

	// tasks for keycloak & k8s
	pns, err := p.usecase.GetProjectNamespaces(r.Context(), organizationId, projectId, nil)
	if err != nil {
		log.Error(r.Context(), err)
		ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
		return
	}
	stackIds := make(map[string]struct{})
	for _, pn := range pns {
		stackIds[pn.StackId] = struct{}{}
	}

	err = p.usecase.UnassignKeycloakClientRoleToMember(r.Context(), organizationId, projectId, keycloak.DefaultClientID, projectMemberId)
	if err != nil {
		log.Error(r.Context(), err)
		ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
		return
	}

	for stackId := range stackIds {
		if err := p.usecase.UnassignKeycloakClientRoleToMember(r.Context(), organizationId, projectId, stackId+"-k8s-api", projectMemberId); err != nil {
			log.Error(r.Context(), err)
			ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
			return
		}
	}

	pm, err := p.usecase.GetProjectMember(r.Context(), projectMemberId)
	if err != nil {
		log.Error(r.Context(), err)
		ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
		return
	}

	err = p.authUsecase.UpdateExpiredTimeOnToken(r.Context(), organizationId, pm.ProjectUserId.String())
	if err != nil {
		log.Error(r.Context(), err)
		ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
		return
	}

	if err := p.usecase.RemoveProjectMember(r.Context(), organizationId, projectMemberId); err != nil {
		ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
		return

	}
	ResponseJSON(w, r, http.StatusOK, domain.CommonProjectResponse{Result: "OK"})
}

// RemoveProjectMembers godoc
//
//	@Tags			Projects
//	@Summary		Remove project members to project
//	@Description	Remove project members to project
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string								true	"Organization ID"
//	@Param			projectId		path		string								true	"Project ID"
//	@Param			request			body		domain.RemoveProjectMemberRequest	true	"Request body to remove project member"
//	@Success		200				{object}	domain.CommonProjectResponse
//	@Router			/organizations/{organizationId}/projects/{projectId}/members [delete]
//	@Security		JWT
func (p ProjectHandler) RemoveProjectMembers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	log.Debugf(r.Context(), "organizationId = [%v]\n", organizationId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"),
			"C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	projectId, ok := vars["projectId"]
	log.Debugf(r.Context(), "projectId = [%v]\n", projectId)
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

	// tasks for keycloak & k8s
	pns, err := p.usecase.GetProjectNamespaces(r.Context(), organizationId, projectId, nil)
	if err != nil {
		log.Error(r.Context(), err)
		ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
		return
	}
	stackIds := make(map[string]struct{})
	for _, pn := range pns {
		stackIds[pn.StackId] = struct{}{}
	}

	// TODO: change multi row delete
	for _, pmId := range projectMemberReq.ProjectMember {
		if err := p.usecase.UnassignKeycloakClientRoleToMember(r.Context(), organizationId, projectId, keycloak.DefaultClientID, pmId.ProjectMemberId); err != nil {
			log.Error(r.Context(), err)
			ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
			return
		}

		// tasks for keycloak & k8s
		for stackId := range stackIds {
			if err := p.usecase.UnassignKeycloakClientRoleToMember(r.Context(), organizationId, projectId, stackId+"-k8s-api", pmId.ProjectMemberId); err != nil {
				log.Error(r.Context(), err)
				ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
				return
			}
		}

		pm, err := p.usecase.GetProjectMember(r.Context(), pmId.ProjectMemberId)
		if err != nil {
			log.Error(r.Context(), err)
			ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
			return
		}
		err = p.authUsecase.UpdateExpiredTimeOnToken(r.Context(), organizationId, pm.ProjectUserId.String())
		if err != nil {
			log.Error(r.Context(), err)
			ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
			return
		}

		if err := p.usecase.RemoveProjectMember(r.Context(), organizationId, pmId.ProjectMemberId); err != nil {
			ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
			return
		}
	}
	ResponseJSON(w, r, http.StatusOK, domain.CommonProjectResponse{Result: "OK"})
}

// UpdateProjectMemberRole godoc
//
//	@Tags			Projects
//	@Summary		Update project  member   Role
//	@Description	Update project  member   Role
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string									true	"Organization ID"
//	@Param			projectId		path		string									true	"Project ID"
//	@Param			projectMemberId	path		string									true	"Project Member ID"
//	@Param			request			body		domain.UpdateProjectMemberRoleRequest	true	"Request body to update project member role"
//	@Success		200				{object}	domain.CommonProjectResponse
//	@Router			/organizations/{organizationId}/projects/{projectId}/members/{projectMemberId}/role [put]
//	@Security		JWT
func (p ProjectHandler) UpdateProjectMemberRole(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	log.Debugf(r.Context(), "organizationId = [%v]\n", organizationId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"),
			"C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	projectId, ok := vars["projectId"]
	log.Debugf(r.Context(), "projectId = [%v]\n", projectId)
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

	now := time.Now()
	pm, err := p.usecase.GetProjectMember(r.Context(), projectMemberId)
	if err != nil {
		log.Error(r.Context(), err)
		ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
		return
	}
	if pm == nil {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid projectMemberId"),
			"C_INVALID_PROJECT_MEMBER_ID", ""))
		return
	}

	if err = p.usecase.UnassignKeycloakClientRoleToMember(r.Context(), organizationId, projectId, keycloak.DefaultClientID, projectMemberId); err != nil {
		log.Error(r.Context(), err)
		ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
		return
	}

	pns, err := p.usecase.GetProjectNamespaces(r.Context(), organizationId, projectId, nil)
	if err != nil {
		log.Error(r.Context(), err)
		ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
		return
	}
	stackIds := make(map[string]struct{})
	for _, pn := range pns {
		stackIds[pn.StackId] = struct{}{}
	}
	// tasks for keycloak & k8s. Unassign old role
	for stackId := range stackIds {
		if err := p.usecase.UnassignKeycloakClientRoleToMember(r.Context(), organizationId, projectId, stackId+"-k8s-api", projectMemberId); err != nil {
			log.Error(r.Context(), err)
			ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
			return
		}
	}

	pm.ProjectRoleId = pmrReq.ProjectRoleId
	pm.ProjectUser = nil
	pm.ProjectRole = nil
	pm.UpdatedAt = &now

	if err := p.usecase.UpdateProjectMemberRole(r.Context(), pm); err != nil {
		ErrorJSON(w, r, err)
		return
	}

	if err := p.usecase.AssignKeycloakClientRoleToMember(r.Context(), organizationId, projectId, keycloak.DefaultClientID, projectMemberId); err != nil {
		log.Error(r.Context(), err)
		ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
		return
	}

	// tasks for keycloak & k8s. Assign new role
	for stackId := range stackIds {
		if err := p.usecase.AssignKeycloakClientRoleToMember(r.Context(), organizationId, projectId, stackId+"-k8s-api", projectMemberId); err != nil {
			log.Error(r.Context(), err)
			ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
			return
		}
	}

	// update token expired time
	err = p.authUsecase.UpdateExpiredTimeOnToken(r.Context(), organizationId, pm.ProjectUserId.String())
	if err != nil {
		log.Error(r.Context(), err)
		ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
		return
	}

	ResponseJSON(w, r, http.StatusOK, domain.CommonProjectResponse{Result: "OK"})
}

// UpdateProjectMembersRole godoc
//
//	@Tags			Projects
//	@Summary		Update project  member   Role
//	@Description	Update project  member   Role
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string									true	"Organization ID"
//	@Param			projectId		path		string									true	"Project ID"
//	@Param			request			body		domain.UpdateProjectMembersRoleRequest	true	"Request body to update project member role"
//	@Success		200				{object}	domain.CommonProjectResponse
//	@Router			/organizations/{organizationId}/projects/{projectId}/members [put]
//	@Security		JWT
func (p ProjectHandler) UpdateProjectMembersRole(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	log.Debugf(r.Context(), "organizationId = [%v]\n", organizationId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"),
			"C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	projectId, ok := vars["projectId"]
	log.Debugf(r.Context(), "projectId = [%v]\n", projectId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid projectId"),
			"C_INVALID_PROJECT_ID", ""))
		return
	}

	now := time.Now()
	var projectMemberReq domain.UpdateProjectMembersRoleRequest
	if err := UnmarshalRequestInput(r, &projectMemberReq); err != nil {
		ErrorJSON(w, r, err)
		return
	}

	pns, err := p.usecase.GetProjectNamespaces(r.Context(), organizationId, projectId, nil)
	if err != nil {
		log.Error(r.Context(), err)
		ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
		return
	}
	stackIds := make(map[string]struct{})
	for _, pn := range pns {
		stackIds[pn.StackId] = struct{}{}
	}

	for _, pmr := range projectMemberReq.ProjectMemberRoleRequests {
		pm, err := p.usecase.GetProjectMember(r.Context(), pmr.ProjectMemberId)
		if err != nil {
			log.Error(r.Context(), err)
			ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
			return
		}
		if pm == nil {
			ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid projectMemberId"),
				"C_INVALID_PROJECT_MEMBER_ID", ""))
			return
		}

		if err = p.usecase.UnassignKeycloakClientRoleToMember(r.Context(), organizationId, projectId, keycloak.DefaultClientID, pm.ID); err != nil {
			log.Error(r.Context(), err)
			ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
			return
		}

		for stackId := range stackIds {
			if err := p.usecase.UnassignKeycloakClientRoleToMember(r.Context(), organizationId, projectId, stackId+"-k8s-api", pm.ID); err != nil {
				log.Error(r.Context(), err)
				ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
				return
			}
		}

		pm.ProjectRoleId = pmr.ProjectRoleId
		pm.ProjectUser = nil
		pm.ProjectRole = nil
		pm.UpdatedAt = &now

		if err := p.usecase.UpdateProjectMemberRole(r.Context(), pm); err != nil {
			ErrorJSON(w, r, err)
			return
		}

		if err := p.usecase.AssignKeycloakClientRoleToMember(r.Context(), organizationId, projectId, keycloak.DefaultClientID, pm.ID); err != nil {
			log.Error(r.Context(), err)
			ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
			return
		}

		for stackId := range stackIds {
			if err := p.usecase.AssignKeycloakClientRoleToMember(r.Context(), organizationId, projectId, stackId+"-k8s-api", pm.ID); err != nil {
				log.Error(r.Context(), err)
				ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
				return
			}
		}

		// update token expired time
		err = p.authUsecase.UpdateExpiredTimeOnToken(r.Context(), organizationId, pm.ProjectUserId.String())
		if err != nil {
			log.Error(r.Context(), err)
			ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
			return
		}
	}

	ResponseJSON(w, r, http.StatusOK, domain.CommonProjectResponse{Result: "OK"})
}

// CreateProjectNamespace godoc
//
//	@Tags			Projects
//	@Summary		Create project namespace
//	@Description	Create project namespace
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string									true	"Organization ID"
//	@Param			projectId		path		string									true	"Project ID"
//	@Param			request			body		domain.CreateProjectNamespaceRequest	true	"Request body to create project namespace"
//	@Success		200				{object}	domain.CommonProjectResponse
//	@Router			/organizations/{organizationId}/projects/{projectId}/namespaces [post]
//	@Security		JWT
func (p ProjectHandler) CreateProjectNamespace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
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

	var projectNamespaceReq domain.CreateProjectNamespaceRequest
	if err := UnmarshalRequestInput(r, &projectNamespaceReq); err != nil {
		ErrorJSON(w, r, err)
		return
	}

	now := time.Now()
	pn := &model.ProjectNamespace{
		StackId:     projectNamespaceReq.StackId,
		Namespace:   projectNamespaceReq.Namespace,
		ProjectId:   projectId,
		Stack:       nil,
		Description: projectNamespaceReq.Description,
		Status:      "RUNNING",
		CreatedAt:   now,
	}

	// tasks for keycloak & k8s
	// ToDo: Check if the namespace is already created
	if err := p.usecase.EnsureNamespaceForCluster(r.Context(), organizationId, projectNamespaceReq.StackId, projectNamespaceReq.Namespace); err != nil {
		ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
		return
	}

	if err := p.usecase.EnsureRequiredSetupForCluster(r.Context(), organizationId, projectId, projectNamespaceReq.StackId); err != nil {
		ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
		return
	}
	if err := p.usecase.CreateK8SNSRoleBinding(r.Context(), organizationId, projectId, projectNamespaceReq.StackId, projectNamespaceReq.Namespace); err != nil {
		ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
		return
	}

	if err := p.usecase.CreateProjectNamespace(r.Context(), organizationId, pn); err != nil {
		ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
		return
	}
	out := domain.CommonProjectResponse{Result: "OK"}
	ResponseJSON(w, r, http.StatusOK, out)
}

// IsProjectNamespaceExist godoc
//
//	@Tags			Projects
//	@Summary		Check project namespace exist
//	@Description	Check project namespace exist
//	@Accept			json
//	@Produce		json
//	@Param			organizationId		path		string	true	"Organization ID"
//	@Param			projectId			path		string	true	"Project ID"
//	@Param			stackId				path		string	true	"Project Stack ID"
//	@Param			projectNamespace	path		string	true	"Project Namespace"
//	@Success		200					{object}	domain.CheckExistedResponse
//	@Router			/organizations/{organizationId}/projects/{projectId}/namespaces/{projectNamespace}/stacks/{stackId}/existence [get]
//	@Security		JWT
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

	exist, err := p.usecase.IsProjectNamespaceExist(r.Context(), organizationId, projectId, stackId, projectNamespace)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var out domain.CheckExistedResponse
	out.Existed = exist

	ResponseJSON(w, r, http.StatusOK, out)
}

// GetProjectNamespaces godoc
//
//	@Tags			Projects
//	@Summary		Get project namespaces
//	@Description	Get project namespaces
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"Organization ID"
//	@Param			projectId		path		string	true	"Project ID"
//	@Success		200				{object}	domain.GetProjectNamespacesResponse
//	@Router			/organizations/{organizationId}/projects/{projectId}/namespaces [get]
//	@Security		JWT
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

	urlParams := r.URL.Query()
	pg := pagination.NewPagination(&urlParams)
	pns, err := p.usecase.GetProjectNamespaces(r.Context(), organizationId, projectId, pg)
	if err != nil {
		log.Error(r.Context(), "Failed to get project namespaces.", err)
		ErrorJSON(w, r, err)
		return
	}

	var out domain.GetProjectNamespacesResponse
	if pns == nil {
		ResponseJSON(w, r, http.StatusNotFound, out)
		return
	}
	pnrs := make([]domain.ProjectNamespaceResponse, 0)
	for _, pn := range pns {
		var pnr domain.ProjectNamespaceResponse
		if err = serializer.Map(r.Context(), pn, &pnr); err != nil {
			log.Error(r.Context(), err)
			ErrorJSON(w, r, err)
			return
		}
		appCount, err := p.usecase.GetAppCount(r.Context(), organizationId, projectId, pn.Namespace)
		if err != nil {
			log.Error(r.Context(), "Failed to retrieve app count", err)
			ErrorJSON(w, r, err)
			return
		}

		pnr.StackName = pn.Stack.Name
		pnr.AppCount = appCount
		pnrs = append(pnrs, pnr)
	}

	out.ProjectNamespaces = pnrs
	ResponseJSON(w, r, http.StatusOK, out)

}

// GetProjectNamespace godoc
//
//	@Tags			Projects
//	@Summary		Get project namespace
//	@Description	Get project namespace
//	@Accept			json
//	@Produce		json
//	@Param			organizationId		path		string	true	"Organization ID"
//	@Param			projectId			path		string	true	"Project ID"
//	@Param			projectNamespace	path		string	true	"Project Namespace"
//	@Param			stackId				path		string	true	"Project Stack ID"
//	@Success		200					{object}	domain.GetProjectNamespaceResponse
//	@Router			/organizations/{organizationId}/projects/{projectId}/namespaces/{projectNamespace}/stacks/{stackId} [get]
//	@Security		JWT
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
	projectNamespace, ok := vars["projectNamespace"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid projectNamespace"),
			"C_INVALID_PROJECT_NAMESPACE", ""))
		return
	}
	stackId, ok := vars["stackId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid stackId"),
			"C_INVALID_STACK_ID", ""))
		return
	}

	pn, err := p.usecase.GetProjectNamespace(r.Context(), organizationId, projectId, projectNamespace, stackId)
	if err != nil {
		log.Error(r.Context(), "Failed to get project namespace.", err)
		ErrorJSON(w, r, err)
		return
	}

	appCount, err := p.usecase.GetAppCount(r.Context(), organizationId, projectId, projectNamespace)
	if err != nil {
		log.Error(r.Context(), "Failed to retrieve app count", err)
		ErrorJSON(w, r, err)
		return
	}

	var out domain.GetProjectNamespaceResponse
	if pn == nil {
		ResponseJSON(w, r, http.StatusNotFound, out)
		return
	}

	var pnr domain.ProjectNamespaceResponse
	if err = serializer.Map(r.Context(), *pn, &pnr); err != nil {
		log.Error(r.Context(), err)
		ErrorJSON(w, r, err)
		return
	}
	pnr.StackName = pn.Stack.Name
	pnr.AppCount = appCount

	out.ProjectNamespace = &pnr
	ResponseJSON(w, r, http.StatusOK, out)

}

// UpdateProjectNamespace godoc
//
//	@Tags			Projects
//	@Summary		Update project namespace
//	@Description	Update project namespace
//	@Accept			json
//	@Produce		json
//	@Param			organizationId		path		string									true	"Organization ID"
//	@Param			projectId			path		string									true	"Project ID"
//	@Param			projectNamespace	path		string									true	"Project Namespace"
//	@Param			stackId				path		string									true	"Project Stack ID"
//	@Param			request				body		domain.UpdateProjectNamespaceRequest	true	"Request body to update project namespace"
//	@Success		200					{object}	domain.CommonProjectResponse
//	@Router			/organizations/{organizationId}/projects/{projectId}/namespaces/{projectNamespace}/stacks/{stackId} [put]
//	@Security		JWT
func (p ProjectHandler) UpdateProjectNamespace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	log.Debugf(r.Context(), "organizationId = [%v]\n", organizationId)
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
	projectNamespace, ok := vars["projectNamespace"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid projectNamespace"),
			"C_INVALID_PROJECT_NAMESPACE", ""))
		return
	}
	stackId, ok := vars["stackId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid stackId"),
			"C_INVALID_STACK_ID", ""))
		return
	}

	var projectNamespaceReq domain.UpdateProjectNamespaceRequest
	if err := UnmarshalRequestInput(r, &projectNamespaceReq); err != nil {
		ErrorJSON(w, r, err)
		return
	}

	now := time.Now()
	pn, err := p.usecase.GetProjectNamespace(r.Context(), organizationId, projectId, projectNamespace, stackId)
	if err != nil {
		log.Error(r.Context(), "Failed to get project namespace.", err)
		ErrorJSON(w, r, err)
		return
	}

	pn.Description = projectNamespaceReq.Description
	pn.UpdatedAt = &now

	if err := p.usecase.UpdateProjectNamespace(r.Context(), pn); err != nil {
		ErrorJSON(w, r, err)
		return
	}

	ResponseJSON(w, r, http.StatusOK, domain.CommonProjectResponse{Result: "OK"})
}

// DeleteProjectNamespace godoc
//
//	@Tags			Projects
//	@Summary		Delete project namespace
//	@Description	Delete project namespace
//	@Accept			json
//	@Produce		json
//	@Param			organizationId		path		string	true	"Organization ID"
//	@Param			projectId			path		string	true	"Project ID"
//	@Param			stackId				path		string	true	"Stack ID"
//	@Param			projectNamespace	path		string	true	"Project Namespace"
//	@Success		200					{object}	domain.CommonProjectResponse
//	@Router			/organizations/{organizationId}/projects/{projectId}/namespaces/{projectNamespace}/stacks/{stackId} [delete]
//	@Security		JWT
func (p ProjectHandler) DeleteProjectNamespace(w http.ResponseWriter, r *http.Request) {

	//ToDo: from donggyu. uncomment lines below after implementing usecase.DeleteProjectNamespace.
	// tasks for keycloak & k8s
	/*if err := p.usecase.DeleteK8SNSRoleBinding(organizationId, projectId, stackId, projectNamespace); err != nil {
		ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
		return
	}
	if err := p.usecase.MayRemoveRequiredSetupForCluster(organizationId, projectId, stackId); err != nil {
		ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
		return
	}*/

	//TODO implement me

	//vars := mux.Vars(r)
	//organizationId, ok := vars["organizationId"]
	//if !ok {
	//	ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("organizationId not found in path"),
	//		"C_INVALID_ORGANIZATION_ID", ""))
	//	return
	//}
	//projectId, ok := vars["projectId"]
	//if !ok {
	//	ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid projectId"),
	//		"C_INVALID_PROJECT_ID", ""))
	//	return
	//}
	//projectNamespace, ok := vars["projectNamespace"]
	//if !ok {
	//	ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid projectNamespace"),
	//		"C_INVALID_PROJECT_NAMESPACE", ""))
	//	return
	//}
	//stackId, ok := vars["stackId"]
	//if !ok {
	//	ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid stackId"),
	//		"C_INVALID_STACK_ID", ""))
	//	return
	//}
	//
	//if err := p.usecase.DeleteProjectNamespace(organizationId, projectId, projectNamespace, stackId); err != nil {
	//	ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
	//	return
	//
	//}
	//ResponseJSON(w, r, http.StatusOK, model.CommonProjectResponse{Result: "OK"})
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

// GetProjectKubeconfig godoc
//
//	@Tags			Projects
//	@Summary		Get project kubeconfig
//	@Description	Get project kubeconfig
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"Organization ID"
//	@Param			projectId		path		string	true	"Project ID"
//	@Success		200				{object}	domain.GetProjectKubeconfigResponse
//	@Router			/organizations/{organizationId}/projects/{projectId}/kubeconfig [get]
//	@Security		JWT
func (p ProjectHandler) GetProjectKubeconfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("organizationId not found in path"),
			"C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	projectId, ok := vars["projectId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("projectId not found in path"),
			"C_INVALID_PROJECT_ID", ""))
		return
	}

	kubeconfig, err := p.usecase.GetProjectKubeconfig(r.Context(), organizationId, projectId)
	if err != nil {
		log.Error(r.Context(), "Failed to get project kubeconfig.", err)
		ErrorJSON(w, r, err)
		return
	}

	out := domain.GetProjectKubeconfigResponse{
		Kubeconfig: kubeconfig,
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// GetProjectNamespaceK8sResources godoc
//
//	@Tags			Projects
//	@Summary		Get k8s resources for project namespace
//	@Description	Get k8s resources for project namespace
//	@Accept			json
//	@Produce		json
//	@Param			organizationId		path		string	true	"Organization ID"
//	@Param			projectId			path		string	true	"Project ID"
//	@Param			stackId				path		string	true	"Stack ID"
//	@Param			projectNamespace	path		string	true	"Project Namespace"
//	@Success		200					{object}	domain.GetProjectNamespaceK8sResourcesResponse
//	@Router			/organizations/{organizationId}/projects/{projectId}/namespaces/{projectNamespace}/stacks/{stackId}/k8s-resources [get]
//	@Security		JWT
func (p ProjectHandler) GetProjectNamespaceK8sResources(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("organizationId not found in path"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	projectId, ok := vars["projectId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("projectId not found in path"), "C_INVALID_PROJECT_ID", ""))
		return
	}

	projectNamespace, ok := vars["projectNamespace"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid projectNamespace"), "C_INVALID_PROJECT_NAMESPACE", ""))
		return
	}
	stackId, ok := vars["stackId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid stackId"), "C_INVALID_STACK_ID", ""))
		return
	}

	k8sResources, err := p.usecase.GetK8sResources(r.Context(), organizationId, projectId, projectNamespace, domain.StackId(stackId))
	if err != nil {
		log.Error(r.Context(), "Failed to get project resources.", err)
		ErrorJSON(w, r, err)
		return
	}

	var out domain.GetProjectNamespaceK8sResourcesResponse
	if err = serializer.Map(r.Context(), k8sResources, &out.K8sResources); err != nil {
		log.Error(r.Context(), err)
	}
	ResponseJSON(w, r, http.StatusOK, out)
}

// GetProjectNamespaceResourcesUsage godoc
//
//	@Tags			Projects
//	@Summary		Get resources usage for project namespace
//	@Description	Get resources usage for project namespace
//	@Accept			json
//	@Produce		json
//	@Param			organizationId		path		string	true	"Organization ID"
//	@Param			projectId			path		string	true	"Project ID"
//	@Param			stackId				path		string	true	"Stack ID"
//	@Param			projectNamespace	path		string	true	"Project Namespace"
//	@Success		200					{object}	domain.GetProjectNamespaceResourcesUsageResponse
//	@Router			/organizations/{organizationId}/projects/{projectId}/namespaces/{projectNamespace}/stacks/{stackId}/resources-usage [get]
//	@Security		JWT
func (p ProjectHandler) GetProjectNamespaceResourcesUsage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("organizationId not found in path"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	projectId, ok := vars["projectId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("projectId not found in path"), "C_INVALID_PROJECT_ID", ""))
		return
	}

	projectNamespace, ok := vars["projectNamespace"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid projectNamespace"), "C_INVALID_PROJECT_NAMESPACE", ""))
		return
	}
	stackId, ok := vars["stackId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid stackId"), "C_INVALID_STACK_ID", ""))
		return
	}

	resourcesUsage, err := p.usecase.GetResourcesUsage(r.Context(), organizationId, projectId, projectNamespace, domain.StackId(stackId))
	if err != nil {
		log.Error(r.Context(), "Failed to get project resources.", err)
		ErrorJSON(w, r, err)
		return
	}

	var out domain.GetProjectNamespaceResourcesUsageResponse
	if err = serializer.Map(r.Context(), resourcesUsage, &out.ResourcesUsage); err != nil {
		log.Error(r.Context(), err)
	}
	ResponseJSON(w, r, http.StatusOK, out)
}

// GetProjectNamespaceKubeconfig godoc
//
//	@Tags			Projects
//	@Summary		Get project namespace kubeconfig
//	@Description	Get project namespace kubeconfig
//	@Accept			json
//	@Produce		json
//	@Param			organizationId		path		string	true	"Organization ID"
//	@Param			projectId			path		string	true	"Project ID"
//	@Param			stackId				path		string	true	"Stack ID"
//	@Param			projectNamespace	path		string	true	"Project Namespace"
//	@Success		200					{object}	domain.GetProjectNamespaceKubeConfigResponse
//	@Router			/organizations/{organizationId}/projects/{projectId}/namespaces/{projectNamespace}/stacks/{stackId}/kubeconfig [get]
//	@Security		JWT
func (p ProjectHandler) GetProjectNamespaceKubeconfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("organizationId not found in path"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	projectId, ok := vars["projectId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("projectId not found in path"), "C_INVALID_PROJECT_ID", ""))
		return
	}

	projectNamespace, ok := vars["projectNamespace"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid projectNamespace"), "C_INVALID_PROJECT_NAMESPACE", ""))
		return
	}
	stackId, ok := vars["stackId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid stackId"), "C_INVALID_STACK_ID", ""))
		return
	}

	kubeconfig, err := p.usecase.GetProjectNamespaceKubeconfig(r.Context(), organizationId, projectId, projectNamespace, domain.StackId(stackId))
	if err != nil {
		log.Error(r.Context(), "Failed to get project kubeconfig.", err)
		ErrorJSON(w, r, err)
		return
	}

	out := domain.GetProjectNamespaceKubeConfigResponse{
		KubeConfig: kubeconfig,
	}

	ResponseJSON(w, r, http.StatusOK, out)
}
