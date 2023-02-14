package domain

import (
	"time"
)

type History = struct {
	Id          string    `json:"id"`
	AccountId   string    `json:"accountId"`
	HistoryType string    `json:"historyType"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
