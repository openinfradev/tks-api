package repository

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/openinfradev/tks-api/internal/domain"
	"github.com/openinfradev/tks-api/internal/helper"
)

// Interfaces
type IContractRepository interface {
	Fetch() (res []domain.Contract, err error)
	Get(contractId string) (res domain.Contract, err error)
	Create(name string, creator uuid.UUID, description string) (string, error)
}

type ContractRepository struct {
	db *gorm.DB
}

func NewContractRepository(db *gorm.DB) IContractRepository {
	return &ContractRepository{
		db: db,
	}
}

// Models
type Contract struct {
	gorm.Model
	ID          string `gorm:"primarykey"`
	Name        string `gorm:"uniqueIndex"`
	Creator     uuid.UUID
	Description string
}

func (c *Contract) BeforeCreate(tx *gorm.DB) (err error) {
	c.ID = helper.GenerateContractId()
	return nil
}

// Logics
func (r *ContractRepository) Fetch() (out []domain.Contract, err error) {
	var contracts []Contract
	out = []domain.Contract{}

	res := r.db.Find(&contracts)
	if res.Error != nil {
		return nil, res.Error
	}
	for _, contract := range contracts {
		outContract := r.reflect(contract)
		out = append(out, outContract)
	}
	return out, nil
}

func (r *ContractRepository) Get(id string) (domain.Contract, error) {
	var contract Contract
	res := r.db.First(&contract, "id = ?", id)
	if res.RowsAffected == 0 || res.Error != nil {
		return domain.Contract{}, fmt.Errorf("Not found contract for %s", id)
	}
	resContract := r.reflect(contract)
	return resContract, nil
}

func (r *ContractRepository) Create(name string, creator uuid.UUID, description string) (string, error) {
	contract := Contract{Name: name, Creator: creator, Description: description}
	err := r.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Create(&contract)
		if res.Error != nil {
			return res.Error
		}
		return nil
	})

	return contract.ID, err
}

func (r *ContractRepository) Delete(contractId string) error {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Delete(&Contract{}, "id = ?", contractId)
		if res.Error != nil {
			return fmt.Errorf("could not delete contract for contractId %s", contractId)
		}
		return nil
	})
	return err
}

func (r *ContractRepository) reflect(contract Contract) domain.Contract {
	return domain.Contract{
		Id:          contract.ID,
		Name:        contract.Name,
		Description: contract.Description,
		Creator:     contract.Creator.String(),
		CreatedAt:   contract.CreatedAt,
		UpdatedAt:   contract.UpdatedAt,
	}
}
