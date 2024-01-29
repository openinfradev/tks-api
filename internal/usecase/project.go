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
	GetProjectRole(id string) (*domain.ProjectRole, error)
	GetProjectRoles(int) ([]domain.ProjectRole, error)
	AddProjectMember(pm *domain.ProjectMember) (string, error)
	GetProjectMemberById(projectMemberId string) (domain.ProjectMember, error)
	GetProjectMembersByProjectId(projectId string) ([]domain.ProjectMember, error)
	RemoveProjectMember(projectMemberId string) error
	UpdateProjectMemberRole(projectMemberId string, projectRoleId string) error
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

func (u *ProjectUsecase) GetProjectMemberById(projectMemberId string) (pm domain.ProjectMember, err error) {
	pm, err = u.projectRepo.GetProjectMemberById(projectMemberId)
	if err != nil {
		log.Error(err)
		return pm, errors.Wrap(err, "Failed to get project member.")
	}

	var uid uuid.UUID
	uid, err = uuid.Parse(pm.UserId)
	if err != nil {
		log.Error(err)
		return pm, errors.Wrap(err, "Failed to parse uuid to string")
	}
	user, err := u.userRepository.GetByUuid(uid)
	if err != nil {
		log.Error(err)
		return pm, errors.Wrap(err, "Failed to retrieve user by id")
	}
	var pu domain.ProjectUser
	if err = serializer.Map(user, &pu); err != nil {
		log.Error(err)
		return pm, err
	}

	pm.User = pu
	return pm, nil
}

func (u *ProjectUsecase) GetProjectMembersByProjectId(projectId string) ([]domain.ProjectMember, error) {
	pms, err := u.projectRepo.GetProjectMembersByProjectId(projectId)
	if err != nil {
		log.Error(err)
		return nil, errors.Wrap(err, "Failed to get project members.")
	}

	var uid uuid.UUID
	for idx, pm := range pms {
		uid, err = uuid.Parse(pm.UserId)
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
		pms[idx].User = pu
	}

	return pms, nil
}

func (u *ProjectUsecase) RemoveProjectMember(projectMemberId string) error {
	if err := u.projectRepo.RemoveProjectMember(projectMemberId); err != nil {
		log.Error(err)
		return errors.Wrap(err, "Failed to remove project member to project.")
	}
	return nil
}

func (u *ProjectUsecase) UpdateProjectMemberRole(projectMemberId string, projectRoleId string) error {
	if err := u.projectRepo.UpdateProjectMemberRole(projectMemberId, projectRoleId); err != nil {
		log.Error(err)
		return errors.Wrap(err, "Failed to remove project member to project.")
	}
	return nil
}
