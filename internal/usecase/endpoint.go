package usecase

import (
	"context"
	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/internal/repository"
)

type IEndpointUsecase interface {
	ListEndpoints(ctx context.Context, pg *pagination.Pagination) ([]*model.Endpoint, error)
}

type EndpointUsecase struct {
	repo repository.IEndpointRepository
}

func NewEndpointUsecase(repo repository.Repository) *EndpointUsecase {
	return &EndpointUsecase{
		repo: repo.Endpoint,
	}
}

func (e EndpointUsecase) ListEndpoints(ctx context.Context, pg *pagination.Pagination) ([]*model.Endpoint, error) {
	return e.repo.List(ctx, pg)
}
