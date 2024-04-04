package domain

import (
	"time"

	"github.com/google/uuid"
)

type UserResponse struct {
	ID                uuid.UUID `gorm:"primarykey;type:uuid" json:"id"`
	AccountId         string    `json:"accountId"`
	Password          string    `gorm:"-:all" json:"password"`
	Name              string    `json:"name"`
	Token             string    `json:"token"`
	RoleId            string
	Roles             []SimpleRoleResponse `json:"roles"`
	OrganizationId    string
	Organization      OrganizationResponse `gorm:"foreignKey:OrganizationId;references:ID" json:"organization"`
	Creator           string               `json:"creator"`
	CreatedAt         time.Time            `json:"createdAt"`
	UpdatedAt         time.Time            `json:"updatedAt"`
	PasswordUpdatedAt time.Time            `json:"passwordUpdatedAt"`
	PasswordExpired   bool                 `json:"passwordExpired"`

	Email       string `json:"email"`
	Department  string `json:"department"`
	Description string `json:"description"`
}

type CreateUserRequest struct {
	AccountId   string             `json:"accountId" validate:"required"`
	Password    string             `json:"password" validate:"required"`
	Name        string             `json:"name" validate:"name"`
	Email       string             `json:"email" validate:"required,email"`
	Department  string             `json:"department" validate:"min=0,max=50"`
	Roles       []UserCreationRole `json:"roles" validate:"required"`
	Description string             `json:"description" validate:"min=0,max=100"`
}

type UserCreationRole struct {
	ID *string `json:"id" validate:"required"`
}

type SimpleUserResponse struct {
	ID         string `json:"id"`
	AccountId  string `json:"accountId"`
	Name       string `json:"name"`
	Department string `json:"department"`
	Email      string `json:"email"`
}

type CreateUserResponse struct {
	User struct {
		ID           string               `json:"id"`
		AccountId    string               `json:"accountId"`
		Name         string               `json:"name"`
		Roles        []SimpleRoleResponse `json:"roles"`
		Organization OrganizationResponse `json:"organization"`
		Email        string               `json:"email"`
		Department   string               `json:"department"`
		Description  string               `json:"description"`
	} `json:"user"`
}

type GetUserResponse struct {
	User struct {
		ID           string               `json:"id"`
		AccountId    string               `json:"accountId"`
		Name         string               `json:"name"`
		Roles        []SimpleRoleResponse `json:"roles"`
		Organization OrganizationResponse `json:"organization"`
		Email        string               `json:"email"`
		Department   string               `json:"department"`
		Description  string               `json:"description"`
		Creator      string               `json:"creator"`
		CreatedAt    time.Time            `json:"createdAt"`
		UpdatedAt    time.Time            `json:"updatedAt"`
	} `json:"user"`
}

type ListUserResponse struct {
	Users      []ListUserBody     `json:"users"`
	Pagination PaginationResponse `json:"pagination"`
}
type ListUserBody struct {
	ID           string               `json:"id"`
	AccountId    string               `json:"accountId"`
	Name         string               `json:"name"`
	Roles        []SimpleRoleResponse `json:"roles"`
	Organization OrganizationResponse `json:"organization"`
	Email        string               `json:"email"`
	Department   string               `json:"department"`
	Description  string               `json:"description"`
	Creator      string               `json:"creator"`
	CreatedAt    time.Time            `json:"createdAt"`
	UpdatedAt    time.Time            `json:"updatedAt"`
}

type UpdateUserRequest struct {
	Name        string             `json:"name" validate:"required"`
	Roles       []UserCreationRole `json:"roles" validate:"required"`
	Email       string             `json:"email" validate:"required,email"`
	Department  string             `json:"department" validate:"min=0,max=50"`
	Description string             `json:"description" validate:"min=0,max=100"`
}

type UpdateUsersRequest struct {
	Users []struct {
		AccountId   string             `json:"accountId" validate:"required"`
		Name        string             `json:"name" validate:"required,name"`
		Roles       []UserCreationRole `json:"roles" validate:"required"`
		Email       string             `json:"email" validate:"required,email"`
		Department  string             `json:"department" validate:"min=0,max=50"`
		Description string             `json:"description" validate:"min=0,max=100"`
	} `json:"users"`
}

type UpdateUserResponse struct {
	User struct {
		ID           string               `json:"id"`
		AccountId    string               `json:"accountId"`
		Name         string               `json:"name"`
		Roles        []SimpleRoleResponse `json:"roles"`
		Organization OrganizationResponse `json:"organization"`
		Email        string               `json:"email"`
		Department   string               `json:"department"`
		Description  string               `json:"description"`
		CreatedAt    time.Time            `json:"createdAt"`
		UpdatedAt    time.Time            `json:"updatedAt"`
	} `json:"user"`
}

type GetMyProfileResponse struct {
	User struct {
		ID           string               `json:"id"`
		AccountId    string               `json:"accountId"`
		Name         string               `json:"name"`
		Roles        []SimpleRoleResponse `json:"roles"`
		Organization OrganizationResponse `json:"organization"`
		Email        string               `json:"email"`
		Department   string               `json:"department"`
	} `json:"user"`
}
type UpdateMyProfileRequest struct {
	Password   string `json:"password" validate:"required"`
	Name       string `json:"name" validate:"required,min=1,max=30"`
	Email      string `json:"email" validate:"required,email"`
	Department string `json:"department" validate:"required,min=0,max=50"`
}

type UpdateMyProfileResponse struct {
	User struct {
		ID           string               `json:"id"`
		AccountId    string               `json:"accountId"`
		Name         string               `json:"name"`
		Roles        []SimpleRoleResponse `json:"roles"`
		Organization OrganizationResponse `json:"organization"`
		Email        string               `json:"email"`
		Department   string               `json:"department"`
	} `json:"user"`
}

type UpdatePasswordRequest struct {
	OriginPassword string `json:"originPassword" validate:"required"`
	NewPassword    string `json:"newPassword" validate:"required"`
}

type CheckExistedResponse struct {
	Existed bool `json:"existed"`
}

type Admin_CreateUserRequest struct {
	AccountId     string             `json:"accountId" validate:"required"`
	Name          string             `json:"name" validate:"name"`
	Email         string             `json:"email" validate:"required,email"`
	Roles         []UserCreationRole `json:"roles" validate:"required"`
	Department    string             `json:"department" validate:"min=0,max=50"`
	Description   string             `json:"description" validate:"min=0,max=100"`
	AdminPassword string             `json:"adminPassword"`
}

type Admin_CreateUserResponse struct {
	ID string `json:"id"`
}

type Admin_UpdateUserRequest struct {
	Name          string             `json:"name" validate:"name"`
	Email         string             `json:"email" validate:"required,email"`
	Department    string             `json:"department" validate:"min=0,max=50"`
	Roles         []UserCreationRole `json:"roles" validate:"required"`
	Description   string             `json:"description" validate:"min=0,max=100"`
	AdminPassword string             `json:"adminPassword"`
}

type Admin_UpdateUserResponse struct {
	User struct {
		ID           string               `json:"id"`
		AccountId    string               `json:"accountId"`
		Name         string               `json:"name"`
		Roles        []SimpleRoleResponse `json:"roles"`
		Organization OrganizationResponse `json:"organization"`
		Email        string               `json:"email"`
		Department   string               `json:"department"`
		Description  string               `json:"description"`
		CreatedAt    time.Time            `json:"createdAt"`
		UpdatedAt    time.Time            `json:"updatedAt"`
	} `json:"user"`
}

type DeleteUserRequest struct {
	AdminPassword string `json:"adminPassword"`
}
