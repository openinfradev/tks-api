package usecase

import (
	"context"

	"github.com/google/uuid"
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
	repo             repository.IAuditRepository
	userRepo         repository.IUserRepository
	organizationRepo repository.IOrganizationRepository
}

func NewAuditUsecase(r repository.Repository) IAuditUsecase {
	return &AuditUsecase{
		repo:             r.Audit,
		userRepo:         r.User,
		organizationRepo: r.Organization,
	}
}

func (u *AuditUsecase) Create(ctx context.Context, dto model.Audit) (auditId uuid.UUID, err error) {
	if dto.OrganizationId != "" {
		organization, err := u.organizationRepo.Get(ctx, dto.OrganizationId)
		if err == nil {
			dto.OrganizationName = organization.Name
		}
	}

	if dto.UserId != nil && *dto.UserId != uuid.Nil {
		user, err := u.userRepo.GetByUuid(ctx, *dto.UserId)
		if err != nil {
			return auditId, err
		}

		userRoles := ""
		for i, role := range user.Roles {
			if i > 0 {
				userRoles = userRoles + ","
			}
			userRoles = userRoles + role.Name
		}
		dto.OrganizationId = user.Organization.ID
		dto.OrganizationName = user.Organization.Name
		dto.UserAccountId = user.AccountId
		dto.UserName = user.Name
		dto.UserRoles = userRoles
	}

	auditId, err = u.repo.Create(ctx, dto)
	if err != nil {
		return uuid.Nil, httpErrors.NewInternalServerError(err, "", "")
	}
	return auditId, nil
}

func (u *AuditUsecase) Get(ctx context.Context, auditId uuid.UUID) (res model.Audit, err error) {
	res, err = u.repo.Get(ctx, auditId)
	if err != nil {
		return model.Audit{}, err
	}
	return
}

func (u *AuditUsecase) Fetch(ctx context.Context, pg *pagination.Pagination) (audits []model.Audit, err error) {
	audits, err = u.repo.Fetch(ctx, pg)
	if err != nil {
		return nil, err
	}
	return
}

func (u *AuditUsecase) Delete(ctx context.Context, dto model.Audit) (err error) {
	err = u.repo.Delete(ctx, dto.ID)
	if err != nil {
		return httpErrors.NewNotFoundError(err, "", "")
	}
	return nil
}
