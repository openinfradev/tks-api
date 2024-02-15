package repository

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/pkg/domain"
	"gorm.io/gorm"
	"math"
)

type IEndpointRepository interface {
	Create(endpoint *domain.Endpoint) error
	List(pg *pagination.Pagination) ([]*domain.Endpoint, error)
	Get(id uuid.UUID) (*domain.Endpoint, error)
	Update(endpoint *domain.Endpoint) error
	Delete(id uuid.UUID) error
}

type EndpointRepository struct {
	db *gorm.DB
}

func NewEndpointRepository(db *gorm.DB) *EndpointRepository {
	return &EndpointRepository{
		db: db,
	}
}

//type Endpoint struct {
//	gorm.Model
//
//	ID uuid.UUID `gorm:"type:uuid;primary_key;"`
//
//	Name         string     `gorm:"type:text;not null;unique"`
//	Group        string     `gorm:"type:text;"`
//	PermissionID uuid.UUID  `gorm:"type:uuid;"`
//	Permission   Permission `gorm:"foreignKey:PermissionID;"`
//}

func (e *EndpointRepository) Create(endpoint *domain.Endpoint) error {
	obj := &domain.Endpoint{
		Name:         endpoint.Name,
		Group:        endpoint.Group,
		PermissionID: endpoint.PermissionID,
	}

	if err := e.db.Create(obj).Error; err != nil {
		return err
	}

	return nil
}

func (e *EndpointRepository) List(pg *pagination.Pagination) ([]*domain.Endpoint, error) {
	var endpoints []*domain.Endpoint
	var objs []*domain.Endpoint

	if pg == nil {
		pg = pagination.NewDefaultPagination()
	}
	filterFunc := CombinedGormFilter("endpoints", pg.GetFilters(), pg.CombinedFilter)
	db := filterFunc(e.db.Model(&domain.Endpoint{}))

	db.Count(&pg.TotalRows)
	pg.TotalPages = int(math.Ceil(float64(pg.TotalRows) / float64(pg.Limit)))

	orderQuery := fmt.Sprintf("%s %s", pg.SortColumn, pg.SortOrder)
	res := db.Preload("Permissions").Offset(pg.GetOffset()).Limit(pg.GetLimit()).Order(orderQuery).Find(&objs)
	if res.Error != nil {
		return nil, res.Error
	}
	for _, obj := range objs {
		endpoints = append(endpoints, &domain.Endpoint{
			ID:           obj.ID,
			Name:         obj.Name,
			Group:        obj.Group,
			PermissionID: obj.PermissionID,
			Permission:   obj.Permission,
		})
	}

	return endpoints, nil
}

func (e *EndpointRepository) Get(id uuid.UUID) (*domain.Endpoint, error) {
	var obj domain.Endpoint

	if err := e.db.Preload("Permission").First(&obj, "id = ?", id).Error; err != nil {
		return nil, err
	}

	return &domain.Endpoint{
		ID:    obj.ID,
		Name:  obj.Name,
		Group: obj.Group,
	}, nil
}

func (e *EndpointRepository) Update(endpoint *domain.Endpoint) error {
	obj := &domain.Endpoint{
		ID:           endpoint.ID,
		Name:         endpoint.Name,
		Group:        endpoint.Group,
		PermissionID: endpoint.PermissionID,
	}

	if err := e.db.Save(obj).Error; err != nil {
		return err
	}

	return nil
}

//
//// domain.Endpoint to repository.Endpoint
//func ConvertDomainToRepoEndpoint(endpoint *domain.Endpoint) *Endpoint {
//	return &Endpoint{
//		ID:           endpoint.ID,
//		Name:         endpoint.Name,
//		Group:        endpoint.Group,
//		PermissionID: endpoint.PermissionID,
//		Permission:   *ConvertDomainToRepoPermission(&endpoint.Permission),
//	}
//}
//
//// repository.Endpoint to domain.Endpoint
//func ConvertRepoToDomainEndpoint(endpoint *Endpoint) *domain.Endpoint {
//	return &domain.Endpoint{
//		ID:           endpoint.ID,
//		Name:         endpoint.Name,
//		Group:        endpoint.Group,
//		PermissionID: endpoint.PermissionID,
//		Permission:   *ConvertRepoToDomainPermission(&endpoint.Permission),
//	}
//}
