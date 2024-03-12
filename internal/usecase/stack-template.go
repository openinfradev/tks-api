package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal"
	"github.com/openinfradev/tks-api/internal/middleware/auth/request"
	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type IStackTemplateUsecase interface {
	Get(ctx context.Context, stackTemplateId uuid.UUID) (model.StackTemplate, error)
	Fetch(ctx context.Context, pg *pagination.Pagination) ([]model.StackTemplate, error)
	FetchWithOrganization(ctx context.Context, organizationId string, pg *pagination.Pagination) ([]model.StackTemplate, error)
	Create(ctx context.Context, dto model.StackTemplate) (stackTemplate uuid.UUID, err error)
	Update(ctx context.Context, dto model.StackTemplate) error
	Delete(ctx context.Context, dto model.StackTemplate) error
	UpdateOrganizations(ctx context.Context, dto model.StackTemplate) error
	GetByName(ctx context.Context, name string) (model.StackTemplate, error)
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

func (u *StackTemplateUsecase) Create(ctx context.Context, dto model.StackTemplate) (stackTemplateId uuid.UUID, err error) {
	user, ok := request.UserFrom(ctx)
	if !ok {
		return uuid.Nil, httpErrors.NewBadRequestError(fmt.Errorf("Invalid token"), "", "")
	}
	userId := user.GetUserId()
	dto.CreatorId = &userId
	dto.UpdatorId = &userId

	pg := pagination.NewPaginationWithFilter("name", "", "$eq", []string{dto.Name})
	stackTemplates, _ := u.Fetch(ctx, pg)
	if len(stackTemplates) > 0 {
		return uuid.Nil, httpErrors.NewBadRequestError(fmt.Errorf("duplicate stackTemplate name"), "ST_CREATE_ALREADY_EXISTED_NAME", "")
	}

	services := "["
	for _, serviceId := range dto.ServiceIds {
		switch serviceId {
		case "LMA":
			services = services + internal.SERVICE_LMA + ","
		case "SERVICE_MESH":
			services = services + internal.SERVICE_SERVICE_MESH
		}
	}
	services = services + "]"
	dto.Services = []byte(services)

	stackTemplateId, err = u.repo.Create(ctx, dto)
	if err != nil {
		return uuid.Nil, httpErrors.NewInternalServerError(err, "", "")
	}
	log.InfoWithContext(ctx, "newly created StackTemplate ID:", stackTemplateId)

	dto.ID = stackTemplateId
	err = u.UpdateOrganizations(ctx, dto)
	if err != nil {
		return uuid.Nil, err
	}

	return stackTemplateId, nil
}

func (u *StackTemplateUsecase) Update(ctx context.Context, dto model.StackTemplate) error {
	_, err := u.repo.Get(ctx, dto.ID)
	if err != nil {
		return httpErrors.NewBadRequestError(err, "ST_NOT_EXISTED_STACK_TEMPLATE", "")
	}

	err = u.repo.Update(ctx, dto)
	if err != nil {
		return err
	}
	return nil
}

func (u *StackTemplateUsecase) Get(ctx context.Context, stackTemplateId uuid.UUID) (res model.StackTemplate, err error) {
	res, err = u.repo.Get(ctx, stackTemplateId)
	if err != nil {
		return res, err
	}
	return
}

func (u *StackTemplateUsecase) GetByName(ctx context.Context, name string) (out model.StackTemplate, err error) {
	out, err = u.repo.GetByName(ctx, name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return out, httpErrors.NewNotFoundError(err, "ST_FAILED_FETCH_STACK_TEMPLATE", "")
		}
		return out, err
	}

	return
}

func (u *StackTemplateUsecase) Fetch(ctx context.Context, pg *pagination.Pagination) (res []model.StackTemplate, err error) {
	res, err = u.repo.Fetch(ctx, pg)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (u *StackTemplateUsecase) FetchWithOrganization(ctx context.Context, organizationId string, pg *pagination.Pagination) (res []model.StackTemplate, err error) {
	res, err = u.repo.FetchWithOrganization(ctx, organizationId, pg)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (u *StackTemplateUsecase) Delete(ctx context.Context, dto model.StackTemplate) (err error) {
	return nil
}

func (u *StackTemplateUsecase) UpdateOrganizations(ctx context.Context, dto model.StackTemplate) error {
	_, err := u.repo.Get(ctx, dto.ID)
	if err != nil {
		return httpErrors.NewBadRequestError(err, "ST_NOT_EXISTED_STACK_TEMPLATE", "")
	}

	organizations := make([]model.Organization, 0)
	for _, organizationId := range dto.OrganizationIds {
		organization, err := u.organizationRepo.Get(ctx, organizationId)
		if err == nil {
			organizations = append(organizations, organization)
		}
	}

	err = u.repo.UpdateOrganizations(ctx, dto.ID, organizations)
	if err != nil {
		return httpErrors.NewBadRequestError(err, "ST_FAILED_UPDATE_ORGANIZATION", "")
	}

	return nil
}
