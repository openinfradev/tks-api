package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/log"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type IPolicyTemplateRepository interface {
	Create(ctx context.Context, policyTemplate model.PolicyTemplate) (policyTemplateId uuid.UUID, err error)
	Update(ctx context.Context, policyTemplateId uuid.UUID, updateMap map[string]interface{}, permittedOrganizations *[]model.Organization) (err error)
	Fetch(ctx context.Context, pg *pagination.Pagination) (out []model.PolicyTemplate, err error)
	FetchForOrganization(ctx context.Context, organizationId string, pg *pagination.Pagination) (out []model.PolicyTemplate, err error)
	GetByName(ctx context.Context, policyTemplateName string) (out *model.PolicyTemplate, err error)
	GetByKind(ctx context.Context, policyTemplateKind string) (out *model.PolicyTemplate, err error)
	GetByID(ctx context.Context, policyTemplateId uuid.UUID) (out *model.PolicyTemplate, err error)
	Delete(ctx context.Context, policyTemplateId uuid.UUID) (err error)
	ExistByName(ctx context.Context, policyTemplateName string) (exist bool, err error)
	ExistByKind(ctx context.Context, policyTemplateKind string) (exist bool, err error)
	ExistByNameInOrganization(ctx context.Context, organizationId string, policyTemplateName string) (exist bool, err error)
	ExistByKindInOrganization(ctx context.Context, organizationId string, policyTemplateKind string) (exist bool, err error)
	ExistByID(ctx context.Context, policyTemplateId uuid.UUID) (exist bool, err error)
	ListPolicyTemplateVersions(ctx context.Context, policyTemplateId uuid.UUID) (policyTemplateVersionsReponse *domain.ListPolicyTemplateVersionsResponse, err error)
	GetPolicyTemplateVersion(ctx context.Context, policyTemplateId uuid.UUID, version string) (policyTemplateVersionsReponse *model.PolicyTemplate, err error)
	DeletePolicyTemplateVersion(ctx context.Context, policyTemplateId uuid.UUID, version string) (err error)
	CreatePolicyTemplateVersion(ctx context.Context, policyTemplateId uuid.UUID, newVersion string,
		schema []*domain.ParameterDef, rego string, libs []string, syncKinds *[]string, syncJson *string) (version string, err error)
	GetLatestTemplateVersion(ctx context.Context, policyTemplateId uuid.UUID) (version string, err error)
	CountTksTemplateByOrganization(ctx context.Context, organizationId string) (count int64, err error)
	CountOrganizationTemplate(ctx context.Context, organizationId string) (count int64, err error)
	CountPolicyFromOrganizationTemplate(ctx context.Context, organizationId string) (count int64, err error)
	GetPolicyTemplateByOrganizationIdOrTKS(ctx context.Context, organizationId string) ([]model.PolicyTemplate, error)
}

type PolicyTemplateRepository struct {
	db *gorm.DB
}

func NewPolicyTemplateRepository(db *gorm.DB) IPolicyTemplateRepository {
	return &PolicyTemplateRepository{
		db: db,
	}
}

func (r *PolicyTemplateRepository) Create(ctx context.Context, dto model.PolicyTemplate) (policyTemplateId uuid.UUID, err error) {
	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 이미 org가 존재하므로 many2many 레코드를 추가하지 않고 관계만 업데이트하도록 보장
		if err := tx.Omit("PermittedOrganizations").Create(&dto).Error; err != nil {
			return err
		}

		if err := tx.Model(&dto).Association("PermittedOrganizations").
			Append(dto.PermittedOrganizations); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return uuid.Nil, err
	}

	return dto.ID, nil
}

func (r *PolicyTemplateRepository) Update(ctx context.Context, policyTemplateId uuid.UUID,
	updateMap map[string]interface{}, permittedOrganizations *[]model.Organization) (err error) {

	var policyTemplate model.PolicyTemplate
	policyTemplate.ID = policyTemplateId

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if permittedOrganizations != nil {
			err = tx.WithContext(ctx).Model(&policyTemplate).Limit(1).
				Association("PermittedOrganizations").Replace(permittedOrganizations)

			if err != nil {
				return err
			}
		}

		if len(updateMap) > 0 {
			err = tx.WithContext(ctx).Omit("PermittedOrganizations").Model(&policyTemplate).Limit(1).
				Where("id = ?", policyTemplateId).Where("type = ?", "tks").
				Updates(updateMap).Error

			if err != nil {
				return err
			}
		}

		// return nil will commit the whole transaction
		return nil
	})
}

