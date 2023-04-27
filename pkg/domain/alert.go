package domain

import (
	"time"

	"github.com/google/uuid"
)

// enum
type AlertActionStatus int32

const (
	AlertActionStatus_INPROGRESS AlertActionStatus = iota
	AlertActionStatus_CLOSED
	AlertActionStatus_ERROR
)

var alertStatus = [...]string{
	"INPROGRESS",
	"CLOSED",
	"ERROR",
}

func (m AlertActionStatus) String() string { return alertStatus[(m)] }
func (m AlertActionStatus) FromString(s string) AlertActionStatus {
	for i, v := range alertStatus {
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
	ClusterId      ClusterId
	Cluster        Cluster
	GrafanaUrl     string
	Status         string
	AlertActions   []AlertAction
	CreatorId      *uuid.UUID
	Creator        User
	UpdatorId      *uuid.UUID
	Updator        User
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type AlertAction struct {
	ID          uuid.UUID
	AlertId     uuid.UUID
	Content     string
	Status      AlertActionStatus
	TakerId     *uuid.UUID
	Taker       User
	StartedAt   time.Time
	CompletedAt time.Time
}

type AlertResponse struct {
	ID             string                `json:"id"`
	OrganizationId string                `json:"organizationId"`
	Name           string                `json:"name"`
	Description    string                `json:"description"`
	Code           string                `json:"code"`
	Grade          string                `json:"grade"`
	Cluster        SimpleClusterResponse `json:"cluster"`
	GrafanaUrl     string                `json:"grafanaUrl"`
	FiredAt        time.Time             `json:"firedAt"`
	TakedAt        time.Time             `json:"takedAt"`
	ClosedAt       time.Time             `json:"closedAt"`
	Status         string                `json:"status"`
	ProcessingSec  int                   `json:"processingSec"`
	TakedTimeSec   int                   `json:"takedSec"`
	AlertActions   []AlertActionResponse `json:"alertActions"`
	CreatedAt      time.Time             `json:"createdAt"`
	UpdatedAt      time.Time             `json:"updatedAt"`
}

type AlertActionResponse struct {
	ID          uuid.UUID          `json:"id"`
	AlertId     uuid.UUID          `json:"alertId"`
	Content     string             `json:"content"`
	Status      string             `json:"status"`
	Taker       SimpleUserResponse `json:"taker"`
	StartedAt   time.Time          `json:"startedAt"`
	CompletedAt time.Time          `json:"completedAt"`
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
