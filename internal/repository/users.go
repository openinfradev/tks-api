package repository

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Id        uuid.UUID `gorm:"primarykey;type:uuid;"`
	AccountId string    `gorm:"uniqueIndex"`
	Name      string
	Password  string
	Tutorial  bool `gorm:"default:true"`
	Role      string
	Initial   bool `gorm:"default:true"`
}

func (g *User) BeforeCreate(tx *gorm.DB) (err error) {
	g.Id = uuid.New()
	return nil
}

func (r *Repository) CreateUser(user *User) (err error) {
	err = r.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Create(user)
		if res.Error != nil {
			return res.Error
		}
		return nil
	})

	return nil
}

func (r *Repository) GetUserByAccountId(user *User, accountId string) (err error) {
	res := r.db.Find(&user, "account_id = ?", accountId)
	if res.RowsAffected == 0 || res.Error != nil {
		return fmt.Errorf("Not found user")
	}

	return nil
}

func (r *Repository) GetUserById(user *User, userId uuid.UUID) (err error) {
	res := r.db.Find(&user, "id = ?", userId)
	if res.RowsAffected == 0 || res.Error != nil {
		return fmt.Errorf("Not found user")
	}

	return nil
}

func (r *Repository) GetUserByAccountIdAndPassword(user *User, accountId string, hashedPwd string) (err error) {
	res := r.db.Find(&user, "account_id = ? and password = ? ", accountId, hashedPwd)
	if res.RowsAffected == 0 || res.Error != nil {
		return fmt.Errorf("Not found user")
	}

	return nil
}

func (r *Repository) GetUsers(users *[]User) (err error) {
	res := r.db.Find(&users)
	if res.RowsAffected == 0 || res.Error != nil {
		return fmt.Errorf("No users")
	}

	return nil
}

func (r *Repository) UpdatePassword(id uuid.UUID, newPassword string) (err error) {
	res := r.db.Model(&User{}).
		Where("id = ?", id).
		//Update("password", newPassword)
		Updates(map[string]interface{}{"password": newPassword, "initial": false})

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("could not update password")
	}
	return nil

}

func (r *Repository) UpdateRole(id uuid.UUID, role string) (err error) {
	res := r.db.Model(&User{}).
		Where("id = ?", id).
		Update("role", role)

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("could not update password")
	}
	return nil

}
