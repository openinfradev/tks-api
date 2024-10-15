package usecase

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal"
	"github.com/openinfradev/tks-api/internal/middleware/auth/request"
	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/kubernetes"
	"github.com/openinfradev/tks-api/pkg/log"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type IStackTemplateUsecase interface {
	Get(ctx context.Context, stackTemplateId uuid.UUID) (model.StackTemplate, error)
	Fetch(ctx context.Context, pg *pagination.Pagination) ([]model.StackTemplate, error)
	FetchWithOrganization(ctx context.Context, organizationId string, pg *pagination.Pagination) ([]model.StackTemplate, error)
	Create(ctx context.Context, dto model.StackTemplate) (stackTemplate uuid.UUID, err error)
	Update(ctx context.Context, dto model.StackTemplate) error
	Delete(ctx context.Context, stackTemplateId uuid.UUID) error
	UpdateOrganizations(ctx context.Context, dto model.StackTemplate) error
	GetByName(ctx context.Context, name string) (model.StackTemplate, error)
	AddOrganizationStackTemplates(ctx context.Context, organizationId string, stackTemplateIds []string) error
	RemoveOrganizationStackTemplates(ctx context.Context, organizationId string, stackTemplateIds []string) error
	GetTemplateIds(ctx context.Context) ([]string, error)
	GetCloudServices(ctx context.Context, organizationId string) ([]string, error)
}

type StackTemplateUsecase struct {
	repo             repository.IStackTemplateRepository
	organizationRepo repository.IOrganizationRepository
	clusterRepo      repository.IClusterRepository
}

func NewStackTemplateUsecase(r repository.Repository) IStackTemplateUsecase {
	return &StackTemplateUsecase{
		repo:             r.StackTemplate,
		organizationRepo: r.Organization,
		clusterRepo:      r.Cluster,
	}
}

func (u *StackTemplateUsecase) Create(ctx context.Context, dto model.StackTemplate) (stackTemplateId uuid.UUID, err error) {
	user, ok := request.UserFrom(ctx)
	if !ok {
		return uuid.Nil, httpErrors.NewBadRequestError(fmt.Errorf("Invalid token"), "", "")
	}
	userId := user.GetUserId()
	dto.CreatorId = &userId
	dto.UpdatorId = &userId

	if _, err = u.GetByName(ctx, dto.Name); err == nil {
		return uuid.Nil, httpErrors.NewBadRequestError(fmt.Errorf("duplicate stackTemplate name"), "ST_CREATE_ALREADY_EXISTED_NAME", "")
	}

	dto.Services = servicesFromIds(dto.ServiceIds)
	stackTemplateId, err = u.repo.Create(ctx, dto)
	if err != nil {
		return uuid.Nil, httpErrors.NewInternalServerError(err, "", "")
	}
	log.Info(ctx, "newly created StackTemplate ID:", stackTemplateId)

	dto.ID = stackTemplateId
	err = u.UpdateOrganizations(ctx, dto)
	if err != nil {
		return uuid.Nil, err
	}

	return stackTemplateId, nil
}

func (u *StackTemplateUsecase) Update(ctx context.Context, dto model.StackTemplate) error {
	_, err := u.repo.Get(ctx, dto.ID)
	if err != nil {
		return httpErrors.NewBadRequestError(err, "ST_NOT_EXISTED_STACK_TEMPLATE", "")
	}

	dto.Services = servicesFromIds(dto.ServiceIds)
	err = u.repo.Update(ctx, dto)
	if err != nil {
		return err
	}

	err = u.UpdateOrganizations(ctx, dto)
	if err != nil {
		return err
	}

	return nil
}

func (u *StackTemplateUsecase) Get(ctx context.Context, stackTemplateId uuid.UUID) (res model.StackTemplate, err error) {
	res, err = u.repo.Get(ctx, stackTemplateId)
	if err != nil {
		return res, err
	}
	return
}

func (u *StackTemplateUsecase) GetByName(ctx context.Context, name string) (out model.StackTemplate, err error) {
	out, err = u.repo.GetByName(ctx, name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return out, httpErrors.NewNotFoundError(err, "ST_FAILED_FETCH_STACK_TEMPLATE", "")
		}
		return out, err
	}

	return
}

