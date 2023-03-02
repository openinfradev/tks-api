package domain

import (
	"time"
)

type User = struct {
	ID            string         `json:"id"`
	AccountId     string         `json:"accountId"`
	Password      string         `json:"password"`
	Name          string         `json:"name"`
	Token         string         `json:"token"`
	Authorized    bool           `json:"authorized"`
	Roles         []Role         `json:"roles"`
	Organizations []Organization `json:"organizations"`
	Creator       string         `json:"creator"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
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
