package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/internal/serializer"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/log"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IPolicyTemplateRepository interface {
	Create(ctx context.Context, dto model.PolicyTemplate) (policyTemplateId uuid.UUID, err error)
	Update(ctx context.Context, dto domain.UpdatePolicyTemplateUpdate) (err error)
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
	jsonByte, err := json.Marshal(dto.ParametersSchema)

	if err != nil {
		return uuid.Nil, err
	}

	policyTemplate := model.PolicyTemplate{
		Type:    "tks",
		Name:    dto.TemplateName,
		Version: "v1.0.0",
		SupportedVersions: []model.PolicyTemplateSupportedVersion{
			{
				Version:         "v1.0.0",
				ParameterSchema: string(jsonByte),
				Rego:            dto.Rego,
			},
		},
		Description: dto.Description,
		Kind:        dto.Kind,
		Deprecated:  false,
		Mandatory:   false, // Tks 인 경우에는 무시
		Severity:    dto.Severity,
	}

	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.WithContext(ctx).Create(&policyTemplate).Error

		if err != nil {
			return err
		}

		if dto.PermittedOrganizationIds != nil {
			permittedOrganizations := make([]model.Organization, len(dto.PermittedOrganizationIds))
			for i, permittedOrganizationId := range dto.PermittedOrganizationIds {
				permittedOrganizations[i] = model.Organization{ID: permittedOrganizationId}
			}

			err = tx.WithContext(ctx).Model(&policyTemplate).Association("PermittedOrganizations").Replace(permittedOrganizations)

			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return uuid.Nil, err
	}

	return policyTemplate.ID, nil
}

func (r *PolicyTemplateRepository) Update(ctx context.Context, dto domain.UpdatePolicyTemplateUpdate) (err error) {
	updateMap := make(map[string]interface{})

	updateMap["updator_id"] = dto.UpdatorId

	if dto.Description != nil {
		updateMap["description"] = dto.Description
	}

	if dto.Deprecated != nil {
		updateMap["deprecated"] = dto.Deprecated
	}

	if dto.Severity != nil {
		updateMap["severity"] = dto.Severity
	}

	if dto.TemplateName != nil {
		updateMap["name"] = dto.TemplateName
	}

	fmt.Printf("--updateMap=%+v\n--", updateMap)

	var policyTemplate model.PolicyTemplate
	policyTemplate.ID = dto.ID

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if dto.PermittedOrganizationIds != nil {
			permittedOrganizations := make([]model.Organization, len(*dto.PermittedOrganizationIds))
			for i, permittedOrganizationId := range *dto.PermittedOrganizationIds {
				permittedOrganizations[i] = model.Organization{ID: permittedOrganizationId}
			}

			err = r.db.WithContext(ctx).Model(&policyTemplate).Limit(1).
				Association("PermittedOrganizations").Replace(permittedOrganizations)

			if err != nil {
				return err
			}
		}

		if len(updateMap) > 0 {
			err = r.db.WithContext(ctx).Model(&policyTemplate).Limit(1).
				Where("id = ? and type = 'tks'", dto.ID).
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

	_, res := pg.Fetch(r.db.WithContext(ctx).Preload(clause.Associations).Model(&model.PolicyTemplate{}).
		Where("type = 'tks'"), &policyTemplates)
	if res.Error != nil {
		return nil, res.Error
	}

	for _, policyTemplate := range policyTemplates {
		var policyTemplateVersion model.PolicyTemplateSupportedVersion
		res = r.db.WithContext(ctx).
			Where("policy_template_id = ? and version = ?", policyTemplate.ID, policyTemplate.Version).
			First(&policyTemplateVersion)

		if res.Error != nil {
			if errors.Is(res.Error, gorm.ErrRecordNotFound) {
				log.Info(ctx, "Not found policyTemplate version")
			} else {
				log.Error(ctx, res.Error)
			}
		}

		outPolicyTemplate := r.reflectPolicyTemplate(ctx, policyTemplate, policyTemplateVersion)
		out = append(out, outPolicyTemplate)
	}
	return out, nil
}

func (r *PolicyTemplateRepository) reflectPolicyTemplate(ctx context.Context, policyTemplate model.PolicyTemplate, policyTemplateVersion model.PolicyTemplateSupportedVersion) (out model.PolicyTemplate) {
	if err := serializer.Map(ctx, policyTemplate.Model, &out); err != nil {
		log.Error(ctx, err)
	}
	if err := serializer.Map(ctx, policyTemplate, &out); err != nil {
		log.Error(ctx, err)
	}
	out.TemplateName = policyTemplate.Name
	out.ID = policyTemplate.ID

	var schemas []domain.ParameterDef

	if len(policyTemplateVersion.ParameterSchema) > 0 {
		if err := json.Unmarshal([]byte(policyTemplateVersion.ParameterSchema), &schemas); err != nil {
			log.Error(ctx, err)
		} else {
			out.ParametersSchema = schemas
		}
	}

	// ktkfree : 이 부분은 재 구현 부탁 드립니다.
	// PermittedOrganizations 필드가 model 과 response 를 위한 객체의 형이 다르네요.
	// 아울러 reflect 는 repository 가 아닌 usecase 에 표현되는게 더 좋겠습니다.

	/*
		out.PermittedOrganizations = make([]domain.PermittedOrganization, len(policyTemplate.PermittedOrganizations))
		for i, org := range policyTemplate.PermittedOrganizations {
			out.PermittedOrganizations[i] = domain.PermittedOrganization{
				OrganizationId:   org.ID,
				OrganizationName: org.Name,
				Permitted:        true,
			}
		}
	*/

	out.Rego = policyTemplateVersion.Rego
	out.Libs = strings.Split(policyTemplateVersion.Libs, "---\n")

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
	return r.ExistsBy(ctx, "name", policyTemplateName)
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
	res := r.db.WithContext(ctx).Preload(clause.Associations).Where(query, value).
		First(&policyTemplate)

	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			log.Infof(ctx, "Not found policyTemplate %s='%v'", key, value)
			return nil, nil
		} else {
			log.Error(ctx, res.Error)
			return nil, res.Error
		}
	}

	var policyTemplateVersion model.PolicyTemplateSupportedVersion
	res = r.db.WithContext(ctx).Limit(1).
		Where("policy_template_id = ? and version = ?", policyTemplate.ID, policyTemplate.Version).
		First(&policyTemplateVersion)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			log.Info(ctx, "Not found policyTemplate version")
			return nil, nil
		} else {
			log.Error(ctx, res.Error)
			return nil, res.Error
		}
	}

	fmt.Printf("BBBB %+v\n", policyTemplate.PermittedOrganizations)

	result := r.reflectPolicyTemplate(ctx, policyTemplate, policyTemplateVersion)
	fmt.Printf("2222BBBB %+v\n", result.PermittedOrganizations)

	return &result, nil
}

