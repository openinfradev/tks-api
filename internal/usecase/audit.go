package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/middleware/auth/request"
	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
)

type IAuditUsecase interface {
	Get(ctx context.Context, auditId uuid.UUID) (model.Audit, error)
	Fetch(ctx context.Context, pg *pagination.Pagination) ([]model.Audit, error)
	Create(ctx context.Context, dto model.Audit) (auditId uuid.UUID, err error)
	Delete(ctx context.Context, dto model.Audit) error
}

type AuditUsecase struct {
	repo repository.IAuditRepository
}

func NewAuditUsecase(r repository.Repository) IAuditUsecase {
	return &AuditUsecase{
		repo: r.Audit,
	}
}

func (u *AuditUsecase) Create(ctx context.Context, dto model.Audit) (auditId uuid.UUID, err error) {
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

func (u *AuditUsecase) Get(ctx context.Context, auditId uuid.UUID) (res model.Audit, err error) {
	res, err = u.repo.Get(auditId)
	if err != nil {
		return model.Audit{}, err
	}
	return
}

func (u *AuditUsecase) Fetch(ctx context.Context, pg *pagination.Pagination) (audits []model.Audit, err error) {
	audits, err = u.repo.Fetch(pg)
	if err != nil {
		return nil, err
	}
	return
}

func (u *AuditUsecase) Delete(ctx context.Context, dto model.Audit) (err error) {
	err = u.repo.Delete(dto.ID)
	if err != nil {
		return httpErrors.NewNotFoundError(err, "", "")
	}
	return nil
}
