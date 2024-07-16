package usecase

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Nerzal/gocloak/v13"
	"github.com/openinfradev/tks-api/internal/keycloak"

	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/helper"
	"github.com/openinfradev/tks-api/internal/middleware/auth/request"
	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-api/internal/serializer"
	argowf "github.com/openinfradev/tks-api/pkg/argo-client"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/kubernetes"
	"github.com/openinfradev/tks-api/pkg/log"
	gcache "github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	byoh "github.com/vmware-tanzu/cluster-api-provider-bringyourownhost/apis/infrastructure/v1beta1"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type IClusterUsecase interface {
	WithTrx(*gorm.DB) IClusterUsecase
	Fetch(ctx context.Context, organizationId string, pg *pagination.Pagination) ([]model.Cluster, error)
	FetchByCloudAccountId(ctx context.Context, cloudAccountId uuid.UUID, pg *pagination.Pagination) (out []model.Cluster, err error)
	Create(ctx context.Context, dto model.Cluster) (clusterId domain.ClusterId, err error)
	Import(ctx context.Context, dto model.Cluster) (clusterId domain.ClusterId, err error)
	Bootstrap(ctx context.Context, dto model.Cluster) (clusterId domain.ClusterId, err error)
	Install(ctx context.Context, clusterId domain.ClusterId) (err error)
	Resume(ctx context.Context, clusterId domain.ClusterId) (err error)
	Get(ctx context.Context, clusterId domain.ClusterId) (out model.Cluster, err error)
	GetClusterSiteValues(ctx context.Context, clusterId domain.ClusterId) (out domain.ClusterSiteValuesResponse, err error)
	Delete(ctx context.Context, clusterId domain.ClusterId) (err error)
	CreateBootstrapKubeconfig(ctx context.Context, clusterId domain.ClusterId) (out domain.BootstrapKubeconfig, err error)
	GetBootstrapKubeconfig(ctx context.Context, clusterId domain.ClusterId) (out domain.BootstrapKubeconfig, err error)
	GetNodes(ctx context.Context, clusterId domain.ClusterId) (out []domain.ClusterNode, err error)
}

type ClusterUsecase struct {
	repo              repository.IClusterRepository
	appGroupRepo      repository.IAppGroupRepository
	cloudAccountRepo  repository.ICloudAccountRepository
	stackTemplateRepo repository.IStackTemplateRepository
	organizationRepo  repository.IOrganizationRepository
	argo              argowf.ArgoClient
	cache             *gcache.Cache
	kc                keycloak.IKeycloak
}

func NewClusterUsecase(r repository.Repository, argoClient argowf.ArgoClient, cache *gcache.Cache, kc keycloak.IKeycloak) IClusterUsecase {
	return &ClusterUsecase{
		repo:              r.Cluster,
		appGroupRepo:      r.AppGroup,
		cloudAccountRepo:  r.CloudAccount,
		stackTemplateRepo: r.StackTemplate,
		organizationRepo:  r.Organization,
		argo:              argoClient,
		cache:             cache,
		kc:                kc,
	}
}

/*
var azPerRegion = map[string]int{
	"ec2.af-south-1.amazonaws.com":     3,
	"ec2.eu-north-1.amazonaws.com":     3,
	"ec2.ap-south-1.amazonaws.com":     3,
	"ec2.eu-west-3.amazonaws.com":      3,
	"ec2.eu-west-2.amazonaws.com":      3,
	"ec2.eu-south-1.amazonaws.com":     3,
	"ec2.eu-west-1.amazonaws.com":      3,
	"ec2.ap-northeast-3.amazonaws.com": 3,
	"ec2.ap-northeast-2.amazonaws.com": 4,
	"ec2.me-south-1.amazonaws.com":     3,
	"ec2.ap-northeast-1.amazonaws.com": 4,
	"ec2.sa-east-1.amazonaws.com":      3,
	"ec2.ca-central-1.amazonaws.com":   3,
	"ec2.ap-east-1.amazonaws.com":      3,
	"ec2.ap-southeast-1.amazonaws.com": 3,
	"ec2.ap-southeast-2.amazonaws.com": 3,
	"ec2.eu-central-1.amazonaws.com":   3,
	"ec2.ap-southeast-3.amazonaws.com": 3,
	"ec2.us-east-1.amazonaws.com":      6,
	"ec2.us-east-2.amazonaws.com":      3,
	"ec2.us-west-1.amazonaws.com":      3,
	"ec2.us-west-2.amazonaws.com":      4,
}

const MAX_SIZE_PER_AZ = 99
*/

func (u *ClusterUsecase) WithTrx(trxHandle *gorm.DB) IClusterUsecase {
	u.repo = u.repo.WithTrx(trxHandle)
	return u
}

