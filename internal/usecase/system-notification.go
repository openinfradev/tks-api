package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/mail"
	"github.com/openinfradev/tks-api/internal/middleware/auth/request"
	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type ISystemNotificationUsecase interface {
	Get(ctx context.Context, systemNotificationId uuid.UUID) (model.SystemNotification, error)
	GetByName(ctx context.Context, organizationId string, name string) (model.SystemNotification, error)
	FetchSystemNotifications(ctx context.Context, organizationId string, pg *pagination.Pagination) ([]model.SystemNotification, error)
	FetchPolicyNotifications(ctx context.Context, organizationId string, pg *pagination.Pagination) ([]model.SystemNotification, error)
	Create(ctx context.Context, dto domain.CreateSystemNotificationRequest) (err error)
	Update(ctx context.Context, dto model.SystemNotification) error
	Delete(ctx context.Context, dto model.SystemNotification) error

	CreateSystemNotificationAction(ctx context.Context, dto model.SystemNotificationAction) (systemNotificationActionId uuid.UUID, err error)
}

type SystemNotificationUsecase struct {
	repo                       repository.ISystemNotificationRepository
	clusterRepo                repository.IClusterRepository
	organizationRepo           repository.IOrganizationRepository
	appGroupRepo               repository.IAppGroupRepository
	systemNotificationRuleRepo repository.ISystemNotificationRuleRepository
	userRepo                   repository.IUserRepository
}

func NewSystemNotificationUsecase(r repository.Repository) ISystemNotificationUsecase {
	return &SystemNotificationUsecase{
		repo:                       r.SystemNotification,
		clusterRepo:                r.Cluster,
		appGroupRepo:               r.AppGroup,
		organizationRepo:           r.Organization,
		systemNotificationRuleRepo: r.SystemNotificationRule,
		userRepo:                   r.User,
	}
}

func (u *SystemNotificationUsecase) Create(ctx context.Context, input domain.CreateSystemNotificationRequest) (err error) {
	if input.SystemNotifications == nil || len(input.SystemNotifications) == 0 {
		return fmt.Errorf("No data found")
	}

	allClusters, err := u.clusterRepo.Fetch(ctx, nil)
	if err != nil {
		return fmt.Errorf("No clusters")
	}

	for _, systemNotification := range input.SystemNotifications {
		clusterId := systemNotification.Labels.TacoCluster
		organizationId, err := u.getOrganizationFromCluster(&allClusters, clusterId)
		if err != nil {
			log.Error(ctx, err)
			continue
		}

		organization, err := u.organizationRepo.Get(ctx, organizationId)
		if err != nil {
			log.Error(ctx, err)
			continue
		}
		primaryCluster, err := u.clusterRepo.Get(ctx, domain.ClusterId(organization.PrimaryClusterId))
		if err != nil {
			log.Error(ctx, err)
			continue
		}

		rawData, err := json.Marshal(systemNotification)
		if err != nil {
			rawData = []byte{}
		}

		/*
			target := ""
			// discriminative 에 target 에 대한 정보가 있다.
			// discriminative: $labels.taco_cluster, $labels.instance
			discriminative := systemNotification.Annotations.Discriminative
			if discriminative != "" {
				trimed := strings.TrimLeft(discriminative, " ")
				trimed = strings.TrimLeft(trimed, "$")
				arr := strings.Split(trimed, ",")

				for _, refer := range arr {

				}
			}
		*/

		node := ""
		if strings.Contains(systemNotification.Labels.AlertName, "node") {
			node = systemNotification.Labels.Instance
		}

		var systemNotificationRuleId *uuid.UUID
		if systemNotification.Annotations.SystemNotificationRuleId != "" {
			id, err := uuid.Parse(systemNotification.Annotations.SystemNotificationRuleId)
			if err == nil {
				systemNotificationRuleId = &id
			}
		}

		dto := model.SystemNotification{
			OrganizationId:           organizationId,
			Name:                     systemNotification.Labels.AlertName,
			Severity:                 systemNotification.Labels.Severity,
			Node:                     node,
			MessageTitle:             systemNotification.Annotations.Message,
			MessageContent:           systemNotification.Annotations.Description,
			MessageActionProposal:    systemNotification.Annotations.Checkpoint,
			Summary:                  systemNotification.Annotations.Summary,
			ClusterId:                domain.ClusterId(clusterId),
			GrafanaUrl:               u.makeGrafanaUrl(ctx, primaryCluster, systemNotification, domain.ClusterId(clusterId)),
			RawData:                  rawData,
			SystemNotificationRuleId: systemNotificationRuleId,
		}

		_, err = u.repo.Create(ctx, dto)
		if err != nil {
			log.Error(ctx, "Failed to create systemNotification ", err)
			continue
		}

		// 사용자가 생성한 알림
		if systemNotificationRuleId != nil {
			rule, err := u.systemNotificationRuleRepo.Get(ctx, *systemNotificationRuleId)
			if err != nil {
				log.Error(ctx, "Failed to get systemNotificationRule ", err)
				continue
			}

			if rule.SystemNotificationCondition.EnableEmail {
				to := []string{}
				for _, user := range rule.TargetUsers {
					to = append(to, user.Email)
				}
				message, err := mail.MakeSystemNotificationMessage(ctx, organizationId, systemNotification.Annotations.Message, to)
				if err != nil {
					log.Error(ctx, fmt.Sprintf("Failed to make email content. err : %s", err.Error()))
					continue
				}
				mailer := mail.New(message)
				if err := mailer.SendMail(ctx); err != nil {
					log.Error(ctx, fmt.Sprintf("Failed to send email to %s. err : %s", to, err.Error()))
					continue
				}
			}
		}

	}

	return nil
}

