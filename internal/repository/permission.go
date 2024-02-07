package repository

import (
	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/pkg/domain"
	"gorm.io/gorm"
)

type IPermissionRepository interface {
	Create(permission *domain.Permission) error
	List() ([]*domain.Permission, error)
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

type Permission struct {
	gorm.Model

	ID   uuid.UUID `gorm:"primarykey;type:uuid;"`
	Name string

	IsAllowed *bool `gorm:"type:boolean;"`
	RoleID    *uuid.UUID
	Role      *Role       `gorm:"foreignKey:RoleID;references:ID;"`
	Endpoints []*Endpoint `gorm:"one2many:endpoints;"`

	ParentID *uuid.UUID
	Parent   *Permission   `gorm:"foreignKey:ParentID;references:ID;"`
	Children []*Permission `gorm:"foreignKey:ParentID;references:ID;"`
}

func (p *Permission) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
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

	permission := ConvertDomainToRepoPermission(p)

	return r.db.Create(permission).Error
}

func (r PermissionRepository) List() ([]*domain.Permission, error) {
	var permissions []*Permission
	var outs []*domain.Permission

	err := r.db.Preload("Children.Children.Children.Children").Where("parent_id IS NULL").Find(&permissions).Error
	if err != nil {
		return nil, err
	}

	for _, permission := range permissions {
		outs = append(outs, ConvertRepoToDomainPermission(permission))
	}
	return outs, nil
}

func (r PermissionRepository) Get(id uuid.UUID) (*domain.Permission, error) {
	permission := &Permission{}
	result := r.db.Preload("Children.Children.Children").Preload("Parent").First(&permission, "id = ?", id)
	if result.Error != nil {
		return nil, result.Error
	}

	return ConvertRepoToDomainPermission(permission), nil
}

func (r PermissionRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&Permission{}, id).Error
}

func (r PermissionRepository) Update(p *domain.Permission) error {
	permission := ConvertDomainToRepoPermission(p)

	return r.db.Updates(permission).Error

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
	//
	//permission := &Permission{}
	//
	//result := r.db.First(&permission, "id = ?", p.ID)
	//if result.Error != nil {
	//	return result.Error
	//}
	//
	//permission.Name = p.Name
	//permission.Parent = parent
	//permission.Children = children
	//permission.IsAllowed = p.IsAllowed
	//
	//return r.db.Save(permission).Error
}

// repository.Permission to domain.Permission
func ConvertRepoToDomainPermission(repoPerm *Permission) *domain.Permission {
	if repoPerm == nil {
		return nil
	}

	if repoPerm.Endpoints == nil {
		repoPerm.Endpoints = []*Endpoint{}
	}
	var domainEndpoints []*domain.Endpoint
	for _, endpoint := range repoPerm.Endpoints {
		domainEndpoints = append(domainEndpoints, ConvertRepoToDomainEndpoint(endpoint))
	}

	// Domain Permission 객체 생성
	domainPerm := &domain.Permission{
		ID:        repoPerm.ID,
		Name:      repoPerm.Name,
		ParentID:  repoPerm.ParentID,
		IsAllowed: repoPerm.IsAllowed,
		Endpoints: domainEndpoints,
	}

	// 자식 권한들 변환
	for _, child := range repoPerm.Children {
		domainChild := ConvertRepoToDomainPermission(child)
		domainPerm.Children = append(domainPerm.Children, domainChild)
	}

	// 부모 권한 변환 (부모 권한이 있을 경우만)
	if repoPerm.Parent != nil {
		domainPerm.Parent = ConvertRepoToDomainPermission(repoPerm.Parent)
	}

	return domainPerm
}

// domain.Permission to repository.Permission
func ConvertDomainToRepoPermission(domainPerm *domain.Permission) *Permission {
	if domainPerm == nil {
		return nil
	}

	if domainPerm.Endpoints == nil {
		domainPerm.Endpoints = []*domain.Endpoint{}
	}
	var repoEndpoints []*Endpoint
	for _, endpoint := range domainPerm.Endpoints {
		repoEndpoints = append(repoEndpoints, ConvertDomainToRepoEndpoint(endpoint))
	}

	// Domain Permission 객체 생성
	repoPerm := &Permission{
		ID:        domainPerm.ID,
		Name:      domainPerm.Name,
		ParentID:  domainPerm.ParentID,
		IsAllowed: domainPerm.IsAllowed,
		Endpoints: repoEndpoints,
	}

	// 자식 권한들 변환
	for _, child := range domainPerm.Children {
		repoChild := ConvertDomainToRepoPermission(child)
		repoPerm.Children = append(repoPerm.Children, repoChild)
	}

	// 부모 권한 변환 (부모 권한이 있을 경우만)
	if domainPerm.Parent != nil {
		repoPerm.Parent = ConvertDomainToRepoPermission(domainPerm.Parent)
	}

	return repoPerm
}
