package usecase

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/keycloak"
	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-api/internal/serializer"
	argowf "github.com/openinfradev/tks-api/pkg/argo-client"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/kubernetes"
	"github.com/openinfradev/tks-api/pkg/log"
	thanos "github.com/openinfradev/tks-api/pkg/thanos-client"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ProjectAll int = iota
	ProjectLeader
	ProjectMember
	ProjectViewer
)

type IProjectUsecase interface {
	CreateProject(ctx context.Context, p *model.Project) (string, error)
	GetProjects(ctx context.Context, organizationId string, userId string, onlyMyProject bool, projectName string, pg *pagination.Pagination) ([]domain.ProjectResponse, error)
	GetProject(ctx context.Context, organizationId string, projectId string) (*model.Project, error)
	GetProjectWithLeader(ctx context.Context, organizationId string, projectId string) (*model.Project, error)
	IsProjectNameExist(ctx context.Context, organizationId string, projectName string) (bool, error)
	UpdateProject(ctx context.Context, p *model.Project, newLeaderId string) error
	GetProjectRole(ctx context.Context, id string) (*model.ProjectRole, error)
	GetProjectRoles(ctx context.Context, query int) ([]model.ProjectRole, error)
	AddProjectMember(ctx context.Context, organizationId string, pm *model.ProjectMember) (string, error)
	GetProjectUser(ctx context.Context, projectUserId string) (*model.ProjectUser, error)
	GetProjectMember(ctx context.Context, projectMemberId string) (*model.ProjectMember, error)
	GetProjectMemberByUserId(ctx context.Context, projectId string, userId string) (*model.ProjectMember, error)
	GetProjectMembers(ctx context.Context, projectId string, query int, pg *pagination.Pagination) ([]model.ProjectMember, error)
	GetProjectMemberCount(ctx context.Context, projectMemberId string) (*domain.GetProjectMemberCountResponse, error)
	RemoveProjectMember(ctx context.Context, organizationId string, projectMemberId string) error
	UpdateProjectMemberRole(ctx context.Context, pm *model.ProjectMember) error
	CreateProjectNamespace(ctx context.Context, organizationId string, pn *model.ProjectNamespace) error
	IsProjectNamespaceExist(ctx context.Context, organizationId string, projectId string, stackId string, projectNamespace string) (bool, error)
	GetProjectNamespaces(ctx context.Context, organizationId string, projectId string, pg *pagination.Pagination) ([]model.ProjectNamespace, error)
	GetProjectNamespace(ctx context.Context, organizationId string, projectId string, projectNamespace string, stackId string) (*model.ProjectNamespace, error)
	UpdateProjectNamespace(ctx context.Context, pn *model.ProjectNamespace) error
	DeleteProjectNamespace(ctx context.Context, organizationId string, projectId string, projectNamespace string, stackId string) error
	GetAppCount(ctx context.Context, organizationId string, projectId string, namespace string) (appCount int, err error)
	EnsureNamespaceForCluster(ctx context.Context, organizationId string, stackId string, namespaceName string) error
	EnsureRequiredSetupForCluster(ctx context.Context, organizationId string, projectId string, stackId string) error
	MayRemoveRequiredSetupForCluster(ctx context.Context, organizationId string, projectId string, stackId string) error
	CreateK8SNSRoleBinding(ctx context.Context, organizationId string, projectId string, stackId string, namespace string) error
	DeleteK8SNSRoleBinding(ctx context.Context, organizationId string, projectId string, stackId string, namespace string) error
	GetProjectNamespaceKubeconfig(ctx context.Context, organizationId string, projectId string, namespace string, stackId domain.StackId) (string, error)
	GetProjectKubeconfig(ctx context.Context, organizationId string, projectId string) (string, error)
	GetK8sResources(ctx context.Context, organizationId string, projectId string, namespace string, stackId domain.StackId) (out domain.ProjectNamespaceK8sResources, err error)
	GetResourcesUsage(ctx context.Context, thanosClient thanos.ThanosClient, organizationId string, projectId string, namespace string, stackId domain.StackId) (out domain.ProjectNamespaceResourcesUsage, err error)
	AssignKeycloakClientRoleToMember(ctx context.Context, organizationId string, projectId string, clientId string, projectMemberId string) error
	UnassignKeycloakClientRoleToMember(ctx context.Context, organizationId string, projectId string, clientId string, projectMemberId string) error
}

type ProjectUsecase struct {
	projectRepo            repository.IProjectRepository
	userRepository         repository.IUserRepository
	authRepository         repository.IAuthRepository
	clusterRepository      repository.IClusterRepository
	appgroupRepository     repository.IAppGroupRepository
	organizationRepository repository.IOrganizationRepository
	argo                   argowf.ArgoClient
	kc                     keycloak.IKeycloak
}

