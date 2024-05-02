package api_test

import (
	"github.com/openinfradev/tks-api/internal/delivery/api"
	"github.com/openinfradev/tks-api/internal/model"
	"testing"
)

func TestEndpointsUsage(t *testing.T) {
	var allEndpoints []string
	for _, v := range api.MapWithEndpoint {
		allEndpoints = append(allEndpoints, v.Name)
	}
	//allEndpoints := []Endpoint{
	//	Login, Logout, RefreshToken, FindId, // 계속해서 모든 Endpoint 추가
	//	// 나머지 Endpoint 상수들을 여기에 추가
	//}
	usageCount := make(map[string]int)
	ps := model.NewAdminPermissionSet()

	permissions := []*model.Permission{
		ps.Dashboard,
		ps.Notification,
		ps.Configuration,
		ps.ProjectManagement,
		ps.Stack,
		ps.Policy,
		ps.Common,
	}

	leafPermissions := make([]*model.Permission, 0)

	for _, perm := range permissions {
		leafPermissions = model.GetEdgePermission(perm, leafPermissions, nil)
	}

	// Permission 설정에서 Endpoint 사용 횟수 카운트
	for _, perm := range leafPermissions {
		countEndpoints(perm, usageCount)
	}

	var unusedEndpoints, duplicatedEndpoints []string

	// 미사용 또는 중복 사용된 Endpoint 확인 및 출력
	for _, endpoint := range allEndpoints {
		count, exists := usageCount[endpoint]
		if !exists {
			unusedEndpoints = append(unusedEndpoints, endpoint)
		} else if count > 1 {
			duplicatedEndpoints = append(duplicatedEndpoints, endpoint)
		}
	}

	for _, endpoint := range unusedEndpoints {
		t.Logf("Unused Endpoint: %s", endpoint)
	}

	t.Logf("\n")
	for _, endpoint := range duplicatedEndpoints {
		t.Logf("Duplicated Endpoint: %s", endpoint)
	}

}

func countEndpoints(perm *model.Permission, usageCount map[string]int) {
	for _, endpoint := range perm.Endpoints {
		usageCount[endpoint.Name]++
	}
}
