package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/openinfradev/tks-api/internal/domain"
)

// Interface
type IUserRepository interface {
	Get(userId uuid.UUID) (user domain.User, err error)
	Signin(accountId string, password string) (user domain.User, err error)
}

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) IUserRepository {
	return &UserRepository{
		db: db,
	}
}

// Models
type User struct {
	Id        uuid.UUID `gorm:"primarykey;type:uuid;"`
	AccountId string    `gorm:"uniqueIndex"`
	Name      string
	Password  string
}

func (g *User) BeforeCreate(tx *gorm.DB) (err error) {
	g.Id = uuid.New()
	return nil
}

// Logics
func (r *UserRepository) Get(userId uuid.UUID) (res domain.User, err error) {
	/*
		resHistories := make([]domain.History, 0)

		for _, history := range histories {
			outHistory := domain.History{}

			u.reflect(&outHistory, history)
			resHistories = append(resHistories, outHistory)
		}
	*/

	return domain.User{}, nil
}

func (r *UserRepository) Signin(accountId string, password string) (user domain.User, err error) {
	/*
		resHistories := make([]domain.History, 0)

		for _, history := range histories {
			outHistory := domain.History{}

			u.reflect(&outHistory, history)
			resHistories = append(resHistories, outHistory)
		}
	*/

	return domain.User{}, nil
}
