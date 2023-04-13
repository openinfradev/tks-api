package domain

import (
	"time"
)

type User = struct {
	ID           string       `json:"id"`
	AccountId    string       `json:"accountId"`
	Password     string       `json:"password"`
	Name         string       `json:"name"`
	Token        string       `json:"token"`
	Role         Role         `json:"role"`
	Organization Organization `json:"organization"`
	Creator      string       `json:"creator"`
	CreatedAt    time.Time    `json:"createdAt"`
	UpdatedAt    time.Time    `json:"updatedAt"`

	Email       string `json:"email"`
	Department  string `json:"department"`
	Description string `json:"description"`
}

type Role = struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Creator     string    `json:"creator"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type Policy = struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Create           bool      `json:"create"`
	CreatePriviledge string    `json:"createPriviledge"`
	Update           bool      `json:"update"`
	UpdatePriviledge string    `json:"updatePriviledge"`
	Read             bool      `json:"read"`
	ReadPriviledge   string    `json:"readPriviledge"`
	Delete           bool      `json:"delete"`
	DeletePriviledge string    `json:"deletePriviledge"`
	Creator          string    `json:"creator"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

type LoginRequest struct {
	AccountId      string `json:"accountId" validate:"required"`
	Password       string `json:"password" validate:"required"`
	OrganizationId string `json:"organizationId" validate:"required"`
}

type LoginResponse struct {
	User struct {
		AccountId    string       `json:"accountId"`
		Name         string       `json:"name"`
		Token        string       `json:"token"`
		Role         Role         `json:"role"`
		Organization Organization `json:"organization"`
	} `json:"user"`
}

type LogoutRequest struct {
	//TODO implement me
	AccountId string `json:"accountId"`
}

type FindIdRequest struct {
	//TODO implement me
}

type FindPasswordRequest struct {
	//TODO implement me
}

type CreateUserRequest struct {
	AccountId   string `json:"accountId" validate:"required"`
	Password    string `json:"password" validate:"required"`
	Name        string `json:"name" validate:"omitempty,min=0,max=20"`
	Email       string `json:"email" validate:"omitempty,email"`
	Department  string `json:"department" validate:"omitempty,min=0,max=20"`
	Role        string `json:"role" validate:"required,oneof=admin user"`
	Description string `json:"description" validate:"omitempty,min=0,max=100"`
}

type SimpleUserResponse struct {
	ID        string `json:"id"`
	AccountId string `json:"accountId"`
	Name      string `json:"name"`
}

type CreateUserResponse struct {
	User struct {
		ID           string       `json:"id"`
		AccountId    string       `json:"accountId"`
		Name         string       `json:"name"`
		Role         Role         `json:"role"`
		Organization Organization `json:"organization"`
		Email        string       `json:"email"`
		Department   string       `json:"department"`
		Description  string       `json:"description"`
	} `json:"user"`
}

type GetUserResponse struct {
	User struct {
		ID           string       `json:"id"`
		AccountId    string       `json:"accountId"`
		Name         string       `json:"name"`
		Role         Role         `json:"role"`
		Organization Organization `json:"organization"`
		Email        string       `json:"email"`
		Department   string       `json:"department"`
		Description  string       `json:"description"`
		Creator      string       `json:"creator"`
		CreatedAt    time.Time    `json:"createdAt"`
		UpdatedAt    time.Time    `json:"updatedAt"`
	} `json:"user"`
}

type ListUserResponse struct {
	Users []ListUserBody `json:"users"`
}
type ListUserBody struct {
	ID           string       `json:"id"`
	AccountId    string       `json:"accountId"`
	Name         string       `json:"name"`
	Role         Role         `json:"role"`
	Organization Organization `json:"organization"`
	Email        string       `json:"email"`
	Department   string       `json:"department"`
	Description  string       `json:"description"`
	Creator      string       `json:"creator"`
	CreatedAt    time.Time    `json:"createdAt"`
	UpdatedAt    time.Time    `json:"updatedAt"`
}

type UpdateUserRequest struct {
	Name        string `json:"name" validate:"omitempty,min=0,max=20"`
	Role        string `json:"role" validate:"oneof=admin user"`
	Email       string `json:"email" validate:"omitempty,email"`
	Department  string `json:"department" validate:"omitempty,min=0,max=20"`
	Description string `json:"description" validate:"omitempty,min=0,max=100"`
}

type UpdateUserResponse struct {
	User struct {
		ID           string       `json:"id"`
		AccountId    string       `json:"accountId"`
		Name         string       `json:"name"`
		Role         Role         `json:"role"`
		Organization Organization `json:"organization"`
		Email        string       `json:"email"`
		Department   string       `json:"department"`
		Description  string       `json:"description"`
	} `json:"user"`
}

type UpdatePasswordRequest struct {
	Password string `json:"password" validate:"required"`
}

type UpdatePasswordResponse struct {
}

type CheckExistedIdRequest struct {
	AccountId string `json:"accountId" validate:"required"`
}

type CheckExistedIdResponse struct {
	Existed bool `json:"existed"`
}
