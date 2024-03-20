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

type ISystemNotificationTemplateUsecase interface {
	Get(ctx context.Context, alertId uuid.UUID) (model.SystemNotificationTemplate, error)
	GetByName(ctx context.Context, name string) (model.SystemNotificationTemplate, error)
	Fetch(ctx context.Context, pg *pagination.Pagination) ([]model.SystemNotificationTemplate, error)
	Create(ctx context.Context, dto model.SystemNotificationTemplate) (systemNotificationTemplate uuid.UUID, err error)
	Update(ctx context.Context, dto model.SystemNotificationTemplate) error
	Delete(ctx context.Context, dto model.SystemNotificationTemplate) error
}

type SystemNotificationTemplateUsecase struct {
	repo             repository.ISystemNotificationTemplateRepository
	clusterRepo      repository.IClusterRepository
	organizationRepo repository.IOrganizationRepository
	appGroupRepo     repository.IAppGroupRepository
}

func NewSystemNotificationTemplateUsecase(r repository.Repository) ISystemNotificationTemplateUsecase {
	return &SystemNotificationTemplateUsecase{
		repo:             r.SystemNotificationTemplate,
		clusterRepo:      r.Cluster,
		appGroupRepo:     r.AppGroup,
		organizationRepo: r.Organization,
	}
}

func (u *SystemNotificationTemplateUsecase) Create(ctx context.Context, dto model.SystemNotificationTemplate) (systemNotificationTemplate uuid.UUID, err error) {
	user, ok := request.UserFrom(ctx)
	if !ok {
		return uuid.Nil, httpErrors.NewBadRequestError(fmt.Errorf("Invalid token"), "", "")
	}
	userId := user.GetUserId()
	dto.CreatorId = &userId
	dto.UpdatorId = &userId

	if _, err = u.GetByName(ctx, dto.Name); err == nil {
		return uuid.Nil, httpErrors.NewBadRequestError(fmt.Errorf("duplicate systemNotificationTemplate name"), "SNT_CREATE_ALREADY_EXISTED_NAME", "")
	}

	systemNotificationTemplate, err = u.repo.Create(ctx, dto)
	if err != nil {
		return uuid.Nil, httpErrors.NewInternalServerError(err, "", "")
	}
	log.Info(ctx, "newly created SystemNotificationTemplate ID:", systemNotificationTemplate)

	dto.ID = systemNotificationTemplate
	err = u.UpdateOrganizations(ctx, dto)
	if err != nil {
		return uuid.Nil, err
	}

	return systemNotificationTemplate, nil
}

func (u *SystemNotificationTemplateUsecase) Update(ctx context.Context, dto model.SystemNotificationTemplate) error {
	_, err := u.repo.Get(ctx, dto.ID)
	if err != nil {
		return httpErrors.NewBadRequestError(err, "SNT_NOT_EXISTED_ALERT_TEMPLATE", "")
	}

	err = u.repo.Update(ctx, dto)
	if err != nil {
		return err
	}

	err = u.UpdateOrganizations(ctx, dto)
	if err != nil {
		return err
	}
	return nil
}

func (u *SystemNotificationTemplateUsecase) Get(ctx context.Context, systemNotificationTemplate uuid.UUID) (alert model.SystemNotificationTemplate, err error) {
	alert, err = u.repo.Get(ctx, systemNotificationTemplate)
	if err != nil {
		return alert, err
	}
	return
}

func (u *SystemNotificationTemplateUsecase) GetByName(ctx context.Context, name string) (out model.SystemNotificationTemplate, err error) {
	out, err = u.repo.GetByName(ctx, name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return out, httpErrors.NewNotFoundError(err, "SNT_FAILED_FETCH_ALERT_TEMPLATE", "")
		}
		return out, err
	}

	return
}

func (u *SystemNotificationTemplateUsecase) Fetch(ctx context.Context, pg *pagination.Pagination) (alerts []model.SystemNotificationTemplate, err error) {
	alerts, err = u.repo.Fetch(ctx, pg)
	if err != nil {
		return nil, err
	}
	return
}

func (u *SystemNotificationTemplateUsecase) Delete(ctx context.Context, dto model.SystemNotificationTemplate) (err error) {
	return nil
}

func (u *SystemNotificationTemplateUsecase) UpdateOrganizations(ctx context.Context, dto model.SystemNotificationTemplate) error {
	_, err := u.repo.Get(ctx, dto.ID)
	if err != nil {
		return httpErrors.NewBadRequestError(err, "SNT_NOT_EXISTED_ALERT_TEMPLATE", "")
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
		return httpErrors.NewBadRequestError(err, "SNT_FAILED_UPDATE_ORGANIZATION", "")
	}

	return nil
}
