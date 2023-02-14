package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/openinfradev/tks-api/internal/domain"
)

// Interface
type IHistoryRepository interface {
	Fetch() (res []domain.History, err error)
}

type HistoryRepository struct {
	db *gorm.DB
}

func NewHistoryRepository(db *gorm.DB) IHistoryRepository {
	return &HistoryRepository{
		db: db,
	}
}

// Models
type History struct {
	gorm.Model
	Id          uuid.UUID `gorm:"primarykey;type:uuid;"`
	UserId      uuid.UUID
	HistoryType string
	ProjectId   string
	Description string
}

type HistoryWithUser struct {
	gorm.Model
	Id          uuid.UUID `gorm:"primarykey;type:uuid;"`
	UserId      uuid.UUID
	AccountId   string
	HistoryType string
	ProjectId   string
	Description string
}

func (g *History) BeforeCreate(tx *gorm.DB) (err error) {
	g.Id = uuid.New()
	return nil
}

// Logics
func (r *HistoryRepository) Fetch() (res []domain.History, err error) {
	/*
		resHistories := make([]domain.History, 0)

		for _, history := range histories {
			outHistory := domain.History{}

			u.reflect(&outHistory, history)
			resHistories = append(resHistories, outHistory)
		}
	*/

	return nil, nil
}

func (u *HistoryRepository) reflect(out *domain.History, history History) {
	out.Id = history.Id.String()
	out.Description = history.Description
	out.HistoryType = history.HistoryType
	out.CreatedAt = history.CreatedAt
	out.UpdatedAt = history.UpdatedAt
}
