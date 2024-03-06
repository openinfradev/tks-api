package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/domain"
	"github.com/openinfradev/tks-api/internal/keycloak"
	"github.com/openinfradev/tks-api/internal/kubernetes"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-api/internal/serializer"
	argowf "github.com/openinfradev/tks-api/pkg/argo-client"
	out_domain "github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/log"
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
	CreateProject(ctx context.Context, p *domain.Project) (string, error)
	GetProjects(ctx context.Context, organizationId string, userId string, onlyMyProject bool, pg *pagination.Pagination) ([]out_domain.ProjectResponse, error)
	GetProject(ctx context.Context, organizationId string, projectId string) (*domain.Project, error)
	GetProjectWithLeader(ctx context.Context, organizationId string, projectId string) (*domain.Project, error)
	IsProjectNameExist(ctx context.Context, organizationId string, projectName string) (bool, error)
	UpdateProject(ctx context.Context, p *domain.Project, newLeaderId string) error
	GetProjectRole(ctx context.Context, id string) (*domain.ProjectRole, error)
	GetProjectRoles(ctx context.Context, query int) ([]domain.ProjectRole, error)
	AddProjectMember(ctx context.Context, pm *domain.ProjectMember) (string, error)
	GetProjectUser(ctx context.Context, projectUserId string) (*domain.ProjectUser, error)
	GetProjectMember(ctx context.Context, projectMemberId string) (*domain.ProjectMember, error)
	GetProjectMembers(ctx context.Context, projectId string, query int, pg *pagination.Pagination) ([]domain.ProjectMember, error)
	GetProjectMemberCount(ctx context.Context, projectMemberId string) (*out_domain.GetProjectMemberCountResponse, error)
	RemoveProjectMember(ctx context.Context, projectMemberId string) error
	UpdateProjectMemberRole(ctx context.Context, pm *domain.ProjectMember) error
	CreateProjectNamespace(ctx context.Context, organizationId string, pn *domain.ProjectNamespace) error
	IsProjectNamespaceExist(ctx context.Context, organizationId string, projectId string, stackId string, projectNamespace string) (bool, error)
	GetProjectNamespaces(ctx context.Context, organizationId string, projectId string, pg *pagination.Pagination) ([]domain.ProjectNamespace, error)
	GetProjectNamespace(ctx context.Context, organizationId string, projectId string, projectNamespace string, stackId string) (*domain.ProjectNamespace, error)
	UpdateProjectNamespace(ctx context.Context, pn *domain.ProjectNamespace) error
	DeleteProjectNamespace(ctx context.Context, organizationId string, projectId string, projectNamespace string, stackId string) error
	GetAppCount(ctx context.Context, organizationId string, projectId string, namespace string) (appCount int, err error)
	EnsureRequiredSetupForCluster(ctx context.Context, organizationId string, projectId string, stackId string) error
	MayRemoveRequiredSetupForCluster(ctx context.Context, organizationId string, projectId string, stackId string) error
	CreateK8SNSRoleBinding(ctx context.Context, organizationId string, projectId string, stackId string, namespace string) error
	DeleteK8SNSRoleBinding(ctx context.Context, organizationId string, projectId string, stackId string, namespace string) error
	GetProjectKubeconfig(ctx context.Context, organizationId string, projectId string) (string, error)
	GetK8sResources(ctx context.Context, organizationId string, projectId string, namespace string, stackId out_domain.StackId) (out out_domain.ProjectNamespaceK8sResources, err error)
	GetResourcesUsage(ctx context.Context, organizationId string, projectId string, namespace string, stackId out_domain.StackId) (out out_domain.ProjectNamespaceResourcesUsage, err error)
	AssignKeycloakClientRoleToMember(ctx context.Context, organizationId string, projectId string, stackId string, projectMemberId string) error
	UnassignKeycloakClientRoleToMember(ctx context.Context, organizationId string, projectId string, stackId string, projectMemberId string) error
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

func (u *ProjectUsecase) CreateProject(ctx context.Context, p *domain.Project) (string, error) {
	projectId, err := u.projectRepo.CreateProject(p)
	if err != nil {
		log.ErrorWithContext(ctx, err)
		return "", errors.Wrap(err, "Failed to create project.")
	}

	return projectId, nil
}