func (r *PolicyTemplateRepository) Fetch(ctx context.Context, pg *pagination.Pagination) (out []model.PolicyTemplate, err error) {
	var policyTemplates []model.PolicyTemplate
	if pg == nil {
		pg = pagination.NewPagination(nil)
	}

	_, res := pg.Fetch(r.db.WithContext(ctx).
		Preload("SupportedVersions", func(db *gorm.DB) *gorm.DB {
			// 최신 버전만
			return db.Order("policy_template_supported_versions.version DESC")
		}).
		Preload("PermittedOrganizations").Preload("Creator").Preload("Updator").
		Model(&model.PolicyTemplate{}).
		Where("type = 'tks'"), &policyTemplates)
	if res.Error != nil {
		return nil, res.Error
	}

	return policyTemplates, nil
}

func (r *PolicyTemplateRepository) FetchForOrganization(ctx context.Context, organizationId string, pg *pagination.Pagination) (out []model.PolicyTemplate, err error) {
	var policyTemplates []model.PolicyTemplate
	if pg == nil {
		pg = pagination.NewPagination(nil)
	}

	// 다음과 같은 쿼리를 생성해서 tks 템플릿에 대해선 PermittedOrganizations가 비거나, PermittedOrganizations에 해당 organizations이 속하는 템플릿을 찾음
	// organization 템플릿은 organizationId가 매칭되는 것을 찾음, 이를 통해 해당 사용자가 사용할 수 있는 모든 템플릿을 fetch
	// select id from policy_templates where
	// 	(
	// 		type = 'tks'
	// 		and (
	// 			id not in (select policy_template_id from policy_template_permitted_organizations) -- PermitedOrganizations이 빈 경우, 모두에게 허용
	// 			or id in (select policy_template_id from policy_template_permitted_organizations organization where organization_id = 'orgid') -- PermitedOrganizations 허용된 경우
	// 		)
	// 	)
	// 	or (type = 'organization' and organization_id='orgid')
	subQueryAloowedAll := r.db.Table("policy_template_permitted_organizations").Select("policy_template_id")
	subQueryMatchId := r.db.Table("policy_template_permitted_organizations").Select("policy_template_id").
		Where("organization_id = ?", organizationId)

	_, res := pg.Fetch(r.db.WithContext(ctx).
		Preload("SupportedVersions", func(db *gorm.DB) *gorm.DB {
			// 최신 버전만
			return db.Order("policy_template_supported_versions.version DESC")
		}).
		Preload("Creator").Preload("Updator"). // organization을 기준으로 조회할 때에는 PermittedOrganizations는 로딩하지 않아도 됨
		Model(&model.PolicyTemplate{}).
		Where(
			// tks 템플릿인 경우
			r.db.Where("type = ?", "tks").
				Where(
					// permitted_organizations이 비어있거나
					r.db.Where("id not in (?)", subQueryAloowedAll).
						Or("id in (?)", subQueryMatchId),
					// permitted_organization에 매칭되는 템플릿 아이디가 있거나
				),
		).
		Or(
			// organization 타입 템플릿이면서 organization_id가 매칭
			r.db.Where("type = ?", "organization").
				Where("organization_id = ?", organizationId),
		),
		&policyTemplates)

	if res.Error != nil {
		return nil, res.Error
	}

	return policyTemplates, nil
}