func NewProjectUsecase(r repository.Repository, kc keycloak.IKeycloak, argoClient argowf.ArgoClient) IProjectUsecase {
	return &ProjectUsecase{
		projectRepo:            r.Project,
		userRepository:         r.User,
		authRepository:         r.Auth,
		clusterRepository:      r.Cluster,
		appgroupRepository:     r.AppGroup,
		organizationRepository: r.Organization,
		argo:                   argoClient,
		kc:                     kc,
	}
}

func (u *ProjectUsecase) CreateProject(ctx context.Context, p *model.Project) (string, error) {
	projectId, err := u.projectRepo.CreateProject(ctx, p)
	if err != nil {
		log.Error(ctx, err)
		return "", errors.Wrap(err, "Failed to create project.")
	}

	prs, err := u.GetProjectRoles(ctx, ProjectAll)
	if err != nil {
		log.Error(ctx, err)
		return "", errors.Wrap(err, "Failed to retrieve project roles.")
	}
	log.Debugf(ctx, "Project roles: %v", prs)
	for _, pr := range prs {
		log.Debugf(ctx, "Start to create Project role: %s. orgId: %s, projectId: %s", pr.Name, p.OrganizationId, projectId)
		err = u.kc.EnsureClientRoleWithClientName(ctx, p.OrganizationId, keycloak.DefaultClientID, pr.Name+"@"+projectId)
		if err != nil {
			log.Error(ctx, err)
			return "", errors.Wrap(err, "Failed to create project setting on keycloak.")
		}
	}

	return projectId, nil
}

func (u *ProjectUsecase) GetProjects(ctx context.Context, organizationId string, userId string, onlyMyProject bool, projectName string, pg *pagination.Pagination) (pr []domain.ProjectResponse, err error) {
	if userId == "" {
		if pr, err = u.projectRepo.GetAllProjects(ctx, organizationId, projectName, pg); err != nil {
			log.Error(ctx, err)
			return nil, errors.Wrap(err, "Failed to get projects.")
		}
	} else {
		userUuid, err := uuid.Parse(userId)
		if err != nil {
			log.Error(ctx, err)
			return nil, errors.Wrap(err, "Failed to parse uuid to string")
		}
		if !onlyMyProject {
			pr, err = u.projectRepo.GetProjects(ctx, organizationId, userUuid, projectName, pg)
		} else {
			pr, err = u.projectRepo.GetProjectsByUserId(ctx, organizationId, userUuid, projectName, pg)
		}
		if err != nil {
			log.Error(ctx, err)
			return nil, errors.Wrap(err, "Failed to get projects.")
		}
	}

	return pr, err
}

func (u *ProjectUsecase) GetProject(ctx context.Context, organizationId string, projectId string) (*model.Project, error) {
	p, err := u.projectRepo.GetProjectById(ctx, organizationId, projectId)
	if err != nil {
		log.Error(ctx, err)
		return nil, errors.Wrap(err, "Failed to get projects.")
	}
	return p, err
}

func (u *ProjectUsecase) GetProjectWithLeader(ctx context.Context, organizationId string, projectId string) (*model.Project, error) {
	p, err := u.projectRepo.GetProjectByIdAndLeader(ctx, organizationId, projectId)
	if err != nil {
		log.Error(ctx, err)
		return nil, errors.Wrap(err, "Failed to get projects.")
	}
	return p, err
}

func (u *ProjectUsecase) IsProjectNameExist(ctx context.Context, organizationId string, projectName string) (bool, error) {
	exist := true
	p, err := u.projectRepo.GetProjectByName(ctx, organizationId, projectName)
	if err != nil {
		log.Error(ctx, err)
		exist = false
		return exist, errors.Wrap(err, "Failed to retrieve project name.")
	}
	if p == nil {
		exist = false
	}
	return exist, nil
}

