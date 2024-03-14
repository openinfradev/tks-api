package repository

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/internal/serializer"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/log"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IPolicyTemplateRepository interface {
	Create(dto domain.PolicyTemplate) (policyTemplateId uuid.UUID, err error)
	Update(dto domain.UpdatePolicyTemplateUpdate) (err error)
	Fetch(pg *pagination.Pagination) (out []domain.PolicyTemplate, err error)
	GetByName(policyTemplateName string) (out *domain.PolicyTemplate, err error)
	GetByKind(policyTemplateKind string) (out *domain.PolicyTemplate, err error)
	GetByID(policyTemplateId uuid.UUID) (out *domain.PolicyTemplate, err error)
	Delete(policyTemplateId uuid.UUID) (err error)
	ExistByName(policyTemplateName string) (exist bool, err error)
	ExistByKind(policyTemplateKind string) (exist bool, err error)
	ExistByID(policyTemplateId uuid.UUID) (exist bool, err error)
	ListPolicyTemplateVersions(policyTemplateId uuid.UUID) (policyTemplateVersionsReponse *domain.ListPolicyTemplateVersionsResponse, err error)
	GetPolicyTemplateVersion(policyTemplateId uuid.UUID, version string) (policyTemplateVersionsReponse *domain.PolicyTemplate, err error)
	DeletePolicyTemplateVersion(policyTemplateId uuid.UUID, version string) (err error)
	CreatePolicyTemplateVersion(policyTemplateId uuid.UUID, newVersion string, schema []domain.ParameterDef, rego string, libs []string) (version string, err error)
}

type PolicyTemplateRepository struct {
	db *gorm.DB
}

func NewPolicyTemplateRepository(db *gorm.DB) IPolicyTemplateRepository {
	return &PolicyTemplateRepository{
		db: db,
	}
}

type PolicyTemplateSupportedVersion struct {
	gorm.Model

	PolicyTemplateId uuid.UUID `gorm:"index:template_version,unique"`
	Version          string    `gorm:"index:template_version,unique"`

	ParameterSchema string `gorm:"type:text"`
	Rego            string `gorm:"type:text"`
	Libs            string `gorm:"type:text"`
}

type PolicyTemplate struct {
	gorm.Model

	ID                     uuid.UUID `gorm:"primarykey;type:varchar(36);not null"`
	Type                   string    // Org or Tks
	Name                   string
	Version                string
	SupportedVersions      []PolicyTemplateSupportedVersion `gorm:"foreignKey:PolicyTemplateId"`
	Description            string
	Kind                   string
	Deprecated             bool
	Mandatory              bool // Tks 인 경우에는 무시
	Severity               string
	PermittedOrganizations []domain.Organization `gorm:"many2many:policy_template_permitted_organiations;"`
	CreatorId              *uuid.UUID            `gorm:"type:uuid"`
	Creator                domain.User           `gorm:"foreignKey:CreatorId"`
	UpdatorId              *uuid.UUID            `gorm:"type:uuid"`
	Updator                domain.User           `gorm:"foreignKey:UpdatorId"`
}

func (c *PolicyTemplate) BeforeCreate(tx *gorm.DB) (err error) {
	c.ID = uuid.New()
	return nil
}