func (u *SystemNotificationUsecase) Update(ctx context.Context, dto model.SystemNotification) error {
	return nil
}

func (u *SystemNotificationUsecase) Get(ctx context.Context, systemNotificationId uuid.UUID) (systemNotification model.SystemNotification, err error) {
	userInfo, ok := request.UserFrom(ctx)
	if !ok {
		return systemNotification, httpErrors.NewUnauthorizedError(fmt.Errorf("Invalid token"), "A_INVALID_TOKEN", "")
	}

	systemNotification, err = u.repo.Get(ctx, systemNotificationId)
	if err != nil {
		return systemNotification, err
	}
	u.makeAdditionalInfo(&systemNotification, userInfo.GetUserId())

	user, err := u.userRepo.GetByUuid(ctx, userInfo.GetUserId())
	if err == nil {
		err = u.repo.UpdateRead(ctx, systemNotificationId, user)
		if err != nil {
			return systemNotification, err
		}
	}

	return
}

func (u *SystemNotificationUsecase) GetByName(ctx context.Context, organizationId string, name string) (out model.SystemNotification, err error) {
	out, err = u.repo.GetByName(ctx, organizationId, name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return out, httpErrors.NewNotFoundError(err, "", "")
		}
		return out, err
	}
	return
}

func (u *SystemNotificationUsecase) FetchSystemNotifications(ctx context.Context, organizationId string, pg *pagination.Pagination) (systemNotifications []model.SystemNotification, err error) {
	userInfo, ok := request.UserFrom(ctx)
	if !ok {
		return systemNotifications, httpErrors.NewUnauthorizedError(fmt.Errorf("Invalid token"), "A_INVALID_TOKEN", "")
	}

	systemNotifications, err = u.repo.FetchSystemNotifications(ctx, organizationId, pg)
	if err != nil {
		return nil, err
	}

	for i := range systemNotifications {
		u.makeAdditionalInfo(&systemNotifications[i], userInfo.GetUserId())
	}

	return systemNotifications, nil
}

func (u *SystemNotificationUsecase) FetchPolicyNotifications(ctx context.Context, organizationId string, pg *pagination.Pagination) (systemNotifications []model.SystemNotification, err error) {
	userInfo, ok := request.UserFrom(ctx)
	if !ok {
		return systemNotifications, httpErrors.NewUnauthorizedError(fmt.Errorf("Invalid token"), "A_INVALID_TOKEN", "")
	}

	systemNotifications, err = u.repo.FetchPolicyNotifications(ctx, organizationId, pg)
	if err != nil {
		return nil, err
	}

	for i := range systemNotifications {
		u.makeAdditionalInfo(&systemNotifications[i], userInfo.GetUserId())
	}

	return systemNotifications, nil
}

func (u *SystemNotificationUsecase) Delete(ctx context.Context, dto model.SystemNotification) (err error) {
	_, ok := request.UserFrom(ctx)
	if !ok {
		return httpErrors.NewUnauthorizedError(fmt.Errorf("Invalid token"), "A_INVALID_TOKEN", "")
	}

	_, err = u.Get(ctx, dto.ID)
	if err != nil {
		return httpErrors.NewNotFoundError(err, "EV_NOT_FOUND_EVENT", "")
	}

	err = u.repo.Delete(ctx, dto)
	if err != nil {
		return err
	}

	return nil
}

func (u *SystemNotificationUsecase) CreateSystemNotificationAction(ctx context.Context, dto model.SystemNotificationAction) (systemNotificationActionId uuid.UUID, err error) {
	user, ok := request.UserFrom(ctx)
	if !ok {
		return uuid.Nil, httpErrors.NewUnauthorizedError(fmt.Errorf("Invalid token"), "A_INVALID_TOKEN", "")
	}

	_, err = u.repo.Get(ctx, dto.SystemNotificationId)
	if err != nil {
		return uuid.Nil, httpErrors.NewBadRequestError(fmt.Errorf("Not found systemNotification"), "EV_NOT_FOUND_EVENT", "")
	}

	userId := user.GetUserId()
	dto.TakerId = &userId
	dto.CreatedAt = time.Now()

	systemNotificationActionId, err = u.repo.CreateSystemNotificationAction(ctx, dto)
	if err != nil {
		return uuid.Nil, err
	}
	log.Info(ctx, "newly created systemNotificationActionId:", systemNotificationActionId)

	return
}

