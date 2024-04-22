package domain

import (
	"time"
)

type PolicyNotificationResponse struct {
	ID                    string                `json:"id"`
	OrganizationId        string                `json:"organizationId"`
	Severity              string                `json:"severity"`
	MessageTitle          string                `json:"messageTitle"`
	MessageContent        string                `json:"messageContent"`
	MessageActionProposal string                `json:"messageActionProposal"`
	Cluster               SimpleClusterResponse `json:"cluster"`
	GrafanaUrl            string                `json:"grafanaUrl"`
	Status                string                `json:"status"`
	RawData               string                `json:"rawData"`
	NotificationType      string                `json:"notificationType"`
	Read                  bool                  `json:"read"`
	CreatedAt             time.Time             `json:"createdAt"`
	UpdatedAt             time.Time             `json:"updatedAt"`
}

type GetPolicyNotificationsResponse struct {
	PolicyNotifications []PolicyNotificationResponse `json:"policyNotifications"`
	Pagination          PaginationResponse           `json:"pagination"`
}

type GetPolicyNotificationResponse struct {
	PolicyNotification PolicyNotificationResponse `json:"policyNotification"`
}
