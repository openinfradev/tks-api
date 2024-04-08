package domain

import (
	"time"
)

// enum
type SystemNotificationRuleStatus int32

const (
	SystemNotificationRuleStatus_PENDING SystemNotificationRuleStatus = iota
	SystemNotificationRuleStatus_APPLIED
	SystemNotificationRuleStatus_ERROR
)

var systemNotificationRuleStatus = [...]string{
	"PENDING",
	"APPLIED",
	"ERROR",
}

func (m SystemNotificationRuleStatus) String() string { return systemNotificationRuleStatus[(m)] }
func (m SystemNotificationRuleStatus) FromString(s string) SystemNotificationRuleStatus {
	for i, v := range systemNotificationRuleStatus {
		if v == s {
			return SystemNotificationRuleStatus(i)
		}
	}
	return SystemNotificationRuleStatus_PENDING
}

type SystemNotificationRuleResponse struct {
	ID                          string                                   `json:"id"`
	Name                        string                                   `json:"name"`
	Description                 string                                   `json:"description"`
	MessageTitle                string                                   `json:"messageTitle"`
	MessageContent              string                                   `json:"messageContent"`
	MessageActionProposal       string                                   `json:"messageActionProposal"`
	TargetUsers                 []SimpleUserResponse                     `json:"targetUsers"`
	SystemNotificationTemplate  SimpleSystemNotificationTemplateResponse `json:"systemNotificationTemplate"`
	SystemNotificationCondition SystemNotificationConditionResponse      `json:"systemNotificationCondition"`
	IsSystem                    bool                                     `json:"isSystem"`
	Creator                     SimpleUserResponse                       `json:"creator"`
	Updator                     SimpleUserResponse                       `json:"updator"`
	CreatedAt                   time.Time                                `json:"createdAt"`
	UpdatedAt                   time.Time                                `json:"updatedAt"`
}

type SystemNotificationParameter struct {
	Order    int    `json:"order"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

type SystemNotificationConditionResponse struct {
	SystemNotificationRuleId string                        `json:"systemNotificationRuleId"`
	Severity                 string                        `json:"severity"`
	Duration                 string                        `json:"duration"`
	Parameters               []SystemNotificationParameter `json:"parameters"`
	EnableEmail              bool                          `json:"enableEmail"`
	EnablePortal             bool                          `json:"enablePortal"`
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
	Name                         string   `json:"name" validate:"required,name"`
	Description                  string   `json:"description"`
	MessageTitle                 string   `json:"messageTitle" validate:"required"`
	MessageContent               string   `json:"messageContent" validate:"required"`
	MessageActionProposal        string   `json:"messageActionProposal"`
	TargetUserIds                []string `json:"targetUserIds"`
	SystemNotificationTemplateId string   `json:"systemNotificationTemplateId" validate:"required"`
	SystemNotificationCondition  struct {
		Severity     string                        `json:"severity"`
		Duration     string                        `json:"duration"`
		Parameters   []SystemNotificationParameter `json:"parameters"`
		EnableEmail  bool                          `json:"enableEmail"`
		EnablePortal bool                          `json:"enablePortal"`
	} `json:"systemNotificationCondition"`
}

type CreateSystemNotificationRuleResponse struct {
	ID string `json:"id"`
}

type UpdateSystemNotificationRuleRequest struct {
	Name                         string   `json:"name" validate:"required,name"`
	Description                  string   `json:"description"`
	MessageTitle                 string   `json:"messageTitle" validate:"required"`
	MessageContent               string   `json:"messageContent" validate:"required"`
	MessageActionProposal        string   `json:"messageActionProposal"`
	TargetUserIds                []string `json:"targetUserIds"`
	SystemNotificationTemplateId string   `json:"systemNotificationTemplateId" validate:"required"`
	SystemNotificationCondition  struct {
		SystemNotificationRuleId string                        `json:"systemNotificationRuleId"`
		Severity                 string                        `json:"severity"`
		Duration                 string                        `json:"duration"`
		Parameters               []SystemNotificationParameter `json:"parameters"`
		EnableEmail              bool                          `json:"enableEmail"`
		EnablePortal             bool                          `json:"enablePortal"`
	} `json:"systemNotificationCondition"`
}

type CheckSystemNotificationRuleNameResponse struct {
	Existed bool `json:"existed"`
}
