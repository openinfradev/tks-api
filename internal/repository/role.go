package repository

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/pkg/domain"
	"gorm.io/gorm"
	"math"
)

//
//type Role struct {
//	gorm.Model
//
//	ID             string `gorm:"primarykey;"`
//	Name           string
//	OrganizationID string
//	Organization   Organization `gorm:"foreignKey:OrganizationID;references:ID;"`
//	Type           string
//	Creator        uuid.UUID
//	Description    string
//}
//
//func (r *Role) BeforeCreate(tx *gorm.DB) (err error) {
//	r.ID = uuid.New().String()
//	return nil
//}
//
//type TksRole struct {
//	RoleID string `gorm:"primarykey;"`
//	Role   Role   `gorm:"foreignKey:RoleID;references:ID;"`
//}

//type ProjectRole struct {
//	RoleID    string `gorm:"primarykey;"`
//	Role      Role   `gorm:"foreignKey:RoleID;references:ID;"`
//	ProjectID string
//	Project   domain.Project `gorm:"foreignKey:ProjectID;references:ID;"`
//}

type IRoleRepository interface {
	Create(roleObj interface{}) error
	List(pg *pagination.Pagination) ([]*domain.Role, error)
	ListTksRoles(organizationId string, pg *pagination.Pagination) ([]*domain.TksRole, error)
	ListProjectRoles(projectId string, pg *pagination.Pagination) ([]*domain.ProjectRole, error)
	Get(id uuid.UUID) (*domain.Role, error)
	GetTksRole(id uuid.UUID) (*domain.TksRole, error)
	GetProjectRole(id uuid.UUID) (*domain.ProjectRole, error)
	DeleteCascade(id uuid.UUID) error
	Update(roleObj interface{}) error
}

type RoleRepository struct {
	db *gorm.DB
}

func (r RoleRepository) Create(roleObj interface{}) error {
	if roleObj == nil {
		return fmt.Errorf("roleObj is nil")
	}
	switch roleObj.(type) {
	case *domain.TksRole:
		inputRole := roleObj.(*domain.TksRole)
		role := ConvertDomainToRepoTksRole(inputRole)
		if err := r.db.Create(role).Error; err != nil {
			return err
		}

	case *domain.ProjectRole:
		inputRole := roleObj.(*domain.ProjectRole)
		//role := ConvertDomainToRepoProjectRole(inputRole)
		//if err := r.db.Create(role).Error; err != nil {
		//	return err
		//}
		if err := r.db.Create(inputRole).Error; err != nil {
			return err
		}
	}

	return nil
}

func (r RoleRepository) List(pg *pagination.Pagination) ([]*domain.Role, error) {
	var roles []*domain.Role
	var objs []*domain.Role

	if pg == nil {
		pg = pagination.NewDefaultPagination()
	}
	filterFunc := CombinedGormFilter("roles", pg.GetFilters(), pg.CombinedFilter)
	db := filterFunc(r.db.Model(&domain.Role{}))

	db.Count(&pg.TotalRows)
	pg.TotalPages = int(math.Ceil(float64(pg.TotalRows) / float64(pg.Limit)))

	orderQuery := fmt.Sprintf("%s %s", pg.SortColumn, pg.SortOrder)
	//res := db.Joins("JOIN roles as r on r.id = tks_roles.role_id").
	//	Offset(pg.GetOffset()).Limit(pg.GetLimit()).Order(orderQuery).Find(&objs)
	res := db.Offset(pg.GetOffset()).Limit(pg.GetLimit()).Order(orderQuery).Find(&objs)

	if res.Error != nil {
		return nil, res.Error
	}
	for _, role := range objs {
		roles = append(roles, ConvertRepoToDomainRole(role))
	}

	return roles, nil
}

func (r RoleRepository) ListTksRoles(organizationId string, pg *pagination.Pagination) ([]*domain.TksRole, error) {
	var roles []*domain.TksRole
	var objs []*domain.TksRole

	if pg == nil {
		pg = pagination.NewDefaultPagination()
	}
	filterFunc := CombinedGormFilter("roles", pg.GetFilters(), pg.CombinedFilter)
	db := filterFunc(r.db.Model(&domain.TksRole{}))

	db.Count(&pg.TotalRows)
	pg.TotalPages = int(math.Ceil(float64(pg.TotalRows) / float64(pg.Limit)))

	orderQuery := fmt.Sprintf("%s %s", pg.SortColumn, pg.SortOrder)
	res := db.Joins("JOIN roles as r on r.id = tks_roles.role_id").
		Where("r.organization_id = ?", organizationId).
		Offset(pg.GetOffset()).
		Limit(pg.GetLimit()).
		Order(orderQuery).
		Find(&objs)
	//res := db.Preload("Role").Offset(pg.GetOffset()).Limit(pg.GetLimit()).Order(orderQuery).Find(&objs)
	if res.Error != nil {
		return nil, res.Error
	}
	for _, role := range objs {
		roles = append(roles, ConvertRepoToDomainTksRole(role))
	}

	return roles, nil
}