func (u *ClusterUsecase) Fetch(ctx context.Context, organizationId string, pg *pagination.Pagination) (out []model.Cluster, err error) {
	user, ok := request.UserFrom(ctx)
	if !ok {
		return out, httpErrors.NewBadRequestError(fmt.Errorf("Invalid token"), "", "")
	}

	if organizationId == "" {
		// [TODO] 사용자가 속한 organization 리스트
		out, err = u.repo.Fetch(ctx, pg)
	} else {
		out, err = u.repo.FetchByOrganizationId(ctx, organizationId, user.GetUserId(), pg)
	}

	if err != nil {
		return nil, err
	}
	return out, nil
}

func (u *ClusterUsecase) FetchByCloudAccountId(ctx context.Context, cloudAccountId uuid.UUID, pg *pagination.Pagination) (out []model.Cluster, err error) {
	if cloudAccountId == uuid.Nil {
		return nil, fmt.Errorf("Invalid cloudAccountId")
	}

	out, err = u.repo.Fetch(ctx, pg)

	if err != nil {
		return nil, err
	}
	return out, nil
}

func (u *ClusterUsecase) Create(ctx context.Context, dto model.Cluster) (clusterId domain.ClusterId, err error) {
	user, ok := request.UserFrom(ctx)
	if !ok {
		return "", httpErrors.NewBadRequestError(fmt.Errorf("Invalid token"), "", "")
	}

	_, err = u.repo.GetByName(ctx, dto.OrganizationId, dto.Name)
	if err == nil {
		return "", httpErrors.NewBadRequestError(httpErrors.DuplicateResource, "", "")
	}

	// check cloudAccount
	cloudAccounts, err := u.cloudAccountRepo.Fetch(ctx, dto.OrganizationId, nil)
	if err != nil {
		return "", httpErrors.NewBadRequestError(fmt.Errorf("Failed to get cloudAccounts"), "", "")
	}

	tksCloudAccountId := dto.CloudAccountId.String()
	isExist := false
	for _, ca := range cloudAccounts {
		if ca.ID == *dto.CloudAccountId {

			// FOR TEST. ADD MAGIC KEYWORD
			if strings.Contains(ca.Name, domain.CLOUD_ACCOUNT_INCLUSTER) {
				tksCloudAccountId = "NULL"
			}
			isExist = true
			break
		}
	}
	if !isExist {
		return "", httpErrors.NewBadRequestError(fmt.Errorf("Not found cloudAccountId[%s] in organization[%s]", dto.CloudAccountId, dto.OrganizationId), "", "")
	}

	// check stackTemplate
	stackTemplate, err := u.stackTemplateRepo.Get(ctx, dto.StackTemplateId)
	if err != nil {
		return "", httpErrors.NewBadRequestError(errors.Wrap(err, "Invalid stackTemplateId"), "", "")
	}
	if stackTemplate.CloudService != dto.CloudService {
		return "", httpErrors.NewBadRequestError(fmt.Errorf("Invalid cloudService for stackTemplate "), "", "")
	}

	userId := user.GetUserId()
	dto.CreatorId = &userId
	clusterId, err = u.repo.Create(ctx, dto)
	if err != nil {
		return "", errors.Wrap(err, "Failed to create cluster")
	}

	workflowId, err := u.argo.SumbitWorkflowFromWftpl(
		ctx,
		"create-tks-usercluster",
		argowf.SubmitOptions{
			Parameters: []string{
				fmt.Sprintf("tks_api_url=%s", viper.GetString("external-address")),
				"contract_id=" + dto.OrganizationId,
				"cluster_id=" + clusterId.String(),
				"site_name=" + clusterId.String(),
				"template_name=" + stackTemplate.Template,
				"git_account=" + viper.GetString("git-account"),
				"creator=" + user.GetUserId().String(),
				"cloud_account_id=" + tksCloudAccountId,
				"base_repo_branch=" + viper.GetString("revision"),
				"keycloak_url=" + viper.GetString("keycloak-address"),
				"policy_ids=" + strings.Join(dto.PolicyIds, ","),
			},
		})
	if err != nil {
		log.Error(ctx, "failed to submit argo workflow template. err : ", err)
		return "", err
	}
	log.Info(ctx, "Successfully submited workflow: ", workflowId)

	if err := u.repo.InitWorkflow(ctx, clusterId, workflowId, domain.ClusterStatus_INSTALLING); err != nil {
		return "", errors.Wrap(err, "Failed to initialize status")
	}

	return clusterId, nil
}

