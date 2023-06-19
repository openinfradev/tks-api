package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-api/pkg/domain"
)

type IStackTemplateUsecase interface {
	Get(ctx context.Context, stackTemplate uuid.UUID) (domain.StackTemplate, error)
	Fetch(ctx context.Context) ([]domain.StackTemplate, error)
	Create(ctx context.Context, dto domain.StackTemplate) (stackTemplate uuid.UUID, err error)
	Update(ctx context.Context, dto domain.StackTemplate) error
	Delete(ctx context.Context, dto domain.StackTemplate) error
}

type StackTemplateUsecase struct {
	repo repository.IStackTemplateRepository
}

func NewStackTemplateUsecase(r repository.Repository) IStackTemplateUsecase {
	return &StackTemplateUsecase{
		repo: r.StackTemplate,
	}
}

func (u *StackTemplateUsecase) Create(ctx context.Context, dto domain.StackTemplate) (stackTemplate uuid.UUID, err error) {
	return uuid.Nil, nil
}

func (u *StackTemplateUsecase) Update(ctx context.Context, dto domain.StackTemplate) error {
	return nil
}

func (u *StackTemplateUsecase) Get(ctx context.Context, stackTemplate uuid.UUID) (res domain.StackTemplate, err error) {
	res, err = u.repo.Get(stackTemplate)
	if err != nil {
		return domain.StackTemplate{}, err
	}
	return
}

func (u *StackTemplateUsecase) Fetch(ctx context.Context) (res []domain.StackTemplate, err error) {
	res, err = u.repo.Fetch()
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (u *StackTemplateUsecase) Delete(ctx context.Context, dto domain.StackTemplate) (err error) {
	return nil
}
