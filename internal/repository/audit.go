package repository

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/pagination"
)

// Interfaces
type IAuditRepository interface {
	Get(auditId uuid.UUID) (model.Audit, error)
	Fetch(organizationId string, pg *pagination.Pagination) ([]model.Audit, error)
	Create(dto model.Audit) (auditId uuid.UUID, err error)
	Delete(auditId uuid.UUID) (err error)
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
func (r *AuditRepository) Get(auditId uuid.UUID) (out model.Audit, err error) {
	res := r.db.Preload(clause.Associations).First(&out, "id = ?", auditId)
	if res.Error != nil {
		return
	}
	return
}

func (r *AuditRepository) Fetch(organizationId string, pg *pagination.Pagination) (out []model.Audit, err error) {
	if pg == nil {
		pg = pagination.NewPagination(nil)
	}

	db := r.db.Model(&model.Audit{}).Preload(clause.Associations).Where("audits.organization_id = ?", organizationId)
	_, res := pg.Fetch(db, &out)
	if res.Error != nil {
		return nil, res.Error
	}
	return
}

func (r *AuditRepository) Create(dto model.Audit) (auditId uuid.UUID, err error) {
	audit := model.Audit{
		OrganizationId: dto.OrganizationId,
		Group:          dto.Group,
		Message:        dto.Message,
		Description:    dto.Description,
		ClientIP:       dto.ClientIP,
		UserId:         dto.UserId}
	res := r.db.Create(&audit)
	if res.Error != nil {
		return uuid.Nil, res.Error
	}
	return audit.ID, nil
}

func (r *AuditRepository) Delete(auditId uuid.UUID) (err error) {
	return fmt.Errorf("to be implemented")
}