func (u *ClusterUsecase) Import(ctx context.Context, dto model.Cluster) (clusterId domain.ClusterId, err error) {
	user, ok := request.UserFrom(ctx)
	if !ok {
		return "", httpErrors.NewBadRequestError(fmt.Errorf("Invalid token"), "", "")
	}

	_, err = u.repo.GetByName(ctx, dto.OrganizationId, dto.Name)
	if err == nil {
		return "", httpErrors.NewBadRequestError(httpErrors.DuplicateResource, "", "")
	}

	_, err = u.organizationRepo.Get(ctx, dto.OrganizationId)
	if err != nil {
		return "", httpErrors.NewBadRequestError(fmt.Errorf("Invalid organizationId"), "", "")
	}

	// check stackTemplate
	stackTemplate, err := u.stackTemplateRepo.Get(ctx, dto.StackTemplateId)
	if err != nil {
		return "", httpErrors.NewBadRequestError(errors.Wrap(err, "Invalid stackTemplateId"), "", "")
	}

	userId := user.GetUserId()
	dto.CreatorId = &userId
	if dto.ClusterType == domain.ClusterType_ADMIN {
		dto.ID = "tks-admin"
		dto.Name = "tks-admin"
	}
	clusterId, err = u.repo.Create(ctx, dto)
	if err != nil {
		return "", errors.Wrap(err, "Failed to create cluster")
	}

	kubeconfigBase64 := base64.StdEncoding.EncodeToString([]byte(dto.Kubeconfig))

	workflowId, err := u.argo.SumbitWorkflowFromWftpl(
		ctx,
		"import-tks-usercluster",
		argowf.SubmitOptions{
			Parameters: []string{
				fmt.Sprintf("tks_api_url=%s", viper.GetString("external-address")),
				"contract_id=" + dto.OrganizationId,
				"cluster_id=" + clusterId.String(),
				"template_name=" + stackTemplate.Template,
				"kubeconfig=" + kubeconfigBase64,
				"git_account=" + viper.GetString("git-account"),
				"keycloak_url=" + strings.TrimSuffix(viper.GetString("keycloak-address"), "/auth"),
				"base_repo_branch=" + viper.GetString("revision"),
			},
		})
	if err != nil {
		log.Error(ctx, "failed to submit argo workflow template. err : ", err)
		return "", err
	}
	log.Info(ctx, "Successfully submited workflow: ", workflowId)

	if err := u.repo.InitWorkflow(ctx, clusterId, workflowId, domain.ClusterStatus_INSTALLING); err != nil {
		return "", errors.Wrap(err, "Failed to initialize status")
	}

	// keycloak setting
	log.Debugf(ctx, "Create keycloak client for %s", dto.ID)
	// Create keycloak client
	clientUUID, err := u.kc.CreateClient(ctx, dto.OrganizationId, dto.ID.String()+"-k8s-api", "", nil)
	if err != nil {
		log.Errorf(ctx, "Failed to create keycloak client for %s", dto.ID)
		return "", err
	}
	// Create keycloak client protocol mapper
	_, err = u.kc.CreateClientProtocolMapper(ctx, dto.OrganizationId, clientUUID, gocloak.ProtocolMapperRepresentation{
		Name:            gocloak.StringP("k8s-role-mapper"),
		Protocol:        gocloak.StringP("openid-connect"),
		ProtocolMapper:  gocloak.StringP("oidc-usermodel-client-role-mapper"),
		ConsentRequired: gocloak.BoolP(false),
		Config: &map[string]string{
			"usermodel.clientRoleMapping.clientId": dto.ID.String() + "-k8s-api",
			"claim.name":                           "groups",
			"access.token.claim":                   "false",
			"id.token.claim":                       "true",
			"userinfo.token.claim":                 "true",
			"multivalued":                          "true",
			"jsonType.label":                       "String",
		},
	})
	if err != nil {
		log.Errorf(ctx, "Failed to create keycloak client protocol mapper for %s", dto.ID)
		return "", err
	}
	// Create keycloak client role
	err = u.kc.CreateClientRole(ctx, dto.OrganizationId, clientUUID, "cluster-admin-create")
	if err != nil {
		log.Errorf(ctx, "Failed to create keycloak client role named %s for %s", "cluster-admin-create", dto.ID)
		return "", err
	}
	err = u.kc.CreateClientRole(ctx, dto.OrganizationId, clientUUID, "cluster-admin-read")
	if err != nil {
		log.Errorf(ctx, "Failed to create keycloak client role named %s for %s", "cluster-admin-read", dto.ID)
		return "", err
	}
	err = u.kc.CreateClientRole(ctx, dto.OrganizationId, clientUUID, "cluster-admin-update")
	if err != nil {
		log.Errorf(ctx, "Failed to create keycloak client role named %s for %s", "cluster-admin-update", dto.ID)
		return "", err
	}
	err = u.kc.CreateClientRole(ctx, dto.OrganizationId, clientUUID, "cluster-admin-delete")
	if err != nil {
		log.Errorf(ctx, "Failed to create keycloak client role named %s for %s", "cluster-admin-delete", dto.ID)
		return "", err
	}

	return clusterId, nil
}