func (r *PolicyTemplateRepository) CountTksTemplateByOrganization(ctx context.Context, organizationId string) (count int64, err error) {
	subQueryAloowedAll := r.db.Table("policy_template_permitted_organizations").Select("policy_template_id")
	subQueryMatchId := r.db.Table("policy_template_permitted_organizations").Select("policy_template_id").
		Where("organization_id = ?", organizationId)

	err = r.db.WithContext(ctx).
		Model(&model.PolicyTemplate{}).
		Where(
			// tks 템플릿인 경우
			r.db.Where("type = ?", "tks").
				Where(
					// permitted_organizations이 비어있거나
					r.db.Where("id not in (?)", subQueryAloowedAll).
						Or("id in (?)", subQueryMatchId),
					// permitted_organization에 매칭되는 템플릿 아이디가 있거나
				),
		).Count(&count).Error

	return
}

func (r *PolicyTemplateRepository) CountOrganizationTemplate(ctx context.Context, organizationId string) (count int64, err error) {
	err = r.db.WithContext(ctx).
		Model(&model.PolicyTemplate{}).
		Where("type = ?", "organization").
		Where("organization_id = ?", organizationId).
		Count(&count).Error

	return
}

func (r *PolicyTemplateRepository) CountPolicyFromOrganizationTemplate(ctx context.Context, organizationId string) (count int64, err error) {
	subQuery := r.db.Table("policy_templates").Select("id").
		Where("type = ?", "organization").
		Where("organization_id = ?", organizationId)

	err = r.db.WithContext(ctx).
		Model(&model.Policy{}).
		Where(
			r.db.Where("template_id in (?)", subQuery),
		).
		Count(&count).Error

	return
}

func (r *PolicyTemplateRepository) ExistsBy(ctx context.Context, key string, value interface{}) (exists bool, err error) {
	query := fmt.Sprintf("%s = ?", key)

	var policyTemplate model.PolicyTemplate
	res := r.db.WithContext(ctx).Where(query, value).
		First(&policyTemplate)

	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			log.Infof(ctx, "Not found policyTemplate %s='%v'", key, value)
			return false, nil
		} else {
			log.Error(ctx, res.Error)
			return false, res.Error
		}
	}

	return true, nil
}

func (r *PolicyTemplateRepository) ExistByName(ctx context.Context, policyTemplateName string) (exist bool, err error) {
	return r.ExistsBy(ctx, "template_name", policyTemplateName)
}

func (r *PolicyTemplateRepository) ExistByKind(ctx context.Context, policyTemplateKind string) (exist bool, err error) {
	return r.ExistsBy(ctx, "kind", policyTemplateKind)
}

func (r *PolicyTemplateRepository) ExistsByInOrganization(ctx context.Context, organizationId string, key string, value interface{}) (exists bool, err error) {

	var policyTemplate model.PolicyTemplate
	// query := fmt.Sprintf("%s = ? and (type = 'tks' or organization_id = ?)", key)
	// res := r.db.WithContext(ctx).Where(query, value, organizationId).
	// 	First(&policyTemplate)

	res := r.db.WithContext(ctx).Where(fmt.Sprintf("%s = ?", key), value).
		Where(
			r.db.Where("type = ?", "tks").Or("organization_id = ?", organizationId),
		).First(&policyTemplate)

	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			log.Infof(ctx, "Not found policyTemplate %s='%v'", key, value)
			return false, nil
		} else {
			log.Error(ctx, res.Error)
			return false, res.Error
		}
	}

	return true, nil
}

func (r *PolicyTemplateRepository) ExistByNameInOrganization(ctx context.Context, organizationId string, policyTemplateName string) (exist bool, err error) {
	return r.ExistsByInOrganization(ctx, organizationId, "template_name", policyTemplateName)
}

func (r *PolicyTemplateRepository) ExistByKindInOrganization(ctx context.Context, organizationId string, policyTemplateKind string) (exist bool, err error) {
	return r.ExistsByInOrganization(ctx, organizationId, "kind", policyTemplateKind)
}

func (r *PolicyTemplateRepository) ExistByID(ctx context.Context, policyTemplateId uuid.UUID) (exist bool, err error) {
	return r.ExistsBy(ctx, "id", policyTemplateId)
}

