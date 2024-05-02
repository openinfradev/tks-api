package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/model"
	"gorm.io/gorm"
)

type IPermissionRepository interface {
	Create(ctx context.Context, permission *model.Permission) error
	List(ctx context.Context, roleId string) ([]*model.Permission, error)
	Get(ctx context.Context, id uuid.UUID) (*model.Permission, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Update(ctx context.Context, permission *model.Permission) error
	EdgeKeyOverwrite(ctx context.Context, permission *model.Permission) error
}

type PermissionRepository struct {
	db *gorm.DB
}

func NewPermissionRepository(db *gorm.DB) *PermissionRepository {
	return &PermissionRepository{
		db: db,
	}
}

func (r PermissionRepository) Create(ctx context.Context, p *model.Permission) error {
	//var parent *Permission
	//var children []*Permission
	//
	//if p.Parent != nil {
	//	parent = &Permission{}
	//	result := r.db.First(&parent, "id = ?", p.Parent.ID)
	//	if result.Error != nil {
	//		return result.Error
	//	}
	//}
	//if p.Children != nil {
	//	for _, child := range p.Children {
	//		newChild := &Permission{}
	//		result := r.db.First(&newChild, "id = ?", child.ID)
	//		if result.Error != nil {
	//			return result.Error
	//		}
	//		children = append(children, newChild)
	//	}
	//}

	p.ID = uuid.New()
	return r.db.WithContext(ctx).Create(p).Error
}

func (r PermissionRepository) List(ctx context.Context, roleId string) ([]*model.Permission, error) {
	var permissions []*model.Permission

	err := r.db.WithContext(ctx).Preload("Children.Children.Children.Children").
		Where("parent_id IS NULL AND role_id = ?", roleId).Find(&permissions).Error
	if err != nil {
		return nil, err
	}

	return permissions, nil
}

func (r PermissionRepository) Get(ctx context.Context, id uuid.UUID) (*model.Permission, error) {
	permission := &model.Permission{}
	result := r.db.WithContext(ctx).Preload("Children.Children.Children").Preload("Parent").First(&permission, "id = ?", id)
	if result.Error != nil {
		return nil, result.Error
	}

	return permission, nil
}

func (r PermissionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Permission{}, "id = ?", id).Error
}

func (r PermissionRepository) Update(ctx context.Context, p *model.Permission) error {
	// update on is_allowed
	return r.db.WithContext(ctx).Model(&model.Permission{}).Where("id = ?", p.ID).Updates(map[string]interface{}{"is_allowed": p.IsAllowed}).Error
}

func (r PermissionRepository) EdgeKeyOverwrite(ctx context.Context, p *model.Permission) error {
	return r.db.WithContext(ctx).Model(&model.Permission{}).Where("id = ?", p.ID).Updates(map[string]interface{}{"edge_key": p.EdgeKey}).Error
}