func (u *StackTemplateUsecase) Fetch(ctx context.Context, pg *pagination.Pagination) (res []model.StackTemplate, err error) {
	res, err = u.repo.Fetch(ctx, pg)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (u *StackTemplateUsecase) FetchWithOrganization(ctx context.Context, organizationId string, pg *pagination.Pagination) (res []model.StackTemplate, err error) {
	res, err = u.repo.FetchWithOrganization(ctx, organizationId, pg)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (u *StackTemplateUsecase) Delete(ctx context.Context, stackTemplateId uuid.UUID) (err error) {
	stackTemplate, err := u.repo.Get(ctx, stackTemplateId)
	if err != nil {
		return err
	}

	user, ok := request.UserFrom(ctx)
	if !ok {
		return httpErrors.NewBadRequestError(fmt.Errorf("Invalid token"), "", "")
	}
	userId := user.GetUserId()
	stackTemplate.UpdatorId = &userId

	// check if used
	pg := pagination.NewPaginationWithFilter("stack_template_id", "", "$eq", []string{stackTemplateId.String()})
	res, err := u.clusterRepo.Fetch(ctx, pg)
	if err != nil {
		return err
	}
	if len(res) > 0 {
		return httpErrors.NewBadRequestError(fmt.Errorf("Failed to delete stackTemplate %s", stackTemplateId.String()), "ST_FAILED_DELETE_EXIST_CLUSTERS", "")
	}

	err = u.repo.Delete(ctx, stackTemplate)
	if err != nil {
		return err
	}

	return nil
}

func (u *StackTemplateUsecase) UpdateOrganizations(ctx context.Context, dto model.StackTemplate) error {
	_, err := u.repo.Get(ctx, dto.ID)
	if err != nil {
		return httpErrors.NewBadRequestError(err, "ST_NOT_EXISTED_STACK_TEMPLATE", "")
	}

	organizations := make([]model.Organization, 0)
	for _, organizationId := range dto.OrganizationIds {
		organization, err := u.organizationRepo.Get(ctx, organizationId)
		if err == nil {
			organizations = append(organizations, organization)
		}
	}

	err = u.repo.UpdateOrganizations(ctx, dto.ID, organizations)
	if err != nil {
		return httpErrors.NewBadRequestError(err, "ST_FAILED_UPDATE_ORGANIZATION", "")
	}

	return nil
}

func (u *StackTemplateUsecase) AddOrganizationStackTemplates(ctx context.Context, organizationId string, stackTemplateIds []string) error {
	_, err := u.organizationRepo.Get(ctx, organizationId)
	if err != nil {
		return httpErrors.NewBadRequestError(err, "O_NOT_EXISTED_NAME", "")
	}

	stackTemplates := make([]model.StackTemplate, 0)
	for _, strId := range stackTemplateIds {
		stackTemplateId, _ := uuid.Parse(strId)
		stackTemplate, err := u.repo.Get(ctx, stackTemplateId)
		if err == nil {
			stackTemplates = append(stackTemplates, stackTemplate)
		}
	}

	err = u.organizationRepo.AddStackTemplates(ctx, organizationId, stackTemplates)
	if err != nil {
		return httpErrors.NewBadRequestError(err, "ST_FAILED_ADD_ORGANIZATION_STACK_TEMPLATE", "")
	}

	return nil
}

func (u *StackTemplateUsecase) RemoveOrganizationStackTemplates(ctx context.Context, organizationId string, stackTemplateIds []string) error {
	_, err := u.organizationRepo.Get(ctx, organizationId)
	if err != nil {
		return httpErrors.NewBadRequestError(err, "O_NOT_EXISTED_NAME", "")
	}

	stackTemplates := make([]model.StackTemplate, 0)
	for _, strId := range stackTemplateIds {
		stackTemplateId, _ := uuid.Parse(strId)
		stackTemplate, err := u.repo.Get(ctx, stackTemplateId)
		if err == nil {
			stackTemplates = append(stackTemplates, stackTemplate)
		}
	}

	err = u.organizationRepo.RemoveStackTemplates(ctx, organizationId, stackTemplates)
	if err != nil {
		return httpErrors.NewBadRequestError(err, "ST_FAILED_REMOVE_ORGANIZATION_STACK_TEMPLATE", "")
	}

	return nil
}

func (u *StackTemplateUsecase) GetTemplateIds(ctx context.Context) (out []string, err error) {
	clientset_admin, err := kubernetes.GetClientAdminCluster(ctx)
	if err != nil {
		return out, errors.Wrap(err, "Failed to get client set for admin cluster")
	}

	secrets, err := clientset_admin.CoreV1().Secrets("argo").Get(context.TODO(), "git-svc-token", metav1.GetOptions{})
	if err != nil {
		log.Error(ctx, "cannot found git-svc-token. so use default hard-corded values")
		return out, err
	}

	gitSvcUrl := string(secrets.Data["GIT_SVC_URL"])
	username := string(secrets.Data["USERNAME"])
	branch := string(secrets.Data["GIT_BASE_BRANCH"])
	url := fmt.Sprintf("%s/%s/decapod-site/src/branch/%s", gitSvcUrl, username, branch)
	log.Info(ctx, "git url : ", url)

	rsp, err := http.Get(url)
	if err != nil {
		return out, err
	}
	defer rsp.Body.Close()

	html, err := goquery.NewDocumentFromReader(rsp.Body)
	if err != nil {
		return out, err
	}

	wrapper := html.Find("#repo-files-table > tbody")
	items := wrapper.Find("a.muted")
	items.Each(func(idx int, sel *goquery.Selection) {
		href, _ := sel.Attr("href")
		if strings.Contains(href, "reference") {
			arr := strings.Split(href, "/")
			log.Info(ctx, arr[len(arr)-1])

			out = append(out, arr[len(arr)-1])
		}

	})

	return
}

func (u *StackTemplateUsecase) GetCloudServices(ctx context.Context, organizationId string) (out []string, err error) {
	stackTemplates, err := u.repo.FetchWithOrganization(ctx, organizationId, nil)
	if err != nil {
		return nil, err
	}

	for _, stackTemplate := range stackTemplates {
		bExist := false
		for _, val := range out {
			if val == stackTemplate.CloudService {
				bExist = true
				break
			}
		}

		if !bExist {
			out = append(out, stackTemplate.CloudService)
		}
	}
	return
}

func servicesFromIds(serviceIds []string) []byte {
	services := "["
	for i, serviceId := range serviceIds {
		if i > 0 {
			services = services + ","
		}
		switch serviceId {
		case "LMA":
			services = services + internal.SERVICE_LMA
		case "SERVICE_MESH":
			services = services + internal.SERVICE_SERVICE_MESH
		}
	}
	services = services + "]"
	return []byte(services)
}
