package repository

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/internal/serializer"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/log"
)

// Interfaces
type IAuditRepository interface {
	Get(auditId uuid.UUID) (domain.Audit, error)
	Fetch(organizationId string, pg *pagination.Pagination) ([]domain.Audit, error)
	Create(dto domain.Audit) (auditId uuid.UUID, err error)
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

// Models
type Audit struct {
	gorm.Model

	ID             uuid.UUID `gorm:"primarykey"`
	OrganizationId string
	Organization   Organization `gorm:"foreignKey:OrganizationId"`
	Group          string
	Message        string
	Description    string
	ClientIP       string
	UserId         *uuid.UUID `gorm:"type:uuid"`
	User           User       `gorm:"foreignKey:UserId"`
}

func (c *Audit) BeforeCreate(tx *gorm.DB) (err error) {
	c.ID = uuid.New()
	return nil
}

// Logics
func (r *AuditRepository) Get(auditId uuid.UUID) (out domain.Audit, err error) {
	var audit Audit
	res := r.db.Preload(clause.Associations).First(&audit, "id = ?", auditId)
	if res.Error != nil {
		return
	}
	out = reflectAudit(audit)
	return
}

func (r *AuditRepository) Fetch(organizationId string, pg *pagination.Pagination) (out []domain.Audit, err error) {
	var audits []Audit

	if pg == nil {
		pg = pagination.NewDefaultPagination()
	}

	db := r.db.Model(&Audit{}).Preload(clause.Associations).Where("organization_id = ?", organizationId)
	_, res := pg.Fetch(db, &audits)
	if res.Error != nil {
		return nil, res.Error
	}

	for _, audit := range audits {
		out = append(out, reflectAudit(audit))
	}

	return
}

func (r *AuditRepository) Create(dto domain.Audit) (auditId uuid.UUID, err error) {
	audit := Audit{
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

func reflectAudit(audit Audit) (out domain.Audit) {
	if err := serializer.Map(audit.Model, &out); err != nil {
		log.Error(err)
	}
	if err := serializer.Map(audit, &out); err != nil {
		log.Error(err)
	}
	return
}
