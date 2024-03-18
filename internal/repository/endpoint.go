package repository

import (
	"context"

	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/pagination"
	"gorm.io/gorm"
)

type IEndpointRepository interface {
	Create(ctx context.Context, endpoint *model.Endpoint) error
	List(ctx context.Context, pg *pagination.Pagination) ([]*model.Endpoint, error)
	Get(ctx context.Context, id uint) (*model.Endpoint, error)
}

type EndpointRepository struct {
	db *gorm.DB
}

func NewEndpointRepository(db *gorm.DB) *EndpointRepository {
	return &EndpointRepository{
		db: db,
	}
}

func (e *EndpointRepository) Create(ctx context.Context, endpoint *model.Endpoint) error {
	obj := &model.Endpoint{
		Name:  endpoint.Name,
		Group: endpoint.Group,
	}

	if err := e.db.WithContext(ctx).Create(obj).Error; err != nil {
		return err
	}

	return nil
}

func (e *EndpointRepository) List(ctx context.Context, pg *pagination.Pagination) ([]*model.Endpoint, error) {
	var endpoints []*model.Endpoint

	if pg == nil {
		pg = pagination.NewPagination(nil)
	}

	_, res := pg.Fetch(e.db.WithContext(ctx).Model(&model.Endpoint{}), &endpoints)
	if res.Error != nil {
		return nil, res.Error
	}

	return endpoints, nil
}

func (e *EndpointRepository) Get(ctx context.Context, id uint) (*model.Endpoint, error) {
	var obj model.Endpoint

	if err := e.db.WithContext(ctx).Preload("Permission").First(&obj, "id = ?", id).Error; err != nil {
		return nil, err
	}

	return &model.Endpoint{
		Name:  obj.Name,
		Group: obj.Group,
	}, nil
}
