package repository

import (
	"database/sql"

	"github.com/google/uuid"

	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/log"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type IProjectRepository interface {
	CreateProject(p *model.Project) (string, error)
	GetProjects(organizationId string, userId uuid.UUID, pg *pagination.Pagination) ([]domain.ProjectResponse, error)
	GetProjectsByUserId(organizationId string, userId uuid.UUID, pg *pagination.Pagination) ([]domain.ProjectResponse, error)
	GetAllProjects(organizationId string, pg *pagination.Pagination) (pr []domain.ProjectResponse, err error)
	GetProjectById(organizationId string, projectId string) (*model.Project, error)
	GetProjectByIdAndLeader(organizationId string, projectId string) (*model.Project, error)
	GetProjectByName(organizationId string, projectName string) (*model.Project, error)
	UpdateProject(p *model.Project) error
	GetAllProjectRoles() ([]model.ProjectRole, error)
	GetProjectRoleByName(name string) (*model.ProjectRole, error)
	GetProjectRoleById(id string) (*model.ProjectRole, error)
	AddProjectMember(*model.ProjectMember) (string, error)
	GetProjectMembersByProjectId(projectId string, pg *pagination.Pagination) ([]model.ProjectMember, error)
	GetProjectMembersByProjectIdAndRoleName(projectId string, memberRole string, pg *pagination.Pagination) ([]model.ProjectMember, error)
	GetProjectMemberCountByProjectId(projectId string) (*domain.GetProjectMemberCountResponse, error)
	GetProjectMemberById(projectMemberId string) (*model.ProjectMember, error)
	GetProjectMemberByUserId(projectId string, projectUserId string) (pm *model.ProjectMember, err error)
	RemoveProjectMember(projectMemberId string) error
	UpdateProjectMemberRole(pm *model.ProjectMember) error
	CreateProjectNamespace(organizationId string, pn *model.ProjectNamespace) error
	GetProjectNamespaceByName(organizationId string, projectId string, stackId string, projectNamespace string) (*model.ProjectNamespace, error)
	GetProjectNamespaces(organizationId string, projectId string, pg *pagination.Pagination) ([]model.ProjectNamespace, error)
	GetProjectNamespaceByPrimaryKey(organizationId string, projectId string, projectNamespace string, stackId string) (*model.ProjectNamespace, error)
	UpdateProjectNamespace(pn *model.ProjectNamespace) error
	DeleteProjectNamespace(organizationId string, projectId string, projectNamespace string, stackId string) error
	GetAppCountByProjectId(organizationId string, projectId string) (int, error)
	GetAppCountByNamespace(organizationId string, projectId string, namespace string) (int, error)
}

type ProjectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) IProjectRepository {
	return &ProjectRepository{
		db: db,
	}
}

func (r *ProjectRepository) CreateProject(p *model.Project) (string, error) {
	res := r.db.Create(&p)
	if res.Error != nil {
		return "", res.Error
	}

	return p.ID, nil
}

