package domain

import (
	"time"
)

type Organization = struct {
	Id                string    `json:"id"`
	Name              string    `json:"name"`
	Description       string    `json:"description"`
	Status            string    `json:"status"`
	StatusDescription string    `json:"statusDescription"`
	Creator           string    `json:"creator"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}