func (r RoleRepository) ListProjectRoles(projectId string, pg *pagination.Pagination) ([]*domain.ProjectRole, error) {
	var roles []*domain.ProjectRole
	var objs []*domain.ProjectRole

	if pg == nil {
		pg = pagination.NewDefaultPagination()
	}
	filterFunc := CombinedGormFilter("roles", pg.GetFilters(), pg.CombinedFilter)
	db := filterFunc(r.db.Model(&domain.ProjectRole{}))

	db.Count(&pg.TotalRows)
	pg.TotalPages = int(math.Ceil(float64(pg.TotalRows) / float64(pg.Limit)))

	orderQuery := fmt.Sprintf("%s %s", pg.SortColumn, pg.SortOrder)
	res := db.Joins("JOIN roles as r on r.id = project_roles.role_id").
		Where("project_roles.project_id = ?", projectId).
		Offset(pg.GetOffset()).Limit(pg.GetLimit()).Order(orderQuery).Find(&objs)
	//res := db.Preload("Role").Preload("Project").Offset(pg.GetOffset()).Limit(pg.GetLimit()).Order(orderQuery).Find(&objs)
	if res.Error != nil {
		return nil, res.Error
	}
	for _, role := range objs {
		roles = append(roles, role)
	}

	return roles, nil
}

func (r RoleRepository) Get(id uuid.UUID) (*domain.Role, error) {
	var role domain.Role
	if err := r.db.First(&role, "id = ?", id).Error; err != nil {
		return nil, err
	}

	return ConvertRepoToDomainRole(&role), nil
}

func (r RoleRepository) GetTksRole(id uuid.UUID) (*domain.TksRole, error) {
	var role domain.TksRole
	if err := r.db.Preload("Role").First(&role, "role_id = ?", id).Error; err != nil {
		return nil, err
	}

	return ConvertRepoToDomainTksRole(&role), nil
}

func (r RoleRepository) GetProjectRole(id uuid.UUID) (*domain.ProjectRole, error) {
	var role domain.ProjectRole
	if err := r.db.Preload("Role").Preload("Project").First(&role, "role_id = ?", id).Error; err != nil {
		return nil, err
	}

	return &role, nil
}

func (r RoleRepository) DeleteCascade(id uuid.UUID) error {
	// manual cascade delete
	if err := r.db.Delete(&domain.TksRole{}, "role_id = ?", id).Error; err != nil {
		return err
	}
	if err := r.db.Delete(&domain.ProjectRole{}, "role_id = ?", id).Error; err != nil {
		return err
	}

	if err := r.db.Delete(&domain.Role{}, "id = ?", id).Error; err != nil {
		return err
	}

	return nil
}

func (r RoleRepository) Update(roleObj interface{}) error {
	switch roleObj.(type) {
	case *domain.TksRole:
		inputRole := roleObj.(*domain.TksRole)
		role := ConvertRepoToDomainTksRole(inputRole)
		if err := r.db.Model(&domain.TksRole{}).Where("id = ?", role.RoleID).Updates(domain.Role{
			Name:        role.Role.Name,
			Description: role.Role.Description,
		}).Error; err != nil {
			return err
		}

	case *domain.ProjectRole:
		inputRole := roleObj.(*domain.ProjectRole)
		//projectRole := ConvertRepoToDomainProjectRole(inputRole)
		// update role
		if err := r.db.Model(&domain.ProjectRole{}).Where("role_id = ?", inputRole.RoleID).Updates(domain.Role{
			Name:        inputRole.Role.Name,
			Description: inputRole.Role.Description,
		}).Error; err != nil {
			return err
		}
	}

	return nil
}

func NewRoleRepository(db *gorm.DB) IRoleRepository {
	return &RoleRepository{
		db: db,
	}
}

// domain.Role to repository.Role
func ConverDomainToRepoRole(domainRole *domain.Role) *domain.Role {
	return &domain.Role{
		ID:             domainRole.ID,
		Name:           domainRole.Name,
		OrganizationID: domainRole.OrganizationID,
		Type:           domainRole.Type,
		Creator:        domainRole.Creator,
		Description:    domainRole.Description,
	}
}

// repository.Role to domain.Role
func ConvertRepoToDomainRole(repoRole *domain.Role) *domain.Role {
	return &domain.Role{
		ID:             repoRole.ID,
		Name:           repoRole.Name,
		OrganizationID: repoRole.OrganizationID,
		Type:           repoRole.Type,
		Creator:        repoRole.Creator,
		Description:    repoRole.Description,
	}
}

// domain.TksRole to repository.TksRole
func ConvertDomainToRepoTksRole(domainRole *domain.TksRole) *domain.TksRole {
	return &domain.TksRole{
		RoleID: domainRole.Role.ID,
		Role:   *ConverDomainToRepoRole(&domainRole.Role),
	}
}

// repository.TksRole to domain.TksRole
func ConvertRepoToDomainTksRole(repoRole *domain.TksRole) *domain.TksRole {
	return &domain.TksRole{
		RoleID: repoRole.RoleID,
		Role:   *ConvertRepoToDomainRole(&repoRole.Role),
	}
}

//// domain.ProjectRole to repository.ProjectRole
//func ConvertDomainToRepoProjectRole(domainRole *domain.ProjectRole) *ProjectRole {
//	return &ProjectRole{
//		RoleID:    domainRole.RoleID,
//		ProjectID: domainRole.ProjectID,
//		Role:      *ConverDomainToRepoRole(&domainRole.Role),
//		Project:   domainRole.Project,
//	}
//}
//
//// repository.ProjectRole to domain.ProjectRole
//func ConvertRepoToDomainProjectRole(repoRole *ProjectRole) *domain.ProjectRole {
//	return &domain.ProjectRole{
//		RoleID:    repoRole.RoleID,
//		ProjectID: repoRole.ProjectID,
//		Role:      *ConvertRepoToDomainRole(&repoRole.Role),
//		Project:   repoRole.Project,
//	}
//}