func (u *ProjectUsecase) UpdateProject(ctx context.Context, p *model.Project, newLeaderId string) error {
	//TODO: [donggyu] have to implementation about un/assigning client roles

	var currentMemberId, currentLeaderId, projectRoleId string
	for _, pm := range p.ProjectMembers {
		currentMemberId = pm.ID
		currentLeaderId = pm.ProjectUser.ID.String()
		projectRoleId = pm.ProjectRole.ID
	}
	p.ProjectNamespaces = nil
	p.ProjectMembers = nil

	// only project table update
	if err := u.projectRepo.UpdateProject(ctx, p); err != nil {
		log.Error(ctx, err)
		return errors.Wrap(err, "Failed to update project.")
	}

	if newLeaderId != "" && currentLeaderId != newLeaderId {
		pu, err := u.GetProjectUser(ctx, newLeaderId)
		if err != nil {
			return err
		}
		if pu == nil {
			return errors.Wrap(err, "The user doesn't exist.")
		}

		// If project leader exists, remove leader
		if currentMemberId != "" {
			if err := u.RemoveProjectMember(ctx, p.OrganizationId, currentMemberId); err != nil {
				log.Error(ctx, err)
				return errors.Wrap(err, "Failed to remove project member.")
			}
		}

		if projectRoleId == "" {
			pr, err := u.projectRepo.GetProjectRoleByName(ctx, "project-leader")
			if err != nil {
				return err
			}
			projectRoleId = pr.ID
		}

		// If the Member does not exist in the Project, insert it, but if it does, update it.
		pm, err := u.projectRepo.GetProjectMemberByUserId(ctx, p.ID, newLeaderId)
		if err != nil {
			return err
		}
		if pm == nil {
			newPm := &model.ProjectMember{
				ProjectId:       p.ID,
				ProjectUserId:   pu.ID,
				ProjectUser:     nil,
				ProjectRoleId:   projectRoleId,
				ProjectRole:     nil,
				IsProjectLeader: true,
				CreatedAt:       *p.UpdatedAt,
			}
			res, err := u.AddProjectMember(ctx, p.OrganizationId, newPm)
			if err != nil {
				return err
			}
			log.Infof(ctx, "Added project member: %s", res)
		} else {
			pm.ProjectUserId = pu.ID
			pm.ProjectRoleId = projectRoleId
			pm.IsProjectLeader = true
			pm.UpdatedAt = p.UpdatedAt
			pm.ProjectUser = nil
			pm.ProjectRole = nil
			err := u.UpdateProjectMemberRole(ctx, pm)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (u *ProjectUsecase) GetProjectRole(ctx context.Context, id string) (*model.ProjectRole, error) {
	pr, err := u.projectRepo.GetProjectRoleById(ctx, id)
	if err != nil {
		log.Error(ctx, err)
		return nil, errors.Wrap(err, "Failed to get project roles.")
	}

	return pr, nil
}

func (u *ProjectUsecase) GetProjectRoles(ctx context.Context, query int) (prs []model.ProjectRole, err error) {
	var pr *model.ProjectRole

	if query == ProjectLeader {
		pr, err = u.projectRepo.GetProjectRoleByName(ctx, "project-leader")
	} else if query == ProjectMember {
		pr, err = u.projectRepo.GetProjectRoleByName(ctx, "project-member")
	} else if query == ProjectViewer {
		pr, err = u.projectRepo.GetProjectRoleByName(ctx, "project-viewer")
	} else {
		prs, err = u.projectRepo.GetAllProjectRoles(ctx)
	}
	if err != nil {
		log.Error(ctx, err)
		return nil, errors.Wrap(err, "Failed to get project roles.")
	}

	if pr != nil {
		prs = append(prs, *pr)
	}

	return prs, nil
}

func (u *ProjectUsecase) AddProjectMember(ctx context.Context, organizationId string, pm *model.ProjectMember) (string, error) {
	projectMemberId, err := u.projectRepo.AddProjectMember(ctx, pm)
	if err != nil {
		log.Error(ctx, err)
		return "", errors.Wrap(err, "Failed to add project member to project.")
	}

	return projectMemberId, nil
}

func (u *ProjectUsecase) GetProjectUser(ctx context.Context, projectUserId string) (*model.ProjectUser, error) {
	var uid uuid.UUID
	uid, err := uuid.Parse(projectUserId)
	if err != nil {
		log.Error(ctx, err)
		return nil, errors.Wrap(err, "Failed to parse uuid to string")
	}

	user, err := u.userRepository.GetByUuid(ctx, uid)
	if err != nil {
		log.Error(ctx, err)
		return nil, errors.Wrap(err, "Failed to retrieve user by id")
	}
	var pu model.ProjectUser
	if err = serializer.Map(ctx, user, &pu); err != nil {
		log.Error(ctx, err)
		return nil, err
	}
	return &pu, nil
}

func (u *ProjectUsecase) GetProjectMember(ctx context.Context, projectMemberId string) (pm *model.ProjectMember, err error) {
	pm, err = u.projectRepo.GetProjectMemberById(ctx, projectMemberId)
	if err != nil {
		log.Error(ctx, err)
		return pm, errors.Wrap(err, "Failed to get project member.")
	}

	return pm, nil
}

func (u *ProjectUsecase) GetProjectMemberByUserId(ctx context.Context, projectId string, userId string) (pm *model.ProjectMember, err error) {
	pm, err = u.projectRepo.GetProjectMemberByUserId(ctx, projectId, userId)
	if err != nil {
		log.Error(ctx, err)
		return pm, errors.Wrap(err, "Failed to get project member.")
	}

	return pm, nil
}

func (u *ProjectUsecase) GetProjectMembers(ctx context.Context, projectId string, query int, pg *pagination.Pagination) (pms []model.ProjectMember, err error) {
	if query == ProjectLeader {
		pms, err = u.projectRepo.GetProjectMembersByProjectIdAndRoleName(ctx, projectId, "project-leader", pg)
	} else if query == ProjectMember {
		pms, err = u.projectRepo.GetProjectMembersByProjectIdAndRoleName(ctx, projectId, "project-member", pg)
	} else if query == ProjectViewer {
		pms, err = u.projectRepo.GetProjectMembersByProjectIdAndRoleName(ctx, projectId, "project-viewer", pg)
	} else {
		pms, err = u.projectRepo.GetProjectMembersByProjectId(ctx, projectId, pg)
	}
	if err != nil {
		log.Error(ctx, err)
		return nil, errors.Wrap(err, "Failed to get project members.")
	}

	return pms, nil
}

func (u *ProjectUsecase) GetProjectMemberCount(ctx context.Context, projectMemberId string) (pmcr *domain.GetProjectMemberCountResponse, err error) {
	pmcr, err = u.projectRepo.GetProjectMemberCountByProjectId(ctx, projectMemberId)
	if err != nil {
		log.Error(ctx, err)
		return pmcr, errors.Wrap(err, "Failed to get project member count.")
	}

	return pmcr, nil
}

func (u *ProjectUsecase) RemoveProjectMember(ctx context.Context, organizationId string, projectMemberId string) error {
	pm, err := u.GetProjectMember(ctx, projectMemberId)
	if err != nil {
		log.Error(ctx, err)
		return errors.Wrap(err, "Failed to get project member to project.")
	}
	// unassign keycloak client role to member
	err = u.kc.UnassignClientRoleToUser(ctx, organizationId, pm.ProjectUserId.String(), keycloak.DefaultClientID, pm.ProjectRole.Name+"@"+pm.ProjectId)
	if err != nil {
		log.Error(ctx, err)
		return errors.Wrap(err, "Failed to remove project member to project.")
	}

	if err := u.projectRepo.RemoveProjectMember(ctx, projectMemberId); err != nil {
		log.Error(ctx, err)
		return errors.Wrap(err, "Failed to remove project member to project.")
	}

	return nil
}

func (u *ProjectUsecase) UpdateProjectMemberRole(ctx context.Context, pm *model.ProjectMember) error {

	if err := u.projectRepo.UpdateProjectMemberRole(ctx, pm); err != nil {
		log.Error(ctx, err)
		return errors.Wrap(err, "Failed to remove project member to project.")
	}
	return nil
}

func (u *ProjectUsecase) CreateProjectNamespace(ctx context.Context, organizationId string, pn *model.ProjectNamespace) error {
	if err := u.projectRepo.CreateProjectNamespace(ctx, organizationId, pn); err != nil {
		log.Error(ctx, err)
		return errors.Wrap(err, "Failed to create project namespace.")
	}
	return nil
}

func (u *ProjectUsecase) IsProjectNamespaceExist(ctx context.Context, organizationId string, projectId string, stackId string, projectNamespace string) (bool, error) {
	exist := true
	pn, err := u.projectRepo.GetProjectNamespaceByName(ctx, organizationId, projectId, stackId, projectNamespace)
	if err != nil {
		log.Error(ctx, err)
		exist = false
		return exist, errors.Wrap(err, "Failed to retrieve project namespace.")
	}
	if pn == nil {
		exist = false
	}
	return exist, nil
}

func (u *ProjectUsecase) GetProjectNamespaces(ctx context.Context, organizationId string, projectId string, pg *pagination.Pagination) ([]model.ProjectNamespace, error) {
	pns, err := u.projectRepo.GetProjectNamespaces(ctx, organizationId, projectId, pg)
	if err != nil {
		log.Error(ctx, err)
		return nil, errors.Wrap(err, "Failed to retrieve project namespaces.")
	}

	for _, pn := range pns {
		o, err := u.organizationRepository.Get(ctx, organizationId)
		if err != nil {
			log.Error(ctx, err)
			return nil, errors.Wrap(err, "Failed to get organization.")
		}

		appGroupsInPrimaryCluster, err := u.appgroupRepository.Fetch(ctx, domain.ClusterId(o.PrimaryClusterId), nil)
		if err != nil {
			log.Error(ctx, err)
			return nil, errors.Wrap(err, "Failed to get app groups.")
		}

		for i, appGroup := range appGroupsInPrimaryCluster {
			if appGroup.AppGroupType == domain.AppGroupType_LMA {
				applications, err := u.appgroupRepository.GetApplications(ctx, appGroup.ID, domain.ApplicationType_GRAFANA)
				if err != nil {
					log.Error(ctx, err)
					return nil, errors.Wrap(err, "Failed to get applications.")
				}
				if len(applications) > 0 {
					pns[i].GrafanaUrl = applications[0].Endpoint + "/d/tks_namespace_dashboard/tks-kubernetes-view-namespaces?orgId=" + organizationId + "&var-datasource=Prometheus&var-taco_cluster=" + pn.StackId + "&var-namespace=" + pn.Namespace
				}
			}
		}
	}

	return pns, nil
}

func (u *ProjectUsecase) GetProjectNamespace(ctx context.Context, organizationId string, projectId string, projectNamespace string, stackId string) (*model.ProjectNamespace, error) {
	pn, err := u.projectRepo.GetProjectNamespaceByPrimaryKey(ctx, organizationId, projectId, projectNamespace, stackId)
	if err != nil {
		log.Error(ctx, err)
		return nil, errors.Wrap(err, "Failed to retrieve project namespace.")
	}
	o, err := u.organizationRepository.Get(ctx, organizationId)
	if err != nil {
		log.Error(ctx, err)
		return nil, errors.Wrap(err, "Failed to get organization.")
	}

	appGroupsInPrimaryCluster, err := u.appgroupRepository.Fetch(ctx, domain.ClusterId(o.PrimaryClusterId), nil)
	if err != nil {
		log.Error(ctx, err)
		return nil, errors.Wrap(err, "Failed to get app groups.")
	}

	for _, appGroup := range appGroupsInPrimaryCluster {
		if appGroup.AppGroupType == domain.AppGroupType_LMA {
			applications, err := u.appgroupRepository.GetApplications(ctx, appGroup.ID, domain.ApplicationType_GRAFANA)
			if err != nil {
				log.Error(ctx, err)
				return nil, errors.Wrap(err, "Failed to get applications.")
			}
			if len(applications) > 0 {
				pn.GrafanaUrl = applications[0].Endpoint + "/d/tks_namespace_dashboard/tks-kubernetes-view-namespaces?orgId=" + organizationId + "&var-datasource=Prometheus&var-taco_cluster=" + stackId + "&var-namespace=" + projectNamespace
			}
		}
	}
	return pn, nil
}

func (u *ProjectUsecase) UpdateProjectNamespace(ctx context.Context, pn *model.ProjectNamespace) error {
	if err := u.projectRepo.UpdateProjectNamespace(ctx, pn); err != nil {
		log.Error(ctx, err)
		return errors.Wrap(err, "Failed to update project namespace")
	}
	return nil
}

func (u *ProjectUsecase) DeleteProjectNamespace(ctx context.Context, organizationId string, projectId string,
	stackId string, projectNamespace string) error {
	if err := u.projectRepo.DeleteProjectNamespace(ctx, organizationId, projectId, projectNamespace, stackId); err != nil {
		log.Error(ctx, err)
		return errors.Wrap(err, "Failed to delete project namespace.")
	}
	return nil
}

func (u *ProjectUsecase) GetAppCount(ctx context.Context, organizationId string, projectId string, namespace string) (appCount int, err error) {
	if namespace == "" {
		appCount, err = u.projectRepo.GetAppCountByProjectId(ctx, organizationId, projectId)
	} else {
		appCount, err = u.projectRepo.GetAppCountByNamespace(ctx, organizationId, projectId, namespace)
	}
	if err != nil {
		log.Error(ctx, err)
		return 0, errors.Wrap(err, "Failed to retrieve app count.")
	}

	return appCount, nil
}

func (u *ProjectUsecase) EnsureRequiredSetupForCluster(ctx context.Context, organizationId string, projectId string, stackId string) error {
	_, err := u.projectRepo.GetProjectNamespaces(ctx, organizationId, projectId, nil)
	if err != nil {
		log.Error(ctx, err)
		return errors.Wrap(err, "Failed to get project namespace in database.")
	}

	if err := u.createK8SInitialResource(ctx, organizationId, projectId, stackId); err != nil {
		log.Error(ctx, err)
		return err
	}

	if err := u.createKeycloakClientRoles(ctx, organizationId, projectId, stackId+"-k8s-api"); err != nil {
		log.Error(ctx, err)
		return err
	}

	projectMembers, err := u.GetProjectMembers(ctx, projectId, ProjectAll, nil)
	if err != nil {
		log.Error(ctx, err)
		return err
	}
	for _, pm := range projectMembers {
		err = u.assignEachKeycloakClientRoleToMember(ctx, organizationId, projectId, stackId+"-k8s-api", pm.ProjectUserId.String(), pm.ProjectRole.Name)
		if err != nil {
			log.Error(ctx, err)
			return err
		}
	}

	return nil
}
func (u *ProjectUsecase) MayRemoveRequiredSetupForCluster(ctx context.Context, organizationId string, projectId string, stackId string) error {
	pns, err := u.projectRepo.GetProjectNamespaces(ctx, organizationId, projectId, nil)
	if err != nil {
		return err
	}
	var nsCount int
	for _, pn := range pns {
		if pn.StackId == stackId {
			nsCount++
		}
	}

	// if there are more than one namespace, it means that required setup is needed on the other namespace
	if nsCount > 1 {
		return nil
	}

	if err := u.deleteK8SInitialResource(ctx, organizationId, projectId, stackId); err != nil {
		return err
	}

	projectMembers, err := u.GetProjectMembers(ctx, projectId, ProjectAll, nil)
	if err != nil {
		return err
	}
	for _, pm := range projectMembers {
		err = u.unassignKeycloakClientRoleToMember(ctx, organizationId, projectId, stackId+"-k8s-api", pm.ProjectUserId.String(), pm.ProjectRole.Name)
		if err != nil {
			return err
		}
	}

	if err := u.deleteKeycloakClientRoles(ctx, organizationId, projectId, stackId+"-k8s-api"); err != nil {
		return err
	}

	return nil
}
func (u *ProjectUsecase) createK8SInitialResource(ctx context.Context, organizationId string, projectId string, stackId string) error {
	kubeconfig, err := kubernetes.GetKubeConfig(ctx, stackId, kubernetes.KubeconfigForAdmin)
	if err != nil {
		return errors.Wrap(err, "Failed to get kubeconfig.")
	}

	err = kubernetes.EnsureClusterRole(ctx, kubeconfig, projectId)
	if err != nil {
		return err
	}

	err = kubernetes.EnsureCommonClusterRole(ctx, kubeconfig, projectId)
	if err != nil {
		return err
	}

	err = kubernetes.EnsureCommonClusterRoleBinding(ctx, kubeconfig, projectId)
	if err != nil {
		return err
	}

	return nil
}
func (u *ProjectUsecase) deleteK8SInitialResource(ctx context.Context, organizationId string, projectId string, stackId string) error {
	kubeconfig, err := kubernetes.GetKubeConfig(ctx, stackId, kubernetes.KubeconfigForAdmin)
	if err != nil {
		log.Error(ctx, err)
		return errors.Wrap(err, "Failed to get kubeconfig.")
	}

	err = kubernetes.RemoveClusterRole(ctx, kubeconfig, projectId)
	if err != nil {
		log.Error(ctx, err)
		return errors.Wrap(err, "Failed to remove cluster roles.")
	}

	err = kubernetes.RemoveCommonClusterRole(ctx, kubeconfig, projectId)
	if err != nil {
		log.Error(ctx, err)
		return errors.Wrap(err, "Failed to remove common cluster roles.")
	}

	err = kubernetes.RemoveCommonClusterRoleBinding(ctx, kubeconfig, projectId)
	if err != nil {
		log.Error(ctx, err)
		return errors.Wrap(err, "Failed to remove common cluster role bindings.")
	}

	return nil
}
func (u *ProjectUsecase) createKeycloakClientRoles(ctx context.Context, organizationId string, projectId string, clientId string) error {
	// create Roles in keycloak
	roles := []string{"project-leader", "project-member", "project-viewer"}
	for _, role := range roles {
		err := u.kc.EnsureClientRoleWithClientName(ctx, organizationId, clientId, role+"@"+projectId)
		if err != nil {
			return errors.Wrap(err, "Failed to create keycloak client role.")
		}

	}
	return nil
}
func (u *ProjectUsecase) deleteKeycloakClientRoles(ctx context.Context, organizationId string, projectId string, clientId string) error {
	// first check whether the stac

	// delete Roles in keycloak
	for _, role := range []string{strconv.Itoa(ProjectLeader), strconv.Itoa(ProjectMember), strconv.Itoa(ProjectViewer)} {
		err := u.kc.DeleteClientRoleWithClientName(ctx, organizationId, clientId, role+"@"+projectId)
		if err != nil {
			log.Error(ctx, err)
			return errors.Wrap(err, "Failed to delete keycloak client roles.")
		}
	}
	return nil
}
func (u *ProjectUsecase) CreateK8SNSRoleBinding(ctx context.Context, organizationId string, projectId string, stackId string, namespace string) error {
	kubeconfig, err := kubernetes.GetKubeConfig(ctx, stackId, kubernetes.KubeconfigForAdmin)
	if err != nil {
		log.Error(ctx, err)
		return errors.Wrap(err, "Failed to get kubeconfig.")
	}

	err = kubernetes.EnsureRoleBinding(ctx, kubeconfig, projectId, namespace)
	if err != nil {
		log.Error(ctx, err)
		return errors.Wrap(err, "Failed to create K8s role binding.")
	}

	return nil
}
func (u *ProjectUsecase) DeleteK8SNSRoleBinding(ctx context.Context, organizationId string, projectId string, stackId string, namespace string) error {
	//TODO implement me
	return nil
}

func (u *ProjectUsecase) AssignKeycloakClientRoleToMember(ctx context.Context, organizationId string, projectId string, clientId string, projectMemberId string) error {
	pm, err := u.GetProjectMember(ctx, projectMemberId)
	if err != nil {
		log.Error(ctx, err)
		return errors.Wrap(err, "Failed to get project member.")
	}
	err = u.assignEachKeycloakClientRoleToMember(ctx, organizationId, projectId, clientId, pm.ProjectUserId.String(), pm.ProjectRole.Name)
	if err != nil {
		log.Error(ctx, err)
		return err
	}
	return nil
}

func (u *ProjectUsecase) assignEachKeycloakClientRoleToMember(ctx context.Context, organizationId string, projectId string, clientId string, userId string, roleName string) error {
	err := u.kc.AssignClientRoleToUser(ctx, organizationId, userId, clientId, roleName+"@"+projectId)
	if err != nil {
		log.Error(ctx, err)
		return errors.Wrap(err, "Failed to assign each KeycloakClientRole to member.")
	}
	return nil
}

func (u *ProjectUsecase) UnassignKeycloakClientRoleToMember(ctx context.Context, organizationId string, projectId string, clientId string, projectMemberId string) error {
	pm, err := u.GetProjectMember(ctx, projectMemberId)
	if err != nil {
		log.Error(ctx, err)
		return errors.Wrap(err, "Failed to get project member.")
	}
	err = u.unassignKeycloakClientRoleToMember(ctx, organizationId, projectId, clientId, pm.ProjectUser.ID.String(), pm.ProjectRole.Name)
	if err != nil {
		log.Error(ctx, err)
		return err
	}
	return nil
}

func (u *ProjectUsecase) unassignKeycloakClientRoleToMember(ctx context.Context, organizationId string, projectId string, clientId string, userId string, roleName string) error {
	err := u.kc.UnassignClientRoleToUser(ctx, organizationId, userId, clientId, roleName+"@"+projectId)
	if err != nil {
		log.Error(ctx, err)
		return errors.Wrap(err, "Failed to un-assign each KeycloakClientRole to member.")
	}
	return nil
}

func (u *ProjectUsecase) GetProjectNamespaceKubeconfig(ctx context.Context, organizationId string, projectId string, namespace string, stackId domain.StackId) (string, error) {
	kubeconfig, err := kubernetes.GetKubeConfig(ctx, stackId.String(), kubernetes.KubeconfigForUser)
	if err != nil {
		log.Error(ctx, err)
		return "", errors.Wrap(err, "Failed to get kubeconfig.")
	}

	type kubeConfigType struct {
		APIVersion string `yaml:"apiVersion"`
		Kind       string `yaml:"kind"`
		Clusters   []struct {
			Name    string `yaml:"name"`
			Cluster struct {
				Server                   string `yaml:"server"`
				CertificateAuthorityData string `yaml:"certificate-authority-data,omitempty"`
			} `yaml:"cluster"`
		} `yaml:"clusters"`
		Contexts []struct {
			Name    string `yaml:"name"`
			Context struct {
				Cluster   string `yaml:"cluster"`
				User      string `yaml:"user"`
				Namespace string `yaml:"namespace,omitempty"`
			} `yaml:"context"`
		} `yaml:"contexts"`

		Users []interface{} `yaml:"users,omitempty"`
	}

	var config kubeConfigType
	err = yaml.Unmarshal(kubeconfig, &config)
	if err != nil {
		log.Error(ctx, err)
		return "", errors.Wrap(err, "Failed to unmarshal kubeconfig.")
	}
	config.Contexts[0].Context.Namespace = namespace

	kubeconfig, err = yaml.Marshal(config)
	if err != nil {
		log.Error(ctx, err)
		return "", errors.Wrap(err, "Failed to marshal kubeconfig.")
	}

	return string(kubeconfig[:]), nil

}

func (u *ProjectUsecase) GetProjectKubeconfig(ctx context.Context, organizationId string, projectId string) (string, error) {
	projectNamespaces, err := u.projectRepo.GetProjectNamespaces(ctx, organizationId, projectId, nil)
	if err != nil {
		log.Error(ctx, err)
		return "", errors.Wrap(err, "Failed to retrieve project namespaces.")
	}

	type kubeConfigType struct {
		APIVersion string `yaml:"apiVersion"`
		Kind       string `yaml:"kind"`
		Clusters   []struct {
			Name    string `yaml:"name"`
			Cluster struct {
				Server                   string `yaml:"server"`
				CertificateAuthorityData string `yaml:"certificate-authority-data,omitempty"`
			} `yaml:"cluster"`
		} `yaml:"clusters"`
		Contexts []struct {
			Name    string `yaml:"name"`
			Context struct {
				Cluster   string `yaml:"cluster"`
				User      string `yaml:"user"`
				Namespace string `yaml:"namespace,omitempty"`
			} `yaml:"context"`
		} `yaml:"contexts"`

		Users []interface{} `yaml:"users,omitempty"`
	}

	kubeconfigs := make([]string, 0)
	for _, pn := range projectNamespaces {
		kubeconfig, err := kubernetes.GetKubeConfig(ctx, pn.StackId, kubernetes.KubeconfigForUser)
		if err != nil {
			log.Error(ctx, err)
			return "", errors.Wrap(err, "Failed to retrieve kubeconfig.")
		}

		var config kubeConfigType
		err = yaml.Unmarshal(kubeconfig, &config)
		if err != nil {
			log.Error(ctx, err)
			return "", errors.Wrap(err, "Failed to unmarshal kubeconfig.")
		}
		config.Contexts[0].Context.Namespace = pn.Namespace

		kubeconfig, err = yaml.Marshal(config)
		if err != nil {
			log.Error(ctx, err)
			return "", errors.Wrap(err, "Failed to marshal kubeconfig.")
		}

		kubeconfigs = append(kubeconfigs, string(kubeconfig[:]))
	}

	return kubernetes.MergeKubeconfigsWithSingleUser(kubeconfigs)
}

func (u *ProjectUsecase) GetK8sResources(ctx context.Context, organizationId string, projectId string, namespace string, stackId domain.StackId) (out domain.ProjectNamespaceK8sResources, err error) {
	_, err = u.clusterRepository.Get(ctx, domain.ClusterId(stackId))
	if err != nil {
		return out, errors.Wrap(err, fmt.Sprintf("Failed to get cluster : stackId %s", stackId))
	}

	clientset_user, err := kubernetes.GetClientFromClusterId(ctx, stackId.String())
	if err != nil {
		return out, errors.Wrap(err, fmt.Sprintf("Failed to get clientset : stackId %s", stackId))
	}

	out.UpdatedAt = time.Now()

	pods, err := clientset_user.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err == nil {
		out.Pods = len(pods.Items)
	} else {
		log.Error(ctx, "Failed to get pods. err : ", err)
	}

	pvcs, err := clientset_user.CoreV1().PersistentVolumeClaims(namespace).List(ctx, metav1.ListOptions{})
	if err == nil {
		out.PVCs = len(pvcs.Items)
	} else {
		log.Error(ctx, "Failed to get pvcs. err : ", err)
	}

	services, err := clientset_user.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err == nil {
		out.Services = len(services.Items)
	} else {
		log.Error(ctx, "Failed to get services. err : ", err)
	}

	ingresses, err := clientset_user.NetworkingV1().Ingresses(namespace).List(ctx, metav1.ListOptions{})
	if err == nil {
		out.Ingresses = len(ingresses.Items)
	} else {
		log.Error(ctx, "Failed to get ingresses. err : ", err)
	}

	deployments, err := clientset_user.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err == nil {
		out.Deployments = len(deployments.Items)
	} else {
		log.Error(ctx, "Failed to get deployments. err : ", err)
	}

	statefulsets, err := clientset_user.AppsV1().StatefulSets(namespace).List(ctx, metav1.ListOptions{})
	if err == nil {
		out.Statefulsets = len(statefulsets.Items)
	} else {
		log.Error(ctx, "Failed to get statefulsets. err : ", err)
	}

	daemonsets, err := clientset_user.AppsV1().DaemonSets(namespace).List(ctx, metav1.ListOptions{})
	if err == nil {
		out.Daemonsets = len(daemonsets.Items)
	} else {
		log.Error(ctx, "Failed to get daemonsets. err : ", err)
	}

	jobs, err := clientset_user.BatchV1().Jobs(namespace).List(ctx, metav1.ListOptions{})
	if err == nil {
		out.Jobs = len(jobs.Items)
	} else {
		log.Error(ctx, "Failed to get jobs. err : ", err)
	}

	return
}

func (u *ProjectUsecase) GetResourcesUsage(ctx context.Context, thanosClient thanos.ThanosClient, organizationId string, projectId string, namespace string, stackId domain.StackId) (out domain.ProjectNamespaceResourcesUsage, err error) {
	_, err = u.clusterRepository.Get(ctx, domain.ClusterId(stackId))
	if err != nil {
		return out, errors.Wrap(err, fmt.Sprintf("Failed to get cluster : stackId %s", stackId))
	}

	// sum(rate(container_cpu_usage_seconds_total{taco_cluster=\"$taco_cluster\",image!=\"\"}[$__rate_interval])) by (namespace)
	query := fmt.Sprintf("sum(rate(container_cpu_usage_seconds_total{image!=\"\", namespace=\"%s\"}[10m]) ) by (taco_cluster, namespace)", namespace)
	result, err := thanosClient.Get(ctx, query)
	if err != nil {
		return out, err
	}
	for _, val := range result.Data.Result {
		if val.Metric.TacoCluster == stackId.String() {
			if val.Metric.Namespace == namespace {
				if s, err := strconv.ParseFloat(val.Value[1].(string), 32); err == nil {
					out.Cpu = fmt.Sprintf("%0.2f %%", s*100)
				}
			}
		}
	}

	// sum(container_memory_working_set_bytes{taco_cluster=\"$taco_cluster\",image!=\"\"}) by (namespace)
	query = fmt.Sprintf("sum(container_memory_working_set_bytes{image!=\"\", namespace=\"%s\"}) by (taco_cluster, namespace)", namespace)
	result, err = thanosClient.Get(ctx, query)
	if err != nil {
		return out, err
	}
	for _, val := range result.Data.Result {
		if val.Metric.TacoCluster == stackId.String() {
			if val.Metric.Namespace == namespace {
				memory, _ := strconv.Atoi(val.Value[1].(string))
				out.Memory = fmt.Sprintf("%d MiB", memory/1024/1024)
			}
		}
	}

	out.Storage = ""

	return
}

func (u *ProjectUsecase) EnsureNamespaceForCluster(ctx context.Context, organizationId string, stackId string, namespaceName string) error {
	kubeconfig, err := kubernetes.GetKubeConfig(ctx, stackId, kubernetes.KubeconfigForAdmin)
	if err != nil {
		log.Error(ctx, err)
		return errors.Wrap(err, "Failed to get kubeconfig.")
	}

	err = kubernetes.EnsureNamespace(ctx, kubeconfig, namespaceName)
	if err != nil {
		log.Error(ctx, err)
		return errors.Wrap(err, "Failed to create K8s namespace resource.")
	}

	return nil
}