func (r *PolicyTemplateRepository) GetBy(ctx context.Context, key string, value interface{}) (out *model.PolicyTemplate, err error) {
	query := fmt.Sprintf("%s = ?", key)

	var policyTemplate model.PolicyTemplate
	// res := r.db.WithContext(ctx).Preload(clause.Associations).Where(query, value).
	// 	First(&policyTemplate)
	res := r.db.WithContext(ctx).
		Preload("SupportedVersions", func(db *gorm.DB) *gorm.DB {
			// 최신 버전만
			return db.Order("policy_template_supported_versions.version DESC").Limit(1)
		}).
		Preload("PermittedOrganizations").Preload("Creator").Preload("Updator").
		Where(query, value).
		First(&policyTemplate)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			log.Info(ctx, "Not found policyTemplate id")
			return nil, nil
		} else {
			log.Error(ctx, res.Error)
			return nil, res.Error
		}
	}

	if len(policyTemplate.SupportedVersions) == 0 {
		log.Info(ctx, "Not found policyTemplate version")
		return nil, nil
	}

	return &policyTemplate, nil
}

func (r *PolicyTemplateRepository) GetByID(ctx context.Context, policyTemplateId uuid.UUID) (out *model.PolicyTemplate, err error) {
	return r.GetBy(ctx, "id", policyTemplateId)
}

func (r *PolicyTemplateRepository) GetByName(ctx context.Context, policyTemplateName string) (out *model.PolicyTemplate, err error) {
	return r.GetBy(ctx, "name", policyTemplateName)
}

func (r *PolicyTemplateRepository) GetByKind(ctx context.Context, policyTemplateKind string) (out *model.PolicyTemplate, err error) {
	return r.GetBy(ctx, "kind", policyTemplateKind)
}

func (r *PolicyTemplateRepository) Delete(ctx context.Context, policyTemplateId uuid.UUID) (err error) {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("policy_template_id = ?", policyTemplateId).Delete(&model.PolicyTemplateSupportedVersion{}).Error; err != nil {
			return err
		}

		if err := tx.Model(&model.PolicyTemplate{ID: policyTemplateId}).Association("PermittedOrganizations").Clear(); err != nil {
			return err
		}

		if err := tx.Where("id = ?", policyTemplateId).Delete(&model.PolicyTemplate{}).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *PolicyTemplateRepository) ListPolicyTemplateVersions(ctx context.Context, policyTemplateId uuid.UUID) (policyTemplateVersionsReponse *domain.ListPolicyTemplateVersionsResponse, err error) {
	var supportedVersions []model.PolicyTemplateSupportedVersion
	res := r.db.WithContext(ctx).Where("policy_template_id = ?", policyTemplateId).Find(&supportedVersions)

	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			log.Info(ctx, "Not found policyTemplate Id")
			return nil, nil
		} else {
			log.Error(ctx, res.Error)
			return nil, res.Error
		}
	}

	versions := make([]string, len(supportedVersions))

	for i, supportedVersion := range supportedVersions {
		versions[i] = supportedVersion.Version
	}

	result := &domain.ListPolicyTemplateVersionsResponse{
		Versions: versions,
	}

	return result, nil
}

func (r *PolicyTemplateRepository) GetPolicyTemplateVersion(ctx context.Context, policyTemplateId uuid.UUID, version string) (policyTemplateVersionsReponse *model.PolicyTemplate, err error) {
	var policyTemplate model.PolicyTemplate

	res := r.db.WithContext(ctx).Preload("SupportedVersions", "version=?", version).
		Preload("PermittedOrganizations").Preload("Creator").Preload("Updator").
		Where("id = ?", policyTemplateId).
		First(&policyTemplate)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			log.Info(ctx, "Not found policyTemplate id")
			return nil, nil
		} else {
			log.Error(ctx, res.Error)
			return nil, res.Error
		}
	}

	if len(policyTemplate.SupportedVersions) == 0 {
		log.Info(ctx, "Not found policyTemplate version")
		return nil, nil
	}

	return &policyTemplate, nil
}