func (r *PolicyTemplateRepository) Create(dto domain.PolicyTemplate) (policyTemplateId uuid.UUID, err error) {
	jsonByte, err := json.Marshal(dto.ParametersSchema)

	if err != nil {
		return uuid.Nil, err
	}

	policyTemplate := PolicyTemplate{
		Type:    "tks",
		Name:    dto.TemplateName,
		Version: "v1.0.0",
		SupportedVersions: []PolicyTemplateSupportedVersion{
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

	err = r.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Create(&policyTemplate).Error

		if err != nil {
			return err
		}

		if dto.PermittedOrganizationIds != nil {
			permittedOrganizations := make([]domain.Organization, len(dto.PermittedOrganizationIds))
			for i, permittedOrganizationId := range dto.PermittedOrganizationIds {
				permittedOrganizations[i] = domain.Organization{ID: permittedOrganizationId}
			}

			err = tx.Model(&policyTemplate).Association("PermittedOrganizations").Replace(permittedOrganizations)

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

func (r *PolicyTemplateRepository) Update(dto domain.UpdatePolicyTemplateUpdate) (err error) {
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

	var policyTemplate PolicyTemplate
	policyTemplate.ID = dto.ID

	return r.db.Transaction(func(tx *gorm.DB) error {
		if dto.PermittedOrganizationIds != nil {
			permittedOrganizations := make([]domain.Organization, len(*dto.PermittedOrganizationIds))
			for i, permittedOrganizationId := range *dto.PermittedOrganizationIds {
				permittedOrganizations[i] = domain.Organization{ID: permittedOrganizationId}
			}

			err = r.db.Model(&policyTemplate).Limit(1).
				Association("PermittedOrganizations").Replace(permittedOrganizations)

			if err != nil {
				return err
			}
		}

		if len(updateMap) > 0 {
			err = r.db.Model(&policyTemplate).Limit(1).
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

func (r *PolicyTemplateRepository) Fetch(pg *pagination.Pagination) (out []domain.PolicyTemplate, err error) {
	var policyTemplates []PolicyTemplate
	if pg == nil {
		pg = pagination.NewPagination(nil)
	}

	_, res := pg.Fetch(r.db.Preload(clause.Associations).Model(&PolicyTemplate{}).
		Where("type = 'tks'"), &policyTemplates)
	if res.Error != nil {
		return nil, res.Error
	}

	for _, policyTemplate := range policyTemplates {
		var policyTemplateVersion PolicyTemplateSupportedVersion
		res = r.db.
			Where("policy_template_id = ? and version = ?", policyTemplate.ID, policyTemplate.Version).
			First(&policyTemplateVersion)

		if res.Error != nil {
			if errors.Is(res.Error, gorm.ErrRecordNotFound) {
				log.Info("Not found policyTemplate version")
			} else {
				log.Error(res.Error)
			}
		}

		outPolicyTemplate := r.reflectPolicyTemplate(policyTemplate, policyTemplateVersion)
		out = append(out, outPolicyTemplate)
	}
	return out, nil
}

func (r *PolicyTemplateRepository) reflectPolicyTemplate(policyTemplate PolicyTemplate, policyTemplateVersion PolicyTemplateSupportedVersion) (out domain.PolicyTemplate) {
	if err := serializer.Map(policyTemplate.Model, &out); err != nil {
		log.Error(err)
	}
	if err := serializer.Map(policyTemplate, &out); err != nil {
		log.Error(err)
	}
	out.TemplateName = policyTemplate.Name
	out.ID = domain.PolicyTemplateId(policyTemplate.ID.String())

	var schemas []domain.ParameterDef

	if len(policyTemplateVersion.ParameterSchema) > 0 {
		if err := json.Unmarshal([]byte(policyTemplateVersion.ParameterSchema), &schemas); err != nil {
			log.Error(err)
		} else {
			out.ParametersSchema = schemas
		}
	}

	out.PermittedOrganizations = make([]domain.PermittedOrganization, len(policyTemplate.PermittedOrganizations))
	for i, org := range policyTemplate.PermittedOrganizations {
		out.PermittedOrganizations[i] = domain.PermittedOrganization{
			OrganizationId:   org.ID,
			OrganizationName: org.Name,
			Permitted:        true,
		}
	}

	out.Rego = policyTemplateVersion.Rego
	out.Libs = strings.Split(policyTemplateVersion.Libs, "---\n")

	return
}

func (r *PolicyTemplateRepository) ExistsBy(key string, value interface{}) (exists bool, err error) {
	query := fmt.Sprintf("%s = ?", key)

	var policyTemplate PolicyTemplate
	res := r.db.Where(query, value).
		First(&policyTemplate)

	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			log.Infof("Not found policyTemplate %s='%v'", key, value)
			return false, nil
		} else {
			log.Error(res.Error)
			return false, res.Error
		}
	}

	return true, nil
}

func (r *PolicyTemplateRepository) ExistByName(policyTemplateName string) (exist bool, err error) {
	return r.ExistsBy("name", policyTemplateName)
}

func (r *PolicyTemplateRepository) ExistByKind(policyTemplateKind string) (exist bool, err error) {
	return r.ExistsBy("kind", policyTemplateKind)
}

func (r *PolicyTemplateRepository) ExistByID(policyTemplateId uuid.UUID) (exist bool, err error) {
	return r.ExistsBy("id", policyTemplateId)
}

func (r *PolicyTemplateRepository) GetBy(key string, value interface{}) (out *domain.PolicyTemplate, err error) {
	query := fmt.Sprintf("%s = ?", key)

	var policyTemplate PolicyTemplate
	res := r.db.Preload(clause.Associations).Where(query, value).
		First(&policyTemplate)

	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			log.Infof("Not found policyTemplate %s='%v'", key, value)
			return nil, nil
		} else {
			log.Error(res.Error)
			return nil, res.Error
		}
	}

	var policyTemplateVersion PolicyTemplateSupportedVersion
	res = r.db.Limit(1).
		Where("policy_template_id = ? and version = ?", policyTemplate.ID, policyTemplate.Version).
		First(&policyTemplateVersion)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			log.Info("Not found policyTemplate version")
			return nil, nil
		} else {
			log.Error(res.Error)
			return nil, res.Error
		}
	}

	fmt.Printf("BBBB %+v\n", policyTemplate.PermittedOrganizations)

	result := r.reflectPolicyTemplate(policyTemplate, policyTemplateVersion)
	fmt.Printf("2222BBBB %+v\n", result.PermittedOrganizations)

	return &result, nil
}

func (r *PolicyTemplateRepository) GetByID(policyTemplateId uuid.UUID) (out *domain.PolicyTemplate, err error) {
	return r.GetBy("id", policyTemplateId)

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

func (r *PolicyTemplateRepository) GetByName(policyTemplateName string) (out *domain.PolicyTemplate, err error) {
	return r.GetBy("name", policyTemplateName)

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

func (r *PolicyTemplateRepository) GetByKind(policyTemplateKind string) (out *domain.PolicyTemplate, err error) {
	return r.GetBy("kind", policyTemplateKind)
}

func (r *PolicyTemplateRepository) Delete(policyTemplateId uuid.UUID) (err error) {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("policy_template_id = ?", policyTemplateId).Delete(&PolicyTemplateSupportedVersion{}).Error; err != nil {
			return err
		}

		if err := tx.Model(&PolicyTemplate{ID: policyTemplateId}).Association("PermittedOrganizations").Clear(); err != nil {
			return err
		}

		if err := tx.Where("id = ?", policyTemplateId).Delete(&PolicyTemplate{}).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *PolicyTemplateRepository) ListPolicyTemplateVersions(policyTemplateId uuid.UUID) (policyTemplateVersionsReponse *domain.ListPolicyTemplateVersionsResponse, err error) {
	var supportedVersions []PolicyTemplateSupportedVersion
	res := r.db.Where("policy_template_id = ?", policyTemplateId).Find(&supportedVersions)

	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			log.Info("Not found policyTemplate kind")
			return nil, nil
		} else {
			log.Error(res.Error)
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

func (r *PolicyTemplateRepository) GetPolicyTemplateVersion(policyTemplateId uuid.UUID, version string) (policyTemplateVersionsReponse *domain.PolicyTemplate, err error) {
	var policyTemplateVersion PolicyTemplateSupportedVersion
	res := r.db.
		Where("policy_template_id = ? and version = ?", policyTemplateId, version).
		First(&policyTemplateVersion)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			log.Info("Not found policyTemplate version")
			return nil, nil
		} else {
			log.Error(res.Error)
			return nil, res.Error
		}
	}

	var policyTemplate PolicyTemplate
	res = r.db.
		Where("id = ?", policyTemplateId).
		First(&policyTemplate)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			log.Info("Not found policyTemplate id")
			return nil, nil
		} else {
			log.Error(res.Error)
			return nil, res.Error
		}
	}

	result := r.reflectPolicyTemplate(policyTemplate, policyTemplateVersion)

	return &result, nil
}

func (r *PolicyTemplateRepository) DeletePolicyTemplateVersion(policyTemplateId uuid.UUID, version string) (err error) {
	var policyTemplate PolicyTemplate
	res := r.db.Select("version").First(&policyTemplate)

	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			log.Info("Not found policyTemplate id")
			return nil
		} else {
			log.Error(res.Error)
			return res.Error
		}
	}

	// 현재 버전이 템플릿에서 최신 버전으로 사용 중이면 삭제 금지
	if policyTemplate.Version == version {
		return fmt.Errorf("version '%s' is currently in use", version)
	}

	// TODO: Operator에 현재 버전 사용중인 정책이 있는지 체크 필요

	var policyTemplateVersion PolicyTemplateSupportedVersion
	res = r.db.Where("policy_template_id = ? and version = ?", policyTemplateId, version).
		Delete(&policyTemplateVersion)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			log.Info("Not found policyTemplate version")
			return nil
		} else {
			log.Error(res.Error)
			return res.Error
		}
	}

	return nil
}

