package repository

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/pkg/domain"
)

// Interfaces
type IAuditRepository interface {
	Get(auditId uuid.UUID) (domain.Audit, error)
	Fetch(pg *pagination.Pagination) ([]domain.Audit, error)
	Create(dto domain.Audit) (auditId uuid.UUID, err error)
	Delete(dto domain.Audit) (err error)
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
	return out, fmt.Errorf("to be implemented")
}

func (r *AuditRepository) Fetch(pg *pagination.Pagination) (out []domain.Audit, err error) {
	return out, fmt.Errorf("to be implemented")
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

func (r *AuditRepository) Delete(dto domain.Audit) (err error) {
	return fmt.Errorf("to be implemented")
}

/*
func reflectAudit(audit Audit) (out domain.Audit) {
	if err := serializer.Map(audit.Model, &out); err != nil {
		log.Error(err)
	}
	if err := serializer.Map(audit, &out); err != nil {
		log.Error(err)
	}
	return
}
*/
