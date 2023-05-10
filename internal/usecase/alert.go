package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/openinfradev/tks-api/internal/middleware/auth/request"

	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type IAlertUsecase interface {
	Get(alertId uuid.UUID) (domain.Alert, error)
	GetByName(organizationId string, name string) (domain.Alert, error)
	Fetch(organizationId string) ([]domain.Alert, error)
	Create(ctx context.Context, dto domain.CreateAlertRequest) (err error)
	Update(ctx context.Context, dto domain.Alert) error
	Delete(ctx context.Context, dto domain.Alert) error

	CreateAlertAction(ctx context.Context, dto domain.AlertAction) (alertActionId uuid.UUID, err error)
}

type AlertUsecase struct {
	repo        repository.IAlertRepository
	clusterRepo repository.IClusterRepository
}

func NewAlertUsecase(r repository.Repository) IAlertUsecase {
	return &AlertUsecase{
		repo:        r.Alert,
		clusterRepo: r.Cluster,
	}
}

func (u *AlertUsecase) Create(ctx context.Context, input domain.CreateAlertRequest) (err error) {
	if input.Alerts == nil || len(input.Alerts) == 0 {
		return fmt.Errorf("No data found")
	}

	allClusters, err := u.clusterRepo.Fetch()
	if err != nil {
		return fmt.Errorf("No clusters")
	}

	for _, alert := range input.Alerts {
		clusterId := alert.Labels.TacoCluster
		organizationId, err := u.getOrganizationFromCluster(&allClusters, clusterId)
		if err != nil {
			log.Error(err)
			continue
		}

		rawData, err := json.Marshal(alert)
		if err != nil {
			rawData = []byte{}
		}
		dto := domain.Alert{
			OrganizationId: organizationId,
			Name:           alert.Labels.AlertName,
			Code:           alert.Labels.AlertName,
			Grade:          alert.Labels.Severity,
			Message:        alert.Annotations.Message,
			Description:    alert.Annotations.Description,
			Node:           alert.Annotations.Node,
			CheckPoint:     alert.Annotations.Checkpoint,
			Summary:        alert.Annotations.Summary,
			ClusterId:      domain.ClusterId(clusterId),
			GrafanaUrl:     "http://localhost/grafana",
			RawData:        rawData,
		}

		_, err = u.repo.Create(dto)
		if err != nil {
			continue
		}
	}

	return nil
}

func (u *AlertUsecase) Update(ctx context.Context, dto domain.Alert) error {
	return nil
}

func (u *AlertUsecase) Get(alertId uuid.UUID) (alert domain.Alert, err error) {
	alert, err = u.repo.Get(alertId)
	if err != nil {
		return domain.Alert{}, err
	}
	u.makeAdditionalInfo(&alert)

	return
}

func (u *AlertUsecase) GetByName(organizationId string, name string) (res domain.Alert, err error) {
	res, err = u.repo.GetByName(organizationId, name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Alert{}, httpErrors.NewNotFoundError(err)
		}
		return domain.Alert{}, err
	}
	return
}

func (u *AlertUsecase) Fetch(organizationId string) (alerts []domain.Alert, err error) {
	alerts, err = u.repo.Fetch(organizationId)
	if err != nil {
		return nil, err
	}

	for i := range alerts {
		// make data ( FiredAt, TakedAt, ClosedAt )
		u.makeAdditionalInfo(&alerts[i])
	}

	return alerts, nil
}

func (u *AlertUsecase) Delete(ctx context.Context, dto domain.Alert) (err error) {
	user, ok := request.UserFrom(ctx)
	if !ok {
		return httpErrors.NewBadRequestError(fmt.Errorf("Invalid token"))
	}

	_, err = u.Get(dto.ID)
	if err != nil {
		return httpErrors.NewNotFoundError(err)
	}

	*dto.UpdatorId = user.GetUserId()

	err = u.repo.Delete(dto)
	if err != nil {
		return err
	}

	return nil
}

func (u *AlertUsecase) CreateAlertAction(ctx context.Context, dto domain.AlertAction) (alertActionId uuid.UUID, err error) {
	user, ok := request.UserFrom(ctx)
	if !ok {
		return uuid.Nil, httpErrors.NewBadRequestError(fmt.Errorf("Invalid token"))
	}

	_, err = u.repo.Get(dto.AlertId)
	if err != nil {
		return uuid.Nil, httpErrors.NewBadRequestError(fmt.Errorf("Not found alert"))
	}

	userId := user.GetUserId()
	dto.TakerId = &userId
	dto.CreatedAt = time.Now()

	alertActionId, err = u.repo.CreateAlertAction(dto)
	if err != nil {
		return uuid.Nil, err
	}
	log.Info("newly created alertActionId:", alertActionId)

	return
}

func (u *AlertUsecase) getOrganizationFromCluster(clusters *[]domain.Cluster, strId string) (organizationId string, err error) {
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

func (u *AlertUsecase) makeAdditionalInfo(alert *domain.Alert) {
	alert.FiredAt = &alert.CreatedAt
	alert.Status = domain.AlertActionStatus_CREATED

	if len(alert.AlertActions) > 0 {
		alert.TakedAt = &alert.AlertActions[0].CreatedAt
		for _, action := range alert.AlertActions {
			if action.Status == domain.AlertActionStatus_CLOSED {
				alert.ClosedAt = &action.CreatedAt
				alert.ProcessingSec = int((action.CreatedAt).Sub(alert.CreatedAt).Seconds())
			}
		}

		alert.LastTaker = alert.AlertActions[len(alert.AlertActions)-1].Taker
		alert.TakedSec = int((alert.AlertActions[0].CreatedAt).Sub(alert.CreatedAt).Seconds())
		alert.Status = alert.AlertActions[len(alert.AlertActions)-1].Status
	}

	// make data grafana URL
	alert.GrafanaUrl = "TODO url"
	alert.Node = fmt.Sprintf("STACK (%s) / NODE (%s)", alert.Cluster.Name, alert.Node)
}
