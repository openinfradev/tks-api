package repository

import (
	"context"
	"fmt"
	"github.com/pkg/errors"

	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/pagination"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IRoleRepository interface {
	Create(ctx context.Context, roleObj *model.Role) (string, error)
	ListTksRoles(ctx context.Context, organizationId string, pg *pagination.Pagination) ([]*model.Role, error)
	GetTksRole(ctx context.Context, organizationId string, id string) (*model.Role, error)
	GetTksRoleByRoleName(ctx context.Context, organizationId string, roleName string) (*model.Role, error)
	Delete(ctx context.Context, id string) error
	Update(ctx context.Context, roleObj *model.Role) error
}

type RoleRepository struct {
	db *gorm.DB
}

func (r RoleRepository) GetTksRoleByRoleName(ctx context.Context, oragnizationId string, roleName string) (*model.Role, error) {
	var role model.Role
	if err := r.db.WithContext(ctx).First(&role, "organization_id = ? AND name = ?", oragnizationId, roleName).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &role, nil
}

func (r RoleRepository) Create(ctx context.Context, roleObj *model.Role) (string, error) {
	if roleObj == nil {
		return "", fmt.Errorf("roleObj is nil")
	}
	if roleObj.ID == "" {
		roleObj.ID = uuid.New().String()
	}
	if err := r.db.WithContext(ctx).Create(roleObj).Error; err != nil {
		return "", err
	}

	return roleObj.ID, nil
}

func (r RoleRepository) ListTksRoles(ctx context.Context, organizationId string, pg *pagination.Pagination) ([]*model.Role, error) {
	var roles []*model.Role

	if pg == nil {
		pg = pagination.NewPagination(nil)
	}

	_, res := pg.Fetch(r.db.WithContext(ctx).Preload(clause.Associations).Model(&model.Role{}).Where("organization_id = ?", organizationId), &roles)
	if res.Error != nil {
		return nil, res.Error
	}
	return roles, nil
}

func (r RoleRepository) GetTksRole(ctx context.Context, organizationId string, id string) (*model.Role, error) {
	var role model.Role
	if err := r.db.WithContext(ctx).First(&role, "id = ?", id).Error; err != nil {
		return nil, err
	}

	return &role, nil
}

func (r RoleRepository) Update(ctx context.Context, roleObj *model.Role) error {
	if roleObj == nil {
		return fmt.Errorf("roleObj is nil")
	}

	err := r.db.WithContext(ctx).Model(&model.Role{}).Where("id = ?", roleObj.ID).Updates(model.Role{
		Name:        roleObj.Name,
		Description: roleObj.Description,
	}).Error

	if err != nil {
		return err
	}

	return nil
}

func (r RoleRepository) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&model.Role{}, "id = ?", id).Error; err != nil {
		return err
	}

	return nil
}

func NewRoleRepository(db *gorm.DB) IRoleRepository {
	return &RoleRepository{
		db: db,
	}
}