func (u *SystemNotificationUsecase) getOrganizationFromCluster(clusters *[]model.Cluster, strId string) (organizationId string, err error) {
	clusterId := domain.ClusterId(strId)
	if !clusterId.Validate() {
		return "", fmt.Errorf("Invalid clusterId %s", strId)
	}

	for _, cluster := range *clusters {
		if cluster.ID == clusterId {
			return cluster.OrganizationId, nil
		}
	}

	return "", fmt.Errorf("No martched organization %s", strId)
}

func (u *SystemNotificationUsecase) makeAdditionalInfo(systemNotification *model.SystemNotification, userId uuid.UUID) {

	systemNotification.FiredAt = &systemNotification.CreatedAt
	//systemNotification.Status = model.SystemNotificationActionStatus_CREATED

	if len(systemNotification.SystemNotificationActions) > 0 {
		systemNotification.TakedAt = &systemNotification.SystemNotificationActions[0].CreatedAt
		for _, action := range systemNotification.SystemNotificationActions {
			if action.Status == domain.SystemNotificationActionStatus_CLOSED {
				systemNotification.ClosedAt = &action.CreatedAt
				systemNotification.ProcessingSec = int((action.CreatedAt).Sub(systemNotification.CreatedAt).Seconds())
			}
		}

		systemNotification.LastTaker = systemNotification.SystemNotificationActions[len(systemNotification.SystemNotificationActions)-1].Taker
		systemNotification.TakedSec = int((systemNotification.SystemNotificationActions[0].CreatedAt).Sub(systemNotification.CreatedAt).Seconds())
		//systemNotification.Status = systemNotification.SystemNotificationActions[len(systemNotification.SystemNotificationActions)-1].Status
	}

	systemNotification.Read = false
	for _, v := range systemNotification.Readers {
		if v.ID == userId {
			systemNotification.Read = true
			break
		}
	}
}

func (u *SystemNotificationUsecase) makeGrafanaUrl(ctx context.Context, primaryCluster model.Cluster, systemNotification domain.SystemNotificationRequest, clusterId domain.ClusterId) (url string) {
	primaryGrafanaEndpoint := ""
	appGroups, err := u.appGroupRepo.Fetch(ctx, primaryCluster.ID, nil)
	if err == nil {
		for _, appGroup := range appGroups {
			if appGroup.AppGroupType == domain.AppGroupType_LMA {
				applications, err := u.appGroupRepo.GetApplications(ctx, appGroup.ID, domain.ApplicationType_GRAFANA)
				if err != nil {
					return ""
				}
				if len(applications) > 0 {
					primaryGrafanaEndpoint = applications[0].Endpoint
				}
			}
		}
	}

	//// check type
	//url = primaryGrafanaEndpoint

	// tks_node_dashboard/tks-kubernetes-view-nodes?orgId=1&refresh=30s&var-datasource=default&var-taco_cluster=c19rjkn4j&var-job=prometheus-node-exporter&var-hostname=All&var-node=10.0.168.71:9100&var-device=All&var-maxmount=%2F&var-show_hostname=prometheus-node-exporter-xt4vb

	switch systemNotification.Labels.AlertName {
	case "node-memory-high-utilization":
		url = primaryGrafanaEndpoint + "/d/tks_node_dashboard/tks-kubernetes-view-nodes?var-taco_cluster=" + clusterId.String() + "&kiosk"
	case "node-cpu-high-load":
		url = primaryGrafanaEndpoint + "/d/tks_node_dashboard/tks-kubernetes-view-nodes?var-taco_cluster=" + clusterId.String() + "&kiosk"
	case "node-disk-full":
		url = primaryGrafanaEndpoint + "/d/tks_node_dashboard/tks-kubernetes-view-nodes?var-taco_cluster=" + clusterId.String() + "&kiosk"
	case "pod-restart-frequently":
		url = primaryGrafanaEndpoint + "/d/tks_podv1_dashboard/tks-kubernetes-view-pods-v1?var-taco_cluster=" + clusterId.String() + "&kiosk"
	case "pvc-full":
		url = primaryGrafanaEndpoint + "/d/tks_cluster_dashboard/tks-kubernetes-view-cluster-global?var-taco_cluster=" + clusterId.String() + "&kiosk"
	default:
		url = primaryGrafanaEndpoint + "/d/tks_cluster_dashboard"
	}

	return url
}
