package domain

import (
	"time"

	"github.com/google/uuid"
)

// // Deprecated: Policy is deprecated, use Permission instead.
//
//	type Policy = struct {
//		ID               string    `json:"id"`
//		Name             string    `json:"name"`
//		Create           bool      `json:"create"`
//		CreatePriviledge string    `json:"createPriviledge"`
//		Update           bool      `json:"update"`
//		UpdatePriviledge string    `json:"updatePriviledge"`
//		Read             bool      `json:"read"`
//		ReadPriviledge   string    `json:"readPriviledge"`
//		Delete           bool      `json:"delete"`
//		DeletePriviledge string    `json:"deletePriviledge"`
//		Creator          string    `json:"creator"`
//		CreatedAt        time.Time `json:"createdAt"`
//		UpdatedAt        time.Time `json:"updatedAt"`
//	}
type UserResponse struct {
	ID                uuid.UUID `gorm:"primarykey;type:uuid" json:"id"`
	AccountId         string    `json:"accountId"`
	Password          string    `gorm:"-:all" json:"password"`
	Name              string    `json:"name"`
	Token             string    `json:"token"`
	RoleId            string
	Role              RoleResponse `gorm:"foreignKey:RoleId;references:ID" json:"role"`
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
	AccountId   string `json:"accountId" validate:"required"`
	Password    string `json:"password" validate:"required"`
	Name        string `json:"name" validate:"name"`
	Email       string `json:"email" validate:"required,email"`
	Department  string `json:"department" validate:"min=0,max=50"`
	Role        string `json:"role" validate:"required,oneof=admin user"`
	Description string `json:"description" validate:"min=0,max=100"`
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
		Role         RoleResponse         `json:"role"`
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
		Role         RoleResponse         `json:"role"`
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
	Role         RoleResponse         `json:"role"`
	Organization OrganizationResponse `json:"organization"`
	Email        string               `json:"email"`
	Department   string               `json:"department"`
	Description  string               `json:"description"`
	Creator      string               `json:"creator"`
	CreatedAt    time.Time            `json:"createdAt"`
	UpdatedAt    time.Time            `json:"updatedAt"`
}

type UpdateUserRequest struct {
	Name        string `json:"name" validate:"omitempty,min=1,max=30"`
	Role        string `json:"role" validate:"oneof=admin user"`
	Email       string `json:"email" validate:"omitempty,email"`
	Department  string `json:"department" validate:"omitempty,min=0,max=50"`
	Description string `json:"description" validate:"omitempty,min=0,max=100"`
}

type UpdateUserResponse struct {
	User struct {
		ID           string               `json:"id"`
		AccountId    string               `json:"accountId"`
		Name         string               `json:"name"`
		Role         RoleResponse         `json:"role"`
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
		Role         RoleResponse         `json:"role"`
		Organization OrganizationResponse `json:"organization"`
		Email        string               `json:"email"`
		Department   string               `json:"department"`
	} `json:"user"`
}
type UpdateMyProfileRequest struct {
	Password   string `json:"password" validate:"required"`
	Name       string `json:"name" validate:"omitempty,min=1,max=30"`
	Email      string `json:"email" validate:"omitempty,email"`
	Department string `json:"department" validate:"omitempty,min=0,max=50"`
}

type UpdateMyProfileResponse struct {
	User struct {
		ID           string               `json:"id"`
		AccountId    string               `json:"accountId"`
		Name         string               `json:"name"`
		Role         RoleResponse         `json:"role"`
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
