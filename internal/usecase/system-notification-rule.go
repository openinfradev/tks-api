package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/helper"
	"github.com/openinfradev/tks-api/internal/middleware/auth/request"
	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type ISystemNotificationRuleUsecase interface {
	Get(ctx context.Context, systemNotificationRuleId uuid.UUID) (model.SystemNotificationRule, error)
	Fetch(ctx context.Context, organizationId string, pg *pagination.Pagination) ([]model.SystemNotificationRule, error)
	Create(ctx context.Context, dto model.SystemNotificationRule) (systemNotificationRule uuid.UUID, err error)
	Update(ctx context.Context, dto model.SystemNotificationRule) error
	Delete(ctx context.Context, dto model.SystemNotificationRule) error
	GetByName(ctx context.Context, name string) (model.SystemNotificationRule, error)
}

type SystemNotificationRuleUsecase struct {
	repo                           repository.ISystemNotificationRuleRepository
	organizationRepo               repository.IOrganizationRepository
	userRepo                       repository.IUserRepository
	systemNotificationTemplateRepo repository.ISystemNotificationTemplateRepository
}

func NewSystemNotificationRuleUsecase(r repository.Repository) ISystemNotificationRuleUsecase {
	return &SystemNotificationRuleUsecase{
		repo:                           r.SystemNotificationRule,
		organizationRepo:               r.Organization,
		userRepo:                       r.User,
		systemNotificationTemplateRepo: r.SystemNotificationTemplate,
	}
}

func (u *SystemNotificationRuleUsecase) Create(ctx context.Context, dto model.SystemNotificationRule) (systemNotificationRuleId uuid.UUID, err error) {
	user, ok := request.UserFrom(ctx)
	if !ok {
		return uuid.Nil, httpErrors.NewBadRequestError(fmt.Errorf("Invalid token"), "", "")
	}
	userId := user.GetUserId()
	dto.CreatorId = &userId
	dto.UpdatorId = &userId

	if _, err = u.GetByName(ctx, dto.Name); err == nil {
		return uuid.Nil, httpErrors.NewBadRequestError(fmt.Errorf("duplicate systemNotificationRule name"), "SNR_CREATE_ALREADY_EXISTED_NAME", "")
	}

	// Users
	dto.TargetUsers = make([]model.User, 0)
	for _, strId := range dto.TargetUserIds {
		userId, err := uuid.Parse(strId)
		if err == nil {
			user, err := u.userRepo.GetByUuid(ctx, userId)
			if err == nil {
				dto.TargetUsers = append(dto.TargetUsers, user)
			}
		}
	}

	// Make parameters
	for i, condition := range dto.SystemNotificationConditions {
		dto.SystemNotificationConditions[i].Parameter = []byte(helper.ModelToJson(condition.Parameters))
	}

	systemNotificationRuleId, err = u.repo.Create(ctx, dto)
	if err != nil {
		return uuid.Nil, err
	}

	// [TODO] update kubernetes resources

	return
}

func (u *SystemNotificationRuleUsecase) Update(ctx context.Context, dto model.SystemNotificationRule) error {
	_, err := u.repo.Get(ctx, dto.ID)
	if err != nil {
		return httpErrors.NewBadRequestError(err, "SNR_NOT_EXISTED_STACK_TEMPLATE", "")
	}

	// Users
	dto.TargetUsers = make([]model.User, 0)
	for _, strId := range dto.TargetUserIds {
		userId, err := uuid.Parse(strId)
		if err == nil {
			user, err := u.userRepo.GetByUuid(ctx, userId)
			if err == nil {
				dto.TargetUsers = append(dto.TargetUsers, user)
			}
		}
	}

	for i, condition := range dto.SystemNotificationConditions {
		dto.SystemNotificationConditions[i].SystemNotificationRuleId = dto.ID
		dto.SystemNotificationConditions[i].Parameter = []byte(helper.ModelToJson(condition.Parameters))
	}

	err = u.repo.Update(ctx, dto)
	if err != nil {
		return err
	}

	// [TODO] update kubernetes resources

	return nil
}

func (u *SystemNotificationRuleUsecase) Get(ctx context.Context, systemNotificationRuleId uuid.UUID) (res model.SystemNotificationRule, err error) {
	res, err = u.repo.Get(ctx, systemNotificationRuleId)
	if err != nil {
		return res, err
	}
	return
}

func (u *SystemNotificationRuleUsecase) GetByName(ctx context.Context, name string) (out model.SystemNotificationRule, err error) {
	out, err = u.repo.GetByName(ctx, name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return out, httpErrors.NewNotFoundError(err, "SNR_FAILED_FETCH_SYSTEM_NOTIFICATION_RULE", "")
		}
		return out, err
	}

	return
}

func (u *SystemNotificationRuleUsecase) Fetch(ctx context.Context, organizationId string, pg *pagination.Pagination) (res []model.SystemNotificationRule, err error) {
	res, err = u.repo.FetchWithOrganization(ctx, organizationId, pg)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (u *SystemNotificationRuleUsecase) Delete(ctx context.Context, dto model.SystemNotificationRule) (err error) {
	return nil
}
