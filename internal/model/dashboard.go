package model

import (
	"time"

	"github.com/openinfradev/tks-api/pkg/domain"
)

// 내부
type DashboardChart struct {
	ChartType      domain.ChartType
	OrganizationId string
	Name           string
	Description    string
	Duration       string // 1d, 7d, 30d ...
	Interval       string // 1h, 1d, ...
	Year           string
	Month          string
	ChartData      domain.ChartData
	UpdatedAt      time.Time
}

type DashboardStack struct {
	ID          domain.StackId
	Name        string
	Description string
	Status      string
	StatusDesc  string
	Cpu         string
	Memory      string
	Storage     string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