func (r *ProjectRepository) GetProjects(organizationId string, userId uuid.UUID, pg *pagination.Pagination) (pr []domain.ProjectResponse, err error) {
	if pg == nil {
		pg = pagination.NewPagination(nil)
	}
	res := r.db.Raw(""+
		"select distinct p.id as id, p.organization_id as organization_id, p.name as name, p.description as description, p.created_at as created_at, "+
		"       true as is_my_project, pm.project_role_id as project_role_id, pm.pr_name as project_role_name, "+
		"       pn.count as namespace_count, asa.count as app_count, pm_count.count as member_count "+
		"  from projects as p "+
		"  left join "+
		"       (select pm.project_id as project_id, pm.project_user_id as project_user_id, pm.project_role_id as project_role_id, "+
		"               pm.created_at as created_at, pm.is_project_leader as is_project_leader, "+
		"               pr.name as pr_name "+
		"          from project_members as pm "+
		"          left join project_roles as pr on pr.id = pm.project_role_id "+
		"          left join users on users.id = pm.project_user_id "+
		"         where pm.project_user_id = @userId) as pm on p.id = pm.project_id "+
		"  left join "+
		"       (select p.id as project_id, count(pn.stack_id || pn.project_id) as count "+
		"          from project_namespaces as pn "+
		"          left join projects as p on pn.project_id = p.id "+
		"          left join project_members as pm on pn.project_id = pm.project_id "+
		"         where p.organization_id = @organizationId "+
		"           and pm.project_user_id = @userId "+
		"         group by p.id) as pn on p.id = pn.project_id "+
		"  left join "+
		"       (select p.id as project_id, count(asa.id) as count "+
		"          from app_serve_apps as asa "+
		"          left join projects as p on asa.project_id = p.id "+
		"          left join project_members as pm on asa.project_id = pm.project_id "+
		"         where p.organization_id = @organizationId "+
		"           and pm.project_user_id = @userId "+
		"         group by p.id) as asa on p.id = asa.project_id "+
		"  left join "+
		"       (select p.id as project_id, count(pm.id) as count "+
		"          from project_members as pm "+
		"          left join projects as p on pm.project_id = p.id "+
		"         where p.organization_id = @organizationId "+
		"           and pm.project_user_id = @userId "+
		"         group by p.id) as pm_count on p.id = pm_count.project_id "+
		" where p.id = pm.project_id "+
		"   and p.organization_id = @organizationId "+
		"union "+
		"select distinct p.id as id, p.organization_id as organization_id, p.name as name, p.description as description, p.created_at as created_at, "+
		"       false as is_my_project, '' as project_role_id, '' as project_role_name, "+
		"       pn.count as namespace_count, asa.count as app_count, pm_count.count as member_count "+
		"  from projects as p "+
		"  left join "+
		"       (select pm.project_id as project_id, pm.project_user_id as project_user_id, pm.project_role_id as project_role_id, "+
		"               pm.created_at as created_at, pm.is_project_leader as is_project_leader, "+
		"               pr.name as pr_name "+
		"          from project_members as pm "+
		"          left join project_roles as pr on pr.id = pm.project_role_id "+
		"          left join users on users.id = pm.project_user_id "+
		"         where pm.project_user_id <> @userId) as pm on p.id = pm.project_id "+
		"  left join "+
		"       (select p.id as project_id, count(pn.stack_id || pn.project_id) as count "+
		"          from project_namespaces as pn "+
		"          left join projects as p on pn.project_id = p.id "+
		"          left join project_members as pm on pn.project_id = pm.project_id "+
		"         where p.organization_id = @organizationId "+
		"           and pm.project_user_id <> @userId "+
		"         group by p.id) as pn on p.id = pn.project_id "+
		"  left join "+
		"       (select p.id as project_id, count(asa.id) as count "+
		"          from app_serve_apps as asa "+
		"          left join projects as p on asa.project_id = p.id "+
		"          left join project_members as pm on asa.project_id = pm.project_id "+
		"         where p.organization_id = @organizationId "+
		"           and pm.project_user_id <> @userId "+
		"         group by p.id) as asa on p.id = asa.project_id "+
		"  left join "+
		"       (select p.id as project_id, count(pm.id) as count "+
		"          from project_members as pm "+
		"          left join projects as p on pm.project_id = p.id "+
		"         where p.organization_id = @organizationId "+
		"           and pm.project_user_id <> @userId "+
		"         group by p.id) as pm_count on p.id = pm_count.project_id"+
		" where p.id = pm.project_id "+
		"   and p.organization_id = @organizationId "+
		"   and p.id not in (select projects.id "+
		"                      from projects "+
		"                      left join project_members on project_members.project_id = projects.id "+
		"                     where project_members.project_user_id = @userId) ",
		sql.Named("organizationId", organizationId), sql.Named("userId", userId)).
		Scan(&pr)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			log.Info("Cannot find project")
		}
	}
	return pr, nil
}

func (r *ProjectRepository) GetProjectsByUserId(organizationId string, userId uuid.UUID, pg *pagination.Pagination) (pr []domain.ProjectResponse, err error) {
	if pg == nil {
		pg = pagination.NewPagination(nil)
	}
	res := r.db.Raw(""+
		"select distinct p.id as id, p.organization_id as organization_id, p.name as name, p.description as description, p.created_at as created_at, "+
		"       true as is_my_project, pm.project_role_id as project_role_id, pm.pr_name as project_role_name, "+
		"       pn.count as namespace_count, asa.count as app_count, pm_count.count as member_count "+
		"  from projects as p "+
		"  left join "+
		"       (select pm.project_id as project_id, pm.project_user_id as project_user_id, pm.project_role_id as project_role_id, "+
		"               pm.created_at as created_at, pm.is_project_leader as is_project_leader, "+
		"               pr.name as pr_name "+
		"          from project_members as pm "+
		"          left join project_roles as pr on pr.id = pm.project_role_id "+
		"          left join users on users.id = pm.project_user_id "+
		"         where pm.project_user_id = @userId) as pm on p.id = pm.project_id "+
		"  left join "+
		"       (select p.id as project_id, count(pn.stack_id || pn.project_id) as count "+
		"          from project_namespaces as pn "+
		"          left join projects as p on pn.project_id = p.id "+
		"          left join project_members as pm on pn.project_id = pm.project_id "+
		"         where p.organization_id = @organizationId "+
		"           and pm.project_user_id = @userId "+
		"         group by p.id) as pn on p.id = pn.project_id "+
		"  left join "+
		"       (select p.id as project_id, count(asa.id) as count "+
		"          from app_serve_apps as asa "+
		"          left join projects as p on asa.project_id = p.id "+
		"          left join project_members as pm on asa.project_id = pm.project_id "+
		"         where p.organization_id = @organizationId "+
		"           and pm.project_user_id = @userId "+
		"         group by p.id) as asa on p.id = asa.project_id "+
		"  left join "+
		"       (select p.id as project_id, count(pm.id) as count "+
		"          from project_members as pm "+
		"          left join projects as p on pm.project_id = p.id "+
		"         where p.organization_id = @organizationId "+
		"           and pm.project_user_id = @userId "+
		"         group by p.id) as pm_count on p.id = pm_count.project_id "+
		" where p.id = pm.project_id "+
		"   and p.organization_id = @organizationId", sql.Named("organizationId", organizationId), sql.Named("userId", userId)).
		Scan(&pr)

	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			log.Info("Cannot find project")
			return nil, nil
		} else {
			log.Error(res.Error)
			return nil, res.Error
		}
	}

	return pr, nil
}

