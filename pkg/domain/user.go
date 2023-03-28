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
	AccountId      string `json:"accountId"`
	Password       string `json:"password"`
	OrganizationId string `json:"organizationId"`
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
	AccountId   string `json:"accountId"`
	Password    string `json:"password"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	Department  string `json:"department"`
	Role        string `json:"role"`
	Description string `json:"description"`
}

func (r *CreateUserRequest) ToUser() *User {
	return &User{
		ID:           "",
		AccountId:    r.AccountId,
		Password:     r.Password,
		Name:         r.Name,
		Token:        "",
		Role:         Role{Name: r.Role},
		Organization: Organization{},
		Creator:      "",
		CreatedAt:    time.Time{},
		UpdatedAt:    time.Time{},
		Email:        r.Email,
		Department:   r.Department,
		Description:  r.Description,
	}
}

type UpdateUserRequest struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	Department  string `json:"department"`
	Role        string `json:"role"`
	Description string `json:"description"`
}

func (r *UpdateUserRequest) ToUser() *User {
	return &User{
		ID:           "",
		AccountId:    "",
		Password:     "",
		Name:         r.Name,
		Token:        "",
		Role:         Role{Name: r.Role},
		Organization: Organization{},
		Creator:      "",
		CreatedAt:    time.Time{},
		UpdatedAt:    time.Time{},
		Email:        r.Email,
		Department:   r.Department,
		Description:  r.Description,
	}
}

type UpdatePasswordRequest struct {
	Password string `json:"password"`
}

func (u *UpdatePasswordRequest) ToUser() *User {
	return &User{
		Password: u.Password,
	}
}

type CheckDuplicatedIdRequest struct {
	AccountId string `json:"accountId"`
}