func (u *ProjectUsecase) GetProjects(ctx context.Context, organizationId string, userId string, onlyMyProject bool, pg *pagination.Pagination) (pr []out_domain.ProjectResponse, err error) {
	userUuid, err := uuid.Parse(userId)
	if err != nil {
		log.ErrorWithContext(ctx, err)
		return nil, errors.Wrap(err, "Failed to parse uuid to string")
	}
	if onlyMyProject == false {
		pr, err = u.projectRepo.GetProjects(organizationId, userUuid, pg)
		if err != nil {
			log.ErrorWithContext(ctx, err)
			return nil, errors.Wrap(err, "Failed to get projects.")
		}
	} else {
		pr, err = u.projectRepo.GetProjectsByUserId(organizationId, userUuid, pg)
		if err != nil {
			log.ErrorWithContext(ctx, err)
			return nil, errors.Wrap(err, "Failed to get projects.")
		}
	}
	return pr, err
}

func (u *ProjectUsecase) GetProject(ctx context.Context, organizationId string, projectId string) (*domain.Project, error) {
	p, err := u.projectRepo.GetProjectById(organizationId, projectId)
	if err != nil {
		log.ErrorWithContext(ctx, err)
		return nil, errors.Wrap(err, "Failed to get projects.")
	}
	return p, err
}

func (u *ProjectUsecase) GetProjectWithLeader(ctx context.Context, organizationId string, projectId string) (*domain.Project, error) {
	p, err := u.projectRepo.GetProjectByIdAndLeader(organizationId, projectId)
	if err != nil {
		log.ErrorWithContext(ctx, err)
		return nil, errors.Wrap(err, "Failed to get projects.")
	}
	return p, err
}

func (u *ProjectUsecase) IsProjectNameExist(ctx context.Context, organizationId string, projectName string) (bool, error) {
	exist := true
	p, err := u.projectRepo.GetProjectByName(organizationId, projectName)
	if err != nil {
		log.ErrorWithContext(ctx, err)
		exist = false
		return exist, errors.Wrap(err, "Failed to retrieve project name.")
	}
	if p == nil {
		exist = false
	}
	return exist, nil
}

