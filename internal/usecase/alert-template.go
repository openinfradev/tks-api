package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/middleware/auth/request"
	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type IAlertTemplateUsecase interface {
	Get(ctx context.Context, alertId uuid.UUID) (model.AlertTemplate, error)
	GetByName(ctx context.Context, name string) (model.AlertTemplate, error)
	Fetch(ctx context.Context, pg *pagination.Pagination) ([]model.AlertTemplate, error)
	Create(ctx context.Context, dto model.AlertTemplate) (alertTemplateId uuid.UUID, err error)
	Update(ctx context.Context, dto model.AlertTemplate) error
	Delete(ctx context.Context, dto model.AlertTemplate) error
}

type AlertTemplateUsecase struct {
	repo             repository.IAlertTemplateRepository
	clusterRepo      repository.IClusterRepository
	organizationRepo repository.IOrganizationRepository
	appGroupRepo     repository.IAppGroupRepository
}

func NewAlertTemplateUsecase(r repository.Repository) IAlertTemplateUsecase {
	return &AlertTemplateUsecase{
		repo:             r.AlertTemplate,
		clusterRepo:      r.Cluster,
		appGroupRepo:     r.AppGroup,
		organizationRepo: r.Organization,
	}
}

func (u *AlertTemplateUsecase) Create(ctx context.Context, dto model.AlertTemplate) (alertTemplateId uuid.UUID, err error) {
	user, ok := request.UserFrom(ctx)
	if !ok {
		return uuid.Nil, httpErrors.NewBadRequestError(fmt.Errorf("Invalid token"), "", "")
	}
	userId := user.GetUserId()
	dto.CreatorId = &userId
	dto.UpdatorId = &userId

	if _, err = u.GetByName(ctx, dto.Name); err == nil {
		return uuid.Nil, httpErrors.NewBadRequestError(fmt.Errorf("duplicate alertTemplate name"), "AT_CREATE_ALREADY_EXISTED_NAME", "")
	}

	alertTemplateId, err = u.repo.Create(dto)
	if err != nil {
		return uuid.Nil, httpErrors.NewInternalServerError(err, "", "")
	}
	log.InfoWithContext(ctx, "newly created AlertTemplate ID:", alertTemplateId)

	dto.ID = alertTemplateId
	err = u.UpdateOrganizations(ctx, dto)
	if err != nil {
		return uuid.Nil, err
	}

	return alertTemplateId, nil
}

func (u *AlertTemplateUsecase) Update(ctx context.Context, dto model.AlertTemplate) error {
	_, err := u.repo.Get(dto.ID)
	if err != nil {
		return httpErrors.NewBadRequestError(err, "AT_NOT_EXISTED_ALERT_TEMPLATE", "")
	}

	err = u.repo.Update(dto)
	if err != nil {
		return err
	}

	err = u.UpdateOrganizations(ctx, dto)
	if err != nil {
		return err
	}
	return nil
}

func (u *AlertTemplateUsecase) Get(ctx context.Context, alertTemplateId uuid.UUID) (alert model.AlertTemplate, err error) {
	alert, err = u.repo.Get(alertTemplateId)
	if err != nil {
		return alert, err
	}
	return
}

func (u *AlertTemplateUsecase) GetByName(ctx context.Context, name string) (out model.AlertTemplate, err error) {
	out, err = u.repo.GetByName(name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return out, httpErrors.NewNotFoundError(err, "AT_FAILED_FETCH_ALERT_TEMPLATE", "")
		}
		return out, err
	}

	return
}

func (u *AlertTemplateUsecase) Fetch(ctx context.Context, pg *pagination.Pagination) (alerts []model.AlertTemplate, err error) {
	alerts, err = u.repo.Fetch(pg)
	if err != nil {
		return nil, err
	}
	return
}

func (u *AlertTemplateUsecase) Delete(ctx context.Context, dto model.AlertTemplate) (err error) {
	return nil
}

func (u *AlertTemplateUsecase) UpdateOrganizations(ctx context.Context, dto model.AlertTemplate) error {
	_, err := u.repo.Get(dto.ID)
	if err != nil {
		return httpErrors.NewBadRequestError(err, "AT_NOT_EXISTED_ALERT_TEMPLATE", "")
	}

	organizations := make([]model.Organization, 0)
	for _, organizationId := range dto.OrganizationIds {
		organization, err := u.organizationRepo.Get(organizationId)
		if err == nil {
			organizations = append(organizations, organization)
		}
	}

	err = u.repo.UpdateOrganizations(dto.ID, organizations)
	if err != nil {
		return httpErrors.NewBadRequestError(err, "AT_FAILED_UPDATE_ORGANIZATION", "")
	}

	return nil
}
