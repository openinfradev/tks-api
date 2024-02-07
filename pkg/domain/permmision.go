package domain

import "github.com/google/uuid"

type Permission struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`

	IsAllowed *bool       `json:"is_allowed,omitempty"`
	RoleID    *uuid.UUID  `json:"role_id,omitempty"`
	Role      *Role       `json:"role,omitempty"`
	Endpoints []*Endpoint `json:"endpoints,omitempty"`
	// omit empty

	ParentID *uuid.UUID    `json:"parent_id,omitempty"`
	Parent   *Permission   `json:"parent,omitempty"`
	Children []*Permission `json:"children,omitempty"`
}
