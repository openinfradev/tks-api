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
	GetByName(ctx context.Context, policyTemplateName string) (out *model.PolicyTemplate, err error)
	GetByKind(ctx context.Context, policyTemplateKind string) (out *model.PolicyTemplate, err error)
	GetByID(ctx context.Context, policyTemplateId uuid.UUID) (out *model.PolicyTemplate, err error)
	Delete(ctx context.Context, policyTemplateId uuid.UUID) (err error)
	ExistByName(ctx context.Context, policyTemplateName string) (exist bool, err error)
	ExistByKind(ctx context.Context, policyTemplateKind string) (exist bool, err error)
	ExistByID(ctx context.Context, policyTemplateId uuid.UUID) (exist bool, err error)
	ListPolicyTemplateVersions(ctx context.Context, policyTemplateId uuid.UUID) (policyTemplateVersionsReponse *domain.ListPolicyTemplateVersionsResponse, err error)
	GetPolicyTemplateVersion(ctx context.Context, policyTemplateId uuid.UUID, version string) (policyTemplateVersionsReponse *model.PolicyTemplate, err error)
	DeletePolicyTemplateVersion(ctx context.Context, policyTemplateId uuid.UUID, version string) (err error)
	CreatePolicyTemplateVersion(ctx context.Context, policyTemplateId uuid.UUID, newVersion string, schema []domain.ParameterDef, rego string, libs []string) (version string, err error)
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
	err = r.db.WithContext(ctx).Create(&dto).Error

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
			err = r.db.WithContext(ctx).Model(&policyTemplate).Limit(1).
				Association("PermittedOrganizations").Replace(permittedOrganizations)

			if err != nil {
				return err
			}
		}

		if len(updateMap) > 0 {
			err = r.db.WithContext(ctx).Model(&policyTemplate).Limit(1).
				Where("id = ? and type = 'tks'", policyTemplateId).
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
		res = r.db.WithContext(ctx).Unscoped().Where("policy_template_id = ? and version = ?", policyTemplateId, version).
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

func (r *PolicyTemplateRepository) CreatePolicyTemplateVersion(ctx context.Context, policyTemplateId uuid.UUID, newVersion string, schema []domain.ParameterDef, rego string, libs []string) (version string, err error) {
	var policyTemplateVersion model.PolicyTemplateSupportedVersion
	res := r.db.WithContext(ctx).Limit(1).
		Where("policy_template_id = ? and version = ?", policyTemplateId, version).
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

	newPolicyTemplateVersion := &model.PolicyTemplateSupportedVersion{
		PolicyTemplateId: policyTemplateId,
		Version:          newVersion,
		Rego:             rego,
		Libs:             libsString,
		ParameterSchema:  string(jsonBytes),
	}

	if err := r.db.WithContext(ctx).Create(newPolicyTemplateVersion).Error; err != nil {
		return "", err
	}

	return newVersion, nil
}