func (r *ProjectRepository) GetAllProjects(organizationId string, pg *pagination.Pagination) (pr []domain.ProjectResponse, err error) {
	if pg == nil {
		pg = pagination.NewPagination(nil)
	}
	res := r.db.Raw(""+
		"select distinct p.id as id, p.organization_id as organization_id, p.name as name, p.description as description, p.created_at as created_at, "+
		"       false as is_my_project, pm.project_role_id as project_role_id, pm.pr_name as project_role_name, "+
		"       pn.count as namespace_count, asa.count as app_count, pm_count.count as member_count "+
		"  from projects as p "+
		"  left join "+
		"       (select distinct pm.project_id as project_id, '' as project_user_id, '' as project_role_id, "+
		"               pm.created_at as created_at, pm.is_project_leader as is_project_leader, "+
		"               '' as pr_name "+
		"          from project_members as pm "+
		"          left join project_roles as pr on pr.id = pm.project_role_id "+
		"          left join users on users.id = pm.project_user_id) as pm on p.id = pm.project_id "+
		"  left join "+
		"       (select p.id as project_id, count(pn.stack_id || pn.project_id) as count "+
		"          from project_namespaces as pn "+
		"          left join projects as p on pn.project_id = p.id "+
		"          left join project_members as pm on pn.project_id = pm.project_id "+
		"         where p.organization_id = @organizationId "+
		"         group by p.id) as pn on p.id = pn.project_id "+
		"  left join "+
		"       (select p.id as project_id, count(asa.id) as count "+
		"          from app_serve_apps as asa "+
		"          left join projects as p on asa.project_id = p.id "+
		"          left join project_members as pm on asa.project_id = pm.project_id "+
		"         where p.organization_id = @organizationId "+
		"         group by p.id) as asa on p.id = asa.project_id "+
		"  left join "+
		"       (select p.id as project_id, count(pm.id) as count "+
		"          from project_members as pm "+
		"          left join projects as p on pm.project_id = p.id "+
		"         where p.organization_id = @organizationId "+
		"         group by p.id) as pm_count on p.id = pm_count.project_id "+
		" where p.id = pm.project_id "+
		"   and p.organization_id = @organizationId", sql.Named("organizationId", organizationId)).
		Scan(&pr)

	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			log.Info("Cannot find project")
			return nil, nil
		} else {
			log.Error(res.Error)
			return nil, res.Error
		}
	}

	return pr, nil
}

