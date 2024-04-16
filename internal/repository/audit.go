package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/pagination"
)

// Interfaces
type IAuditRepository interface {
	Get(ctx context.Context, auditId uuid.UUID) (model.Audit, error)
	Fetch(ctx context.Context, pg *pagination.Pagination) ([]model.Audit, error)
	Create(ctx context.Context, dto model.Audit) (auditId uuid.UUID, err error)
	Delete(ctx context.Context, auditId uuid.UUID) (err error)
}

type AuditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) IAuditRepository {
	return &AuditRepository{
		db: db,
	}
}

// Logics
func (r *AuditRepository) Get(ctx context.Context, auditId uuid.UUID) (out model.Audit, err error) {
	res := r.db.WithContext(ctx).Preload(clause.Associations).First(&out, "id = ?", auditId)
	if res.Error != nil {
		return
	}
	return
}

func (r *AuditRepository) Fetch(ctx context.Context, pg *pagination.Pagination) (out []model.Audit, err error) {
	if pg == nil {
		pg = pagination.NewPagination(nil)
	}

	db := r.db.WithContext(ctx).Model(&model.Audit{}).Preload(clause.Associations).Preload("User.Roles")

	_, res := pg.Fetch(db, &out)
	if res.Error != nil {
		return nil, res.Error
	}
	return
}

func (r *AuditRepository) Create(ctx context.Context, dto model.Audit) (auditId uuid.UUID, err error) {
	audit := model.Audit{
		ID:             uuid.New(),
		OrganizationId: dto.OrganizationId,
		Group:          dto.Group,
		Message:        dto.Message,
		Description:    dto.Description,
		ClientIP:       dto.ClientIP,
		UserId:         dto.UserId}
	res := r.db.WithContext(ctx).Create(&audit)
	if res.Error != nil {
		return uuid.Nil, res.Error
	}
	return audit.ID, nil
}

func (r *AuditRepository) Delete(ctx context.Context, auditId uuid.UUID) (err error) {
	return fmt.Errorf("to be implemented")
}
