package usecase

import (
	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-api/internal/serializer"
	argowf "github.com/openinfradev/tks-api/pkg/argo-client"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/log"
	"github.com/pkg/errors"
)

const (
	ProjectAll int = iota
	ProjectLeader
	ProjectMember
	ProjectViewer
)

type IProjectUsecase interface {
	CreateProject(*domain.Project) (string, error)
	GetProjects(organizationId string) ([]domain.Project, error)
	GetProject(organizationId string, projectId string) (*domain.Project, error)
	GetProjectWithLeader(organizationId string, projectId string) (*domain.Project, error)
	UpdateProject(p *domain.Project) error
	GetProjectRole(id string) (*domain.ProjectRole, error)
	GetProjectRoles(int) ([]domain.ProjectRole, error)
	AddProjectMember(pm *domain.ProjectMember) (string, error)
	GetProjectUser(projectUserId string) (*domain.ProjectUser, error)
	GetProjectMember(projectMemberId string) (*domain.ProjectMember, error)
	GetProjectMembers(projectId string, query int) ([]domain.ProjectMember, error)
	GetProjectMemberCount(projectMemberId string) (*domain.GetProjectMemberCountResponse, error)
	RemoveProjectMember(projectMemberId string) error
	UpdateProjectMemberRole(pm *domain.ProjectMember) error
	CreateProjectNamespace(organizationId string, pn *domain.ProjectNamespace) error
	IsProjectNamespaceExist(organizationId string, projectId string, stackId string, projectNamespace string) (bool, error)
	GetProjectNamespaces(organizationId string, projectId string) ([]domain.ProjectNamespace, error)
	GetProjectNamespace(organizationId string, projectId string, projectNamespace string, stackId string) (*domain.ProjectNamespace, error)
	DeleteProjectNamespace(organizationId string, projectId string, projectNamespace string, stackId string) error
}

type ProjectUsecase struct {
	projectRepo            repository.IProjectRepository
	userRepository         repository.IUserRepository
	authRepository         repository.IAuthRepository
	clusterRepository      repository.IClusterRepository
	appgroupRepository     repository.IAppGroupRepository
	organizationRepository repository.IOrganizationRepository
	argo                   argowf.ArgoClient
}

func NewProjectUsecase(r repository.Repository, argoClient argowf.ArgoClient) IProjectUsecase {
	return &ProjectUsecase{
		projectRepo:            r.Project,
		userRepository:         r.User,
		authRepository:         r.Auth,
		clusterRepository:      r.Cluster,
		appgroupRepository:     r.AppGroup,
		organizationRepository: r.Organization,
		argo:                   argoClient,
	}
}

func (u *ProjectUsecase) CreateProject(p *domain.Project) (string, error) {
	projectId, err := u.projectRepo.CreateProject(p)
	if err != nil {
		log.Error(err)
		return "", errors.Wrap(err, "Failed to create project.")
	}
	return projectId, nil
}

func (u *ProjectUsecase) GetProjects(organizationId string) (ps []domain.Project, err error) {
	ps, err = u.projectRepo.GetProjects(organizationId)
	if err != nil {
		log.Error(err)
		return nil, errors.Wrap(err, "Failed to get projects.")
	}
	return ps, err
}

func (u *ProjectUsecase) GetProject(organizationId string, projectId string) (*domain.Project, error) {
	p, err := u.projectRepo.GetProjectById(organizationId, projectId)
	if err != nil {
		log.Error(err)
		return nil, errors.Wrap(err, "Failed to get projects.")
	}
	return p, err
}

func (u *ProjectUsecase) GetProjectWithLeader(organizationId string, projectId string) (*domain.Project, error) {
	p, err := u.projectRepo.GetProjectByIdAndLeader(organizationId, projectId)
	if err != nil {
		log.Error(err)
		return nil, errors.Wrap(err, "Failed to get projects.")
	}
	return p, err
}

func (u *ProjectUsecase) UpdateProject(p *domain.Project) error {
	if err := u.projectRepo.UpdateProject(p); err != nil {
		log.Error(err)
		return errors.Wrap(err, "Failed to update project.")
	}
	return nil
}

func (u *ProjectUsecase) GetProjectRole(id string) (*domain.ProjectRole, error) {
	pr, err := u.projectRepo.GetProjectRoleById(id)
	if err != nil {
		log.Error(err)
		return nil, errors.Wrap(err, "Failed to get project roles.")
	}

	return pr, nil
}

func (u *ProjectUsecase) GetProjectRoles(query int) (prs []domain.ProjectRole, err error) {
	var pr *domain.ProjectRole

	if query == ProjectLeader {
		pr, err = u.projectRepo.GetProjectRoleByName("project-leader")
	} else if query == ProjectMember {
		pr, err = u.projectRepo.GetProjectRoleByName("project-member")
	} else if query == ProjectViewer {
		pr, err = u.projectRepo.GetProjectRoleByName("project-viewer")
	} else {
		prs, err = u.projectRepo.GetAllProjectRoles()
	}
	if err != nil {
		log.Error(err)
		return nil, errors.Wrap(err, "Failed to get project roles.")
	}

	if pr != nil {
		prs = append(prs, *pr)
	}

	return prs, nil
}