func (u *ClusterUsecase) Bootstrap(ctx context.Context, dto model.Cluster) (clusterId domain.ClusterId, err error) {
	user, ok := request.UserFrom(ctx)
	if !ok {
		return "", httpErrors.NewBadRequestError(fmt.Errorf("Invalid token"), "", "")
	}

	_, err = u.repo.GetByName(ctx, dto.OrganizationId, dto.Name)
	if err == nil {
		return "", httpErrors.NewBadRequestError(httpErrors.DuplicateResource, "", "")
	}

	stackTemplate, err := u.stackTemplateRepo.Get(ctx, dto.StackTemplateId)
	if err != nil {
		return "", httpErrors.NewBadRequestError(errors.Wrap(err, "Invalid stackTemplateId"), "", "")
	}
	log.Infof(ctx, "%s %s", stackTemplate.CloudService, dto.CloudService)
	if stackTemplate.CloudService != dto.CloudService {
		return "", httpErrors.NewBadRequestError(fmt.Errorf("Invalid cloudService for stackTemplate "), "", "")
	}

	userId := user.GetUserId()
	dto.CreatorId = &userId
	clusterId, err = u.repo.Create(ctx, dto)
	if err != nil {
		return "", errors.Wrap(err, "Failed to create cluster")
	}

	workflow := "create-byoh-bootstrapkubeconfig"
	workflowId, err := u.argo.SumbitWorkflowFromWftpl(ctx, workflow, argowf.SubmitOptions{
		Parameters: []string{
			fmt.Sprintf("tks_api_url=%s", viper.GetString("external-address")),
			"cluster_id=" + clusterId.String(),
		},
	})
	if err != nil {
		log.Error(ctx, "failed to submit argo workflow template. err : ", err)
		return "", err
	}
	log.Info(ctx, "Successfully submited workflow: ", workflowId)

	if err := u.repo.InitWorkflow(ctx, clusterId, workflowId, domain.ClusterStatus_BOOTSTRAPPING); err != nil {
		return "", errors.Wrap(err, "Failed to initialize status")
	}

	return clusterId, nil
}

func (u *ClusterUsecase) Install(ctx context.Context, clusterId domain.ClusterId) (err error) {
	user, ok := request.UserFrom(ctx)
	if !ok {
		return httpErrors.NewBadRequestError(fmt.Errorf("Invalid token"), "", "")
	}

	cluster, err := u.repo.Get(ctx, clusterId)
	if err != nil {
		return httpErrors.NewBadRequestError(fmt.Errorf("Invalid clusterId"), "C_INVALID_CLUSTER_ID", "")
	}
	if cluster.CloudService != domain.CloudService_BYOH {
		return httpErrors.NewBadRequestError(fmt.Errorf("Invalid cloudService"), "C_INVALID_CLOUD_SERVICE", "")
	}

	stackTemplate, err := u.stackTemplateRepo.Get(ctx, cluster.StackTemplateId)
	if err != nil {
		return httpErrors.NewBadRequestError(errors.Wrap(err, "Invalid stackTemplateId"), "", "")
	}
	if stackTemplate.CloudService != cluster.CloudService {
		return httpErrors.NewBadRequestError(fmt.Errorf("Invalid cloudService for stackTemplate "), "", "")
	}

	workflowId, err := u.argo.SumbitWorkflowFromWftpl(
		ctx,
		"create-tks-usercluster",
		argowf.SubmitOptions{
			Parameters: []string{
				fmt.Sprintf("tks_api_url=%s", viper.GetString("external-address")),
				"contract_id=" + cluster.OrganizationId,
				"cluster_id=" + cluster.ID.String(),
				"site_name=" + cluster.ID.String(),
				"template_name=" + stackTemplate.Template,
				"git_account=" + viper.GetString("git-account"),
				"creator=" + user.GetUserId().String(),
				"cloud_account_id=NULL",
				"base_repo_branch=" + viper.GetString("revision"),
				"keycloak_url=" + viper.GetString("keycloak-address"),
				"policy_ids=",
				//"manifest_repo_url=" + viper.GetString("git-base-url") + "/" + viper.GetString("git-account") + "/" + clusterId + "-manifests",
			},
		})
	if err != nil {
		log.Error(ctx, "failed to submit argo workflow template. err : ", err)
		return err
	}
	log.Info(ctx, "Successfully submited workflow: ", workflowId)

	if err := u.repo.InitWorkflow(ctx, cluster.ID, workflowId, domain.ClusterStatus_INSTALLING); err != nil {
		return errors.Wrap(err, "Failed to initialize status")
	}

	return nil
}

func (u *ClusterUsecase) Resume(ctx context.Context, clusterId domain.ClusterId) (err error) {
	cluster, err := u.Get(ctx, clusterId)
	if err != nil {
		return httpErrors.NewBadRequestError(fmt.Errorf("Invalid stackId"), "S_INVALID_STACK_ID", "")
	}

	if cluster.CloudService != domain.CloudService_BYOH {
		return httpErrors.NewBadRequestError(fmt.Errorf("Invalid cloud service"), "S_INVALID_CLOUD_SERVICE", "")
	}

	if cluster.WorkflowId == "" {
		return httpErrors.NewInternalServerError(fmt.Errorf("Invalid workflow id"), "", "")
	}

	workflow, err := u.argo.ResumeWorkflow(ctx, "argo", cluster.WorkflowId)
	if err != nil {
		log.Error(ctx, err)
		return httpErrors.NewInternalServerError(err, "S_FAILED_TO_CALL_WORKFLOW", "")
	}
	log.Debug(ctx, "Resume workflow: ", workflow)

	if err := u.repo.InitWorkflow(ctx, cluster.ID, cluster.WorkflowId, domain.ClusterStatus_INSTALLING); err != nil {
		return errors.Wrap(err, "Failed to initialize status")
	}
	return nil
}

