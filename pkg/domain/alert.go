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

// 내부
type Alert struct {
	ID             uuid.UUID
	OrganizationId string
	Organization   Organization
	Name           string
	Description    string
	Code           string
	Grade          string
	Message        string
	ClusterId      ClusterId
	Cluster        Cluster
	Instance       string
	CheckPoint     string
	Summary        string
	GrafanaUrl     string
	FiredAt        *time.Time
	TakedAt        *time.Time
	ClosedAt       *time.Time
	TakedSec       int
	ProcessingSec  int
	Status         AlertActionStatus
	AlertActions   []AlertAction
	LastTaker      User
	RawData        []byte
	CreatorId      *uuid.UUID
	Creator        User
	UpdatorId      *uuid.UUID
	Updator        User
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type AlertAction struct {
	ID        uuid.UUID
	AlertId   uuid.UUID
	Content   string
	Status    AlertActionStatus
	TakerId   *uuid.UUID
	Taker     User
	CreatedAt time.Time
	UpdatedAt time.Time
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
	Source         string                `json:"source"`
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
	Alerts []AlertResponse `json:"alerts"`
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
