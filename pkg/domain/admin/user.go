package admin

import (
	"github.com/openinfradev/tks-api/pkg/domain"
	"time"
)

type CreateUserRequest struct {
	AccountId     string `json:"accountId" validate:"required"`
	Name          string `json:"name" validate:"name"`
	Email         string `json:"email" validate:"required,email"`
	Role          string `json:"role" validate:"required"`
	Department    string `json:"department" validate:"min=0,max=50"`
	Description   string `json:"description" validate:"min=0,max=100"`
	AdminPassword string `json:"adminPassword"`
}

type CreateUserResponse struct {
	ID string `json:"id"`
}

type ListUserResponse struct {
	Users      []domain.ListUserBody     `json:"users"`
	Pagination domain.PaginationResponse `json:"pagination"`
}

type GetUserResponse struct {
	User struct {
		ID           string              `json:"id"`
		AccountId    string              `json:"accountId"`
		Name         string              `json:"name"`
		Role         domain.Role         `json:"role"`
		Organization domain.Organization `json:"organization"`
		Email        string              `json:"email"`
		Department   string              `json:"department"`
		Description  string              `json:"description"`
		Creator      string              `json:"creator"`
		CreatedAt    time.Time           `json:"createdAt"`
		UpdatedAt    time.Time           `json:"updatedAt"`
	} `json:"user"`
}

type UpdateUserRequest struct {
	Name          string `json:"name" validate:"name"`
	Email         string `json:"email" validate:"required,email"`
	Department    string `json:"department" validate:"min=0,max=50"`
	Role          string `json:"role" validate:"required"`
	Description   string `json:"description" validate:"min=0,max=100"`
	AdminPassword string `json:"adminPassword"`
}

type UpdateUserResponse struct {
	User struct {
		ID           string              `json:"id"`
		AccountId    string              `json:"accountId"`
		Name         string              `json:"name"`
		Role         domain.Role         `json:"role"`
		Organization domain.Organization `json:"organization"`
		Email        string              `json:"email"`
		Department   string              `json:"department"`
		Description  string              `json:"description"`
		CreatedAt    time.Time           `json:"createdAt"`
		UpdatedAt    time.Time           `json:"updatedAt"`
	} `json:"user"`
}

type DeleteUserRequest struct {
	AdminPassword string `json:"adminPassword"`
}
