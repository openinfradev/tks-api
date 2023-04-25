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
	AlertType      string
	ClusterId      ClusterId
	GrafanaUrl     string
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
	Contents    string
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
	AlertType      string                `json:"alertType"`
	ClusterId      ClusterId             `json:"clusterId"`
	GrafanaUrl     string                `json:"grafanaUrl"`
	FiredAt        time.Time             `json:"firedAt"`
	TakedAt        time.Time             `json:"takedAt"`
	ClosedAt       time.Time             `json:"closedAt"`
	ProcessingSec  int                   `json:"processingSec"`
	TakedTimeSec   int                   `json:"takedSec"`
	AlertActions   []AlertActionResponse `json:"alertActions"`
	Creator        SimpleUserResponse    `json:"creator"`
	Updator        SimpleUserResponse    `json:"updator"`
	CreatedAt      time.Time             `json:"createdAt"`
	UpdatedAt      time.Time             `json:"updatedAt"`
}

type AlertActionResponse struct {
	ID          uuid.UUID          `json:"id"`
	AlertId     uuid.UUID          `json:"alertId"`
	Contents    string             `json:"contents"`
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
	Contents string `json:"contents"`
	Status   string `json:"status" validate:"oneof=INPROGRESS CLOSED"`
}

type CreateAlertActionResponse struct {
	ID string `json:"id"`
}
