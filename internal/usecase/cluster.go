package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/openinfradev/tks-api/internal/middleware/auth/request"

	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-api/internal/serializer"
	argowf "github.com/openinfradev/tks-api/pkg/argo-client"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
	gcache "github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

type IClusterUsecase interface {
	WithTrx(*gorm.DB) IClusterUsecase
	Fetch(ctx context.Context, organizationId string, pg *pagination.Pagination) ([]domain.Cluster, error)
	FetchByCloudAccountId(ctx context.Context, cloudAccountId uuid.UUID, pg *pagination.Pagination) (out []domain.Cluster, err error)
	Create(ctx context.Context, dto domain.Cluster) (clusterId domain.ClusterId, err error)
	CreateByoh(ctx context.Context, dto domain.Cluster) (clusterId domain.ClusterId, err error)
	Get(ctx context.Context, clusterId domain.ClusterId) (out domain.Cluster, err error)
	GetClusterSiteValues(ctx context.Context, clusterId domain.ClusterId) (out domain.ClusterSiteValuesResponse, err error)
	Delete(ctx context.Context, clusterId domain.ClusterId) (err error)
}

type ClusterUsecase struct {
	repo              repository.IClusterRepository
	appGroupRepo      repository.IAppGroupRepository
	cloudAccountRepo  repository.ICloudAccountRepository
	stackTemplateRepo repository.IStackTemplateRepository
	argo              argowf.ArgoClient
	cache             *gcache.Cache
}

