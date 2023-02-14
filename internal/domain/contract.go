package domain

import (
	"time"
)

type Contract struct {
	Id          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Creator     string    `json:"creator"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type Quota = struct {
	Cpu    int64 `json:"cpu"`
	Memory int64 `json:"memory"`
}

type Project = struct {
	Id                string    `json:"id"`
	Name              string    `json:"name"`
	Description       string    `json:"description"`
	Status            string    `json:"status"`
	StatusDescription string    `json:"statusDescription"`
	Quota             Quota     `json:"quota"`
	AvailableServices []string  `json:"availableServices"`
	CspIds            []string  `json:"cspIds"`
	ClusterCnt        int       `json:"clusterCnt"`
	GitAccount        string    `json:"gitAccount"`
	Creator           string    `json:"creator"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}
