package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/middleware/auth/request"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
)

type IStackTemplateUsecase interface {
	Get(ctx context.Context, stackTemplateId uuid.UUID) (domain.StackTemplate, error)
	Fetch(ctx context.Context, pg *pagination.Pagination) ([]domain.StackTemplate, error)
	Create(ctx context.Context, dto domain.StackTemplate) (stackTemplate uuid.UUID, err error)
	Update(ctx context.Context, dto domain.StackTemplate) error
	Delete(ctx context.Context, dto domain.StackTemplate) error
	UpdateOrganization(ctx context.Context, stackTemplateId uuid.UUID, organizationIds []string) error
}

type StackTemplateUsecase struct {
	repo             repository.IStackTemplateRepository
	organizationRepo repository.IOrganizationRepository
}

func NewStackTemplateUsecase(r repository.Repository) IStackTemplateUsecase {
	return &StackTemplateUsecase{
		repo:             r.StackTemplate,
		organizationRepo: r.Organization,
	}
}

func (u *StackTemplateUsecase) Create(ctx context.Context, dto domain.StackTemplate) (stackTemplateId uuid.UUID, err error) {
	user, ok := request.UserFrom(ctx)
	if !ok {
		return uuid.Nil, httpErrors.NewBadRequestError(fmt.Errorf("Invalid token"), "", "")
	}

	dto.CreatorId = user.GetUserId()
	dto.UpdatorId = user.GetUserId()
	stackTemplateId, err = u.repo.Create(dto)
	if err != nil {
		return uuid.Nil, httpErrors.NewInternalServerError(err, "", "")
	}
	log.InfoWithContext(ctx, "newly created StackTemplate ID:", stackTemplateId)

	organizations := make([]domain.Organization, 0)
	for _, organizationId := range dto.OrganizationIds {
		organization, err := u.organizationRepo.Get(organizationId)
		if err == nil {
			organizations = append(organizations, organization)
		}
	}

	err = u.repo.UpdateOrganizations(stackTemplateId, organizations)
	if err != nil {
		return uuid.Nil, err
	}

	return stackTemplateId, nil
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

func (u *StackTemplateUsecase) Fetch(ctx context.Context, pg *pagination.Pagination) (res []domain.StackTemplate, err error) {
	res, err = u.repo.Fetch(pg)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (u *StackTemplateUsecase) Delete(ctx context.Context, dto domain.StackTemplate) (err error) {
	return nil
}

func (u *StackTemplateUsecase) UpdateOrganization(ctx context.Context, stackTemplateId uuid.UUID, organizationIds []string) error {
	return nil
}
