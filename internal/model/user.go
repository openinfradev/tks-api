package model

import (
	"time"

	"gorm.io/gorm"

	"github.com/google/uuid"
)

type User struct {
	gorm.Model

	ID                uuid.UUID `gorm:"primarykey;type:uuid" json:"id"`
	AccountId         string    `json:"accountId"`
	Password          string    `gorm:"-:all" json:"password"`
	Name              string    `json:"name"`
	Token             string    `json:"token"`
	Roles             []Role    `gorm:"many2many:user_roles;" json:"roles"`
	OrganizationId    string
	Organization      Organization `gorm:"foreignKey:OrganizationId;references:ID" json:"organization"`
	Creator           string       `json:"creator"`
	CreatedAt         time.Time    `json:"createdAt"`
	UpdatedAt         time.Time    `json:"updatedAt"`
	PasswordUpdatedAt time.Time    `json:"passwordUpdatedAt"`
	PasswordExpired   bool         `json:"passwordExpired"`

	Email       string `json:"email"`
	Department  string `json:"department"`
	Description string `json:"description"`
}

func (u *User) BeforeDelete(db *gorm.DB) (err error) {
	err = db.Table("user_roles").Unscoped().Where("user_id = ?", u.ID).Delete(nil).Error
	if err != nil {
		return err
	}
	return nil
}
