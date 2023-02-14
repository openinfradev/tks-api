package usecase

import (
	"github.com/openinfradev/tks-api/internal/domain"
	"github.com/openinfradev/tks-api/internal/repository"
)

type IHistoryUsecase interface {
	Fetch() ([]domain.History, error)
}

type HistoryUsecase struct {
	historyRepo repository.IHistoryRepository
}

func NewHistoryUsecase(r repository.IHistoryRepository) IHistoryUsecase {
	return &HistoryUsecase{
		historyRepo: r,
	}
}

func (u *HistoryUsecase) Fetch() (out []domain.History, err error) {
	histories, err := u.historyRepo.Fetch()
	if err != nil {
		return nil, err
	}

	return histories, nil
}
