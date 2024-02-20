package repository

import (
	"database/sql"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/log"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type IProjectRepository interface {
	CreateProject(p *domain.Project) (string, error)
	GetProjects(organizationId string) ([]domain.Project, error)
	GetProjectById(organizationId string, projectId string) (*domain.Project, error)
	GetProjectByIdAndLeader(organizationId string, projectId string) (*domain.Project, error)
	GetProjectByName(organizationId string, projectName string) (*domain.Project, error)
	UpdateProject(p *domain.Project) error
	GetAllProjectRoles() ([]domain.ProjectRole, error)
	GetProjectRoleByName(name string) (*domain.ProjectRole, error)
	GetProjectRoleById(id string) (*domain.ProjectRole, error)
	AddProjectMember(*domain.ProjectMember) (string, error)
	GetProjectMembersByProjectId(projectId string) ([]domain.ProjectMember, error)
	GetProjectMembersByProjectIdAndRoleName(projectId string, memberRole string) ([]domain.ProjectMember, error)
	GetProjectMemberCountByProjectId(projectId string) (*domain.GetProjectMemberCountResponse, error)
	GetProjectMemberById(projectMemberId string) (*domain.ProjectMember, error)
	RemoveProjectMember(projectMemberId string) error
	UpdateProjectMemberRole(pm *domain.ProjectMember) error
	CreateProjectNamespace(organizationId string, pn *domain.ProjectNamespace) error
	GetProjectNamespaceByName(organizationId string, projectId string, stackId string, projectNamespace string) (*domain.ProjectNamespace, error)
	GetProjectNamespaces(organizationId string, projectId string) ([]domain.ProjectNamespace, error)
	GetProjectNamespaceByPrimaryKey(organizationId string, projectId string, projectNamespace string, stackId string) (*domain.ProjectNamespace, error)
	UpdateProjectNamespace(pn *domain.ProjectNamespace) error
	DeleteProjectNamespace(organizationId string, projectId string, projectNamespace string, stackId string) error
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

func (r *ProjectRepository) GetProjects(organizationId string) (ps []domain.Project, err error) {
	res := r.db.Where("organization_id = ?", organizationId).
		Preload("ProjectMembers").
		Preload("ProjectMembers.ProjectRole").
		Preload("ProjectMembers.ProjectUser").
		Preload("ProjectNamespaces").
		Find(&ps)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			log.Info("Cannot find project")
			return nil, nil
		} else {
			log.Error(res.Error)
			return nil, res.Error
		}
	}

	return ps, nil
}

func (r *ProjectRepository) GetProjectById(organizationId string, projectId string) (p *domain.Project, err error) {
	res := r.db.Limit(1).Where("organization_id = ? and id = ?", organizationId, projectId).First(&p)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			log.Info("Cannot find project")
			return nil, nil
		} else {
			log.Error(res.Error)
			return nil, res.Error
		}
	}

	return p, nil
}

func (r *ProjectRepository) GetProjectByIdAndLeader(organizationId string, projectId string) (p *domain.Project, err error) {
	res := r.db.Limit(1).
		Preload("ProjectMembers", "is_project_leader = ?", true).
		Preload("ProjectMembers.ProjectRole").
		Preload("ProjectMembers.ProjectUser").
		First(&p, "organization_id = ? and id = ?", organizationId, projectId)

	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			log.Info("Cannot find project")
			return nil, nil
		} else {
			log.Error(res.Error)
			return nil, res.Error
		}
	}

	return p, nil
}

func (r *ProjectRepository) GetProjectByName(organizationId string, projectName string) (p *domain.Project, err error) {
	res := r.db.Limit(1).
		Where("organization_id = ? and name = ?", organizationId, projectName).
		First(&p)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			log.Info("Not found project name")
			return nil, nil
		} else {
			log.Error(res.Error)
			return nil, res.Error
		}
	}

	return p, nil
}

func (r *ProjectRepository) UpdateProject(p *domain.Project) error {
	res := r.db.Model(&p).Updates(domain.Project{Name: p.Name, Description: p.Description, UpdatedAt: p.UpdatedAt})
	if res.Error != nil {
		return res.Error
	}

	return nil
}

func (r *ProjectRepository) GetProjectRoleById(id string) (*domain.ProjectRole, error) {
	var pr = &domain.ProjectRole{ID: id}
	res := r.db.First(pr)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			log.Info("Cannot find project role")
			return nil, nil
		} else {
			log.Error(res.Error)
			return nil, res.Error
		}
	}

	return pr, nil
}

func (r *ProjectRepository) GetAllProjectRoles() (prs []domain.ProjectRole, err error) {
	res := r.db.Find(&prs)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			log.Info("Cannot find project roles")
			return nil, nil
		} else {
			log.Error(res.Error)
			return nil, res.Error
		}
	}

	return prs, nil
}

func (r *ProjectRepository) GetProjectRoleByName(name string) (pr *domain.ProjectRole, err error) {
	res := r.db.Where("name = ?", name).First(&pr)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			log.Info("Cannot find project roles")
			return nil, nil
		} else {
			log.Error(res.Error)
			return nil, res.Error
		}
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
	//res := r.db.Preload("ProjectUser").
	//	Joins("ProjectRole").Where("project_id = ?", projectId).Find(&pms)
	res := r.db.Joins("ProjectUser").
		Joins("ProjectRole").
		Where("project_members.project_id = ?", projectId).
		Order("project_members.created_at ASC").
		Find(&pms)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			log.Info("Cannot find project member")
			return nil, nil
		} else {
			log.Error(res.Error)
			return nil, res.Error
		}
	}

	return pms, nil
}

