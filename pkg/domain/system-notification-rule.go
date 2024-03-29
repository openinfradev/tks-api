package domain

import (
	"time"
)

type SystemNotificationRuleResponse struct {
	ID                           string                                   `json:"id"`
	Name                         string                                   `json:"name"`
	Description                  string                                   `json:"description"`
	MessageTitle                 string                                   `json:"messageTitle"`
	MessageContent               string                                   `json:"messageContent"`
	MessageCondition             string                                   `json:"messageCondition"`
	MessageActionProposal        string                                   `json:"messageActionProposal"`
	TargetUsers                  []SimpleUserResponse                     `json:"targetUsers"`
	SystemNotificationTemplate   SimpleSystemNotificationTemplateResponse `json:"systemNotificationTemplate"`
	SystemNotificationConditions []SystemNotificationConditionResponse    `json:"systemNotificationConditions"`
	Creator                      SimpleUserResponse                       `json:"creator"`
	Updator                      SimpleUserResponse                       `json:"updator"`
	CreatedAt                    time.Time                                `json:"createdAt"`
	UpdatedAt                    time.Time                                `json:"updatedAt"`
}

type SystemNotificationParameter struct {
	Order    int    `json:"order"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

type SystemNotificationConditionResponse struct {
	Order        int                           `json:"order"`
	Severity     string                        `json:"severity"`
	Duration     int                           `json:"duration"`
	Parameters   []SystemNotificationParameter `json:"parameters"`
	EnableEmail  bool                          `json:"enableEmail"`
	EnablePortal bool                          `json:"enablePortal"`
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
	MessageCondition             string   `json:"messageCondition" validate:"required"`
	MessageActionProposal        string   `json:"messageActionProposal"`
	TargetUserIds                []string `json:"targetUserIds"`
	SystemNotificationTemplateId string   `json:"systemNotificationTemplateId" validate:"required"`
	SystemNotificationConditions []struct {
		Order        int                           `json:"order"`
		Severity     string                        `json:"severity"`
		Duration     int                           `json:"duration"`
		Parameters   []SystemNotificationParameter `json:"parameters"`
		EnableEmail  bool                          `json:"enableEmail"`
		EnablePortal bool                          `json:"enablePortal"`
	} `json:"systemNotificationConditions"`
}

type CreateSystemNotificationRuleResponse struct {
	ID string `json:"id"`
}

type UpdateSystemNotificationRuleRequest struct {
	Name                         string   `json:"name" validate:"required,name"`
	Description                  string   `json:"description"`
	MessageTitle                 string   `json:"messageTitle" validate:"required"`
	MessageContent               string   `json:"messageContent" validate:"required"`
	MessageCondition             string   `json:"messageCondition" validate:"required"`
	MessageActionProposal        string   `json:"messageActionProposal"`
	TargetUserIds                []string `json:"targetUserIds"`
	SystemNotificationTemplateId string   `json:"systemNotificationTemplateId" validate:"required"`
	SystemNotificationConditions []struct {
		Order        int                           `json:"order"`
		Severity     string                        `json:"severity"`
		Duration     int                           `json:"duration"`
		Parameters   []SystemNotificationParameter `json:"parameters"`
		EnableEmail  bool                          `json:"enableEmail"`
		EnablePortal bool                          `json:"enablePortal"`
	} `json:"systemNotificationConditions"`
}

type CheckSystemNotificationRuleNameResponse struct {
	Existed bool `json:"existed"`
}