func (r *PolicyTemplateRepository) DeletePolicyTemplateVersion(ctx context.Context, policyTemplateId uuid.UUID, version string) (err error) {
	// TODO: Operator에 현재 버전 사용중인 정책이 있는지 체크 필요

	var policyTemplateVersion model.PolicyTemplateSupportedVersion

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		res := r.db.WithContext(ctx).Model(&policyTemplateVersion).Where("policy_template_id = ?", policyTemplateId).Count(&count)
		if res.Error != nil {
			return res.Error
		}

		// 마지막으로 존재하는 버전은 삭제할 수 없으므로 해당 id의 버전 카운트가 2 이상이어야 함
		if count < 2 {
			return errors.New("Unable to delete last single version")
		}

		// relaton을 unscoped로 삭제하지 않으면 동일한 키로 다시 생성할 때 키가 같은 레코드가 deleted 상태로 존재하므로 unscoped delete
		res = r.db.WithContext(ctx).Unscoped().Where("policy_template_id = ?", policyTemplateId).Where("version = ?", version).
			Delete(&policyTemplateVersion)
		if res.Error != nil {
			if errors.Is(res.Error, gorm.ErrRecordNotFound) {
				log.Info(ctx, "Not found policyTemplate version")
				return nil
			} else {
				log.Error(ctx, res.Error)
				return res.Error
			}
		}

		return nil
	})
}

func (r *PolicyTemplateRepository) CreatePolicyTemplateVersion(ctx context.Context, policyTemplateId uuid.UUID, newVersion string,
	schema []*domain.ParameterDef, rego string, libs []string, syncKinds *[]string, syncJson *string) (version string, err error) {
	var policyTemplateVersion model.PolicyTemplateSupportedVersion
	res := r.db.WithContext(ctx).Limit(1).
		Where("policy_template_id = ?", policyTemplateId).Where("version = ?", version).
		First(&policyTemplateVersion)

	if res.Error == nil {
		err = errors.Errorf("Version %s already exists for the policyTemplate", newVersion)

		log.Error(ctx, res.Error)

		return "", err
	}

	if !errors.Is(res.Error, gorm.ErrRecordNotFound) {
		log.Error(ctx, res.Error)
		return "", res.Error
	}

	libsString := ""
	if len(libs) > 0 {
		libsString = strings.Join(libs, model.FILE_DELIMETER)
	}

	jsonBytes, err := json.Marshal(schema)

	if err != nil {
		parseErr := errors.Errorf("Unable to parse parameter schema: %v", err)

		log.Error(ctx, parseErr)

		return "", parseErr
	}

	var syncKindsString *string = nil

	if syncKinds != nil {
		syncJsonBytes, err := json.Marshal(syncKinds)

		if err != nil {
			parseErr := errors.Errorf("Unable to parse parameter schema: %v", err)

			log.Error(ctx, parseErr)

			return "", parseErr
		}

		syncStr := string(syncJsonBytes)

		syncKindsString = &syncStr
	}

	newPolicyTemplateVersion := &model.PolicyTemplateSupportedVersion{
		PolicyTemplateId: policyTemplateId,
		Version:          newVersion,
		Rego:             rego,
		Libs:             libsString,
		ParameterSchema:  string(jsonBytes),
		SyncJson:         syncJson,
		SyncKinds:        syncKindsString,
	}

	if err := r.db.WithContext(ctx).Create(newPolicyTemplateVersion).Error; err != nil {
		return "", err
	}

	return newVersion, nil
}

func (r *PolicyTemplateRepository) GetLatestTemplateVersion(ctx context.Context, policyTemplateId uuid.UUID) (version string, err error) {
	var policyTemplateVersion model.PolicyTemplateSupportedVersion

	err = r.db.WithContext(ctx).
		Where("policy_template_id = ?", policyTemplateId).
		Order("created_at desc").First(&policyTemplateVersion).Error

	if err != nil {
		return
	}

	return policyTemplateVersion.Version, nil
}

func (r *PolicyTemplateRepository) GetPolicyTemplateByOrganizationIdOrTKS(ctx context.Context, organizationId string) (out []model.PolicyTemplate, err error) {
	res := r.db.WithContext(ctx).
		Select("id", "type", "template_name").
		Where("organization_id = ? or type = ?", organizationId, "tks").
		Find(&out)

	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			log.Info(ctx, "Cannot find policytemplate")
			return nil, nil
		} else {
			log.Error(ctx, res.Error)
			return nil, res.Error
		}
	}
	return out, nil
}