func NewClusterUsecase(r repository.Repository, argoClient argowf.ArgoClient, cache *gcache.Cache) IClusterUsecase {
	return &ClusterUsecase{
		repo:              r.Cluster,
		appGroupRepo:      r.AppGroup,
		cloudAccountRepo:  r.CloudAccount,
		stackTemplateRepo: r.StackTemplate,
		argo:              argoClient,
		cache:             cache,
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

func (u *ClusterUsecase) Fetch(ctx context.Context, organizationId string, pg *pagination.Pagination) (out []domain.Cluster, err error) {
	if organizationId == "" {
		// [TODO] 사용자가 속한 organization 리스트
		out, err = u.repo.Fetch(pg)
	} else {
		out, err = u.repo.FetchByOrganizationId(organizationId, pg)
	}

	if err != nil {
		return nil, err
	}
	return out, nil
}

func (u *ClusterUsecase) FetchByCloudAccountId(ctx context.Context, cloudAccountId uuid.UUID, pg *pagination.Pagination) (out []domain.Cluster, err error) {
	if cloudAccountId == uuid.Nil {
		return nil, fmt.Errorf("Invalid cloudAccountId")
	}

	out, err = u.repo.Fetch(pg)

	if err != nil {
		return nil, err
	}
	return out, nil
}

func (u *ClusterUsecase) Create(ctx context.Context, dto domain.Cluster) (clusterId domain.ClusterId, err error) {
	user, ok := request.UserFrom(ctx)
	if !ok {
		return "", httpErrors.NewBadRequestError(fmt.Errorf("Invalid token"), "", "")
	}

	_, err = u.repo.GetByName(dto.OrganizationId, dto.Name)
	if err == nil {
		return "", httpErrors.NewBadRequestError(httpErrors.DuplicateResource, "", "")
	}

	// check cloudAccount
	cloudAccounts, err := u.cloudAccountRepo.Fetch(dto.OrganizationId, nil)
	if err != nil {
		return "", httpErrors.NewBadRequestError(fmt.Errorf("Failed to get cloudAccounts"), "", "")
	}

	tksCloudAccountId := dto.CloudAccountId.String()
	isExist := false
	for _, ca := range cloudAccounts {
		if ca.ID == dto.CloudAccountId {

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
	stackTemplate, err := u.stackTemplateRepo.Get(dto.StackTemplateId)
	if err != nil {
		return "", httpErrors.NewBadRequestError(errors.Wrap(err, "Invalid stackTemplateId"), "", "")
	}

	userId := user.GetUserId()
	dto.CreatorId = &userId
	clusterId, err = u.repo.Create(dto)
	if err != nil {
		return "", errors.Wrap(err, "Failed to create cluster")
	}

	workflowId, err := u.argo.SumbitWorkflowFromWftpl(
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
				//"manifest_repo_url=" + viper.GetString("git-base-url") + "/" + viper.GetString("git-account") + "/" + clusterId + "-manifests",
			},
		})
	if err != nil {
		log.ErrorWithContext(ctx, "failed to submit argo workflow template. err : ", err)
		return "", err
	}
	log.InfoWithContext(ctx, "Successfully submited workflow: ", workflowId)

	if err := u.repo.InitWorkflow(clusterId, workflowId, domain.ClusterStatus_INSTALLING); err != nil {
		return "", errors.Wrap(err, "Failed to initialize status")
	}

	return clusterId, nil
}

func (u *ClusterUsecase) CreateByoh(ctx context.Context, dto domain.Cluster) (clusterId domain.ClusterId, err error) {
	user, ok := request.UserFrom(ctx)
	if !ok {
		return "", httpErrors.NewBadRequestError(fmt.Errorf("Invalid token"), "", "")
	}

	_, err = u.repo.GetByName(dto.OrganizationId, dto.Name)
	if err == nil {
		return "", httpErrors.NewBadRequestError(httpErrors.DuplicateResource, "", "")
	}

	stackTemplate, err := u.stackTemplateRepo.Get(dto.StackTemplateId)
	if err != nil {
		return "", httpErrors.NewBadRequestError(errors.Wrap(err, "Invalid stackTemplateId"), "", "")
	}

	userId := user.GetUserId()
	dto.CreatorId = &userId
	clusterId, err = u.repo.Create(dto)
	if err != nil {
		return "", errors.Wrap(err, "Failed to create cluster")
	}

	workflowId := ""
	workflowId, err = u.argo.SumbitWorkflowFromWftpl(
		"bootstrap-tks-usercluster",
		argowf.SubmitOptions{
			Parameters: []string{
				fmt.Sprintf("tks_api_url=%s", viper.GetString("external-address")),
				"contract_id=" + dto.OrganizationId,
				"cluster_id=" + clusterId.String(),
				"site_name=" + clusterId.String(),
				"template_name=" + stackTemplate.Template,
				"git_account=" + viper.GetString("git-account"),
				"creator=" + user.GetUserId().String(),
				"cloud_account_id=NULL",
				"base_repo_branch=" + viper.GetString("revision"),
			},
		})
	if err != nil {
		log.ErrorWithContext(ctx, "failed to submit argo workflow template. err : ", err)
		return "", err
	}
	log.InfoWithContext(ctx, "Successfully submited workflow: ", workflowId)

	if err := u.repo.InitWorkflow(clusterId, workflowId, domain.ClusterStatus_INSTALLING); err != nil {
		return "", errors.Wrap(err, "Failed to initialize status")
	}

	return clusterId, nil
}

func (u *ClusterUsecase) Get(ctx context.Context, clusterId domain.ClusterId) (out domain.Cluster, err error) {
	cluster, err := u.repo.Get(clusterId)
	if err != nil {
		return domain.Cluster{}, err
	}

	return cluster, nil
}

func (u *ClusterUsecase) Delete(ctx context.Context, clusterId domain.ClusterId) (err error) {
	cluster, err := u.repo.Get(clusterId)
	if err != nil {
		return httpErrors.NewNotFoundError(err, "", "")
	}

	if cluster.Status != domain.ClusterStatus_RUNNING {
		return fmt.Errorf("The cluster can not be deleted. cluster status : %s", cluster.Status)
	}

	resAppGroups, err := u.appGroupRepo.Fetch(clusterId, nil)
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
	cloudAccount, err := u.cloudAccountRepo.Get(cluster.CloudAccountId)
	if err != nil {
		return httpErrors.NewInternalServerError(fmt.Errorf("Failed to get cloudAccount"), "", "")
	}
	tksCloudAccountId := cluster.CloudAccountId.String()
	if strings.Contains(cloudAccount.Name, domain.CLOUD_ACCOUNT_INCLUSTER) {
		tksCloudAccountId = "NULL"
	}

	workflowId, err := u.argo.SumbitWorkflowFromWftpl(
		"tks-remove-usercluster",
		argowf.SubmitOptions{
			Parameters: []string{
				"app_group=tks-cluster-aws",
				"tks_api_url=http://tks-api.tks.svc:9110",
				"cluster_id=" + clusterId.String(),
				"cloud_account_id=" + tksCloudAccountId,
			},
		})
	if err != nil {
		log.ErrorWithContext(ctx, "failed to submit argo workflow template. err : ", err)
		return errors.Wrap(err, "Failed to call argo workflow")
	}

	log.DebugWithContext(ctx, "submited workflow name : ", workflowId)

	if err := u.repo.InitWorkflow(clusterId, workflowId, domain.ClusterStatus_DELETING); err != nil {
		return errors.Wrap(err, "Failed to initialize status")
	}

	return nil
}

func (u *ClusterUsecase) GetClusterSiteValues(ctx context.Context, clusterId domain.ClusterId) (out domain.ClusterSiteValuesResponse, err error) {
	cluster, err := u.repo.Get(clusterId)
	if err != nil {
		return domain.ClusterSiteValuesResponse{}, errors.Wrap(err, "Failed to get cluster")
	}

	out.SshKeyName = "tks-seoul"
	out.ClusterRegion = "ap-northeast-2"

	if err := serializer.Map(cluster.Conf, &out); err != nil {
		log.ErrorWithContext(ctx, err)
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

/*
func (u *ClusterUsecase) constructClusterConf(rawConf *domain.ClusterConf) (clusterConf *domain.ClusterConf, err error) {
	region := "ap-northeast-2"
	if rawConf != nil && rawConf.Region != "" {
		region = rawConf.Region
	}

	numOfAz := 1
	if rawConf != nil && rawConf.NumOfAz != 0 {
		numOfAz = int(rawConf.NumOfAz)

		if numOfAz > 3 {
			log.ErrorWithContext(ctx,"Error: numOfAz cannot exceed 3.")
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
			log.DebugWithContext(ctx,"Found region : ", key)
			maxAzForSelectedRegion = val
			log.DebugWithContext(ctx,"Trimmed azNum var: ", maxAzForSelectedRegion)
			found = true
		}
	}

	if !found {
		log.ErrorWithContext(ctx,"Couldn't find entry for region ", region)
	}

	if numOfAz > maxAzForSelectedRegion {
		log.ErrorWithContext(ctx,"Invalid numOfAz: exceeded the number of Az in region ", region)
		temp_err := fmt.Errorf("Invalid numOfAz: exceeded the number of Az in region %s", region)
		return nil, temp_err
	}

	// Validate if machineReplicas is multiple of number of AZ
	replicas := int(rawConf.MachineReplicas)
	if replicas == 0 {
		log.DebugWithContext(ctx,"No machineReplicas param. Using default values..")
	} else {
		if remainder := replicas % numOfAz; remainder != 0 {
			log.ErrorWithContext(ctx,"Invalid machineReplicas: it should be multiple of numOfAz ", numOfAz)
			temp_err := fmt.Errorf("Invalid machineReplicas: it should be multiple of numOfAz %d", numOfAz)
			return nil, temp_err
		} else {
			log.DebugWithContext(ctx,"Valid replicas and numOfAz. Caculating minSize & maxSize..")
			minSizePerAz = int(replicas / numOfAz)
			maxSizePerAz = minSizePerAz * 5

			// Validate if maxSizePerAx is within allowed range
			if maxSizePerAz > MAX_SIZE_PER_AZ {
				fmt.Printf("maxSizePerAz exceeded maximum value %d, so adjusted to %d", MAX_SIZE_PER_AZ, MAX_SIZE_PER_AZ)
				maxSizePerAz = MAX_SIZE_PER_AZ
			}
			log.DebugWithContext(ctx,"Derived minSizePerAz: ", minSizePerAz)
			log.DebugWithContext(ctx,"Derived maxSizePerAz: ", maxSizePerAz)
		}
	}

	// Construct cluster conf
	tempConf := domain.ClusterConf{
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
