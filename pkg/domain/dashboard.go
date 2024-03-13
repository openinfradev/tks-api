package domain

import (
	"time"
)

// enum
type ChartType int32

const (
	ChartType_ALL ChartType = iota
	ChartType_TRAFFIC
	ChartType_CPU
	ChartType_POD
	ChartType_MEMORY
	ChartType_POD_CALENDAR
	ChartType_ERROR
)

var chartType = [...]string{
	"ALL",
	"TRAFFIC",
	"CPU",
	"POD",
	"MEMORY",
	"POD_CALENDAR",
	"ERROR",
}

func (m ChartType) String() string { return chartType[(m)] }
func (m ChartType) FromString(s string) ChartType {
	for i, v := range chartType {
		if v == s {
			return ChartType(i)
		}
	}
	return ChartType_ERROR
}

// 내부
type DashboardChart struct {
	ChartType      ChartType
	OrganizationId string
	Name           string
	Description    string
	Duration       string // 1d, 7d, 30d ...
	Interval       string // 1h, 1d, ...
	Year           string
	Month          string
	ChartData      ChartData
	UpdatedAt      time.Time
}

type DashboardStack struct {
	ID          StackId
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

// [TODO]
func (m ChartType) All() (out []string) {
	for _, v := range chartType {
		out = append(out, v)
	}
	return
}

type Unit struct {
	Name string   `json:"name"`
	Data []string `json:"data"`
}

type Axis struct {
	Data []string `json:"data"`
}

type PodCount struct {
	Day   int `json:"day"`
	Value int `json:"value"`
}

type ChartData struct {
	XAxis     *Axis      `json:"xAxis,omitempty"`
	YAxis     *Axis      `json:"yAxis,omitempty"`
	Series    []Unit     `json:"series,omitempty"`
	PodCounts []PodCount `json:"podCounts,omitempty"`
}

type DashboardChartResponse struct {
	ChartType      string    `json:"chartType"`
	OrganizationId string    `json:"organizationId"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	Duration       string    `json:"duration"`
	Interval       string    `json:"interval"`
	Year           string    `json:"year"`
	Month          string    `json:"month"`
	ChartData      ChartData `json:"chartData"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type GetDashboardChartsResponse struct {
	Charts []DashboardChartResponse `json:"charts"`
}

type GetDashboardChartResponse struct {
	Chart DashboardChartResponse `json:"chart"`
}

type DashboardResource struct {
	Stack   string `json:"stack"`
	Cpu     string `json:"cpu"`
	Memory  string `json:"memory"`
	Storage string `json:"storage"`
}

type GetDashboardResourcesResponse struct {
	Resources DashboardResource `json:"resources"`
}

type DashboardStackResponse struct {
	ID          StackId   `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	StatusDesc  string    `json:"statusDesc"`
	Cpu         string    `json:"cpu"`
	Memory      string    `json:"memory"`
	Storage     string    `json:"storage"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type GetDashboardStacksResponse struct {
	Stacks []DashboardStackResponse `json:"stacks"`
}
