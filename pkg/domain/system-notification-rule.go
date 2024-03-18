package domain

import (
	"time"
)

type SystemNotificationRuleResponse struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Creator     SimpleUserResponse `json:"creator"`
	Updator     SimpleUserResponse `json:"updator"`
	CreatedAt   time.Time          `json:"createdAt"`
	UpdatedAt   time.Time          `json:"updatedAt"`
}

type SimpleSystemNotificationRuleResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type GetSystemNotificationRulesResponse struct {
	SystemNotificationRules []SystemNotificationRuleResponse `json:"systemNotificationRules"`
	Pagination              PaginationResponse               `json:"pagination"`
}

type GetSystemNotificationRuleResponse struct {
	SystemNotificationRule SystemNotificationRuleResponse `json:"systemNotificationRule"`
}

type CreateSystemNotificationRuleRequest struct {
	Name        string `json:"name" validate:"required,name"`
	Description string `json:"description"`
}

type CreateSystemNotificationRuleResponse struct {
	ID string `json:"id"`
}

type UpdateSystemNotificationRuleRequest struct {
	Description string `json:"description"`
}

type CheckSystemNotificationRuleNameResponse struct {
	Existed bool `json:"existed"`
}
