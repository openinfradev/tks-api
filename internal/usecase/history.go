package usecase

import (
	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-api/pkg/domain"
)

type IHistoryUsecase interface {
	Fetch() ([]domain.History, error)
}

type HistoryUsecase struct {
	historyRepo repository.IHistoryRepository
}

func NewHistoryUsecase(r repository.Repository) IHistoryUsecase {
	return &HistoryUsecase{
		historyRepo: r.History,
	}
}

func (u *HistoryUsecase) Fetch() (out []domain.History, err error) {
	histories, err := u.historyRepo.Fetch()
	if err != nil {
		return nil, err
	}

	return histories, nil
}
