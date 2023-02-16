package usecase

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/domain"
	"github.com/openinfradev/tks-api/internal/repository"
	argowf "github.com/openinfradev/tks-api/pkg/argo-client"
	"github.com/openinfradev/tks-api/pkg/log"
	"github.com/spf13/viper"
)

type IContractUsecase interface {
	Fetch() ([]domain.Contract, error)
	Get(contractId string) (domain.Contract, error)
	Create(domain.Contract) (contractId string, err error)
}

type ContractUsecase struct {
	repo repository.IContractRepository
	argo argowf.ArgoClient
}

func NewContractUsecase(r repository.IContractRepository, argoClient argowf.ArgoClient) IContractUsecase {
	return &ContractUsecase{
		repo: r,
		argo: argoClient,
	}
}

func (u *ContractUsecase) Fetch() (out []domain.Contract, err error) {
	contracts, err := u.repo.Fetch()
	if err != nil {
		return nil, err
	}
	return contracts, nil
}

func (u *ContractUsecase) Create(in domain.Contract) (contractId string, err error) {
	creator := uuid.Nil
	if in.Creator != "" {
		creator, err = uuid.Parse(in.Creator)
		if err != nil {
			return "", err
		}
	}

	contractId, err = u.repo.Create(in.Name, creator, in.Description)
	if err != nil {
		return "", err
	}
	log.Info("newly created Contract Id:", contractId)

	workflowId, err := u.argo.SumbitWorkflowFromWftpl(
		"tks-create-contract-repo",
		argowf.SubmitOptions{
			Parameters: []string{
				"contract_id=" + contractId,
				"revision=" + viper.GetString("revision"),
			},
		})
	if err != nil {
		log.Error("failed to submit argo workflow template. err : ", err)
		return "", fmt.Errorf("Failed to call argo workflow : %s", err)
	}
	log.Info("submited workflow :", workflowId)

	return contractId, nil
}

func (u *ContractUsecase) Get(contractId string) (res domain.Contract, err error) {
	res, err = u.repo.Get(contractId)
	if err != nil {
		return domain.Contract{}, err
	}
	return res, nil
}
