package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/middleware/auth/request"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
)

type IAuditUsecase interface {
	Get(ctx context.Context, auditId uuid.UUID) (domain.Audit, error)
	Fetch(ctx context.Context, organizationId string, pg *pagination.Pagination) ([]domain.Audit, error)
	Create(ctx context.Context, dto domain.Audit) (auditId uuid.UUID, err error)
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
	if dto.UserId == nil {
		user, ok := request.UserFrom(ctx)
		if ok {
			userId := user.GetUserId()
			dto.UserId = &userId
		}
	}
	auditId, err = u.repo.Create(dto)
	if err != nil {
		return uuid.Nil, httpErrors.NewInternalServerError(err, "", "")
	}
	return auditId, nil
}

func (u *AuditUsecase) Get(ctx context.Context, auditId uuid.UUID) (res domain.Audit, err error) {
	res, err = u.repo.Get(auditId)
	if err != nil {
		return domain.Audit{}, err
	}
	return
}

func (u *AuditUsecase) Fetch(ctx context.Context, organizationId string, pg *pagination.Pagination) (audits []domain.Audit, err error) {
	audits, err = u.repo.Fetch(organizationId, pg)
	if err != nil {
		return nil, err
	}
	return
}

func (u *AuditUsecase) Delete(ctx context.Context, dto domain.Audit) (err error) {
	err = u.repo.Delete(dto.ID)
	if err != nil {
		return httpErrors.NewNotFoundError(err, "", "")
	}
	return nil
}