func (r *PolicyTemplateRepository) GetByID(ctx context.Context, policyTemplateId uuid.UUID) (out *model.PolicyTemplate, err error) {
	return r.GetBy(ctx, "id", policyTemplateId)

	// var policyTemplate PolicyTemplate
	// res := r.db.Preload(clause.Associations).Where("id = ?", policyTemplateId).
	// 	First(&policyTemplate)

	// if res.Error != nil {
	// 	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
	// 		log.Info("Not found policyTemplate id")
	// 		return nil, nil
	// 	} else {
	// 		log.Error(res.Error)
	// 		return nil, res.Error
	// 	}
	// }

	// result := r.reflect(policyTemplate)

	// return &result, nil
}

func (r *PolicyTemplateRepository) GetByName(ctx context.Context, policyTemplateName string) (out *model.PolicyTemplate, err error) {
	return r.GetBy(ctx, "name", policyTemplateName)

	// var policyTemplate PolicyTemplate
	// res := r.db.Limit(1).
	// 	Where("name = ?", policyTemplateName).
	// 	First(&policyTemplate)
	// if res.Error != nil {
	// 	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
	// 		log.Info("Not found policyTemplate name")
	// 		return nil, nil
	// 	} else {
	// 		log.Error(res.Error)
	// 		return nil, res.Error
	// 	}
	// }

	// result := r.reflect(policyTemplate)

	// return &result, nil
}

func (r *PolicyTemplateRepository) GetByKind(ctx context.Context, policyTemplateKind string) (out *model.PolicyTemplate, err error) {
	return r.GetBy(ctx, "kind", policyTemplateKind)
}

func (r *PolicyTemplateRepository) Delete(ctx context.Context, policyTemplateId uuid.UUID) (err error) {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Where("policy_template_id = ?", policyTemplateId).Delete(&model.PolicyTemplateSupportedVersion{}).Error; err != nil {
			return err
		}

		if err := tx.WithContext(ctx).Model(&model.PolicyTemplate{ID: policyTemplateId}).Association("PermittedOrganizations").Clear(); err != nil {
			return err
		}

		if err := tx.WithContext(ctx).Where("id = ?", policyTemplateId).Delete(&model.PolicyTemplate{}).Error; err != nil {
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
			log.Info(ctx, "Not found policyTemplate kind")
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
	var policyTemplateVersion model.PolicyTemplateSupportedVersion
	res := r.db.WithContext(ctx).
		Where("policy_template_id = ? and version = ?", policyTemplateId, version).
		First(&policyTemplateVersion)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			log.Info(ctx, "Not found policyTemplate version")
			return nil, nil
		} else {
			log.Error(ctx, res.Error)
			return nil, res.Error
		}
	}

	var policyTemplate model.PolicyTemplate
	res = r.db.WithContext(ctx).
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

	result := r.reflectPolicyTemplate(ctx, policyTemplate, policyTemplateVersion)

	return &result, nil
}

func (r *PolicyTemplateRepository) DeletePolicyTemplateVersion(ctx context.Context, policyTemplateId uuid.UUID, version string) (err error) {
	var policyTemplate model.PolicyTemplate
	res := r.db.WithContext(ctx).Select("version").First(&policyTemplate)

	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			log.Info(ctx, "Not found policyTemplate id")
			return nil
		} else {
			log.Error(ctx, res.Error)
			return res.Error
		}
	}

	// 현재 버전이 템플릿에서 최신 버전으로 사용 중이면 삭제 금지
	if policyTemplate.Version == version {
		return fmt.Errorf("version '%s' is currently in use", version)
	}

	// TODO: Operator에 현재 버전 사용중인 정책이 있는지 체크 필요

	var policyTemplateVersion model.PolicyTemplateSupportedVersion
	res = r.db.WithContext(ctx).Where("policy_template_id = ? and version = ?", policyTemplateId, version).
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
		libsString = strings.Join(libs, "---\n")
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

	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Create(newPolicyTemplateVersion).Error; err != nil {
			return err
		}

		if err := tx.WithContext(ctx).Model(&model.PolicyTemplate{}).Where("id = ?", policyTemplateId).Update("version", newVersion).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	return newVersion, nil
}