func (u *ClusterUsecase) Get(ctx context.Context, clusterId domain.ClusterId) (out model.Cluster, err error) {
	cluster, err := u.repo.Get(ctx, clusterId)
	if err != nil {
		return model.Cluster{}, err
	}

	return cluster, nil
}

func (u *ClusterUsecase) Delete(ctx context.Context, clusterId domain.ClusterId) (err error) {
	cluster, err := u.repo.Get(ctx, clusterId)
	if err != nil {
		return httpErrors.NewNotFoundError(err, "", "")
	}

	if cluster.Status != domain.ClusterStatus_RUNNING {
		return fmt.Errorf("The cluster can not be deleted. cluster status : %s", cluster.Status)
	}

	resAppGroups, err := u.appGroupRepo.Fetch(ctx, clusterId, nil)
	if err != nil {
		return errors.Wrap(err, "Failed to get appgroup")
	}

	for _, resAppGroup := range resAppGroups {
		if resAppGroup.Status != domain.AppGroupStatus_DELETED {
			return fmt.Errorf("Undeleted services remain. %s", resAppGroup.ID)
		}
	}

	// FOR TEST. ADD MAGIC KEYWORD
	// check cloudAccount
	tksCloudAccountId := "NULL"
	if cluster.CloudService != domain.CloudService_BYOH {
		cloudAccount, err := u.cloudAccountRepo.Get(ctx, cluster.CloudAccount.ID)
		if err != nil {
			return httpErrors.NewInternalServerError(fmt.Errorf("Failed to get cloudAccount"), "", "")
		}
		tksCloudAccountId = cluster.CloudAccount.ID.String()
		if strings.Contains(cloudAccount.Name, domain.CLOUD_ACCOUNT_INCLUSTER) {
			tksCloudAccountId = "NULL"
		}
	}

	workflowId, err := u.argo.SumbitWorkflowFromWftpl(
		ctx,
		"tks-remove-usercluster",
		argowf.SubmitOptions{
			Parameters: []string{
				"app_group=tks-cluster-aws",
				"tks_api_url=http://tks-api.tks.svc:9110",
				"cluster_id=" + clusterId.String(),
				"cloud_account_id=" + tksCloudAccountId,
				"keycloak_url=" + viper.GetString("keycloak-address"),
				"contract_id=" + cluster.OrganizationId,
			},
		})
	if err != nil {
		log.Error(ctx, "failed to submit argo workflow template. err : ", err)
		return errors.Wrap(err, "Failed to call argo workflow")
	}

	log.Debug(ctx, "submited workflow name : ", workflowId)

	if err := u.repo.InitWorkflow(ctx, clusterId, workflowId, domain.ClusterStatus_DELETING); err != nil {
		return errors.Wrap(err, "Failed to initialize status")
	}

	return nil
}

func (u *ClusterUsecase) GetClusterSiteValues(ctx context.Context, clusterId domain.ClusterId) (out domain.ClusterSiteValuesResponse, err error) {
	cluster, err := u.repo.Get(ctx, clusterId)
	if err != nil {
		return domain.ClusterSiteValuesResponse{}, errors.Wrap(err, "Failed to get cluster")
	}

	out.SshKeyName = "tks-seoul"
	out.ClusterRegion = "ap-northeast-2"

	if err := serializer.Map(ctx, cluster, &out); err != nil {
		log.Error(ctx, err)
	}
	out.Domains = make([]domain.ClusterDomain, len(cluster.Domains))
	for i, domain := range cluster.Domains {
		if err = serializer.Map(ctx, domain, &out.Domains[i]); err != nil {
			log.Info(ctx, err)
		}
	}

	if cluster.StackTemplate.CloudService == "AWS" && cluster.StackTemplate.KubeType == "AWS" {
		out.TksUserNode = cluster.TksUserNode / domain.MAX_AZ_NUM
		out.TksUserNodeMax = cluster.TksUserNodeMax / domain.MAX_AZ_NUM
	}

	/*
		// 기능 변경 : 20230614 : machine deployment 사용하지 않음. 단, aws-standard 는 사용할 여지가 있으므로 주석처리해둔다.
		const MAX_AZ_NUM = 4
		if cluster.Conf.UserNodeCnt <= MAX_AZ_NUM {
			out.MdNumOfAz = cluster.Conf.UserNodeCnt
			out.MdMinSizePerAz = 1
			out.MdMaxSizePerAz = cluster.Conf.UserNodeCnt
		} else {
			out.MdNumOfAz = MAX_AZ_NUM
			out.MdMinSizePerAz = int(cluster.Conf.UserNodeCnt / MAX_AZ_NUM)
			out.MdMaxSizePerAz = cluster.Conf.UserNodeCnt * 5
		}
	*/
	return
}

