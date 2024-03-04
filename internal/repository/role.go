package repository

import (
	"fmt"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/pkg/domain"
	"gorm.io/gorm"
	"math"
)

type IRoleRepository interface {
	Create(roleObj *domain.Role) (string, error)
	List(pg *pagination.Pagination) ([]*domain.Role, error)
	ListTksRoles(organizationId string, pg *pagination.Pagination) ([]*domain.Role, error)
	Get(id string) (*domain.Role, error)
	GetTksRole(id string) (*domain.Role, error)
	GetTksRoleByRoleName(roleName string) (*domain.Role, error)
	Delete(id string) error
	Update(roleObj *domain.Role) error
}

type RoleRepository struct {
	db *gorm.DB
}

func (r RoleRepository) GetTksRoleByRoleName(roleName string) (*domain.Role, error) {
	var role domain.Role
	if err := r.db.Preload("Role").First(&role, "Role.name = ?", roleName).Error; err != nil {
		return nil, err
	}

	return &role, nil
}

func (r RoleRepository) Create(roleObj *domain.Role) (string, error) {
	if roleObj == nil {
		return "", fmt.Errorf("roleObj is nil")
	}
	if err := r.db.Create(roleObj).Error; err != nil {
		return "", err
	}

	return roleObj.ID, nil
}

func (r RoleRepository) List(pg *pagination.Pagination) ([]*domain.Role, error) {
	var roles []*domain.Role

	if pg == nil {
		pg = pagination.NewDefaultPagination()
	}
	filterFunc := CombinedGormFilter("roles", pg.GetFilters(), pg.CombinedFilter)
	db := filterFunc(r.db.Model(&domain.Role{}))

	db.Count(&pg.TotalRows)
	pg.TotalPages = int(math.Ceil(float64(pg.TotalRows) / float64(pg.Limit)))

	orderQuery := fmt.Sprintf("%s %s", pg.SortColumn, pg.SortOrder)
	res := db.Offset(pg.GetOffset()).Limit(pg.GetLimit()).Order(orderQuery).Find(&roles)

	if res.Error != nil {
		return nil, res.Error
	}

	return roles, nil
}

func (r RoleRepository) ListTksRoles(organizationId string, pg *pagination.Pagination) ([]*domain.Role, error) {
	var roles []*domain.Role

	if pg == nil {
		pg = pagination.NewDefaultPagination()
	}
	filterFunc := CombinedGormFilter("roles", pg.GetFilters(), pg.CombinedFilter)
	db := filterFunc(r.db.Model(&domain.Role{}))

	db.Count(&pg.TotalRows)
	pg.TotalPages = int(math.Ceil(float64(pg.TotalRows) / float64(pg.Limit)))

	orderQuery := fmt.Sprintf("%s %s", pg.SortColumn, pg.SortOrder)
	res := db.
		Offset(pg.GetOffset()).
		Limit(pg.GetLimit()).
		Order(orderQuery).
		Find(&roles, "organization_id = ?", organizationId)
	//res := db.Preload("Role").Offset(pg.GetOffset()).Limit(pg.GetLimit()).Order(orderQuery).Find(&objs)
	if res.Error != nil {
		return nil, res.Error
	}

	return roles, nil
}

func (r RoleRepository) Get(id string) (*domain.Role, error) {
	var role domain.Role
	if err := r.db.First(&role, "id = ?", id).Error; err != nil {
		return nil, err
	}

	return &role, nil
}

func (r RoleRepository) GetTksRole(id string) (*domain.Role, error) {
	var role domain.Role
	if err := r.db.First(&role, "id = ?", id).Error; err != nil {
		return nil, err
	}

	return &role, nil
}

func (r RoleRepository) Update(roleObj *domain.Role) error {
	if roleObj == nil {
		return fmt.Errorf("roleObj is nil")
	}

	err := r.db.Model(&domain.Role{}).Where("id = ?", roleObj.ID).Updates(domain.Role{
		Name:        roleObj.Name,
		Description: roleObj.Description,
	}).Error

	if err != nil {
		return err
	}

	return nil
}

func (r RoleRepository) Delete(id string) error {
	if err := r.db.Delete(&domain.Role{}, "id = ?", id).Error; err != nil {
		return err
	}

	return nil
}

func NewRoleRepository(db *gorm.DB) IRoleRepository {
	return &RoleRepository{
		db: db,
	}
}
