package domain

import (
	"time"

	"github.com/google/uuid"
)

// enum
type SystemNotificationActionStatus int32

const (
	SystemNotificationActionStatus_CREATED SystemNotificationActionStatus = iota
	SystemNotificationActionStatus_INPROGRESS
	SystemNotificationActionStatus_CLOSED
	SystemNotificationActionStatus_ERROR
)

var systemNotificationActionStatus = [...]string{
	"CREATED",
	"INPROGRESS",
	"CLOSED",
	"ERROR",
}

func (m SystemNotificationActionStatus) String() string { return systemNotificationActionStatus[(m)] }
func (m SystemNotificationActionStatus) FromString(s string) SystemNotificationActionStatus {
	for i, v := range systemNotificationActionStatus {
		if v == s {
			return SystemNotificationActionStatus(i)
		}
	}
	return SystemNotificationActionStatus_ERROR
}

type SystemNotificationRequest struct {
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
		Message                  string `json:"message"`
		Summary                  string `json:"summary"`
		Description              string `json:"description"`
		Checkpoint               string `json:"Checkpoint"`
		Discriminative           string `json:"discriminative"`
		AlertType                string `json:"alertType"`
		SystemNotificationRuleId string `json:"systemNotificationRuleId"`
		PolicyName               string `json:"policyName"`
		PolicyTemplateName       string `json:"policyTemplateName"`
	} `json:"annotations"`
}

type CreateSystemNotificationRequest struct {
	Receiver            string                      `json:"receiver"`
	Status              string                      `json:"status"`
	ExternalURL         string                      `json:"externalURL"`
	Version             string                      `json:"version"`
	GroupKey            string                      `json:"groupKey"`
	TruncatedAlerts     int                         `json:"truncatedAlerts"`
	SystemNotifications []SystemNotificationRequest `json:"alerts"`
	GroupLabels         struct {
		SystemNotificationName string `json:"alertname"`
	} `json:"groupLabels"`
	//CommonLabels      string `json:"commonLabels"`
	//CommonAnnotations string `json:"commonAnnotations"`
}

type SystemNotificationResponse struct {
	ID                        string                             `json:"id"`
	Name                      string                             `json:"name"`
	OrganizationId            string                             `json:"organizationId"`
	Severity                  string                             `json:"severity"`
	MessageTitle              string                             `json:"messageTitle"`
	MessageContent            string                             `json:"messageContent"`
	MessageActionProposal     string                             `json:"messageActionProposal"`
	Cluster                   SimpleClusterResponse              `json:"cluster"`
	GrafanaUrl                string                             `json:"grafanaUrl"`
	Node                      string                             `json:"node"`
	FiredAt                   *time.Time                         `json:"firedAt"`
	TakedAt                   *time.Time                         `json:"takedAt"`
	ClosedAt                  *time.Time                         `json:"closedAt"`
	Status                    string                             `json:"status"`
	ProcessingSec             int                                `json:"processingSec"`
	TakedSec                  int                                `json:"takedSec"`
	SystemNotificationActions []SystemNotificationActionResponse `json:"systemNotificationActions"`
	LastTaker                 SimpleUserResponse                 `json:"lastTaker"`
	RawData                   string                             `json:"rawData"`
	NotificationType          string                             `json:"notificationType"`
	Read                      bool                               `json:"read"`
	CreatedAt                 time.Time                          `json:"createdAt"`
	UpdatedAt                 time.Time                          `json:"updatedAt"`
	PolicyName                string                             `json:"policyName"`
}

type SystemNotificationActionResponse struct {
	ID                   uuid.UUID          `json:"id"`
	SystemNotificationId uuid.UUID          `json:"systemNotificationId"`
	Content              string             `json:"content"`
	Status               string             `json:"status"`
	Taker                SimpleUserResponse `json:"taker"`
	CreatedAt            time.Time          `json:"createdAt"`
	UpdatedAt            time.Time          `json:"updatedAt"`
}

type GetSystemNotificationsResponse struct {
	SystemNotifications []SystemNotificationResponse `json:"systemNotifications"`
	Pagination          PaginationResponse           `json:"pagination"`
}

type GetSystemNotificationResponse struct {
	SystemNotification SystemNotificationResponse `json:"systemNotification"`
}

type CreateSystemNotificationResponse struct {
	ID string `json:"id"`
}

type UpdateSystemNotificationRequest struct {
	Description string `json:"description"`
}

type CreateSystemNotificationActionRequest struct {
	Content string `json:"content"`
	Status  string `json:"status" validate:"oneof=INPROGRESS CLOSED"`
}

type CreateSystemNotificationActionResponse struct {
	ID string `json:"id"`
}
