package domain

import (
	"time"
)

type Organization struct {
	Name        string    `json:"name,omitempty"`
	Id          string    `json:"id,omitempty"`
	Description string    `json:"description,omitempty"`
	Creator     string    `json:"creator,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
