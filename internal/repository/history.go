package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/openinfradev/tks-api/pkg/domain"
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
	ID          uuid.UUID `gorm:"primarykey;type:uuid;"`
	UserId      uuid.UUID
	HistoryType string
	ProjectId   string
	Description string
}

type HistoryWithUser struct {
	gorm.Model
	ID          uuid.UUID `gorm:"primarykey;type:uuid;"`
	UserId      uuid.UUID
	AccountId   string
	HistoryType string
	ProjectId   string
	Description string
}

func (g *History) BeforeCreate(tx *gorm.DB) (err error) {
	g.ID = uuid.New()
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
