package domain

import "time"

type EndpointResponse struct {
	Name      string    `json:"name"`
	Group     string    `json:"group"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