func (u *ProjectUsecase) UpdateProject(ctx context.Context, p *domain.Project, newLeaderId string) error {

	var currentMemberId, currentLeaderId, projectRoleId string
	for _, pm := range p.ProjectMembers {
		currentMemberId = pm.ID
		currentLeaderId = pm.ProjectUser.ID.String()
		projectRoleId = pm.ProjectRole.ID
	}
	p.ProjectNamespaces = nil
	p.ProjectMembers = nil

	if err := u.projectRepo.UpdateProject(p); err != nil {
		log.ErrorWithContext(ctx, err)
		return errors.Wrap(err, "Failed to update project.")
	}

	if newLeaderId != "" && currentLeaderId != newLeaderId {
		if err := u.RemoveProjectMember(ctx, currentMemberId); err != nil {
			log.ErrorWithContext(ctx, err)
			return errors.Wrap(err, "Failed to remove project member.")
		}

		pu, err := u.GetProjectUser(ctx, newLeaderId)
		if err != nil {
			return err
		}
		if pu == nil {
			return errors.Wrap(err, "No userid")
		}

		pm, err := u.projectRepo.GetProjectMemberByUserId(p.ID, newLeaderId)
		if err != nil {
			return err
		}
		if pm == nil {
			newPm := &domain.ProjectMember{
				ProjectId:       p.ID,
				ProjectUserId:   pu.ID,
				ProjectUser:     nil,
				ProjectRoleId:   projectRoleId,
				ProjectRole:     nil,
				IsProjectLeader: true,
				CreatedAt:       *p.UpdatedAt,
			}
			res, err := u.AddProjectMember(ctx, newPm)
			if err != nil {
				return err
			}
			log.Infof("Added project member: %s", res)
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

func (u *ProjectUsecase) GetProjectRole(ctx context.Context, id string) (*domain.ProjectRole, error) {
	pr, err := u.projectRepo.GetProjectRoleById(id)
	if err != nil {
		log.ErrorWithContext(ctx, err)
		return nil, errors.Wrap(err, "Failed to get project roles.")
	}

	return pr, nil
}

func (u *ProjectUsecase) GetProjectRoles(ctx context.Context, query int) (prs []domain.ProjectRole, err error) {
	var pr *domain.ProjectRole

	if query == ProjectLeader {
		pr, err = u.projectRepo.GetProjectRoleByName("project-leader")
	} else if query == ProjectMember {
		pr, err = u.projectRepo.GetProjectRoleByName("project-member")
	} else if query == ProjectViewer {
		pr, err = u.projectRepo.GetProjectRoleByName("project-viewer")
	} else {
		prs, err = u.projectRepo.GetAllProjectRoles()
	}
	if err != nil {
		log.ErrorWithContext(ctx, err)
		return nil, errors.Wrap(err, "Failed to get project roles.")
	}

	if pr != nil {
		prs = append(prs, *pr)
	}

	return prs, nil
}

func (u *ProjectUsecase) AddProjectMember(ctx context.Context, pm *domain.ProjectMember) (string, error) {
	projectMemberId, err := u.projectRepo.AddProjectMember(pm)
	if err != nil {
		log.ErrorWithContext(ctx, err)
		return "", errors.Wrap(err, "Failed to add project member to project.")
	}
	return projectMemberId, nil
}

func (u *ProjectUsecase) GetProjectUser(ctx context.Context, projectUserId string) (*domain.ProjectUser, error) {
	var uid uuid.UUID
	uid, err := uuid.Parse(projectUserId)
	if err != nil {
		log.ErrorWithContext(ctx, err)
		return nil, errors.Wrap(err, "Failed to parse uuid to string")
	}

	user, err := u.userRepository.GetByUuid(uid)
	if err != nil {
		log.ErrorWithContext(ctx, err)
		return nil, errors.Wrap(err, "Failed to retrieve user by id")
	}
	var pu domain.ProjectUser
	if err = serializer.Map(user, &pu); err != nil {
		log.ErrorWithContext(ctx, err)
		return nil, err
	}
	return &pu, nil
}

func (u *ProjectUsecase) GetProjectMember(ctx context.Context, projectMemberId string) (pm *domain.ProjectMember, err error) {
	pm, err = u.projectRepo.GetProjectMemberById(projectMemberId)
	if err != nil {
		log.ErrorWithContext(ctx, err)
		return pm, errors.Wrap(err, "Failed to get project member.")
	}

	return pm, nil
}

func (u *ProjectUsecase) GetProjectMembers(ctx context.Context, projectId string, query int, pg *pagination.Pagination) (pms []domain.ProjectMember, err error) {
	if query == ProjectLeader {
		pms, err = u.projectRepo.GetProjectMembersByProjectIdAndRoleName(projectId, "project-leader", pg)
	} else if query == ProjectMember {
		pms, err = u.projectRepo.GetProjectMembersByProjectIdAndRoleName(projectId, "project-member", pg)
	} else if query == ProjectViewer {
		pms, err = u.projectRepo.GetProjectMembersByProjectIdAndRoleName(projectId, "project-viewer", pg)
	} else {
		pms, err = u.projectRepo.GetProjectMembersByProjectId(projectId, pg)
	}
	if err != nil {
		log.ErrorWithContext(ctx, err)
		return nil, errors.Wrap(err, "Failed to get project members.")
	}

	return pms, nil
}

func (u *ProjectUsecase) GetProjectMemberCount(ctx context.Context, projectMemberId string) (pmcr *out_domain.GetProjectMemberCountResponse, err error) {
	pmcr, err = u.projectRepo.GetProjectMemberCountByProjectId(projectMemberId)
	if err != nil {
		log.ErrorWithContext(ctx, err)
		return pmcr, errors.Wrap(err, "Failed to get project member count.")
	}

	return pmcr, nil
}

func (u *ProjectUsecase) RemoveProjectMember(ctx context.Context, projectMemberId string) error {
	if err := u.projectRepo.RemoveProjectMember(projectMemberId); err != nil {
		log.ErrorWithContext(ctx, err)
		return errors.Wrap(err, "Failed to remove project member to project.")
	}
	return nil
}

func (u *ProjectUsecase) UpdateProjectMemberRole(ctx context.Context, pm *domain.ProjectMember) error {

	if err := u.projectRepo.UpdateProjectMemberRole(pm); err != nil {
		log.ErrorWithContext(ctx, err)
		return errors.Wrap(err, "Failed to remove project member to project.")
	}
	return nil
}

func (u *ProjectUsecase) CreateProjectNamespace(ctx context.Context, organizationId string, pn *domain.ProjectNamespace) error {
	if err := u.projectRepo.CreateProjectNamespace(organizationId, pn); err != nil {
		log.ErrorWithContext(ctx, err)
		return errors.Wrap(err, "Failed to create project namespace.")
	}
	return nil
}

func (u *ProjectUsecase) IsProjectNamespaceExist(ctx context.Context, organizationId string, projectId string, stackId string, projectNamespace string) (bool, error) {
	exist := true
	pn, err := u.projectRepo.GetProjectNamespaceByName(organizationId, projectId, stackId, projectNamespace)
	if err != nil {
		log.ErrorWithContext(ctx, err)
		exist = false
		return exist, errors.Wrap(err, "Failed to retrieve project namespace.")
	}
	if pn == nil {
		exist = false
	}
	return exist, nil
}

func (u *ProjectUsecase) GetProjectNamespaces(ctx context.Context, organizationId string, projectId string, pg *pagination.Pagination) ([]domain.ProjectNamespace, error) {
	pns, err := u.projectRepo.GetProjectNamespaces(organizationId, projectId, pg)
	if err != nil {
		log.ErrorWithContext(ctx, err)
		return nil, errors.Wrap(err, "Failed to retrieve project namespaces.")
	}

	return pns, nil
}

func (u *ProjectUsecase) GetProjectNamespace(ctx context.Context, organizationId string, projectId string, projectNamespace string, stackId string) (*domain.ProjectNamespace, error) {
	pn, err := u.projectRepo.GetProjectNamespaceByPrimaryKey(organizationId, projectId, projectNamespace, stackId)
	if err != nil {
		log.ErrorWithContext(ctx, err)
		return nil, errors.Wrap(err, "Failed to retrieve project namespace.")
	}

	return pn, nil
}

func (u *ProjectUsecase) UpdateProjectNamespace(ctx context.Context, pn *domain.ProjectNamespace) error {
	if err := u.projectRepo.UpdateProjectNamespace(pn); err != nil {
		log.ErrorWithContext(ctx, err)
		return errors.Wrap(err, "Failed to update project namespace")
	}
	return nil
}

func (u *ProjectUsecase) DeleteProjectNamespace(ctx context.Context, organizationId string, projectId string,
	stackId string, projectNamespace string) error {
	if err := u.projectRepo.DeleteProjectNamespace(organizationId, projectId, projectNamespace, stackId); err != nil {
		log.ErrorWithContext(ctx, err)
		return errors.Wrap(err, "Failed to delete project namespace.")
	}
	return nil
}

func (u *ProjectUsecase) GetAppCount(ctx context.Context, organizationId string, projectId string, namespace string) (appCount int, err error) {
	if namespace == "" {
		appCount, err = u.projectRepo.GetAppCountByProjectId(organizationId, projectId)
	} else {
		appCount, err = u.projectRepo.GetAppCountByNamespace(organizationId, projectId, namespace)
	}
	if err != nil {
		log.ErrorWithContext(ctx, err)
		return 0, errors.Wrap(err, "Failed to retrieve app count.")
	}

	return appCount, nil
}

func (u *ProjectUsecase) EnsureRequiredSetupForCluster(ctx context.Context, organizationId string, projectId string, stackId string) error {
	pns, err := u.projectRepo.GetProjectNamespaces(organizationId, projectId, nil)
	if err != nil {
		log.ErrorWithContext(ctx, err)
		return errors.Wrap(err, "Failed to create project namespace.")
	}

	var alreadySetUp bool
	for _, pn := range pns {
		if pn.StackId == stackId {
			alreadySetUp = true
			break
		}
	}

	// if already set up, it means that required setup is already done
	if alreadySetUp {
		return nil
	}

	if err := u.createK8SInitialResource(ctx, organizationId, projectId, stackId); err != nil {
		log.ErrorWithContext(ctx, err)
		return errors.Wrap(err, "Failed to create project namespace.")
	}

	if err := u.createKeycloakClientRoles(ctx, organizationId, projectId, stackId); err != nil {
		log.ErrorWithContext(ctx, err)
		return errors.Wrap(err, "Failed to create project namespace.")
	}

	projectMembers, err := u.GetProjectMembers(ctx, projectId, ProjectAll, nil)
	if err != nil {
		log.ErrorWithContext(ctx, err)
		return errors.Wrap(err, "Failed to create project namespace.")
	}
	for _, pm := range projectMembers {
		err = u.assignEachKeycloakClientRoleToMember(ctx, organizationId, projectId, stackId, pm.ProjectUserId.String(), pm.ProjectRole.Name)
		if err != nil {
			log.ErrorWithContext(ctx, err)
			return errors.Wrap(err, "Failed to create project namespace.")
		}
	}

	return nil
}
func (u *ProjectUsecase) MayRemoveRequiredSetupForCluster(ctx context.Context, organizationId string, projectId string, stackId string) error {
	pns, err := u.projectRepo.GetProjectNamespaces(organizationId, projectId, nil)
	if err != nil {
		log.ErrorWithContext(ctx, err)
		return errors.Wrap(err, "Failed to create project namespace.")
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
		log.ErrorWithContext(ctx, err)
		return errors.Wrap(err, "Failed to create project namespace.")
	}

	projectMembers, err := u.GetProjectMembers(ctx, projectId, ProjectAll, nil)
	if err != nil {
		log.ErrorWithContext(ctx, err)
		return errors.Wrap(err, "Failed to create project namespace.")
	}
	for _, pm := range projectMembers {
		err = u.unassignKeycloakClientRoleToMember(ctx, organizationId, projectId, stackId, pm.ProjectUserId.String(), pm.ProjectRole.Name)
		if err != nil {
			log.ErrorWithContext(ctx, err)
			return errors.Wrap(err, "Failed to create project namespace.")
		}
	}

	if err := u.deleteKeycloakClientRoles(ctx, organizationId, projectId, stackId); err != nil {
		log.ErrorWithContext(ctx, err)
		return errors.Wrap(err, "Failed to create project namespace.")
	}

	return nil
}
func (u *ProjectUsecase) createK8SInitialResource(ctx context.Context, organizationId string, projectId string, stackId string) error {
	kubeconfig, err := kubernetes.GetKubeConfig(stackId)
	if err != nil {
		log.ErrorWithContext(ctx, err)
		return errors.Wrap(err, "Failed to create project namespace.")
	}

	pr, err := u.GetProject(ctx, organizationId, projectId)
	if err != nil {
		log.ErrorWithContext(ctx, err)
		return errors.Wrap(err, "Failed to create project namespace.")
	}

	err = kubernetes.EnsureClusterRole(kubeconfig, pr.Name)
	if err != nil {
		log.ErrorWithContext(ctx, err)
		return errors.Wrap(err, "Failed to create project namespace.")
	}

	err = kubernetes.EnsureCommonClusterRole(kubeconfig, pr.Name)
	if err != nil {
		log.ErrorWithContext(ctx, err)
		return errors.Wrap(err, "Failed to create project namespace.")
	}

	err = kubernetes.EnsureCommonClusterRoleBinding(kubeconfig, pr.Name)
	if err != nil {
		log.ErrorWithContext(ctx, err)
		return errors.Wrap(err, "Failed to create project namespace.")
	}

	return nil
}
func (u *ProjectUsecase) deleteK8SInitialResource(ctx context.Context, organizationId string, projectId string, stackId string) error {
	kubeconfig, err := kubernetes.GetKubeConfig(stackId)
	if err != nil {
		log.ErrorWithContext(ctx, err)
		return errors.Wrap(err, "Failed to create project namespace.")
	}

	pr, err := u.GetProject(ctx, organizationId, projectId)
	if err != nil {
		log.ErrorWithContext(ctx, err)
		return errors.Wrap(err, "Failed to create project namespace.")
	}

	err = kubernetes.RemoveClusterRole(kubeconfig, pr.Name)
	if err != nil {
		log.ErrorWithContext(ctx, err)
		return errors.Wrap(err, "Failed to create project namespace.")
	}

	err = kubernetes.RemoveCommonClusterRole(kubeconfig, pr.Name)
	if err != nil {
		log.ErrorWithContext(ctx, err)
		return errors.Wrap(err, "Failed to create project namespace.")
	}

	err = kubernetes.RemoveCommonClusterRoleBinding(kubeconfig, pr.Name)
	if err != nil {
		log.ErrorWithContext(ctx, err)
		return errors.Wrap(err, "Failed to create project namespace.")
	}

	return nil
}
func (u *ProjectUsecase) createKeycloakClientRoles(ctx context.Context, organizationId string, projectId string, stackId string) error {
	// create Roles in keycloak
	for _, role := range []string{string(ProjectLeader), string(ProjectMember), string(ProjectViewer)} {
		err := u.kc.EnsureClientRoleWithClientName(organizationId, stackId+"-k8s-api", role+"@"+projectId)
		if err != nil {
			log.ErrorWithContext(ctx, err)
			return errors.Wrap(err, "Failed to create project namespace.")
		}
	}
	return nil
}
func (u *ProjectUsecase) deleteKeycloakClientRoles(ctx context.Context, organizationId string, projectId string, stackId string) error {
	// first check whether the stac

	// delete Roles in keycloak
	for _, role := range []string{string(ProjectLeader), string(ProjectMember), string(ProjectViewer)} {
		err := u.kc.DeleteClientRoleWithClientName(organizationId, stackId+"-k8s-api", role+"@"+projectId)
		if err != nil {
			log.ErrorWithContext(ctx, err)
			return errors.Wrap(err, "Failed to create project namespace.")
		}
	}
	return nil
}
func (u *ProjectUsecase) CreateK8SNSRoleBinding(ctx context.Context, organizationId string, projectId string, stackId string, namespace string) error {
	kubeconfig, err := kubernetes.GetKubeConfig(stackId)
	if err != nil {
		log.ErrorWithContext(ctx, err)
		return errors.Wrap(err, "Failed to create project namespace.")
	}

	pr, err := u.GetProject(ctx, organizationId, projectId)
	if err != nil {
		log.ErrorWithContext(ctx, err)
		return errors.Wrap(err, "Failed to create project namespace.")
	}

	err = kubernetes.EnsureRoleBinding(kubeconfig, pr.Name, namespace)
	if err != nil {
		log.ErrorWithContext(ctx, err)
		return errors.Wrap(err, "Failed to create project namespace.")
	}

	return nil
}
func (u *ProjectUsecase) DeleteK8SNSRoleBinding(ctx context.Context, organizationId string, projectId string, stackId string, namespace string) error {
	//TODO implement me
	panic("implement me")
}

func (u *ProjectUsecase) AssignKeycloakClientRoleToMember(ctx context.Context, organizationId string, projectId string, stackId string, projectMemberId string) error {
	pm, err := u.GetProjectMember(ctx, projectMemberId)
	if err != nil {
		log.ErrorWithContext(ctx, err)
		return errors.Wrap(err, "Failed to create project namespace.")
	}
	err = u.assignEachKeycloakClientRoleToMember(ctx, organizationId, projectId, stackId, pm.ProjectUserId.String(), pm.ProjectRole.Name)
	return nil
}

func (u *ProjectUsecase) assignEachKeycloakClientRoleToMember(ctx context.Context, organizationId string, projectId string, stackId string, userId string, roleName string) error {
	err := u.kc.AssignClientRoleToUser(organizationId, userId, stackId+"-k8s-api", roleName+"@"+projectId)
	if err != nil {
		log.ErrorWithContext(ctx, err)
		return errors.Wrap(err, "Failed to create project namespace.")
	}
	return nil
}

func (u *ProjectUsecase) UnassignKeycloakClientRoleToMember(ctx context.Context, organizationId string, projectId string, stackId string, projectMemberId string) error {
	pm, err := u.GetProjectMember(ctx, projectMemberId)
	if err != nil {
		log.ErrorWithContext(ctx, err)
		return errors.Wrap(err, "Failed to create project namespace.")
	}
	err = u.unassignKeycloakClientRoleToMember(ctx, organizationId, projectId, stackId, pm.ProjectUserId.String(), pm.ProjectRole.Name)
	return nil
}

func (u *ProjectUsecase) unassignKeycloakClientRoleToMember(ctx context.Context, organizationId string, projectId string, stackId string, userId string, roleName string) error {
	err := u.kc.UnassignClientRoleToUser(organizationId, userId, stackId+"-k8s-api", roleName+"@"+projectId)
	if err != nil {
		log.ErrorWithContext(ctx, err)
		return errors.Wrap(err, "Failed to create project namespace.")
	}
	return nil
}

func (u *ProjectUsecase) GetProjectKubeconfig(ctx context.Context, organizationId string, projectId string) (string, error) {
	projectNamespaces, err := u.projectRepo.GetProjectNamespaces(organizationId, projectId, nil)
	if err != nil {
		log.ErrorWithContext(ctx, err)
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
		kubeconfig, err := kubernetes.GetKubeConfig(pn.StackId)
		if err != nil {
			log.ErrorWithContext(ctx, err)
			return "", errors.Wrap(err, "Failed to retrieve kubeconfig.")
		}

		var config kubeConfigType
		err = yaml.Unmarshal(kubeconfig, &config)
		if err != nil {
			log.ErrorWithContext(ctx, err)
			return "", errors.Wrap(err, "Failed to unmarshal kubeconfig.")
		}
		config.Contexts[0].Context.Namespace = pn.Namespace

		kubeconfig, err = yaml.Marshal(config)
		if err != nil {
			log.ErrorWithContext(ctx, err)
			return "", errors.Wrap(err, "Failed to marshal kubeconfig.")
		}

		kubeconfigs = append(kubeconfigs, string(kubeconfig[:]))
	}

	return kubernetes.MergeKubeconfigsWithSingleUser(kubeconfigs)
}

func (u *ProjectUsecase) GetK8sResources(ctx context.Context, organizationId string, projectId string, namespace string, stackId out_domain.StackId) (out out_domain.ProjectNamespaceK8sResources, err error) {
	_, err = u.clusterRepository.Get(out_domain.ClusterId(stackId))
	if err != nil {
		return out, errors.Wrap(err, fmt.Sprintf("Failed to get cluster : stackId %s", stackId))
	}

	clientset_user, err := kubernetes.GetClientFromClusterId(stackId.String())
	if err != nil {
		return out, errors.Wrap(err, fmt.Sprintf("Failed to get clientset : stackId %s", stackId))
	}

	out.UpdatedAt = time.Now()

	pods, err := clientset_user.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err == nil {
		out.Pods = len(pods.Items)
	} else {
		log.ErrorWithContext(ctx, "Failed to get pods. err : ", err)
	}

	pvcs, err := clientset_user.CoreV1().PersistentVolumeClaims(namespace).List(ctx, metav1.ListOptions{})
	if err == nil {
		out.PVCs = len(pvcs.Items)
	} else {
		log.ErrorWithContext(ctx, "Failed to get pvcs. err : ", err)
	}

	services, err := clientset_user.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err == nil {
		out.Services = len(services.Items)
	} else {
		log.ErrorWithContext(ctx, "Failed to get services. err : ", err)
	}

	ingresses, err := clientset_user.NetworkingV1().Ingresses(namespace).List(ctx, metav1.ListOptions{})
	if err == nil {
		out.Ingresses = len(ingresses.Items)
	} else {
		log.ErrorWithContext(ctx, "Failed to get ingresses. err : ", err)
	}

	deployments, err := clientset_user.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err == nil {
		out.Deployments = len(deployments.Items)
	} else {
		log.ErrorWithContext(ctx, "Failed to get deployments. err : ", err)
	}

	statefulsets, err := clientset_user.AppsV1().StatefulSets(namespace).List(ctx, metav1.ListOptions{})
	if err == nil {
		out.Statefulsets = len(statefulsets.Items)
	} else {
		log.ErrorWithContext(ctx, "Failed to get statefulsets. err : ", err)
	}

	daemonsets, err := clientset_user.AppsV1().DaemonSets(namespace).List(ctx, metav1.ListOptions{})
	if err == nil {
		out.Daemonsets = len(daemonsets.Items)
	} else {
		log.ErrorWithContext(ctx, "Failed to get daemonsets. err : ", err)
	}

	jobs, err := clientset_user.BatchV1().Jobs(namespace).List(ctx, metav1.ListOptions{})
	if err == nil {
		out.Jobs = len(jobs.Items)
	} else {
		log.ErrorWithContext(ctx, "Failed to get jobs. err : ", err)
	}

	return
}

func (u *ProjectUsecase) GetResourcesUsage(ctx context.Context, organizationId string, projectId string, namespace string, stackId domain.StackId) (out domain.ProjectNamespaceResourcesUsage, err error) {
	_, err = u.clusterRepository.Get(domain.ClusterId(stackId))
	if err != nil {
		return out, errors.Wrap(err, fmt.Sprintf("Failed to get cluster : stackId %s", stackId))
	}

	return
}
