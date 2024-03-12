package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/servicequotas"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/kubernetes"
	"github.com/openinfradev/tks-api/internal/middleware/auth/request"
	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/internal/repository"
	argowf "github.com/openinfradev/tks-api/pkg/argo-client"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

const MAX_WORKFLOW_TIME = 30

type ICloudAccountUsecase interface {
	Get(ctx context.Context, cloudAccountId uuid.UUID) (model.CloudAccount, error)
	GetByName(ctx context.Context, organizationId string, name string) (model.CloudAccount, error)
	GetByAwsAccountId(ctx context.Context, awsAccountId string) (model.CloudAccount, error)
	GetResourceQuota(ctx context.Context, cloudAccountId uuid.UUID) (available bool, out domain.ResourceQuota, err error)
	Fetch(ctx context.Context, organizationId string, pg *pagination.Pagination) ([]model.CloudAccount, error)
	Create(ctx context.Context, dto model.CloudAccount) (cloudAccountId uuid.UUID, err error)
	Update(ctx context.Context, dto model.CloudAccount) error
	Delete(ctx context.Context, dto model.CloudAccount) error
	DeleteForce(ctx context.Context, cloudAccountId uuid.UUID) error
}

type CloudAccountUsecase struct {
	repo        repository.ICloudAccountRepository
	clusterRepo repository.IClusterRepository
	argo        argowf.ArgoClient
}

func NewCloudAccountUsecase(r repository.Repository, argoClient argowf.ArgoClient) ICloudAccountUsecase {
	return &CloudAccountUsecase{
		repo:        r.CloudAccount,
		clusterRepo: r.Cluster,
		argo:        argoClient,
	}
}

func (u *CloudAccountUsecase) Create(ctx context.Context, dto model.CloudAccount) (cloudAccountId uuid.UUID, err error) {
	user, ok := request.UserFrom(ctx)
	if !ok {
		return uuid.Nil, httpErrors.NewBadRequestError(fmt.Errorf("Invalid token"), "", "")
	}
	userId := user.GetUserId()

	dto.Resource = "TODO server result or additional information"
	dto.CreatorId = &userId

	_, err = u.GetByName(ctx, dto.OrganizationId, dto.Name)
	if err == nil {
		return uuid.Nil, httpErrors.NewBadRequestError(httpErrors.DuplicateResource, "", "조직내에 동일한 이름의 클라우드 어카운트가 존재합니다.")
	}
	_, err = u.GetByAwsAccountId(ctx, dto.AwsAccountId)
	if err == nil {
		return uuid.Nil, httpErrors.NewBadRequestError(httpErrors.DuplicateResource, "", "사용 중인 AwsAccountId 입니다. 관리자에게 문의하세요.")
	}

	cloudAccountId, err = u.repo.Create(ctx, dto)
	if err != nil {
		return uuid.Nil, httpErrors.NewInternalServerError(err, "", "")
	}
	log.Info(ctx, "newly created CloudAccount ID:", cloudAccountId)

	// FOR TEST. ADD MAGIC KEYWORD
	if strings.Contains(dto.Name, domain.CLOUD_ACCOUNT_INCLUSTER) {
		if err := u.repo.InitWorkflow(ctx, cloudAccountId, "", domain.CloudAccountStatus_CREATED); err != nil {
			return uuid.Nil, errors.Wrap(err, "Failed to initialize status")
		}
		return cloudAccountId, nil
	}

	workflowId, err := u.argo.SumbitWorkflowFromWftpl(
		"tks-create-aws-cloud-account",
		argowf.SubmitOptions{
			Parameters: []string{
				"aws_region=" + "ap-northeast-2",
				"tks_cloud_account_id=" + cloudAccountId.String(),
				"aws_account_id=" + dto.AwsAccountId,
				"aws_access_key_id=" + dto.AccessKeyId,
				"aws_secret_access_key=" + dto.SecretAccessKey,
				"aws_session_token=" + dto.SessionToken,
			},
		})
	if err != nil {
		log.ErrorWithContext(ctx, "failed to submit argo workflow template. err : ", err)
		return uuid.Nil, fmt.Errorf("Failed to call argo workflow : %s", err)
	}
	log.Info(ctx, "submited workflow :", workflowId)

	if err := u.repo.InitWorkflow(ctx, cloudAccountId, workflowId, domain.CloudAccountStatus_CREATING); err != nil {
		return uuid.Nil, errors.Wrap(err, "Failed to initialize status")
	}

	return cloudAccountId, nil
}

