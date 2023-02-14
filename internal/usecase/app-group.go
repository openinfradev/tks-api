package usecase

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/domain"
	"github.com/openinfradev/tks-api/internal/repository"
)

type IAppGroupUsecase interface {
	Fetch(clusterId string) ([]domain.AppGroup, error)
	Create(clusterId string, name string, appGroupType string, creatorId string, description string) (appGroupId string, err error)
	Get(appGroupId string) (out domain.AppGroup, err error)
	Delete(appGroupId string) (err error)
}

type AppGroupUsecase struct {
	repo repository.IAppGroupRepository
}

func NewAppGroupUsecase(r repository.IAppGroupRepository) IAppGroupUsecase {
	return &AppGroupUsecase{
		repo: r,
	}
}

func (u *AppGroupUsecase) Fetch(clusterId string) (out []domain.AppGroup, err error) {
	out, err = u.repo.Fetch(clusterId)

	if err != nil {
		return nil, err
	}
	return out, nil
}

func (u *AppGroupUsecase) Create(clusterId string, name string, appGroupType string, creatorId string, description string) (appGroupId string, err error) {
	creator := uuid.Nil
	if creatorId != "" {
		creator, err = uuid.Parse(creatorId)
		if err != nil {
			return "", fmt.Errorf("Invalid Creator ID %s", creatorId)
		}
	}

	appGroupId, err = u.repo.Create(clusterId, name, appGroupType, creator, description)
	if err != nil {
		return "", fmt.Errorf("Failed to create appGroup. err %s", err)
	}
	return appGroupId, nil
}

func (u *AppGroupUsecase) Get(appGroupId string) (out domain.AppGroup, err error) {
	appGroup, err := u.repo.Get(appGroupId)
	if err != nil {
		return domain.AppGroup{}, err
	}
	return appGroup, nil
}

func (u *AppGroupUsecase) Delete(appGroupId string) (err error) {
	_, err = u.repo.Get(appGroupId)
	if err != nil {
		return fmt.Errorf("No appGroup for deletiing : %s", appGroupId)
	}

	err = u.repo.Delete(appGroupId)
	if err != nil {
		return fmt.Errorf("Fatiled to deleting appGroup : %s", appGroupId)
	}

	return nil
}
