package repository

import (
	"fmt"
	"math"

	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/pagination"
	"gorm.io/gorm"
)

type IRoleRepository interface {
	Create(roleObj *model.Role) (string, error)
	List(pg *pagination.Pagination) ([]*model.Role, error)
	ListTksRoles(organizationId string, pg *pagination.Pagination) ([]*model.Role, error)
	Get(id string) (*model.Role, error)
	GetTksRole(id string) (*model.Role, error)
	GetTksRoleByRoleName(roleName string) (*model.Role, error)
	Delete(id string) error
	Update(roleObj *model.Role) error
}

type RoleRepository struct {
	db *gorm.DB
}

func (r RoleRepository) GetTksRoleByRoleName(roleName string) (*model.Role, error) {
	var role model.Role
	if err := r.db.Preload("Role").First(&role, "Role.name = ?", roleName).Error; err != nil {
		return nil, err
	}

	return &role, nil
}

func (r RoleRepository) Create(roleObj *model.Role) (string, error) {
	if roleObj == nil {
		return "", fmt.Errorf("roleObj is nil")
	}
	if err := r.db.Create(roleObj).Error; err != nil {
		return "", err
	}

	return roleObj.ID, nil
}

func (r RoleRepository) List(pg *pagination.Pagination) ([]*model.Role, error) {
	var roles []*model.Role

	if pg == nil {
		pg = pagination.NewPagination(nil)
	}
	filterFunc := CombinedGormFilter("roles", pg.GetFilters(), pg.CombinedFilter)
	db := filterFunc(r.db.Model(&model.Role{}))

	db.Count(&pg.TotalRows)
	pg.TotalPages = int(math.Ceil(float64(pg.TotalRows) / float64(pg.Limit)))

	orderQuery := fmt.Sprintf("%s %s", pg.SortColumn, pg.SortOrder)
	res := db.Offset(pg.GetOffset()).Limit(pg.GetLimit()).Order(orderQuery).Find(&roles)

	if res.Error != nil {
		return nil, res.Error
	}

	return roles, nil
}

func (r RoleRepository) ListTksRoles(organizationId string, pg *pagination.Pagination) ([]*model.Role, error) {
	var roles []*model.Role

	if pg == nil {
		pg = pagination.NewPagination(nil)
	}
	filterFunc := CombinedGormFilter("roles", pg.GetFilters(), pg.CombinedFilter)
	db := filterFunc(r.db.Model(&model.Role{}))

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

func (r RoleRepository) Get(id string) (*model.Role, error) {
	var role model.Role
	if err := r.db.First(&role, "id = ?", id).Error; err != nil {
		return nil, err
	}

	return &role, nil
}

func (r RoleRepository) GetTksRole(id string) (*model.Role, error) {
	var role model.Role
	if err := r.db.First(&role, "id = ?", id).Error; err != nil {
		return nil, err
	}

	return &role, nil
}

func (r RoleRepository) Update(roleObj *model.Role) error {
	if roleObj == nil {
		return fmt.Errorf("roleObj is nil")
	}

	err := r.db.Model(&model.Role{}).Where("id = ?", roleObj.ID).Updates(model.Role{
		Name:        roleObj.Name,
		Description: roleObj.Description,
	}).Error

	if err != nil {
		return err
	}

	return nil
}

func (r RoleRepository) Delete(id string) error {
	if err := r.db.Delete(&model.Role{}, "id = ?", id).Error; err != nil {
		return err
	}

	return nil
}

func NewRoleRepository(db *gorm.DB) IRoleRepository {
	return &RoleRepository{
		db: db,
	}
}