func (u *ClusterUsecase) CreateBootstrapKubeconfig(ctx context.Context, clusterId domain.ClusterId) (out domain.BootstrapKubeconfig, err error) {
	_, err = u.repo.Get(ctx, clusterId)
	if err != nil {
		return out, httpErrors.NewNotFoundError(err, "", "")
	}

	workflow := "create-byoh-bootstrapkubeconfig"
	workflowId, err := u.argo.SumbitWorkflowFromWftpl(ctx, workflow, argowf.SubmitOptions{
		Parameters: []string{
			fmt.Sprintf("tks_api_url=%s", viper.GetString("external-address")),
			"cluster_id=" + clusterId.String(),
		},
	})
	if err != nil {
		log.Error(ctx, err)
		return out, httpErrors.NewInternalServerError(err, "S_FAILED_TO_CALL_WORKFLOW", "")
	}
	log.Debug(ctx, "Submitted workflow: ", workflowId)

	// wait & get clusterId ( max 1min 	)
	for i := 0; i < 60; i++ {
		time.Sleep(time.Second * 3)
		workflow, err := u.argo.GetWorkflow(ctx, "argo", workflowId)
		if err != nil {
			return out, err
		}

		log.Debug(ctx, "workflow ", workflow)

		if workflow.Status.Phase == "Succeeded" {
			break
		}
		if workflow.Status.Phase != "" && workflow.Status.Phase != "Running" {
			return out, fmt.Errorf("Invalid workflow status [%s]", workflow.Status.Phase)
		}
	}

	out, err = u.GetBootstrapKubeconfig(ctx, clusterId)
	if err != nil {
		return out, err
	}

	return out, nil
}

func (u *ClusterUsecase) GetBootstrapKubeconfig(ctx context.Context, clusterId domain.ClusterId) (out domain.BootstrapKubeconfig, err error) {
	cluster, err := u.repo.Get(ctx, clusterId)
	if err != nil {
		return out, httpErrors.NewNotFoundError(err, "", "")
	}
	client, err := kubernetes.GetClientAdminCluster(ctx)
	if err != nil {
		return out, err
	}

	kubeconfig := byoh.BootstrapKubeconfig{}
	data, err := client.RESTClient().
		Get().
		AbsPath("/apis/infrastructure.cluster.x-k8s.io/v1beta1").
		Namespace("default").
		Name("bootstrap-kubeconfig-" + cluster.ID.String()).
		Resource("bootstrapkubeconfigs").
		DoRaw(ctx)
	if err != nil {
		return out, err
	}

	if err := json.Unmarshal(data, &kubeconfig); err != nil {
		return out, err
	}

	log.Debug(ctx, helper.ModelToJson(kubeconfig.Status.BootstrapKubeconfigData))

	type BootstrapKubeconfigUser struct {
		Users []struct {
			Name string `yaml:"name"`
			User struct {
				Token string `yaml:"token"`
			} `yaml:"user"`
		} `yaml:"users"`
	}
	bytes := []byte(string(*kubeconfig.Status.BootstrapKubeconfigData))

	kubeconfigData := BootstrapKubeconfigUser{}
	err = yaml.Unmarshal(bytes, &kubeconfigData)
	if err != nil {
		return out, err
	}

	token := kubeconfigData.Users[0].User.Token[:6]
	log.Info(ctx, "token : ", token)

	secrets, err := client.CoreV1().Secrets("kube-system").Get(context.TODO(), "bootstrap-token-"+token, metav1.GetOptions{})
	if err != nil {
		log.Error(ctx, err)
		return out, err
	}

	log.Info(ctx, secrets.Data["expiration"][:])

	// 2023-10-17T11:05:33Z
	now := time.Now()
	expiration, err := time.Parse(time.RFC3339, string(secrets.Data["expiration"][:]))
	if err != nil {
		return out, err
	}

	period, err := time.ParseDuration(expiration.Sub(now).String())
	if err != nil {
		return out, err
	}
	out.Expiration = int(period.Seconds())

	return out, nil
}

