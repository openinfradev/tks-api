package repository

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProjectUser struct {
	gorm.Model
	Id        uuid.UUID `gorm:"primarykey;type:uuid;"`
	ProjectId string
	UserId    uuid.UUID
}

func (g *ProjectUser) BeforeCreate(tx *gorm.DB) (err error) {
	g.Id = uuid.New()
	return nil
}

func (r *Repository) AddUserInProject(userId string, projectId string) (err error) {
	projectUser := ProjectUser{ProjectId: projectId, UserId: uuid.MustParse(userId)}
	err = r.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Create(&projectUser)
		if res.Error != nil {
			return res.Error
		}
		return nil
	})

	return nil
}

func (r *Repository) GetUserInProject(user *User, projectId string, userId string) (err error) {
	res := r.db.Table("users").Select("users.*").Joins("join project_users on project_users.user_id::text = users.id::text").
		Where("project_users.project_id::text = ? AND project_users.user_id::text = ? AND project_users.deleted_at IS NULL", projectId, userId).Scan(user)
	if res.RowsAffected != 0 && res.Error != nil {
		return fmt.Errorf("failed to get user in project %s", res.Error)
	}

	return nil
}

func (r *Repository) GetUsersInProject(users *[]User, projectId string) (err error) {
	//res := r.db.Find(&projectUsers, "projectId = ?", projectId)
	res := r.db.Table("users").Select("users.*").Joins("join project_users on project_users.user_id::text = users.id::text").
		Where("project_users.project_id::text = ? AND project_users.deleted_at IS NULL", projectId).Scan(users)
	if res.RowsAffected != 0 && res.Error != nil {
		return fmt.Errorf("failed to get users in project %s", res.Error)
	}

	return nil
}

func (r *Repository) GetProjectIdsByUser(projectIds *[]string, accountId string) (err error) {
	res := r.db.Table("project_users").Select("project_users.project_id").Joins("join users on users.id::text = project_users.user_id::text").
		Where("users.account_id = ? AND project_users.deleted_at IS NULL", accountId).Scan(projectIds)
	if res.RowsAffected != 0 && res.Error != nil {
		return fmt.Errorf("failed to get projectIds by user %s", res.Error)
	}

	return nil
}

func (r *Repository) RemoveUserFromProject(projectId string, userId string) (err error) {
	res := r.db.Where("user_id=? AND project_id=?", userId, projectId).Delete(&ProjectUser{})
	if res.RowsAffected != 0 && res.Error != nil {
		return fmt.Errorf("failed to remove user from project %s", res.Error)
	}
	return nil
}