func (r *PolicyTemplateRepository) CreatePolicyTemplateVersion(policyTemplateId uuid.UUID, newVersion string, schema []domain.ParameterDef, rego string, libs []string) (version string, err error) {
	var policyTemplateVersion PolicyTemplateSupportedVersion
	res := r.db.Limit(1).
		Where("policy_template_id = ? and version = ?", policyTemplateId, version).
		First(&policyTemplateVersion)

	if res.Error == nil {
		err = errors.Errorf("Version %s already exists for the policyTemplate", newVersion)

		log.Error(res.Error)

		return "", err
	}

	if !errors.Is(res.Error, gorm.ErrRecordNotFound) {
		log.Error(res.Error)
		return "", res.Error
	}

	libsString := ""
	if len(libs) > 0 {
		libsString = strings.Join(libs, "---\n")
	}

	jsonBytes, err := json.Marshal(schema)

	if err != nil {
		parseErr := errors.Errorf("Unable to parse parameter schema: %v", err)

		log.Error(parseErr)

		return "", parseErr
	}

	newPolicyTemplateVersion := &PolicyTemplateSupportedVersion{
		PolicyTemplateId: policyTemplateId,
		Version:          newVersion,
		Rego:             rego,
		Libs:             libsString,
		ParameterSchema:  string(jsonBytes),
	}

	err = r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(newPolicyTemplateVersion).Error; err != nil {
			return err
		}

		if err := tx.Model(&PolicyTemplate{}).Where("id = ?", policyTemplateId).Update("version", newVersion).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	return newVersion, nil
}
