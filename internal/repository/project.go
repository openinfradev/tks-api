package repository

import (
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/log"
	"gorm.io/gorm"
)

type IProjectRepository interface {
	CreateProject(p *domain.Project) (string, error)
	GetAllProjectRoles() ([]domain.ProjectRole, error)
	GetProjectRoleByName(name string) (*domain.ProjectRole, error)
	GetProjectRoleById(id string) (*domain.ProjectRole, error)
	AddProjectMember(pm *domain.ProjectMember) (string, error)
	GetProjectMembersByProjectId(projectId string) ([]domain.ProjectMember, error)
	GetProjectMemberById(projectMemberId string) (pm domain.ProjectMember, err error)
	RemoveProjectMember(projectMemberId string) error
	UpdateProjectMemberRole(projectMemberId string, projectRoleId string) error
}

type ProjectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) IProjectRepository {
	return &ProjectRepository{
		db: db,
	}
}

func (r *ProjectRepository) CreateProject(p *domain.Project) (string, error) {
	res := r.db.Create(&p)
	if res.Error != nil {
		return "", res.Error
	}

	return p.ID, nil
}

func (r *ProjectRepository) GetProjectRoleById(id string) (*domain.ProjectRole, error) {
	var pr = &domain.ProjectRole{ID: id}
	result := r.db.First(pr)
	if result.Error != nil {
		log.Error(result.Error)
		return pr, result.Error
	}
	if result.RowsAffected == 0 {
		log.Info("There is no project_roles table data")
		return pr, nil
	}

	return pr, nil
}

func (r *ProjectRepository) GetAllProjectRoles() (prs []domain.ProjectRole, err error) {
	result := r.db.Find(&prs)
	if result.Error != nil {
		log.Error(result.Error)
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		log.Info("There is no project_roles table data")
		return prs, nil
	}

	return prs, nil
}

func (r *ProjectRepository) GetProjectRoleByName(name string) (pr *domain.ProjectRole, err error) {
	log.Infof("project role name: %s", name)
	result := r.db.Where("name = ?", name).First(&pr)
	if result.Error != nil {
		log.Error(result.Error)
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		log.Info("There is no project_roles table data")
		return pr, nil
	}

	return pr, nil
}

func (r *ProjectRepository) AddProjectMember(pm *domain.ProjectMember) (string, error) {
	res := r.db.Create(&pm)
	if res.Error != nil {
		return "", res.Error
	}

	return pm.ID, nil
}

func (r *ProjectRepository) GetProjectMembersByProjectId(projectId string) (pms []domain.ProjectMember, err error) {
	result := r.db.Joins("ProjectRole").Where("project_id = ?", projectId).Find(&pms)
	if result.Error != nil {
		log.Error(result.Error)
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		log.Info("Cannot find project member")
		return pms, nil
	}

	return pms, nil
}

func (r *ProjectRepository) GetProjectMemberById(projectMemberId string) (pm domain.ProjectMember, err error) {
	result := r.db.Joins("ProjectRole").Where("project_members.id = ?", projectMemberId).First(&pm)
	if result.Error != nil {
		log.Error(result.Error)
		return pm, result.Error
	}
	if result.RowsAffected == 0 {
		log.Info("Cannot find project member")
		return pm, nil
	}

	return pm, nil
}

func (r *ProjectRepository) RemoveProjectMember(projectMemberId string) error {
	res := r.db.Delete(&domain.ProjectMember{ID: projectMemberId})
	if res.Error != nil {
		return res.Error
	}

	return nil
}

func (r *ProjectRepository) UpdateProjectMemberRole(projectMemberId string, projectRoleId string) error {
	res := r.db.Model(&domain.ProjectMember{ID: projectMemberId}).Update("project_role_id", projectRoleId)
	if res.Error != nil {
		return res.Error
	}

	return nil
}
