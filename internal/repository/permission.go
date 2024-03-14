package repository

import (
	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/pkg/domain"
	"gorm.io/gorm"
)

type IPermissionRepository interface {
	Create(permission *domain.Permission) error
	List(roleId string) ([]*domain.Permission, error)
	Get(id uuid.UUID) (*domain.Permission, error)
	Delete(id uuid.UUID) error
	Update(permission *domain.Permission) error
}

type PermissionRepository struct {
	db *gorm.DB
}

func NewPermissionRepository(db *gorm.DB) *PermissionRepository {
	return &PermissionRepository{
		db: db,
	}
}

func (r PermissionRepository) Create(p *domain.Permission) error {
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

	return r.db.Create(p).Error
}

func (r PermissionRepository) List(roleId string) ([]*domain.Permission, error) {
	var permissions []*domain.Permission

	err := r.db.Preload("Children.Children.Children.Children").Where("parent_id IS NULL AND role_id = ?", roleId).Find(&permissions).Error
	if err != nil {
		return nil, err
	}

	return permissions, nil
}

func (r PermissionRepository) Get(id uuid.UUID) (*domain.Permission, error) {
	permission := &domain.Permission{}
	result := r.db.Preload("Children.Children.Children").Preload("Parent").First(&permission, "id = ?", id)
	if result.Error != nil {
		return nil, result.Error
	}

	return permission, nil
}

func (r PermissionRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&domain.Permission{}, "id = ?", id).Error
}

func (r PermissionRepository) Update(p *domain.Permission) error {
	// update on is_allowed
	return r.db.Model(&domain.Permission{}).Where("id = ?", p.ID).Updates(map[string]interface{}{"is_allowed": p.IsAllowed}).Error
}
