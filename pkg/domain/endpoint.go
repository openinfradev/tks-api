package domain

import "github.com/google/uuid"

type Endpoint struct {
	ID           uuid.UUID  `json:"id"`
	Name         string     `json:"name"`
	Group        string     `json:"group"`
	PermissionID uuid.UUID  `json:"permissionId"`
	Permission   Permission `json:"permission"`
}