func (r *ProjectRepository) GetProjectMembersByProjectIdAndRoleName(projectId string, memberRole string) (pms []domain.ProjectMember, err error) {
	res := r.db.Joins("ProjectUser").
		//Joins("ProjectRole").
		//Find(&pms, "project_members.project_id = ? and \"ProjectRole\".name = ?", projectId, memberRole)
		InnerJoins("ProjectRole", r.db.Where(&domain.ProjectRole{Name: memberRole})).
		Order("project_members.created_at ASC").
		Find(&pms, "project_members.project_id = ?", projectId)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			log.Info("Cannot find project member")
			return nil, nil
		} else {
			log.Error(res.Error)
			return nil, res.Error
		}
	}

	return pms, nil
}

func (r *ProjectRepository) GetProjectMemberCountByProjectId(projectId string) (pmcr *domain.GetProjectMemberCountResponse, err error) {
	res := r.db.Raw(""+
		"select (plc.count + pmc.count + pvc.count) as project_member_all_count,"+
		"       plc.count as project_leader_count,"+
		"       pmc.count as project_member_count,"+
		"       pvc.count as project_viewer_count"+
		"  from (select count(project_members.id) as count"+
		"          from project_members"+
		"          left join project_roles on project_roles.id = project_members.project_role_id"+
		"         where project_members.project_id = @projectId"+
		"           and project_roles.name = 'project-leader') as plc,"+
		"       (select count(project_members.id) as count"+
		"          from project_members"+
		"          left join project_roles on project_roles.id = project_members.project_role_id"+
		"         where project_members.project_id = @projectId"+
		"           and project_roles.name = 'project-member') as pmc,"+
		"       (select count(project_members.id) as count"+
		"          from project_members"+
		"          left join project_roles on project_roles.id = project_members.project_role_id"+
		"         where project_members.project_id = @projectId"+
		"           and project_roles.name = 'project-viewer') as pvc", sql.Named("projectId", projectId)).
		Scan(&pmcr)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			log.Info("Cannot find project member count")
			return nil, nil
		} else {
			log.Error(res.Error)
			return nil, res.Error
		}
	}

	return pmcr, nil
}

func (r *ProjectRepository) GetProjectMemberById(projectMemberId string) (pm *domain.ProjectMember, err error) {
	res := r.db.Preload("ProjectUser").
		Joins("ProjectRole").Where("project_members.id = ?", projectMemberId).First(&pm)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			log.Info("Cannot find project member")
			return nil, nil
		} else {
			log.Error(res.Error)
			return nil, res.Error
		}
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

//func (r *ProjectRepository) UpdateProjectMemberRole(projectMemberId string, projectRoleId string) error {
//	res := r.db.Model(&domain.ProjectMember{ID: projectMemberId}).Update("project_role_id", projectRoleId)
//	if res.Error != nil {
//		return res.Error
//	}
//
//	return nil
//}

func (r *ProjectRepository) UpdateProjectMemberRole(pm *domain.ProjectMember) error {
	res := r.db.Model(&pm).Updates(domain.ProjectMember{ProjectRoleId: pm.ProjectRoleId, UpdatedAt: pm.UpdatedAt})
	if res.Error != nil {
		return res.Error
	}

	return nil
}

func (r *ProjectRepository) CreateProjectNamespace(organizationId string, pn *domain.ProjectNamespace) error {
	res := r.db.Create(&pn)
	if res.Error != nil {
		return res.Error
	}

	//return pn.ID, nil
	return nil
}

func (r *ProjectRepository) GetProjectNamespaceByName(organizationId string, projectId string, stackId string,
	projectNamespace string) (pn *domain.ProjectNamespace, err error) {
	res := r.db.Limit(1).
		Where("stack_id = ? and namespace = ? and project_id = ?", stackId, projectNamespace, projectId).
		First(&pn)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			log.Info("Not found project namespace")
			return nil, nil
		} else {
			log.Error(res.Error)
			return nil, res.Error
		}
	}

	return pn, nil
}

func (r *ProjectRepository) GetProjectNamespaces(organizationId string, projectId string) (pns []domain.ProjectNamespace, err error) {
	res := r.db.Where("project_id = ?", projectId).
		Preload("Stack").
		Find(&pns)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			log.Info("Not found project namespaces")
			return nil, nil
		} else {
			log.Error(res.Error)
			return nil, res.Error
		}
	}

	return pns, nil
}

func (r *ProjectRepository) GetProjectNamespaceByPrimaryKey(organizationId string, projectId string,
	projectNamespace string, stackId string) (pn *domain.ProjectNamespace, err error) {
	res := r.db.Limit(1).
		Where("stack_id = ? and namespace = ? and project_id = ?", stackId, projectNamespace, projectId).
		Preload("Stack").
		First(&pn)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			log.Info("Not found project namespace")
			return nil, nil
		} else {
			log.Error(res.Error)
			return nil, res.Error
		}
	}

	return pn, nil
}

func (r *ProjectRepository) UpdateProjectNamespace(pn *domain.ProjectNamespace) error {
	res := r.db.Model(&pn).Updates(domain.ProjectNamespace{Description: pn.Description, UpdatedAt: pn.UpdatedAt})
	if res.Error != nil {
		return res.Error
	}

	return nil
}

func (r *ProjectRepository) DeleteProjectNamespace(organizationId string, projectId string, projectNamespace string,
	stackId string) error {
	res := r.db.Where("stack_id = ? and namespace = ? and project_id = ?", stackId, projectNamespace, projectId).
		Delete(&domain.ProjectNamespace{StackId: stackId, Namespace: projectNamespace})
	if res.Error != nil {
		return res.Error
	}

	return nil
}
