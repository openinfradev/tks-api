package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/helper"
	"github.com/openinfradev/tks-api/internal/middleware/auth/request"
	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type ISystemNotificationRuleUsecase interface {
	Get(ctx context.Context, systemNotificationRuleId uuid.UUID) (model.SystemNotificationRule, error)
	Fetch(ctx context.Context, organizationId string, pg *pagination.Pagination) ([]model.SystemNotificationRule, error)
	Create(ctx context.Context, dto model.SystemNotificationRule) (systemNotificationRule uuid.UUID, err error)
	Update(ctx context.Context, dto model.SystemNotificationRule) error
	Delete(ctx context.Context, systemNotificationRuleId uuid.UUID) error
	GetByName(ctx context.Context, name string) (model.SystemNotificationRule, error)
	MakeDefaultSystemNotificationRules(ctx context.Context, organizationId string, dto *model.Organization) error
}

type SystemNotificationRuleUsecase struct {
	repo                           repository.ISystemNotificationRuleRepository
	organizationRepo               repository.IOrganizationRepository
	userRepo                       repository.IUserRepository
	systemNotificationTemplateRepo repository.ISystemNotificationTemplateRepository
}

func NewSystemNotificationRuleUsecase(r repository.Repository) ISystemNotificationRuleUsecase {
	return &SystemNotificationRuleUsecase{
		repo:                           r.SystemNotificationRule,
		organizationRepo:               r.Organization,
		userRepo:                       r.User,
		systemNotificationTemplateRepo: r.SystemNotificationTemplate,
	}
}

func (u *SystemNotificationRuleUsecase) Create(ctx context.Context, dto model.SystemNotificationRule) (systemNotificationRuleId uuid.UUID, err error) {
	user, ok := request.UserFrom(ctx)
	if !ok {
		return uuid.Nil, httpErrors.NewBadRequestError(fmt.Errorf("Invalid token"), "", "")
	}
	userId := user.GetUserId()
	dto.CreatorId = &userId
	dto.UpdatorId = &userId

	if _, err = u.GetByName(ctx, dto.Name); err == nil {
		return uuid.Nil, httpErrors.NewBadRequestError(fmt.Errorf("duplicate systemNotificationRule name"), "SNR_CREATE_ALREADY_EXISTED_NAME", "")
	}

	// Users
	dto.TargetUsers = make([]model.User, 0)
	for _, strId := range dto.TargetUserIds {
		userId, err := uuid.Parse(strId)
		if err == nil {
			user, err := u.userRepo.GetByUuid(ctx, userId)
			if err == nil {
				dto.TargetUsers = append(dto.TargetUsers, user)
			}
		}
	}

	// Make parameters
	dto.SystemNotificationCondition.Parameter = []byte(helper.ModelToJson(dto.SystemNotificationCondition.Parameters))

	systemNotificationRuleId, err = u.repo.Create(ctx, dto)
	if err != nil {
		return uuid.Nil, err
	}

	return
}

func (u *SystemNotificationRuleUsecase) Update(ctx context.Context, dto model.SystemNotificationRule) error {
	rule, err := u.repo.Get(ctx, dto.ID)
	if err != nil {
		return httpErrors.NewBadRequestError(err, "SNR_NOT_EXISTED_STACK_TEMPLATE", "")
	}

	// Users
	dto.TargetUsers = make([]model.User, 0)
	for _, strId := range dto.TargetUserIds {
		userId, err := uuid.Parse(strId)
		if err == nil {
			user, err := u.userRepo.GetByUuid(ctx, userId)
			if err == nil {
				dto.TargetUsers = append(dto.TargetUsers, user)
			}
		}
	}

	// Make parameters
	dto.SystemNotificationCondition.Parameter = []byte(helper.ModelToJson(dto.SystemNotificationCondition.Parameters))
	dto.SystemNotificationCondition.ID = rule.SystemNotificationCondition.ID

	err = u.repo.Update(ctx, dto)
	if err != nil {
		return err
	}

	// update status for appling kubernetes
	if err = u.repo.UpdateStatus(ctx, dto.ID, domain.SystemNotificationRuleStatus_PENDING); err != nil {
		return err
	}

	return nil
}