func (u *ClusterUsecase) GetNodes(ctx context.Context, clusterId domain.ClusterId) (out []domain.ClusterNode, err error) {
	cluster, err := u.repo.Get(ctx, clusterId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return out, httpErrors.NewNotFoundError(err, "S_FAILED_FETCH_CLUSTER", "")
		}
		return out, err
	}
	if cluster.CloudService != domain.CloudService_BYOH {
		return out, httpErrors.NewBadRequestError(fmt.Errorf("Invalid cloud service"), "", "")
	}

	client, err := kubernetes.GetClientAdminCluster(ctx)
	if err != nil {
		return out, err
	}

	hosts := byoh.ByoHostList{}
	data, err := client.RESTClient().
		Get().
		AbsPath("/apis/infrastructure.cluster.x-k8s.io/v1beta1").
		Namespace("default").
		//Namespace(cluster.ID). [TODO]
		Resource("byohosts").
		DoRaw(ctx)
	if err != nil {
		return out, err
	}

	if err = json.Unmarshal(data, &hosts); err != nil {
		return out, err
	}

	/* FOR DEBUG
	for _, host := range hosts.Items {
		log.Info(host.Name)
		log.Info(host.Labels)
		log.Info(host.Status.Conditions[0].Type)
	}
	*/

	clusterNodeStatus := func(targeted int, registered int) string {
		if targeted <= registered {
			return "COMPLETED"
		}
		return "INPROGRESS"
	}

	tksCpNodeRegistered, tksCpNodeRegistering, tksCpHosts := 0, 0, make([]domain.ClusterHost, 0)
	tksInfraNodeRegistered, tksInfraNodeRegistering, tksInfraHosts := 0, 0, make([]domain.ClusterHost, 0)
	tksUserNodeRegistered, tksUserNodeRegistering, tksUserHosts := 0, 0, make([]domain.ClusterHost, 0)
	for _, host := range hosts.Items {
		label := host.Labels["role"]
		log.Info(ctx, "label : ", label)
		if len(label) < 12 {
			continue
		}
		arr := strings.Split(label, "-")
		if len(arr) < 2 {
			continue
		}
		clusterId := arr[0]
		if label[9] != '-' || clusterId != string(cluster.ID) {
			continue
		}
		role := label[10:]
		/*
			if host.Name == "ip-10-0-12-87.ap-northeast-2.compute.internal" {
				continue
			}

			role := host.Labels["role"] // [FOR TEST]
		*/

		hostStatus := host.Status.Conditions[0].Type
		registered, registering := 0, 0
		// K8sComponentsInstallationSucceeded
		if hostStatus == "K8sNodeBootstrapSucceeded" || hostStatus == "K8sComponentsInstallationSucceeded" {
			registered = 1
		} else {
			registering = 1
		}

		switch role {
		case "control-plane":
			tksCpNodeRegistered = tksCpNodeRegistered + registered
			tksCpNodeRegistering = tksCpNodeRegistering + registering
			tksCpHosts = append(tksCpHosts, domain.ClusterHost{Name: host.Name, Status: string(hostStatus)})
		case "tks":
			tksInfraNodeRegistered = tksInfraNodeRegistered + registered
			tksInfraNodeRegistering = tksInfraNodeRegistering + registering
			tksInfraHosts = append(tksInfraHosts, domain.ClusterHost{Name: host.Name, Status: string(hostStatus)})
		case "worker":
			tksUserNodeRegistered = tksUserNodeRegistered + registered
			tksUserNodeRegistering = tksUserNodeRegistering + registering
			tksUserHosts = append(tksUserHosts, domain.ClusterHost{Name: host.Name, Status: string(hostStatus)})
		}
	}

	bootstrapKubeconfig, err := u.GetBootstrapKubeconfig(ctx, cluster.ID)
	if err != nil {
		return out, err
	}

	command := fmt.Sprintf("curl -fL %s/api/packages/%s/generic/byoh_hostagent_install/%s/byoh_hostagent-install-%s.sh | sh -s -- --role %s-",
		viper.GetString("external-gitea-url"),
		viper.GetString("git-account"),
		string(cluster.ID),
		string(cluster.ID),
		string(cluster.ID))

	out = []domain.ClusterNode{
		{
			Type:        "TKS_CP_NODE",
			Targeted:    cluster.TksCpNode,
			Registered:  tksCpNodeRegistered,
			Registering: tksCpNodeRegistering,
			Status:      clusterNodeStatus(cluster.TksCpNode, tksCpNodeRegistered),
			Command:     command + "control-plane",
			Validity:    bootstrapKubeconfig.Expiration,
			Hosts:       tksCpHosts,
		},
		{
			Type:        "TKS_INFRA_NODE",
			Targeted:    cluster.TksInfraNode,
			Registered:  tksInfraNodeRegistered,
			Registering: tksInfraNodeRegistering,
			Status:      clusterNodeStatus(cluster.TksInfraNode, tksInfraNodeRegistered),
			Command:     command + "tks",
			Validity:    bootstrapKubeconfig.Expiration,
			Hosts:       tksInfraHosts,
		},
		{
			Type:        "TKS_USER_NODE",
			Targeted:    cluster.TksUserNode,
			Registered:  tksUserNodeRegistered,
			Registering: tksUserNodeRegistering,
			Status:      clusterNodeStatus(cluster.TksUserNode, tksUserNodeRegistered),
			Command:     command + "worker",
			Validity:    bootstrapKubeconfig.Expiration,
			Hosts:       tksUserHosts,
		},
	}

	// [TODO] for integration
	/*
		out.Nodes = []domain.StackNodeResponse{
			{
				ID:         "1",
				Type:       "TKS_CP_NODE",
				Targeted:   3,
				Registered: 1,
				Status:     "INPROGRESS",
				Command:    "curl -fL http://192.168.0.77/tks-byoh-hostagent-install.sh | sh -s CLUSTER-ID-control-plane",
				Validity:   3000,
			},
			{
				ID:         "2",
				Type:       "TKS_INFRA_NODE",
				Targeted:   0,
				Registered: 0,
				Status:     "PENDING",
				Command:    "curl -fL http://192.168.0.77/tks-byoh-hostagent-install.sh | sh -s CLUSTER-ID-control-plane",
				Validity:   3000,
			},
			{
				ID:         "3",
				Type:       "TKS_USER_NODE",
				Targeted:   3,
				Registered: 3,
				Status:     "COMPLETED",
				Command:    "curl -fL http://192.168.0.77/tks-byoh-hostagent-install.sh | sh -s CLUSTER-ID-control-plane",
				Validity:   3000,
			},
		}
	*/

	return
}

