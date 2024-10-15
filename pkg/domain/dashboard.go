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
	Stack struct {
		Normal   string `json:"normal"`
		Abnormal string `json:"abnormal"`
	} `json:"stack"`
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

type WidgetResponse struct {
	Key    string `json:"widgetKey"`
	StartX int    `json:"startX"`
	StartY int    `json:"startY"`
	SizeX  int    `json:"sizeX"`
	SizeY  int    `json:"sizeY"`
}

type DashboardContents struct {
	GroupName string           `json:"groupName"`
	SizeX     int              `json:"sizeX"`
	SizeY     int              `json:"sizeY"`
	Widgets   []WidgetResponse `json:"widgets"`
}

type CreateDashboardRequest struct {
	DashboardKey string              `json:"dashboardKey"`
	Contents     []DashboardContents `json:"contents"`
}

type CreateDashboardResponse struct {
	DashboardId string `json:"dashboardId"`
}

type GetDashboardResponse struct {
	GroupName string           `json:"groupName"`
	SizeX     int              `json:"sizeX"`
	SizeY     int              `json:"sizeY"`
	Widgets   []WidgetResponse `json:"widgets"`
}

type UpdateDashboardRequest struct {
	DashboardContents
}

type CommonDashboardResponse struct {
	Result string `json:"result"`
}

type DashboardPolicyStatus struct {
	Normal  int `json:"normal"`
	Warning int `json:"warning"`
	Error   int `json:"error"`
}

type GetDashboardPolicyStatusResponse struct {
	PolicyStatus DashboardPolicyStatus `json:"statuses"`
}

type DashboardPolicyUpdate struct {
	PolicyTemplate int `json:"policyTemplate"`
	Policy         int `json:"policy"`
}

type GetDashboardPolicyUpdateResponse struct {
	PolicyUpdate DashboardPolicyUpdate `json:"updatedResources"`
}

type GetDashboardPolicyEnforcementResponse struct {
	BarChart
	ChartData BarChartData `json:"chartData"`
	UpdatedAt time.Time    `json:"updatedAt"`
}

type GetDashboardPolicyViolationResponse struct {
	BarChart
	ChartData BarChartData `json:"chartData"`
	UpdatedAt time.Time    `json:"updatedAt"`
}

type GetDashboardPolicyViolationLogResponse struct {
	// TODO implement me
}

type GetDashboardPolicyStatisticsResponse struct {
	PolicyStatisticsResponse
}

type WorkloadData struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}
type GetDashboardWorkloadResponse struct {
	Title string         `json:"title"`
	Data  []WorkloadData `json:"data"`
}

type GetDashboardPolicyViolationTop5Response struct {
	GetDashboardPolicyViolationResponse
}

type BarChart struct {
	ChartType      string `json:"chartType"`
	OrganizationId string `json:"organizationId"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	Duration       string `json:"duration"`
	Interval       string `json:"interval"`
}

type BarChartData struct {
	XAxis  *Axis        `json:"xAxis,omitempty"`
	Series []UnitNumber `json:"series,omitempty"`
}

type UnitNumber struct {
	Name string `json:"name"`
	Data []int  `json:"data"`
}
