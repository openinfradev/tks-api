package rbac

import (
	"github.com/openinfradev/tks-api/internal/delivery/api"
	"github.com/openinfradev/tks-api/internal/helper"
	"github.com/openinfradev/tks-api/pkg/domain"
)

type PermissionSet struct {
	Dashboard         *domain.Permission
	Stack             *domain.Permission
	SecurityPolicy    *domain.Permission
	ProjectManagement *domain.Permission
	Notification      *domain.Permission
	Configuration     *domain.Permission
}

func NewDefaultPermission() *PermissionSet {
	return &PermissionSet{
		Dashboard:         newDashboard(),
		Stack:             newStack(),
		SecurityPolicy:    newSecurityPolicy(),
		ProjectManagement: newProjectManagement(),
		Notification:      newNotification(),
		Configuration:     newConfiguration(),
	}
}

func endpointObjects(eps ...api.Endpoint) []*domain.Endpoint {
	var result []*domain.Endpoint
	for _, ep := range eps {
		result = append(result, &domain.Endpoint{
			Name:  api.ApiMap[ep].Name,
			Group: api.ApiMap[ep].Group,
		})
	}
	return result
}

func newDashboard() *domain.Permission {
	dashboard := &domain.Permission{
		Name: "대시보드",
		Children: []*domain.Permission{
			{
				Name: "대시보드",
				Children: []*domain.Permission{
					{
						Name:      "조회",
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.GetChartsDashboard,
							api.GetChartDashboard,
							api.GetStacksDashboard,
							api.GetResourcesDashboard,
						),
					},
				},
			},
			{
				Name: "대시보드 설정",
				Children: []*domain.Permission{
					{
						Name:      "조회",
						IsAllowed: helper.BoolP(false),
					},
					{
						Name:      "생성",
						IsAllowed: helper.BoolP(false),
					},
					{
						Name:      "수정",
						IsAllowed: helper.BoolP(false),
					},
					{
						Name:      "삭제",
						IsAllowed: helper.BoolP(false),
					},
				},
			},
		},
	}

	return dashboard
}

func newStack() *domain.Permission {
	stack := &domain.Permission{
		Name: "스택 관리",
		Children: []*domain.Permission{
			{
				Name:      "조회",
				IsAllowed: helper.BoolP(false),
				Endpoints: endpointObjects(
					api.GetStacks,
					api.GetStack,
					api.CheckStackName,
					api.GetStackStatus,
					api.GetStackKubeConfig,

					api.SetFavoriteStack,
					api.DeleteFavoriteStack,
				),
			},
			{
				Name:      "생성",
				IsAllowed: helper.BoolP(false),
				Endpoints: endpointObjects(
					api.CreateStack,
					api.InstallStack,
				),
			},
			{
				Name:      "수정",
				IsAllowed: helper.BoolP(false),
				Endpoints: endpointObjects(
					api.UpdateStack,
				),
			},
			{
				Name:      "삭제",
				IsAllowed: helper.BoolP(false),
				Endpoints: endpointObjects(
					api.DeleteStack,
				),
			},
		},
	}

	return stack
}

func newSecurityPolicy() *domain.Permission {
	security_policy := &domain.Permission{
		Name: "보안/정책 관리",
		Children: []*domain.Permission{
			{
				Name: "보안/정책",
				Children: []*domain.Permission{
					{
						Name:      "조회",
						IsAllowed: helper.BoolP(false),
					},
					{
						Name:      "생성",
						IsAllowed: helper.BoolP(false),
					},
					{
						Name:      "수정",
						IsAllowed: helper.BoolP(false),
					},
					{
						Name:      "삭제",
						IsAllowed: helper.BoolP(false),
					},
				},
			},
		},
	}

	return security_policy
}

