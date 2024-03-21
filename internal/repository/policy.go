package repository

import (
	"context"
	"fmt"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/pkg/log"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IPolicyRepository interface {
	Create(ctx context.Context, dto model.Policy) (policyId uuid.UUID, err error)
	Update(ctx context.Context, organizationId string, policyId uuid.UUID,
		updateMap map[string]interface{}, TargetClusters *[]model.Cluster) (err error)
	Fetch(ctx context.Context, organizationId string, pg *pagination.Pagination) (out *[]model.Policy, err error)
	ExistByName(ctx context.Context, organizationId string, policyName string) (exist bool, err error)
	ExistByID(ctx context.Context, organizationId string, policyId uuid.UUID) (exist bool, err error)
	GetByName(ctx context.Context, organizationId string, policyName string) (out *model.Policy, err error)
	GetByID(ctx context.Context, organizationId string, policyId uuid.UUID) (out *model.Policy, err error)
	Delete(ctx context.Context, organizationId string, policyId uuid.UUID) (err error)
	UpdatePolicyTargetClusters(ctx context.Context, organizationId string, policyId uuid.UUID, currentClusterIds []string, targetClusters []model.Cluster) (err error)
	SetMandatoryPolicies(ctx context.Context, organizationId string, mandatoryPolicyIds []uuid.UUID, nonMandatoryPolicyIds []uuid.UUID) (err error)
}

type PolicyRepository struct {
	db *gorm.DB
}

func NewPolicyRepository(db *gorm.DB) IPolicyRepository {
	return &PolicyRepository{
		db: db,
	}
}

func (r *PolicyRepository) Create(ctx context.Context, dto model.Policy) (policyId uuid.UUID, err error) {
	err = r.db.WithContext(ctx).Create(&dto).Error

	if err != nil {
		return uuid.Nil, err
	}

	return dto.ID, nil
}

func (r *PolicyRepository) Update(ctx context.Context, organizationId string, policyId uuid.UUID,
	updateMap map[string]interface{}, targetClusters *[]model.Cluster) (err error) {

	var policy model.Policy
	policy.ID = policyId

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if targetClusters != nil {
			err = tx.WithContext(ctx).Model(&policy).Limit(1).
				Association("TargetClusters").Replace(targetClusters)

			if err != nil {
				return err
			}
		}

		if len(updateMap) > 0 {
			err = tx.WithContext(ctx).Model(&policy).Limit(1).
				Where("id = ?", policyId).
				Updates(updateMap).Error

			if err != nil {
				return err
			}
		}

		// return nil will commit the whole transaction
		return nil
	})
}

func (r *PolicyRepository) Fetch(ctx context.Context, organizationId string, pg *pagination.Pagination) (out *[]model.Policy, err error) {
	if pg == nil {
		pg = pagination.NewPagination(nil)
	}

	_, res := pg.Fetch(r.db.WithContext(ctx).Preload(clause.Associations).
		Where("organization_id = ?", organizationId), &out)

	if res.Error != nil {
		return nil, res.Error
	}
	return
}

func (r *PolicyRepository) ExistBy(ctx context.Context, organizationId string, key string, value interface{}) (exists bool, err error) {
	query := fmt.Sprintf("organization_id = ? and %s = ?", value)

	var policy model.Policy
	res := r.db.WithContext(ctx).Where(query, organizationId, value).
		First(&policy)

	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			log.Infof(ctx, "Not found policy %s='%v'", key, value)
			return false, nil
		} else {
			log.Error(ctx, res.Error)
			return false, res.Error
		}
	}

	return true, nil
}

func (r *PolicyRepository) ExistByName(ctx context.Context, organizationId string, policyName string) (exist bool, err error) {
	return r.ExistBy(ctx, "policy_name", organizationId, policyName)
}

func (r *PolicyRepository) ExistByID(ctx context.Context, organizationId string, policyId uuid.UUID) (exist bool, err error) {
	return r.ExistBy(ctx, "id", organizationId, policyId)
}

func (r *PolicyRepository) GetBy(ctx context.Context, organizationId string, key string, value interface{}) (out *model.Policy, err error) {
	query := fmt.Sprintf("organization_id = ? and %s = ?", value)

	var policy model.Policy
	res := r.db.WithContext(ctx).Preload(clause.Associations).
		Where(query, organizationId, value).First(&policy)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			log.Infof(ctx, "Not found policy %s='%v'", key, value)
			return nil, nil
		} else {
			log.Error(ctx, res.Error)
			return nil, res.Error
		}
	}

	return &policy, nil
}

func (r *PolicyRepository) GetByName(ctx context.Context, organizationId string, policyName string) (out *model.Policy, err error) {
	return r.GetBy(ctx, organizationId, "policy_name", policyName)
}

func (r *PolicyRepository) GetByID(ctx context.Context, organizationId string, policyId uuid.UUID) (out *model.Policy, err error) {
	return r.GetBy(ctx, organizationId, "id", policyId)
}

func (r *PolicyRepository) Delete(ctx context.Context, organizationId string, policyId uuid.UUID) (err error) {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Policy{ID: policyId}).Association("TargetClusters").Clear(); err != nil {
			return err
		}

		if err := tx.Where("organization_id = ? and id = ?", organizationId, policyId).Delete(&model.Policy{}).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *PolicyRepository) UpdatePolicyTargetClusters(ctx context.Context, organizationId string, policyId uuid.UUID, currentClusterIds []string, targetClusters []model.Cluster) (err error) {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var policy model.Policy
		res := tx.Preload("TargetClusters").Where("organization_id = ? and id = ?", organizationId, policyId).First(&policy)

		if res.Error != nil {
			return res.Error
		}

		if len(policy.TargetClusterIds) == 0 && len(currentClusterIds) != 0 {
			return errors.New("concurrent modification of target clusters")
		}

		actualCurrentClusterIdSet := mapset.NewSet(policy.TargetClusterIds...)
		knownCurrentClusterIdSet := mapset.NewSet(currentClusterIds...)

		if !actualCurrentClusterIdSet.Equal(knownCurrentClusterIdSet) {
			return errors.New("concurrent modification of target clusters")
		}

		err = tx.Model(&policy).Limit(1).
			Association("TargetClusters").Replace(targetClusters)

		if err != nil {
			return err
		}

		// return nil will commit the whole transaction
		return nil
	})
}

func (r *PolicyRepository) SetMandatoryPolicies(ctx context.Context, organizationId string, mandatoryPolicyIds []uuid.UUID, nonMandatoryPolicyIds []uuid.UUID) (err error) {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err = tx.Model(&model.Policy{}).
			Where("organization_id = ? id in ?", organizationId, mandatoryPolicyIds).
			Update("mandatory", true).Error; err != nil {
			return err
		}

		if err = tx.Model(&model.Policy{}).
			Where("organization_id = ? id in ?", organizationId, nonMandatoryPolicyIds).
			Update("mandatory", false).Error; err != nil {
			return err
		}

		return nil
	})
}