func (r *ProjectRepository) GetProjectById(organizationId string, projectId string) (p *model.Project, err error) {
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

func (r *ProjectRepository) GetProjectByIdAndLeader(organizationId string, projectId string) (p *model.Project, err error) {
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

func (r *ProjectRepository) GetProjectByName(organizationId string, projectName string) (p *model.Project, err error) {
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

func (r *ProjectRepository) UpdateProject(p *model.Project) error {
	res := r.db.Model(&p).Updates(model.Project{Name: p.Name, Description: p.Description, UpdatedAt: p.UpdatedAt})
	if res.Error != nil {
		return res.Error
	}

	return nil
}

func (r *ProjectRepository) GetProjectRoleById(id string) (*model.ProjectRole, error) {
	var pr = &model.ProjectRole{ID: id}
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

func (r *ProjectRepository) GetAllProjectRoles() (prs []model.ProjectRole, err error) {
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

func (r *ProjectRepository) GetProjectRoleByName(name string) (pr *model.ProjectRole, err error) {
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

func (r *ProjectRepository) AddProjectMember(pm *model.ProjectMember) (string, error) {
	res := r.db.Create(&pm)
	if res.Error != nil {
		return "", res.Error
	}

	return pm.ID, nil
}

func (r *ProjectRepository) GetProjectMembersByProjectId(projectId string, pg *pagination.Pagination) (pms []model.ProjectMember, err error) {
	if pg == nil {
		pg = pagination.NewPagination(nil)
	}
	_, res := pg.Fetch(r.db.Joins("ProjectUser").
		Joins("ProjectRole").
		Where("project_members.project_id = ?", projectId).
		Order("project_members.created_at ASC"), &pms)
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

func (r *ProjectRepository) GetProjectMembersByProjectIdAndRoleName(projectId string, memberRole string, pg *pagination.Pagination) (pms []model.ProjectMember, err error) {
	if pg == nil {
		pg = pagination.NewPagination(nil)
	}
	_, res := pg.Fetch(r.db.Joins("ProjectUser").
		InnerJoins("ProjectRole", r.db.Where(&model.ProjectRole{Name: memberRole})).
		Order("project_members.created_at ASC").
		Where("project_members.project_id = ?", projectId), &pms)

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

func (r *ProjectRepository) GetProjectMemberById(projectMemberId string) (pm *model.ProjectMember, err error) {
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

func (r *ProjectRepository) GetProjectMemberByUserId(projectId string, projectUserId string) (pm *model.ProjectMember, err error) {
	res := r.db.Preload("ProjectUser").
		Joins("ProjectRole").Where("project_id = ? and project_user_id = ?", projectId, projectUserId).First(&pm)
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
	res := r.db.Delete(&model.ProjectMember{ID: projectMemberId})
	if res.Error != nil {
		return res.Error
	}

	return nil
}

//func (r *ProjectRepository) UpdateProjectMemberRole(projectMemberId string, projectRoleId string) error {
//	res := r.db.Model(&model.ProjectMember{ID: projectMemberId}).Update("project_role_id", projectRoleId)
//	if res.Error != nil {
//		return res.Error
//	}
//
//	return nil
//}

func (r *ProjectRepository) UpdateProjectMemberRole(pm *model.ProjectMember) error {
	res := r.db.Model(&pm).Updates(
		model.ProjectMember{
			ProjectRoleId:   pm.ProjectRoleId,
			IsProjectLeader: pm.IsProjectLeader,
			UpdatedAt:       pm.UpdatedAt,
		})
	if res.Error != nil {
		return res.Error
	}

	return nil
}

func (r *ProjectRepository) CreateProjectNamespace(organizationId string, pn *model.ProjectNamespace) error {
	res := r.db.Create(&pn)
	if res.Error != nil {
		return res.Error
	}

	//return pn.ID, nil
	return nil
}

func (r *ProjectRepository) GetProjectNamespaceByName(organizationId string, projectId string, stackId string,
	projectNamespace string) (pn *model.ProjectNamespace, err error) {
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

func (r *ProjectRepository) GetProjectNamespaces(organizationId string, projectId string, pg *pagination.Pagination) (pns []model.ProjectNamespace, err error) {
	if pg == nil {
		pg = pagination.NewPagination(nil)
	}
	_, res := pg.Fetch(r.db.Where("project_id = ?", projectId).
		Preload("Stack"), &pns)
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
	projectNamespace string, stackId string) (pn *model.ProjectNamespace, err error) {
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

func (r *ProjectRepository) UpdateProjectNamespace(pn *model.ProjectNamespace) error {
	res := r.db.Model(&pn).Updates(model.ProjectNamespace{Description: pn.Description, UpdatedAt: pn.UpdatedAt})
	if res.Error != nil {
		return res.Error
	}

	return nil
}

func (r *ProjectRepository) DeleteProjectNamespace(organizationId string, projectId string, projectNamespace string,
	stackId string) error {
	res := r.db.Where("stack_id = ? and namespace = ? and project_id = ?", stackId, projectNamespace, projectId).
		Delete(&model.ProjectNamespace{StackId: stackId, Namespace: projectNamespace})
	if res.Error != nil {
		return res.Error
	}

	return nil
}

func (r *ProjectRepository) GetAppCountByProjectId(organizationId string, projectId string) (appCount int, err error) {
	res := r.db.Select("count(*) as app_count").
		Table("app_serve_apps").
		Where("organization_id = ? and project_Id = ?", organizationId, projectId).
		Find(&appCount)
	if res.Error != nil {
		log.Error(res.Error)
		return 0, res.Error
	}

	return appCount, nil
}

func (r *ProjectRepository) GetAppCountByNamespace(organizationId string, projectId string, namespace string) (appCount int, err error) {
	res := r.db.Select("count(*) as app_count").
		Table("app_serve_apps").
		Where("organization_id = ? and project_Id = ? and namespace = ?", organizationId, projectId, namespace).
		Find(&appCount)
	if res.Error != nil {
		log.Error(res.Error)
		return 0, res.Error
	}

	return appCount, nil
}
