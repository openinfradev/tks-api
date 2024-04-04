package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/pagination"
)

// Interfaces
type ISystemNotificationRuleRepository interface {
	Get(ctx context.Context, systemNotificationRuleId uuid.UUID) (model.SystemNotificationRule, error)
	GetByName(ctx context.Context, name string) (model.SystemNotificationRule, error)
	Fetch(ctx context.Context, pg *pagination.Pagination) ([]model.SystemNotificationRule, error)
	FetchWithOrganization(ctx context.Context, organizationId string, pg *pagination.Pagination) (out []model.SystemNotificationRule, err error)
	Create(ctx context.Context, dto model.SystemNotificationRule) (systemNotificationRuleId uuid.UUID, err error)
	Creates(ctx context.Context, dto []model.SystemNotificationRule) (err error)
	Update(ctx context.Context, dto model.SystemNotificationRule) (err error)
	Delete(ctx context.Context, dto model.SystemNotificationRule) (err error)
}

type SystemNotificationRuleRepository struct {
	db *gorm.DB
}

func NewSystemNotificationRuleRepository(db *gorm.DB) ISystemNotificationRuleRepository {
	return &SystemNotificationRuleRepository{
		db: db,
	}
}

// Logics
func (r *SystemNotificationRuleRepository) Get(ctx context.Context, systemNotificationRuleId uuid.UUID) (out model.SystemNotificationRule, err error) {
	res := r.db.WithContext(ctx).Preload(clause.Associations).First(&out, "id = ?", systemNotificationRuleId)
	if res.Error != nil {
		return model.SystemNotificationRule{}, res.Error
	}
	return
}

func (r *SystemNotificationRuleRepository) GetByName(ctx context.Context, name string) (out model.SystemNotificationRule, err error) {
	res := r.db.WithContext(ctx).First(&out, "name = ?", name)
	if res.Error != nil {
		return out, res.Error
	}
	return
}

func (r *SystemNotificationRuleRepository) Fetch(ctx context.Context, pg *pagination.Pagination) (out []model.SystemNotificationRule, err error) {
	if pg == nil {
		pg = pagination.NewPagination(nil)
	}

	_, res := pg.Fetch(r.db.WithContext(ctx).Preload(clause.Associations), &out)
	if res.Error != nil {
		return nil, res.Error
	}
	return
}

func (r *SystemNotificationRuleRepository) FetchWithOrganization(ctx context.Context, organizationId string, pg *pagination.Pagination) (out []model.SystemNotificationRule, err error) {
	if pg == nil {
		pg = pagination.NewPagination(nil)
	}

	_, res := pg.Fetch(
		r.db.WithContext(ctx).Preload(clause.Associations).
			Where("organization_id = ?", organizationId),
		&out)
	if res.Error != nil {
		return nil, res.Error
	}
	return
}

func (r *SystemNotificationRuleRepository) Create(ctx context.Context, dto model.SystemNotificationRule) (systemNotificationRuleId uuid.UUID, err error) {
	dto.ID = uuid.New()
	res := r.db.WithContext(ctx).Create(&dto)
	if res.Error != nil {
		return uuid.Nil, res.Error
	}

	return dto.ID, nil
}

func (r *SystemNotificationRuleRepository) Creates(ctx context.Context, rules []model.SystemNotificationRule) (err error) {
	res := r.db.WithContext(ctx).Create(&rules)
	if res.Error != nil {
		return res.Error
	}

	return nil
}

func (r *SystemNotificationRuleRepository) Update(ctx context.Context, dto model.SystemNotificationRule) (err error) {
	var m model.SystemNotificationRule
	res := r.db.WithContext(ctx).Preload(clause.Associations).First(&m, "id = ?", dto.ID)
	if res.Error != nil {
		return res.Error
	}

	m.Name = dto.Name
	m.Description = dto.Description
	m.SystemNotificationTemplateId = dto.SystemNotificationTemplateId
	m.SystemNotificationCondition = dto.SystemNotificationCondition
	m.MessageTitle = dto.MessageTitle
	m.MessageContent = dto.MessageContent
	m.MessageActionProposal = dto.MessageActionProposal
	m.UpdatorId = dto.UpdatorId

	res = r.db.WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: true}).Updates(&m)
	if res.Error != nil {
		return res.Error
	}

	err = r.db.WithContext(ctx).Model(&m).Association("TargetUsers").Replace(dto.TargetUsers)
	if err != nil {
		return err
	}

	return nil
}

func (r *SystemNotificationRuleRepository) Delete(ctx context.Context, dto model.SystemNotificationRule) (err error) {
	res := r.db.WithContext(ctx).Delete(&model.SystemNotificationRule{}, "id = ?", dto.ID)
	if res.Error != nil {
		return res.Error
	}
	return nil
}