func (u *SystemNotificationRuleUsecase) Get(ctx context.Context, systemNotificationRuleId uuid.UUID) (res model.SystemNotificationRule, err error) {
	res, err = u.repo.Get(ctx, systemNotificationRuleId)
	if err != nil {
		return res, err
	}
	return
}

func (u *SystemNotificationRuleUsecase) GetByName(ctx context.Context, name string) (out model.SystemNotificationRule, err error) {
	out, err = u.repo.GetByName(ctx, name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return out, httpErrors.NewNotFoundError(err, "SNR_FAILED_FETCH_SYSTEM_NOTIFICATION_RULE", "")
		}
		return out, err
	}

	return
}

func (u *SystemNotificationRuleUsecase) Fetch(ctx context.Context, organizationId string, pg *pagination.Pagination) (res []model.SystemNotificationRule, err error) {
	res, err = u.repo.FetchWithOrganization(ctx, organizationId, pg)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (u *SystemNotificationRuleUsecase) Delete(ctx context.Context, systemNotificationRuleId uuid.UUID) (err error) {
	systemNotificationRule, err := u.repo.Get(ctx, systemNotificationRuleId)
	if err != nil {
		return err
	}

	user, ok := request.UserFrom(ctx)
	if !ok {
		return httpErrors.NewBadRequestError(fmt.Errorf("Invalid token"), "", "")
	}
	userId := user.GetUserId()
	systemNotificationRule.UpdatorId = &userId

	err = u.repo.Delete(ctx, systemNotificationRule)
	if err != nil {
		return err
	}
	return
}

func (u *SystemNotificationRuleUsecase) MakeDefaultSystemNotificationRules(ctx context.Context, organizationId string, dto *model.Organization) error {
	organization, err := u.organizationRepo.Get(ctx, organizationId)
	if err != nil {
		return err
	}

	organizationAdmin, err := u.userRepo.GetByUuid(ctx, *organization.AdminId)
	if err != nil {
		return err
	}

	pg := pagination.NewPaginationWithFilter("is_system", "", "$eq", []string{"1"})
	templates, err := u.systemNotificationTemplateRepo.Fetch(ctx, pg)
	if err != nil {
		return err
	}

	rules := make([]model.SystemNotificationRule, 0)
	for _, template := range templates {
		if template.Name == domain.SN_TYPE_NODE_CPU_HIGH_LOAD {
			ruleId := uuid.New()
			rules = append(rules, model.SystemNotificationRule{
				ID:                           ruleId,
				Name:                         domain.SN_TYPE_NODE_CPU_HIGH_LOAD + "-warning",
				Description:                  "",
				OrganizationId:               organizationId,
				NotificationType:             template.NotificationType,
				IsSystem:                     true,
				SystemNotificationTemplateId: template.ID,
				SystemNotificationCondition: model.SystemNotificationCondition{
					SystemNotificationRuleId: ruleId,
					Severity:                 "warning",
					Duration:                 "3m",
					Parameter:                []byte("[{\"order\": 0, \"value\": \"10\", \"operator\": \"<\"}]"),
					EnableEmail:              true,
					EnablePortal:             true,
				},
				TargetUsers:           []model.User{organizationAdmin},
				MessageTitle:          "CPU 사용량이 높습니다",
				MessageContent:        "스택 (<<STACK>>)의 노드(<<INSTANCE>>)의 idle process의 cpu 점유율이 3분 동안 0% 입니다. (현재 사용률 {{$value}}). 워커 노드 CPU가 과부하 상태입니다. 일시적인 서비스 Traffic 증가, Workload의 SW 오류, Server HW Fan Fail등 다양한 원인으로 인해 발생할 수 있습니다.",
				MessageActionProposal: "일시적인 Service Traffic의 증가가 관측되지 않았다면, Alert발생 노드에서 실행 되는 pod중 CPU 자원을 많이 점유하는 pod의 설정을 점검해 보시길 제안드립니다. 예를 들어 pod spec의 limit 설정으로 과도한 CPU자원 점유을 막을 수 있습니다.",
				Status:                domain.SystemNotificationRuleStatus_PENDING,
				CreatorId:             organization.AdminId,
				UpdatorId:             organization.AdminId,
			})
		} else if template.Name == domain.SN_TYPE_NODE_MEMORY_HIGH_UTILIZATION {
			ruleId := uuid.New()
			rules = append(rules, model.SystemNotificationRule{
				ID:                           ruleId,
				Name:                         domain.SN_TYPE_NODE_MEMORY_HIGH_UTILIZATION + "-warning",
				Description:                  "",
				OrganizationId:               organizationId,
				NotificationType:             template.NotificationType,
				IsSystem:                     true,
				SystemNotificationTemplateId: template.ID,
				SystemNotificationCondition: model.SystemNotificationCondition{
					SystemNotificationRuleId: ruleId,
					Severity:                 "warning",
					Duration:                 "3m",
					Parameter:                []byte("[{\"order\": 0, \"value\": \"0.2\", \"operator\": \"<\"}]"),
					EnableEmail:              true,
					EnablePortal:             true,
				},
				TargetUsers:           []model.User{organizationAdmin},
				MessageTitle:          "메모리 사용량이 높습니다",
				MessageContent:        "스택 (<<STACK>>)의 노드(<<INSTANCE>>)의 Memory 사용량이 3분동안 80% 를 넘어서고 있습니다. (현재 사용률 {{$value}}). 워커 노드의 Memory 사용량이 80%를 넘었습니다. 일시적인 서비스 증가 및 SW 오류등 다양한 원인으로 발생할 수 있습니다.",
				MessageActionProposal: "일시적인 Service Traffic의 증가가 관측되지 않았다면, Alert발생 노드에서 실행되는 pod중 Memory 사용량이 높은 pod들에 대한 점검을 제안드립니다.",
				Status:                domain.SystemNotificationRuleStatus_PENDING,
				CreatorId:             organization.AdminId,
				UpdatorId:             organization.AdminId,
			})
		} else if template.Name == domain.SN_TYPE_NODE_DISK_FULL {
			ruleId := uuid.New()
			rules = append(rules, model.SystemNotificationRule{
				ID:                           ruleId,
				Name:                         domain.SN_TYPE_NODE_DISK_FULL + "-critical",
				Description:                  "",
				OrganizationId:               organizationId,
				NotificationType:             template.NotificationType,
				IsSystem:                     true,
				SystemNotificationTemplateId: template.ID,
				SystemNotificationCondition: model.SystemNotificationCondition{
					SystemNotificationRuleId: ruleId,
					Severity:                 "critical",
					Duration:                 "30m",
					Parameter:                []byte("[{\"order\": 0, \"value\": \"0\", \"operator\": \"<\"}]"),
					EnableEmail:              true,
					EnablePortal:             true,
				},
				TargetUsers:           []model.User{organizationAdmin},
				MessageTitle:          "노드 디스크 사용량이 높습니다.",
				MessageContent:        "지난 6시간동안의 추세로 봤을 때, 스택 (<<STACK>>)의 노드(<<INSTANCE>>)의 root 볼륨은 24시간 안에 Disk full이 예상됨. 현재 Disk 사용 추세기준 24시간 내에 Disk 용량이 꽉 찰 것으로 예상됩니다.",
				MessageActionProposal: "Disk 용량 최적화(삭제 및 Backup)을 수행하시길 권고합니다. 삭제할 내역이 없으면 증설 계획을 수립해 주십시요.",
				Status:                domain.SystemNotificationRuleStatus_PENDING,
				CreatorId:             organization.AdminId,
				UpdatorId:             organization.AdminId,
			})
		} else if template.Name == domain.SN_TYPE_PVC_FULL {
			ruleId := uuid.New()
			rules = append(rules, model.SystemNotificationRule{
				ID:                           ruleId,
				Name:                         domain.SN_TYPE_PVC_FULL + "-critical",
				Description:                  "",
				OrganizationId:               organizationId,
				NotificationType:             template.NotificationType,
				IsSystem:                     true,
				SystemNotificationTemplateId: template.ID,
				SystemNotificationCondition: model.SystemNotificationCondition{
					SystemNotificationRuleId: ruleId,
					Severity:                 "critical",
					Duration:                 "30m",
					Parameter:                []byte("[{\"order\": 0, \"value\": \"0\", \"operator\": \"<\"}]"),
					EnableEmail:              true,
					EnablePortal:             true,
				},
				TargetUsers:           []model.User{organizationAdmin},
				MessageTitle:          "PVC 사용량이 높습니다.",
				MessageContent:        "지난 6시간동안의 추세로 봤을 때, 스택 (<<INSTANCE>>)의 파드(<<PVC>>)가 24시간 안에 Disk full이 예상됨. 현재 Disk 사용 추세기준 24시간 내에 Disk 용량이 꽉 찰것으로 예상됩니다. (<<STACK>> 스택, <<PVC>> PVC)",
				MessageActionProposal: "Disk 용량 최적화(삭제 및 Backup)을 수행하시길 권고합니다. 삭제할 내역이 없으면 증설 계획을 수립해 주십시요.",
				Status:                domain.SystemNotificationRuleStatus_PENDING,
				CreatorId:             organization.AdminId,
				UpdatorId:             organization.AdminId,
			})
		} else if template.Name == domain.SN_TYPE_POD_RESTART_FREQUENTLY {
			ruleId := uuid.New()
			rules = append(rules, model.SystemNotificationRule{
				ID:                           ruleId,
				Name:                         domain.SN_TYPE_POD_RESTART_FREQUENTLY + "-critical",
				Description:                  "",
				OrganizationId:               organizationId,
				NotificationType:             template.NotificationType,
				IsSystem:                     true,
				SystemNotificationTemplateId: template.ID,
				SystemNotificationCondition: model.SystemNotificationCondition{
					SystemNotificationRuleId: ruleId,
					Severity:                 "critical",
					Duration:                 "30m",
					Parameter:                []byte("[{\"order\": 0, \"value\": \"2\", \"operator\": \">\"}]"),
					EnableEmail:              true,
					EnablePortal:             true,
				},
				TargetUsers:           []model.User{organizationAdmin},
				MessageTitle:          "스택의 Pod가 재기동되고 있습니다.",
				MessageContent:        "스택 (<<STACK>>)의 파드(<<POD>>)가 30분 동안 5회 이상 재기동 ({{$value}} 회). 특정 Pod가 빈번하게 재기동 되고 있습니다. 점검이 필요합니다. (<<STACK>> 스택, <<POD>> 파드)",
				MessageActionProposal: "pod spec. 에 대한 점검이 필요합니다. pod의 log 및 status를 확인해 주세요.",
				Status:                domain.SystemNotificationRuleStatus_PENDING,
				CreatorId:             organization.AdminId,
				UpdatorId:             organization.AdminId,
			})
		} else if template.Name == domain.SN_TYPE_POLICY_AUDITED {
			ruleId := uuid.New()
			rules = append(rules, model.SystemNotificationRule{
				ID:                           ruleId,
				Name:                         domain.SN_TYPE_POLICY_AUDITED + "-critical",
				Description:                  "",
				OrganizationId:               organizationId,
				NotificationType:             template.NotificationType,
				IsSystem:                     true,
				SystemNotificationTemplateId: template.ID,
				SystemNotificationCondition: model.SystemNotificationCondition{
					SystemNotificationRuleId: ruleId,
					Severity:                 "critical",
					Duration:                 "1m",
					Parameter:                []byte("[{\"order\": 0, \"value\": \"1\", \"operator\": \"==\"}]"),
					EnableEmail:              true,
					EnablePortal:             true,
				},
				TargetUsers:           []model.User{organizationAdmin},
				MessageTitle:          "정책 위반(<<KIND>> / <<NAME>>)",
				MessageContent:        "스택 (<<STACK>>)의 자원(<<VIOLATING_KIND>> - <<VIOLATING_NAMESPACE>> / <<VIOLATING_NAME>>)에서 정책(<<KIND>> / <<NAME>>)위반이 발생했습니다. 메시지 - <<VIOLATION_MSG>>",
				MessageActionProposal: "정책위반이 발생하였습니다.(<<KIND>> / <<NAME>>)",
				Status:                domain.SystemNotificationRuleStatus_PENDING,
				CreatorId:             organization.AdminId,
				UpdatorId:             organization.AdminId,
			})
		} else if template.Name == domain.SN_TYPE_POLICY_BLOCKED {
			ruleId := uuid.New()
			rules = append(rules, model.SystemNotificationRule{
				ID:                           ruleId,
				Name:                         domain.SN_TYPE_POLICY_BLOCKED + "-critical",
				Description:                  "",
				OrganizationId:               organizationId,
				NotificationType:             template.NotificationType,
				IsSystem:                     true,
				SystemNotificationTemplateId: template.ID,
				SystemNotificationCondition: model.SystemNotificationCondition{
					SystemNotificationRuleId: ruleId,
					Severity:                 "critical",
					Duration:                 "1m",
					Parameter:                []byte("[{\"order\": 0, \"value\": \"1\", \"operator\": \"==\"}]"),
					EnableEmail:              true,
					EnablePortal:             true,
				},
				TargetUsers:           []model.User{organizationAdmin},
				MessageTitle:          "정책 위반(<<KIND>> / <<NAME>>) 시도",
				MessageContent:        "스택 (<<STACK>>)의 자원(<<VIOLATING_KIND>> - <<VIOLATING_NAMESPACE>> / <<VIOLATING_NAME>>)에서 정책(<<KIND>> / <<NAME>>)위반 시도가 발생했습니다. 메시지 - <<VIOLATION_MSG>>",
				MessageActionProposal: "정책위반이 시도가 발생하였습니다.(<<KIND>> / <<NAME>>)",
				Status:                domain.SystemNotificationRuleStatus_PENDING,
				CreatorId:             organization.AdminId,
				UpdatorId:             organization.AdminId,
			})
		}
	}

	/*
			          - alert: policy-blocked
		            annotations:
		              Checkpoint: "정책위반이 시도가 발생하였습니다.({{  $labels.kind }} / {{ $labels.name }})"
		              description: "클러스터 ( {{ $labels.taco_cluster }})의 자원({{ $labels.violating_kind }} - {{ $labels.violating_namespace }} / {{  $labels.violating_nam }})에서 정책({{  $labels.kind }} / {{ $labels.name }})위반 시도가 발생했습니다. 메시지 - {{ $labels.violation_msg }}"
		              discriminative: $labels.kind,$labels.name,$labels.taco_cluster,$labels.violating_kind,$labels.violating_name,$labels.violating_namespace,$labels.violation_msg
		              message: 정책 위반({{ $labels.kind }} / {{ $labels.name }}) 시도
		            expr: opa_scorecard_constraint_violations{namespace!='kube-system|taco-system|gatekeeper-system',violation_enforcement=''} == 1
		            for: 1m
		            labels:
		              severity: critical

	*/

	err = u.repo.Creates(ctx, rules)
	if err != nil {
		return err
	}

	return nil
}
