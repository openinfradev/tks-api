package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

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
	res := r.db.WithContext(ctx).First(&out, "id = ?", auditId)
	if res.Error != nil {
		return
	}
	return
}

func (r *AuditRepository) Fetch(ctx context.Context, pg *pagination.Pagination) (out []model.Audit, err error) {
	if pg == nil {
		pg = pagination.NewPagination(nil)
	}

	db := r.db.WithContext(ctx).Model(&model.Audit{})

	_, res := pg.Fetch(db, &out)
	if res.Error != nil {
		return nil, res.Error
	}
	return
}

func (r *AuditRepository) Create(ctx context.Context, dto model.Audit) (auditId uuid.UUID, err error) {
	dto.ID = uuid.New()
	res := r.db.WithContext(ctx).Create(&dto)
	if res.Error != nil {
		return uuid.Nil, res.Error
	}
	return dto.ID, nil
}

func (r *AuditRepository) Delete(ctx context.Context, auditId uuid.UUID) (err error) {
	return fmt.Errorf("to be implemented")
}
