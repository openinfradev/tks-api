package usecase

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/domain"
	"github.com/openinfradev/tks-api/internal/repository"
	argowf "github.com/openinfradev/tks-api/pkg/argo-client"
	"github.com/openinfradev/tks-common/pkg/log"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

type IClusterUsecase interface {
	WithTrx(*gorm.DB) IClusterUsecase
	Fetch(projectId string) ([]domain.Cluster, error)
	Create(projectId string, templateId string, name string, conf domain.ClusterConf, creatorId string, description string) (clusterId string, err error)
	Get(clusterId string) (out domain.Cluster, err error)
	Delete(clusterId string) (err error)
}

type ClusterUsecase struct {
	repo repository.IClusterRepository
	argo argowf.ArgoClient
}

func NewClusterUsecase(r repository.IClusterRepository, argoClient argowf.ArgoClient) IClusterUsecase {
	return &ClusterUsecase{
		repo: r,
		argo: argoClient,
	}
}

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

func (u *ClusterUsecase) WithTrx(trxHandle *gorm.DB) IClusterUsecase {
	u.repo = u.repo.WithTrx(trxHandle)
	return u
}

func (u *ClusterUsecase) Fetch(projectId string) (out []domain.Cluster, err error) {
	if projectId == "" {
		out, err = u.repo.Fetch()
	} else {
		out, err = u.repo.FetchByContractId(projectId)
	}

	if err != nil {
		return nil, err
	}
	return out, nil
}

func (u *ClusterUsecase) Create(projectId string, templateId string, name string, conf domain.ClusterConf, creatorId string, description string) (clusterId string, err error) {
	creator := uuid.Nil
	if creatorId != "" {
		creator, err = uuid.Parse(creatorId)
		if err != nil {
			return "", fmt.Errorf("Invalid Creator ID %s", creatorId)
		}
	}

	/***************************
	 * Pre-process cluster conf *
	 ***************************/
	clConf, err := u.constructClusterConf(&conf)
	if err != nil {
		return "", err
	}

	clusterId, err = u.repo.Create(projectId, templateId, name, clConf, creator, description)
	if err != nil {
		return "", fmt.Errorf("Failed to create cluster")
	}

	// Call argo workflow
	workflowId, err := u.argo.SumbitWorkflowFromWftpl(
		"create-tks-usercluster",
		argowf.SubmitOptions{
			Parameters: []string{
				"contract_id=" + projectId,
				"cluster_id=" + clusterId,
				"site_name=" + clusterId,
				"template_name=" + templateId,
				"git_account=" + viper.GetString("git-account"),
				"manifest_repo_url=" + viper.GetString("git-base-url") + "/" + viper.GetString("git-account") + "/" + clusterId + "-manifests",
				"revision=" + viper.GetString("revision"),
			},
		})
	log.Info("submited workflow: ", workflowId)

	if err != nil {
		log.Error("failed to submit argo workflow template. err : ", err)
		return "", fmt.Errorf("Failed to call argo workflow : %s", err)
	}
	log.Info("Successfully submited workflow: ", workflowId)

	return clusterId, nil
}

func (u *ClusterUsecase) Get(clusterId string) (out domain.Cluster, err error) {
	cluster, err := u.repo.Get(clusterId)
	if err != nil {
		return domain.Cluster{}, err
	}
	return cluster, nil
}

func (u *ClusterUsecase) Delete(clusterId string) (err error) {
	_, err = u.repo.Get(clusterId)
	if err != nil {
		return fmt.Errorf("No cluster for deletiing : %s", clusterId)
	}

	err = u.repo.Delete(clusterId)
	if err != nil {
		return fmt.Errorf("Fatiled to deleting cluster : %s", clusterId)
	}

	return nil
}

func (u *ClusterUsecase) constructClusterConf(rawConf *domain.ClusterConf) (clusterConf *domain.ClusterConf, err error) {
	region := "ap-northeast-2"
	if rawConf != nil && rawConf.Region != "" {
		region = rawConf.Region
	}

	numOfAz := 1
	if rawConf != nil && rawConf.NumOfAz != 0 {
		numOfAz = int(rawConf.NumOfAz)

		if numOfAz > 3 {
			log.Error("Error: numOfAz cannot exceed 3.")
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
			log.Debug("Found region : ", key)
			maxAzForSelectedRegion = val
			log.Debug("Trimmed azNum var: ", maxAzForSelectedRegion)
			found = true
		}
	}

	if !found {
		log.Error("Couldn't find entry for region ", region)
	}

	if numOfAz > maxAzForSelectedRegion {
		log.Error("Invalid numOfAz: exceeded the number of Az in region ", region)
		temp_err := fmt.Errorf("Invalid numOfAz: exceeded the number of Az in region %s", region)
		return nil, temp_err
	}

	// Validate if machineReplicas is multiple of number of AZ
	replicas := int(rawConf.MachineReplicas)
	if replicas == 0 {
		log.Debug("No machineReplicas param. Using default values..")
	} else {
		if remainder := replicas % numOfAz; remainder != 0 {
			log.Error("Invalid machineReplicas: it should be multiple of numOfAz ", numOfAz)
			temp_err := fmt.Errorf("Invalid machineReplicas: it should be multiple of numOfAz %d", numOfAz)
			return nil, temp_err
		} else {
			log.Debug("Valid replicas and numOfAz. Caculating minSize & maxSize..")
			minSizePerAz = int(replicas / numOfAz)
			maxSizePerAz = minSizePerAz * 5

			// Validate if maxSizePerAx is within allowed range
			if maxSizePerAz > MAX_SIZE_PER_AZ {
				fmt.Printf("maxSizePerAz exceeded maximum value %d, so adjusted to %d", MAX_SIZE_PER_AZ, MAX_SIZE_PER_AZ)
				maxSizePerAz = MAX_SIZE_PER_AZ
			}
			log.Debug("Derived minSizePerAz: ", minSizePerAz)
			log.Debug("Derived maxSizePerAz: ", maxSizePerAz)
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