func (u *CloudAccountUsecase) Update(ctx context.Context, dto model.CloudAccount) error {
	user, ok := request.UserFrom(ctx)
	if !ok {
		return httpErrors.NewBadRequestError(fmt.Errorf("Invalid token"), "", "")
	}
	userId := user.GetUserId()

	dto.Resource = "TODO server result or additional information"
	dto.UpdatorId = &userId
	err := u.repo.Update(ctx, dto)
	if err != nil {
		return httpErrors.NewInternalServerError(err, "", "")
	}
	return nil
}

func (u *CloudAccountUsecase) Get(ctx context.Context, cloudAccountId uuid.UUID) (res model.CloudAccount, err error) {
	res, err = u.repo.Get(ctx, cloudAccountId)
	if err != nil {
		return model.CloudAccount{}, err
	}

	res.Clusters = u.getClusterCnt(ctx, cloudAccountId)

	return
}

func (u *CloudAccountUsecase) GetByName(ctx context.Context, organizationId string, name string) (res model.CloudAccount, err error) {
	res, err = u.repo.GetByName(ctx, organizationId, name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.CloudAccount{}, httpErrors.NewNotFoundError(err, "", "")
		}
		return model.CloudAccount{}, err
	}
	res.Clusters = u.getClusterCnt(ctx, res.ID)
	return
}

func (u *CloudAccountUsecase) GetByAwsAccountId(ctx context.Context, awsAccountId string) (res model.CloudAccount, err error) {
	res, err = u.repo.GetByAwsAccountId(ctx, awsAccountId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.CloudAccount{}, httpErrors.NewNotFoundError(err, "", "")
		}
		return model.CloudAccount{}, err
	}
	res.Clusters = u.getClusterCnt(ctx, res.ID)
	return
}

func (u *CloudAccountUsecase) Fetch(ctx context.Context, organizationId string, pg *pagination.Pagination) (cloudAccounts []model.CloudAccount, err error) {
	cloudAccounts, err = u.repo.Fetch(ctx, organizationId, pg)
	if err != nil {
		return nil, err
	}

	for i, cloudAccount := range cloudAccounts {
		cloudAccounts[i].Clusters = u.getClusterCnt(ctx, cloudAccount.ID)
	}
	return
}

func (u *CloudAccountUsecase) Delete(ctx context.Context, dto model.CloudAccount) (err error) {
	user, ok := request.UserFrom(ctx)
	if !ok {
		return httpErrors.NewBadRequestError(fmt.Errorf("Invalid token"), "", "")
	}
	userId := user.GetUserId()

	cloudAccount, err := u.Get(ctx, dto.ID)
	if err != nil {
		return httpErrors.NewNotFoundError(err, "", "")
	}
	dto.UpdatorId = &userId

	if u.getClusterCnt(ctx, dto.ID) > 0 {
		return fmt.Errorf("사용 중인 클러스터가 있어 삭제할 수 없습니다.")
	}

	workflowId, err := u.argo.SumbitWorkflowFromWftpl(
		"tks-delete-aws-cloud-account",
		argowf.SubmitOptions{
			Parameters: []string{
				"aws_region=" + "ap-northeast-2",
				"tks_cloud_account_id=" + dto.ID.String(),
				"aws_account_id=" + cloudAccount.AwsAccountId,
				"aws_access_key_id=" + dto.AccessKeyId,
				"aws_secret_access_key=" + dto.SecretAccessKey,
				"aws_session_token=" + dto.SessionToken,
			},
		})
	if err != nil {
		log.ErrorWithContext(ctx, "failed to submit argo workflow template. err : ", err)
		return fmt.Errorf("Failed to call argo workflow : %s", err)
	}
	log.Info(ctx, "submited workflow :", workflowId)

	if err := u.repo.InitWorkflow(ctx, dto.ID, workflowId, domain.CloudAccountStatus_DELETING); err != nil {
		return errors.Wrap(err, "Failed to initialize status")
	}

	return nil
}