/*
func (u *ClusterUsecase) constructClusterConf(rawConf *model.ClusterConf) (clusterConf *model.ClusterConf, err error) {
	region := "ap-northeast-2"
	if rawConf != nil && rawConf.Region != "" {
		region = rawConf.Region
	}

	numOfAz := 1
	if rawConf != nil && rawConf.NumOfAz != 0 {
		numOfAz = int(rawConf.NumOfAz)

		if numOfAz > 3 {
			log.Error(ctx,"Error: numOfAz cannot exceed 3.")
			temp_err := fmt.Errorf("Error: numOfAz cannot exceed 3.")
			return nil, temp_err
		}
	}

	sshKeyName := "tks-seoul"
	if rawConf != nil && rawConf.SshKeyName != "" {
		sshKeyName = rawConf.SshKeyName
	}

	machineType := "t3.large"
	if rawConf != nil && rawConf.MachineType != "" {
		machineType = rawConf.MachineType
	}

	minSizePerAz := 1
	maxSizePerAz := 5

	// Check if numOfAz is correct based on pre-defined mapping table
	maxAzForSelectedRegion := 0

	var found bool = false
	for key, val := range azPerRegion {
		if strings.Contains(key, region) {
			log.Debug(ctx,"Found region : ", key)
			maxAzForSelectedRegion = val
			log.Debug(ctx,"Trimmed azNum var: ", maxAzForSelectedRegion)
			found = true
		}
	}

	if !found {
		log.Error(ctx,"Couldn't find entry for region ", region)
	}

	if numOfAz > maxAzForSelectedRegion {
		log.Error(ctx,"Invalid numOfAz: exceeded the number of Az in region ", region)
		temp_err := fmt.Errorf("Invalid numOfAz: exceeded the number of Az in region %s", region)
		return nil, temp_err
	}

	// Validate if machineReplicas is multiple of number of AZ
	replicas := int(rawConf.MachineReplicas)
	if replicas == 0 {
		log.Debug(ctx,"No machineReplicas param. Using default values..")
	} else {
		if remainder := replicas % numOfAz; remainder != 0 {
			log.Error(ctx,"Invalid machineReplicas: it should be multiple of numOfAz ", numOfAz)
			temp_err := fmt.Errorf("Invalid machineReplicas: it should be multiple of numOfAz %d", numOfAz)
			return nil, temp_err
		} else {
			log.Debug(ctx,"Valid replicas and numOfAz. Caculating minSize & maxSize..")
			minSizePerAz = int(replicas / numOfAz)
			maxSizePerAz = minSizePerAz * 5

			// Validate if maxSizePerAx is within allowed range
			if maxSizePerAz > MAX_SIZE_PER_AZ {
				fmt.Printf("maxSizePerAz exceeded maximum value %d, so adjusted to %d", MAX_SIZE_PER_AZ, MAX_SIZE_PER_AZ)
				maxSizePerAz = MAX_SIZE_PER_AZ
			}
			log.Debug(ctx,"Derived minSizePerAz: ", minSizePerAz)
			log.Debug(ctx,"Derived maxSizePerAz: ", maxSizePerAz)
		}
	}

	// Construct cluster conf
	tempConf := model.ClusterConf{
		SshKeyName:   sshKeyName,
		Region:       region,
		NumOfAz:      int(numOfAz),
		MachineType:  machineType,
		MinSizePerAz: int(minSizePerAz),
		MaxSizePerAz: int(maxSizePerAz),
	}

	fmt.Printf("Newly constructed cluster conf: %+v\n", &tempConf)
	return &tempConf, nil
}
*/
