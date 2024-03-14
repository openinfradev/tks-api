package repository

import (
	"fmt"
	"math"

	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/pagination"
	"gorm.io/gorm"
)

type IEndpointRepository interface {
	Create(endpoint *model.Endpoint) error
	List(pg *pagination.Pagination) ([]*model.Endpoint, error)
	Get(id uint) (*model.Endpoint, error)
}

type EndpointRepository struct {
	db *gorm.DB
}

func NewEndpointRepository(db *gorm.DB) *EndpointRepository {
	return &EndpointRepository{
		db: db,
	}
}

func (e *EndpointRepository) Create(endpoint *model.Endpoint) error {
	obj := &model.Endpoint{
		Name:  endpoint.Name,
		Group: endpoint.Group,
	}

	if err := e.db.Create(obj).Error; err != nil {
		return err
	}

	return nil
}

func (e *EndpointRepository) List(pg *pagination.Pagination) ([]*model.Endpoint, error) {
	var endpoints []*model.Endpoint

	if pg == nil {
		pg = pagination.NewPagination(nil)
	}
	filterFunc := CombinedGormFilter("endpoints", pg.GetFilters(), pg.CombinedFilter)
	db := filterFunc(e.db.Model(&model.Endpoint{}))

	db.Count(&pg.TotalRows)
	pg.TotalPages = int(math.Ceil(float64(pg.TotalRows) / float64(pg.Limit)))

	orderQuery := fmt.Sprintf("%s %s", pg.SortColumn, pg.SortOrder)
	res := db.Offset(pg.GetOffset()).Limit(pg.GetLimit()).Order(orderQuery).Find(&endpoints)
	if res.Error != nil {
		return nil, res.Error
	}

	return endpoints, nil
}

func (e *EndpointRepository) Get(id uint) (*model.Endpoint, error) {
	var obj model.Endpoint

	if err := e.db.Preload("Permission").First(&obj, "id = ?", id).Error; err != nil {
		return nil, err
	}

	return &model.Endpoint{
		Name:  obj.Name,
		Group: obj.Group,
	}, nil
}