func (u *CloudAccountUsecase) DeleteForce(ctx context.Context, cloudAccountId uuid.UUID) (err error) {
	cloudAccount, err := u.repo.Get(ctx, cloudAccountId)
	if err != nil {
		return err
	}

	if !strings.Contains(cloudAccount.Name, domain.CLOUD_ACCOUNT_INCLUSTER) &&
		cloudAccount.Status != domain.CloudAccountStatus_CREATE_ERROR {
		return fmt.Errorf("The status is not CREATE_ERROR. %s", cloudAccount.Status)
	}

	if u.getClusterCnt(ctx, cloudAccountId) > 0 {
		return fmt.Errorf("사용 중인 클러스터가 있어 삭제할 수 없습니다.")
	}

	err = u.repo.Delete(ctx, cloudAccountId)
	if err != nil {
		return err
	}

	return nil
}

func (u *CloudAccountUsecase) GetResourceQuota(ctx context.Context, cloudAccountId uuid.UUID) (available bool, out domain.ResourceQuota, err error) {
	cloudAccount, err := u.repo.Get(ctx, cloudAccountId)
	if err != nil {
		return false, out, err
	}

	awsAccessKeyId, awsSecretAccessKey, _ := kubernetes.GetAwsSecret()
	if err != nil || awsAccessKeyId == "" || awsSecretAccessKey == "" {
		log.ErrorWithContext(ctx, err)
		return false, out, httpErrors.NewInternalServerError(fmt.Errorf("Invalid aws secret."), "", "")
	}

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID: awsAccessKeyId, SecretAccessKey: awsSecretAccessKey,
			},
		}))
	if err != nil {
		log.ErrorWithContext(ctx, err)
	}

	stsSvc := sts.NewFromConfig(cfg)

	if !strings.Contains(cloudAccount.Name, domain.CLOUD_ACCOUNT_INCLUSTER) {
		log.Info(ctx, "Use assume role. awsAccountId : ", cloudAccount.AwsAccountId)
		creds := stscreds.NewAssumeRoleProvider(stsSvc, "arn:aws:iam::"+cloudAccount.AwsAccountId+":role/controllers.cluster-api-provider-aws.sigs.k8s.io")
		cfg.Credentials = aws.NewCredentialsCache(creds)
	}
	client := servicequotas.NewFromConfig(cfg)

	quotaMap := map[string]string{
		"L-69A177A2": "elasticloadbalancing", // NLB
		"L-E9E9831D": "elasticloadbalancing", // Classic
		"L-A4707A72": "vpc",                  // IGW
		"L-1194D53C": "eks",                  // Cluster
		"L-0263D0A3": "ec2",                  // Elastic IP
	}

	// current usage
	type CurrentUsage struct {
		NLB     int
		CLB     int
		IGW     int
		Cluster int
		EIP     int
	}

	out.Quotas = make([]domain.ResourceQuotaAttr, 0)

	// get current usage
	currentUsage := CurrentUsage{}
	{
		c := elasticloadbalancingv2.NewFromConfig(cfg)
		pageSize := int32(100)
		res, err := c.DescribeLoadBalancers(ctx, &elasticloadbalancingv2.DescribeLoadBalancersInput{
			PageSize: &pageSize,
		}, func(o *elasticloadbalancingv2.Options) {
			o.Region = "ap-northeast-2"
		})
		if err != nil {
			return false, out, err
		}

		for _, elb := range res.LoadBalancers {
			switch elb.Type {
			case "network":
				currentUsage.NLB += 1
			}
		}
	}

	{
		c := elasticloadbalancing.NewFromConfig(cfg)
		pageSize := int32(100)
		res, err := c.DescribeLoadBalancers(ctx, &elasticloadbalancing.DescribeLoadBalancersInput{
			PageSize: &pageSize,
		}, func(o *elasticloadbalancing.Options) {
			o.Region = "ap-northeast-2"
		})
		if err != nil {
			return false, out, err
		}
		currentUsage.CLB = len(res.LoadBalancerDescriptions)
	}

	{
		c := ec2.NewFromConfig(cfg)
		res, err := c.DescribeInternetGateways(ctx, &ec2.DescribeInternetGatewaysInput{}, func(o *ec2.Options) {
			o.Region = "ap-northeast-2"
		})
		if err != nil {
			return false, out, err
		}
		currentUsage.IGW = len(res.InternetGateways)
	}

	{
		c := eks.NewFromConfig(cfg)
		res, err := c.ListClusters(ctx, &eks.ListClustersInput{}, func(o *eks.Options) {
			o.Region = "ap-northeast-2"
		})
		if err != nil {
			return false, out, err
		}
		currentUsage.Cluster = len(res.Clusters)
	}

	{
		c := ec2.NewFromConfig(cfg)
		res, err := c.DescribeAddresses(ctx, &ec2.DescribeAddressesInput{}, func(o *ec2.Options) {
			o.Region = "ap-northeast-2"
		})
		if err != nil {
			log.ErrorWithContext(ctx, err)
			return false, out, err
		}
		currentUsage.EIP = len(res.Addresses)
	}

	for key, val := range quotaMap {
		res, err := getServiceQuota(client, key, val)
		if err != nil {
			return false, out, err
		}
		log.DebugfWithContext(ctx, "%s %s %v", *res.Quota.QuotaName, *res.Quota.QuotaCode, *res.Quota.Value)

		quotaValue := int(*res.Quota.Value)

		// stack 1개 생성하는데 필요한 quota
		// Classic 1
		// Network 5
		// IGW 1
		// EIP 3
		// Cluster 1
		switch key {
		case "L-69A177A2": // NLB
			log.InfofWithContext(ctx, "NLB : usage %d, quota %d", currentUsage.NLB, quotaValue)
			out.Quotas = append(out.Quotas, domain.ResourceQuotaAttr{
				Type:     "NLB",
				Usage:    currentUsage.NLB,
				Quota:    quotaValue,
				Required: 5,
			})
			if quotaValue < currentUsage.NLB+5 {
				available = false
			}
		case "L-E9E9831D": // Classic
			log.InfofWithContext(ctx, "CLB : usage %d, quota %d", currentUsage.CLB, quotaValue)
			out.Quotas = append(out.Quotas, domain.ResourceQuotaAttr{
				Type:     "CLB",
				Usage:    currentUsage.CLB,
				Quota:    quotaValue,
				Required: 1,
			})
			if quotaValue < currentUsage.CLB+1 {
				available = false
			}
		case "L-A4707A72": // IGW
			log.InfofWithContext(ctx, "IGW : usage %d, quota %d", currentUsage.IGW, quotaValue)
			out.Quotas = append(out.Quotas, domain.ResourceQuotaAttr{
				Type:     "IGW",
				Usage:    currentUsage.IGW,
				Quota:    quotaValue,
				Required: 1,
			})
			if quotaValue < currentUsage.IGW+1 {
				available = false
			}
		case "L-1194D53C": // Cluster
			log.InfofWithContext(ctx, "Cluster : usage %d, quota %d", currentUsage.Cluster, quotaValue)
			out.Quotas = append(out.Quotas, domain.ResourceQuotaAttr{
				Type:     "EKS",
				Usage:    currentUsage.Cluster,
				Quota:    quotaValue,
				Required: 1,
			})
			if quotaValue < currentUsage.Cluster+1 {
				available = false
			}
		case "L-0263D0A3": // Elastic IP
			log.InfofWithContext(ctx, "Elastic IP : usage %d, quota %d", currentUsage.EIP, quotaValue)
			out.Quotas = append(out.Quotas, domain.ResourceQuotaAttr{
				Type:     "EIP",
				Usage:    currentUsage.EIP,
				Quota:    quotaValue,
				Required: 3,
			})
			if quotaValue < currentUsage.EIP+3 {
				available = false
			}
		}

	}

	//return fmt.Errorf("Always return err")
	return true, out, nil
}

func (u *CloudAccountUsecase) getClusterCnt(ctx context.Context, cloudAccountId uuid.UUID) (cnt int) {
	cnt = 0

	clusters, err := u.clusterRepo.FetchByCloudAccountId(ctx, cloudAccountId, nil)
	if err != nil {
		log.Error("Failed to get clusters by cloudAccountId. err : ", err)
		return cnt
	}

	for _, cluster := range clusters {
		if cluster.Status != domain.ClusterStatus_DELETED {
			cnt++
		}
	}

	return cnt
}

func getServiceQuota(client *servicequotas.Client, quotaCode string, serviceCode string) (res *servicequotas.GetServiceQuotaOutput, err error) {
	res, err = client.GetServiceQuota(context.TODO(), &servicequotas.GetServiceQuotaInput{
		QuotaCode:   &quotaCode,
		ServiceCode: &serviceCode,
	}, func(o *servicequotas.Options) {
		o.Region = "ap-northeast-2"
	})
	if err != nil {
		return nil, err
	}
	return
}
