package repository

import (
	"fmt"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/pkg/domain"
	"gorm.io/gorm"
	"math"
)

type IEndpointRepository interface {
	Create(endpoint *domain.Endpoint) error
	List(pg *pagination.Pagination) ([]*domain.Endpoint, error)
	Get(id uint) (*domain.Endpoint, error)
}

type EndpointRepository struct {
	db *gorm.DB
}

func NewEndpointRepository(db *gorm.DB) *EndpointRepository {
	return &EndpointRepository{
		db: db,
	}
}

func (e *EndpointRepository) Create(endpoint *domain.Endpoint) error {
	obj := &domain.Endpoint{
		Name:  endpoint.Name,
		Group: endpoint.Group,
	}

	if err := e.db.Create(obj).Error; err != nil {
		return err
	}

	return nil
}

func (e *EndpointRepository) List(pg *pagination.Pagination) ([]*domain.Endpoint, error) {
	var endpoints []*domain.Endpoint

	if pg == nil {
		pg = pagination.NewPagination(nil)
	}
	filterFunc := CombinedGormFilter("endpoints", pg.GetFilters(), pg.CombinedFilter)
	db := filterFunc(e.db.Model(&domain.Endpoint{}))

	db.Count(&pg.TotalRows)
	pg.TotalPages = int(math.Ceil(float64(pg.TotalRows) / float64(pg.Limit)))

	orderQuery := fmt.Sprintf("%s %s", pg.SortColumn, pg.SortOrder)
	res := db.Offset(pg.GetOffset()).Limit(pg.GetLimit()).Order(orderQuery).Find(&endpoints)
	if res.Error != nil {
		return nil, res.Error
	}

	return endpoints, nil
}

func (e *EndpointRepository) Get(id uint) (*domain.Endpoint, error) {
	var obj domain.Endpoint

	if err := e.db.Preload("Permission").First(&obj, "id = ?", id).Error; err != nil {
		return nil, err
	}

	return &domain.Endpoint{
		Name:  obj.Name,
		Group: obj.Group,
	}, nil
}