func newProjectManagement() *domain.Permission {
	projectManagement := &domain.Permission{
		Name: "프로젝트 관리",
		Children: []*domain.Permission{
			{
				Name: "프로젝트",
				Children: []*domain.Permission{
					{
						Name:      "조회",
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.GetProjects,
							api.GetProject,
						),
					},
					{
						Name:      "생성",
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.CreateProject,
						),
					},
				},
			},
			{
				Name: "앱 서빙",
				Children: []*domain.Permission{
					{
						Name:      "조회",
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.GetAppServeApps,
							api.GetAppServeApp,
							api.GetNumOfAppsOnStack,
							api.GetAppServeAppLatestTask,
							api.IsAppServeAppExist,
							api.IsAppServeAppNameExist,
						),
					},
					{
						Name:      "빌드",
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.CreateAppServeApp,
							api.IsAppServeAppExist,
							api.IsAppServeAppNameExist,
							api.UpdateAppServeApp,
							api.UpdateAppServeAppEndpoint,
							api.UpdateAppServeAppStatus,
							api.RollbackAppServeApp,
						),
					},
					{
						Name:      "배포",
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.CreateAppServeApp,
							api.IsAppServeAppExist,
							api.IsAppServeAppNameExist,
							api.UpdateAppServeApp,
							api.UpdateAppServeAppEndpoint,
							api.UpdateAppServeAppStatus,
							api.RollbackAppServeApp,
						),
					},
					{
						Name:      "삭제",
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.DeleteAppServeApp,
						),
					},
				},
			},
			{
				Name: "설정-일반",
				Children: []*domain.Permission{
					{
						Name:      "조회",
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.GetProjects,
							api.GetProject,

							api.GetProjectRoles,
							api.GetProjectRole,
						),
					},
					{
						Name:      "수정",
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.UpdateProject,
						),
					},
					{
						Name:      "삭제",
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.DeleteProject,
						),
					},
				},
			},
			{
				Name: "설정-멤버",
				Children: []*domain.Permission{
					{
						Name:      "조회",
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.GetProjectMembers,
							api.GetProjectMember,
							api.GetProjectRoles,
							api.GetProjectRole,
						),
					},
					{
						Name:      "생성",
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.AddProjectMember,
						),
					},
					{
						Name:      "수정",
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.UpdateProjectMemberRole,
						),
					},
					{
						Name:      "삭제",
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.RemoveProjectMember,
						),
					},
				},
			},
			{
				Name: "설정-네임스페이스",
				Children: []*domain.Permission{
					{
						Name:      "조회",
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.GetProjectNamespaces,
							api.GetProjectNamespace,
						),
					},
					{
						Name:      "생성",
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.CreateProjectNamespace,
						),
					},
					{
						Name:      "수정",
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(),
					},
					{
						Name:      "삭제",
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.DeleteProjectNamespace,
						),
					},
				},
			},
		},
	}

	return projectManagement
}

func newNotification() *domain.Permission {
	notification := &domain.Permission{
		Name: "알림",
		Children: []*domain.Permission{
			{
				Name: "시스템 경고",
				Children: []*domain.Permission{
					{
						Name:      "조회",
						IsAllowed: helper.BoolP(false),
					},
				},
			},
			{
				Name: "보안/정책 감사로그",
				Children: []*domain.Permission{
					{
						Name:      "조회",
						IsAllowed: helper.BoolP(false),
					},
				},
			},
		},
	}

	return notification
}

func newConfiguration() *domain.Permission {
	configuration := &domain.Permission{
		Name: "설정",
		Children: []*domain.Permission{
			{
				Name: "일반",
				Children: []*domain.Permission{
					{
						Name:      "조회",
						IsAllowed: helper.BoolP(false),
					},
					{
						Name:      "수정",
						IsAllowed: helper.BoolP(false),
					},
				},
			},
			{
				Name: "클라우드 계정",
				Children: []*domain.Permission{
					{
						Name:      "조회",
						IsAllowed: helper.BoolP(false),
					},
					{
						Name:      "생성",
						IsAllowed: helper.BoolP(false),
					},
					{
						Name:      "수정",
						IsAllowed: helper.BoolP(false),
					},
					{
						Name:      "삭제",
						IsAllowed: helper.BoolP(false),
					},
				},
			},
			{
				Name: "스택 템플릿",
				Children: []*domain.Permission{
					{
						Name:      "조회",
						IsAllowed: helper.BoolP(false),
					},
				},
			},
			{
				Name: "프로젝트 관리",
				Children: []*domain.Permission{
					{
						Name:      "조회",
						IsAllowed: helper.BoolP(false),
					},
					{
						Name:      "생성",
						IsAllowed: helper.BoolP(false),
					},
					{
						Name:      "수정",
						IsAllowed: helper.BoolP(false),
					},
					{
						Name:      "삭제",
						IsAllowed: helper.BoolP(false),
					},
				},
			},
			{
				Name: "사용자",
				Children: []*domain.Permission{
					{
						Name:      "조회",
						IsAllowed: helper.BoolP(false),
					},
					{
						Name:      "생성",
						IsAllowed: helper.BoolP(false),
					},
					{
						Name:      "수정",
						IsAllowed: helper.BoolP(false),
					},
					{
						Name:      "삭제",
						IsAllowed: helper.BoolP(false),
					},
				},
			},
			{
				Name: "사용자 권한 관리",
				Children: []*domain.Permission{
					{
						Name:      "조회",
						IsAllowed: helper.BoolP(false),
					},
					{
						Name:      "생성",
						IsAllowed: helper.BoolP(false),
					},
					{
						Name:      "수정",
						IsAllowed: helper.BoolP(false),
					},
					{
						Name:      "삭제",
						IsAllowed: helper.BoolP(false),
					},
				},
			},
			{
				Name: "알림 설정",
				Children: []*domain.Permission{
					{
						Name:      "조회",
						IsAllowed: helper.BoolP(false),
					},
					{
						Name:      "생성",
						IsAllowed: helper.BoolP(false),
					},
					{
						Name:      "수정",
						IsAllowed: helper.BoolP(false),
					},
					{
						Name:      "삭제",
						IsAllowed: helper.BoolP(false),
					},
				},
			},
		},
	}

	return configuration
}
