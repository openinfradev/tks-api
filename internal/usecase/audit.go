package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-api/pkg/domain"
)

type IAuditUsecase interface {
	Get(ctx context.Context, auditId uuid.UUID) (domain.Audit, error)
	Fetch(ctx context.Context, organizationId string, pg *pagination.Pagination) ([]domain.Audit, error)
	Create(ctx context.Context, dto domain.Audit) (auditId uuid.UUID, err error)
	Update(ctx context.Context, dto domain.Audit) error
	Delete(ctx context.Context, dto domain.Audit) error
}

type AuditUsecase struct {
	repo repository.IAuditRepository
}

func NewAuditUsecase(r repository.Repository) IAuditUsecase {
	return &AuditUsecase{
		repo: r.Audit,
	}
}

func (u *AuditUsecase) Create(ctx context.Context, dto domain.Audit) (auditId uuid.UUID, err error) {
	// to be implemented
	return
}

func (u *AuditUsecase) Update(ctx context.Context, dto domain.Audit) error {
	// to be implemented
	return nil
}

func (u *AuditUsecase) Get(ctx context.Context, auditId uuid.UUID) (res domain.Audit, err error) {
	// to be implemented
	return
}

func (u *AuditUsecase) Fetch(ctx context.Context, organizationId string, pg *pagination.Pagination) (audits []domain.Audit, err error) {
	// to be implemented
	return
}

func (u *AuditUsecase) Delete(ctx context.Context, dto domain.Audit) (err error) {
	// to be implemented
	return nil
}
