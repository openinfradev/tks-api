package repository

import (
	"context"
	"fmt"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/pkg/domain"
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
	FetchByClusterId(ctx context.Context, clusterId string, pg *pagination.Pagination) (out *[]model.Policy, err error)
	FetchByClusterIdAndTemplaeId(ctx context.Context, clusterId string, templateId uuid.UUID) (out *[]model.Policy, err error)
	ExistByName(ctx context.Context, organizationId string, policyName string) (exist bool, err error)
	ExistByResourceName(ctx context.Context, organizationId string, policyName string) (exist bool, err error)
	ExistByID(ctx context.Context, organizationId string, policyId uuid.UUID) (exist bool, err error)
	GetByName(ctx context.Context, organizationId string, policyName string) (out *model.Policy, err error)
	GetByID(ctx context.Context, organizationId string, policyId uuid.UUID) (out *model.Policy, err error)
	Delete(ctx context.Context, organizationId string, policyId uuid.UUID) (err error)
	UpdatePolicyTargetClusters(ctx context.Context, organizationId string, policyId uuid.UUID, currentClusterIds []string, targetClusters []model.Cluster) (err error)
	SetMandatoryPolicies(ctx context.Context, organizationId string, mandatoryPolicyIds []uuid.UUID, nonMandatoryPolicyIds []uuid.UUID) (err error)
	GetUsageCountByTemplateId(ctx context.Context, organizationId *string, policyTemplateId uuid.UUID) (usageCounts []model.UsageCount, err error)
	CountPolicyByEnforcementAction(ctx context.Context, organizationId string) (policyCount []model.PolicyCount, err error)
	AddPoliciesForClusterID(ctx context.Context, organizationId string, clusterId domain.ClusterId, policies []model.Policy) (err error)
	UpdatePoliciesForClusterID(ctx context.Context, organizationId string, clusterId domain.ClusterId, policies []model.Policy) (err error)
	DeletePoliciesForClusterID(ctx context.Context, organizationId string, clusterId domain.ClusterId, policyIds []uuid.UUID) (err error)
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

func (r *PolicyRepository) FetchByClusterId(ctx context.Context, clusterId string, pg *pagination.Pagination) (out *[]model.Policy, err error) {
	if pg == nil {
		pg = pagination.NewPagination(nil)
	}

	subQueryClusterId := r.db.Table("policy_target_clusters").Select("policy_id").
		Where("cluster_id = ?", clusterId)

	_, res := pg.Fetch(r.db.WithContext(ctx).Preload(clause.Associations).
		Where("id in (?)", subQueryClusterId), &out)

	if res.Error != nil {
		return nil, res.Error
	}
	return
}

func (r *PolicyRepository) FetchByClusterIdAndTemplaeId(ctx context.Context, clusterId string, templateId uuid.UUID) (out *[]model.Policy, err error) {
	subQueryClusterId := r.db.Table("policy_target_clusters").Select("policy_id").
		Where("cluster_id = ?", clusterId)

	res := r.db.WithContext(ctx).Preload(clause.Associations).
		Where("template_id = ?").Where("id in (?)", subQueryClusterId).Find(&out)

	if res.Error != nil {
		return nil, res.Error
	}
	return
}

func (r *PolicyRepository) ExistBy(ctx context.Context, organizationId string, key string, value interface{}) (exists bool, err error) {
	query := fmt.Sprintf("organization_id = ? and %s = ?", key)

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
	return r.ExistBy(ctx, organizationId, "policy_name", policyName)
}

func (r *PolicyRepository) ExistByResourceName(ctx context.Context, organizationId string, policyName string) (exist bool, err error) {
	return r.ExistBy(ctx, organizationId, "policy_resource_name", policyName)
}

func (r *PolicyRepository) ExistByID(ctx context.Context, organizationId string, policyId uuid.UUID) (exist bool, err error) {
	return r.ExistBy(ctx, organizationId, "id", policyId)
}

func (r *PolicyRepository) GetBy(ctx context.Context, organizationId string, key string, value interface{}) (out *model.Policy, err error) {
	query := fmt.Sprintf("organization_id = ? and %s = ?", key)

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
		if len(mandatoryPolicyIds) > 0 {
			if err = tx.Model(&model.Policy{}).
				Where("organization_id = ?", organizationId).
				Where("id in ?", mandatoryPolicyIds).
				Update("mandatory", true).Error; err != nil {
				return err
			}
		}

		if len(nonMandatoryPolicyIds) > 0 {
			if err = tx.Model(&model.Policy{}).
				Where("organization_id = ?", organizationId).
				Where("id in ?", nonMandatoryPolicyIds).
				Update("mandatory", false).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *PolicyRepository) GetUsageCountByTemplateId(ctx context.Context, organizationId *string, policyTemplateId uuid.UUID) (usageCounts []model.UsageCount, err error) {
	// 다음과 같은 쿼리, organization_id 가 nil인 경우는 and organization_id = '...'를 조합하지 않음
	// select organizations.id, organizations.name, count(organizations.id) from policies join organizations
	// 	on policies.organization_id = organizations.id
	// 	where policies.template_id='7f8a9f78-1771-43d4-aa4a-c395b43ebdd6'
	// 	and organization_id = 'ozvnzr3oz'
	// 	group by  organizations.id

	query := r.db.WithContext(ctx).Model(&model.Policy{}).
		Select("organizations.id as organization_id", "organizations.name as organization_name", "count(organizations.id) as usage_count").
		Joins("join organizations on policies.organization_id = organizations.id").
		Where("template_id = ?", policyTemplateId)

	if organizationId != nil {
		query = query.Where("organization_id = ?", organizationId)
	}

	err = query.Group("organizations.id").Scan(&usageCounts).Error

	if err != nil {
		return nil, err
	}

	return
}

func (r *PolicyRepository) CountPolicyByEnforcementAction(ctx context.Context, organizationId string) (policyCount []model.PolicyCount, err error) {

	err = r.db.WithContext(ctx).Model(&model.Policy{}).
		Select("enforcement_action", "count(enforcement_action) as count").
		Where("organization_id = ?", organizationId).
		Group("enforcement_action").Scan(&policyCount).Error

	if err != nil {
		return nil, err
	}

	return
}

func (r *PolicyRepository) AddPoliciesForClusterID(ctx context.Context, organizationId string, clusterId domain.ClusterId, policies []model.Policy) (err error) {
	var cluster model.Cluster
	cluster.ID = clusterId

	err = r.db.WithContext(ctx).Model(&cluster).
		Association("Policies").Append(policies)

	return err
}

func (r *PolicyRepository) UpdatePoliciesForClusterID(ctx context.Context, organizationId string, clusterId domain.ClusterId, policies []model.Policy) (err error) {
	var cluster model.Cluster
	cluster.ID = clusterId

	err = r.db.WithContext(ctx).Model(&cluster).
		Association("Policies").Replace(policies)

	return err
}

func (r *PolicyRepository) DeletePoliciesForClusterID(ctx context.Context, organizationId string, clusterId domain.ClusterId, policyIds []uuid.UUID) (err error) {
	return r.db.WithContext(ctx).
		Where("cluster_id = ?", clusterId).
		Where("policy_id in ?", policyIds).
		Delete(&model.PolicyTargetCluster{}).Error
}