func (u *ProjectUsecase) AddProjectMember(pm *domain.ProjectMember) (string, error) {
	projectMemberId, err := u.projectRepo.AddProjectMember(pm)
	if err != nil {
		log.Error(err)
		return "", errors.Wrap(err, "Failed to add project member to project.")
	}
	return projectMemberId, nil
}

func (u *ProjectUsecase) GetProjectUser(projectUserId string) (*domain.ProjectUser, error) {
	var uid uuid.UUID
	uid, err := uuid.Parse(projectUserId)
	if err != nil {
		log.Error(err)
		return nil, errors.Wrap(err, "Failed to parse uuid to string")
	}

	user, err := u.userRepository.GetByUuid(uid)
	if err != nil {
		log.Error(err)
		return nil, errors.Wrap(err, "Failed to retrieve user by id")
	}
	var pu domain.ProjectUser
	if err = serializer.Map(user, &pu); err != nil {
		log.Error(err)
		return nil, err
	}
	return &pu, nil
}

func (u *ProjectUsecase) GetProjectMember(projectMemberId string) (pm *domain.ProjectMember, err error) {
	pm, err = u.projectRepo.GetProjectMemberById(projectMemberId)
	if err != nil {
		log.Error(err)
		return pm, errors.Wrap(err, "Failed to get project member.")
	}

	return pm, nil
}

func (u *ProjectUsecase) GetProjectMembers(projectId string, query int) (pms []domain.ProjectMember, err error) {
	if query == ProjectLeader {
		pms, err = u.projectRepo.GetProjectMembersByProjectIdAndRoleName(projectId, "project-leader")
	} else if query == ProjectMember {
		pms, err = u.projectRepo.GetProjectMembersByProjectIdAndRoleName(projectId, "project-member")
	} else if query == ProjectViewer {
		pms, err = u.projectRepo.GetProjectMembersByProjectIdAndRoleName(projectId, "project-viewer")
	} else {
		pms, err = u.projectRepo.GetProjectMembersByProjectId(projectId)
	}
	if err != nil {
		log.Error(err)
		return nil, errors.Wrap(err, "Failed to get project members.")
	}

	return pms, nil
}

func (u *ProjectUsecase) GetProjectMemberCount(projectMemberId string) (pmcr *domain.GetProjectMemberCountResponse, err error) {
	pmcr, err = u.projectRepo.GetProjectMemberCountByProjectId(projectMemberId)
	if err != nil {
		log.Error(err)
		return pmcr, errors.Wrap(err, "Failed to get project member count.")
	}

	return pmcr, nil
}

func (u *ProjectUsecase) RemoveProjectMember(projectMemberId string) error {
	if err := u.projectRepo.RemoveProjectMember(projectMemberId); err != nil {
		log.Error(err)
		return errors.Wrap(err, "Failed to remove project member to project.")
	}
	return nil
}

func (u *ProjectUsecase) UpdateProjectMemberRole(pm *domain.ProjectMember) error {

	if err := u.projectRepo.UpdateProjectMemberRole(pm); err != nil {
		log.Error(err)
		return errors.Wrap(err, "Failed to remove project member to project.")
	}
	return nil
}

func (u *ProjectUsecase) CreateProjectNamespace(organizationId string, pn *domain.ProjectNamespace) error {
	if err := u.projectRepo.CreateProjectNamespace(organizationId, pn); err != nil {
		log.Error(err)
		return errors.Wrap(err, "Failed to create project namespace.")
	}
	return nil
}

func (u *ProjectUsecase) IsProjectNamespaceExist(organizationId string, projectId string, stackId string, projectNamespace string) (bool, error) {
	exist := true
	pn, err := u.projectRepo.GetProjectNamespaceByName(organizationId, projectId, stackId, projectNamespace)
	if err != nil {
		log.Error(err)
		exist = false
		return exist, errors.Wrap(err, "Failed to retrieve project namespace.")
	}
	if pn == nil {
		exist = false
	}
	return exist, nil
}

func (u *ProjectUsecase) GetProjectNamespaces(organizationId string, projectId string) ([]domain.ProjectNamespace, error) {
	pns, err := u.projectRepo.GetProjectNamespaces(organizationId, projectId)
	if err != nil {
		log.Error(err)
		return nil, errors.Wrap(err, "Failed to retrieve project namespaces.")
	}

	return pns, nil
}

func (u *ProjectUsecase) GetProjectNamespace(organizationId string, projectId string, projectNamespace string, stackId string) (*domain.ProjectNamespace, error) {
	pn, err := u.projectRepo.GetProjectNamespaceByPrimaryKey(organizationId, projectId, projectNamespace, stackId)
	if err != nil {
		log.Error(err)
		return nil, errors.Wrap(err, "Failed to retrieve project namespace.")
	}

	return pn, nil
}

func (u *ProjectUsecase) DeleteProjectNamespace(organizationId string, projectId string,
	stackId string, projectNamespace string) error {
	if err := u.projectRepo.DeleteProjectNamespace(organizationId, projectId, projectNamespace, stackId); err != nil {
		log.Error(err)
		return errors.Wrap(err, "Failed to delete project namespace.")
	}
	return nil
}
