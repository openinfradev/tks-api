package domain

import (
	"time"

	"github.com/google/uuid"
)

// enum
type AlertActionStatus int32

const (
	AlertActionStatus_CREATED AlertActionStatus = iota
	AlertActionStatus_INPROGRESS
	AlertActionStatus_CLOSED
	AlertActionStatus_ERROR
)

var alertActionStatus = [...]string{
	"CREATED",
	"INPROGRESS",
	"CLOSED",
	"ERROR",
}

func (m AlertActionStatus) String() string { return alertActionStatus[(m)] }
func (m AlertActionStatus) FromString(s string) AlertActionStatus {
	for i, v := range alertActionStatus {
		if v == s {
			return AlertActionStatus(i)
		}
	}
	return AlertActionStatus_ERROR
}

type CreateAlertRequestAlert struct {
	Status       string    `json:"status"`
	GeneratorURL string    `json:"generatorURL"`
	FingerPrint  string    `json:"fingerprint"`
	StartsAt     time.Time `json:"startsAt"`
	EndsAt       time.Time `json:"endsAt"`
	Labels       struct {
		AlertName   string `json:"alertname"`
		Container   string `json:"container"`
		Endpoint    string `json:"endpoint"`
		Job         string `json:"job"`
		Namespace   string `json:"namespace"`
		Pod         string `json:"pod"`
		Prometheus  string `json:"prometheus"`
		Service     string `json:"service"`
		Severity    string `json:"severity"`
		Instance    string `json:"instance"`
		TacoCluster string `json:"taco_cluster"`
	} `json:"labels"`
	Annotations struct {
		Message        string `json:"message"`
		Summary        string `json:"summary"`
		Description    string `json:"description"`
		Checkpoint     string `json:"Checkpoint"`
		Discriminative string `json:"discriminative"`
	} `json:"annotations"`
}

type CreateAlertRequest struct {
	Receiver       string                    `json:"receiver"`
	Status         string                    `json:"status"`
	ExternalURL    string                    `json:"externalURL"`
	Version        string                    `json:"version"`
	GroupKey       string                    `json:"groupKey"`
	TruncateAlerts int                       `json:"truncateAlerts"`
	Alerts         []CreateAlertRequestAlert `json:"alerts"`
	GroupLabels    struct {
		Alertname string `json:"alertname"`
	} `json:"groupLabels"`
	//CommonLabels      string `json:"commonLabels"`
	//CommonAnnotations string `json:"commonAnnotations"`
}

type AlertResponse struct {
	ID             string                `json:"id"`
	OrganizationId string                `json:"organizationId"`
	Name           string                `json:"name"`
	Description    string                `json:"description"`
	Message        string                `json:"message"`
	Code           string                `json:"code"`
	Grade          string                `json:"grade"`
	Cluster        SimpleClusterResponse `json:"cluster"`
	GrafanaUrl     string                `json:"grafanaUrl"`
	Node           string                `json:"node"`
	FiredAt        *time.Time            `json:"firedAt"`
	TakedAt        *time.Time            `json:"takedAt"`
	ClosedAt       *time.Time            `json:"closedAt"`
	Status         string                `json:"status"`
	ProcessingSec  int                   `json:"processingSec"`
	TakedSec       int                   `json:"takedSec"`
	AlertActions   []AlertActionResponse `json:"alertActions"`
	LastTaker      SimpleUserResponse    `json:"lastTaker"`
	RawData        string                `json:"rawData"`
	CreatedAt      time.Time             `json:"createdAt"`
	UpdatedAt      time.Time             `json:"updatedAt"`
}

type AlertActionResponse struct {
	ID        uuid.UUID          `json:"id"`
	AlertId   uuid.UUID          `json:"alertId"`
	Content   string             `json:"content"`
	Status    string             `json:"status"`
	Taker     SimpleUserResponse `json:"taker"`
	CreatedAt time.Time          `json:"createdAt"`
	UpdatedAt time.Time          `json:"updatedAt"`
}

type GetAlertsResponse struct {
	Alerts     []AlertResponse    `json:"alerts"`
	Pagination PaginationResponse `json:"pagination"`
}

type GetAlertResponse struct {
	Alert AlertResponse `json:"alert"`
}

type CreateAlertResponse struct {
	ID string `json:"id"`
}

type UpdateAlertRequest struct {
	Description string `json:"description"`
}

type CreateAlertActionRequest struct {
	Content string `json:"content"`
	Status  string `json:"status" validate:"oneof=INPROGRESS CLOSED"`
}

type CreateAlertActionResponse struct {
	ID string `json:"id"`
}
