package usecase

import (
	"github.com/openinfradev/tks-api/internal/repository"
)

type IProjectUsecase interface {
}

type ProjectUsecase struct {
	userRepository         repository.IUserRepository
	authRepository         repository.IAuthRepository
	clusterRepository      repository.IClusterRepository
	appgroupRepository     repository.IAppGroupRepository
	organizationRepository repository.IOrganizationRepository
}

func NewProjectUsecase(r repository.Repository) IProjectUsecase {
	return &ProjectUsecase{
		userRepository:         r.User,
		authRepository:         r.Auth,
		clusterRepository:      r.Cluster,
		appgroupRepository:     r.AppGroup,
		organizationRepository: r.Organization,
	}
}
