package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	internalApi "github.com/openinfradev/tks-api/internal/delivery/api"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
)

type fnAudit = func(ctx context.Context, out []byte, in []byte, statusCode int) (message string, description string)

var auditMap = map[internalApi.Endpoint]fnAudit{
	internalApi.CreateStack: func(ctx context.Context, out []byte, in []byte, statusCode int) (message string, description string) {
		input := domain.CreateStackRequest{}
		if err := json.Unmarshal(in, &input); err != nil {
			log.Error(ctx, err)
		}
		if isSuccess(statusCode) {
			return fmt.Sprintf("스택 [%s]을 생성하였습니다.", input.Name), ""
		} else {
			return fmt.Sprintf("스택 [%s]을 생성하는데 실패하였습니다.", input.Name), errorText(ctx, out)
		}
	}, internalApi.CreateProject: func(ctx context.Context, out []byte, in []byte, statusCode int) (message string, description string) {
		input := domain.CreateProjectRequest{}
		if err := json.Unmarshal(in, &input); err != nil {
			log.Error(ctx, err)
		}
		if isSuccess(statusCode) {
			return fmt.Sprintf("프로젝트 [%s]를 생성하였습니다.", input.Name), ""
		} else {
			return fmt.Sprintf("프로젝트 [%s]을 생성하는데 실패하였습니다.", input.Name), errorText(ctx, out)
		}
	}, internalApi.CreateCloudAccount: func(ctx context.Context, out []byte, in []byte, statusCode int) (message string, description string) {
		input := domain.CreateCloudAccountRequest{}
		if err := json.Unmarshal(in, &input); err != nil {
			log.Error(ctx, err)
		}
		if isSuccess(statusCode) {
			return fmt.Sprintf("클라우드 어카운트 [%s]를 생성하였습니다.", input.Name), ""
		} else {
			return fmt.Sprintf("클라우드 어카운트 [%s]을 생성하는데 실패하였습니다.", input.Name), errorText(ctx, out)
		}
	}, internalApi.DeleteCloudAccount: func(ctx context.Context, out []byte, in []byte, statusCode int) (message string, description string) {
		if isSuccess(statusCode) {
			output := domain.DeleteCloudAccountResponse{}
			if err := json.Unmarshal(out, &output); err != nil {
				log.Error(ctx, err)
			}
			return fmt.Sprintf("클라우드어카운트 [%s]를 삭제하였습니다.", output.Name), ""
		} else {
			return "클라우드어카운트를 삭제하는데 실패하였습니다. ", errorText(ctx, out)
		}
	}, internalApi.DeleteForceCloudAccount: func(ctx context.Context, out []byte, in []byte, statusCode int) (message string, description string) {
		if isSuccess(statusCode) {
			output := domain.DeleteCloudAccountResponse{}
			if err := json.Unmarshal(out, &output); err != nil {
				log.Error(ctx, err)
			}
			return fmt.Sprintf("클라우드어카운트 [%s]를 강제 삭제하였습니다.", output.Name), ""
		} else {
			return "클라우드어카운트를 강제 삭제하는데 실패하였습니다. ", errorText(ctx, out)
		}
	}, internalApi.CreateUser: func(ctx context.Context, out []byte, in []byte, statusCode int) (message string, description string) {
		input := domain.CreateUserRequest{}
		if err := json.Unmarshal(in, &input); err != nil {
			log.Error(ctx, err)
		}
		if isSuccess(statusCode) {
			return fmt.Sprintf("사용자 [%s]를 생성하였습니다.", input.Name), ""
		} else {
			return fmt.Sprintf("사용자 [%s]을 생성하는데 실패하였습니다.", input.Name), errorText(ctx, out)
		}
	}, internalApi.DeleteUser: func(ctx context.Context, out []byte, in []byte, statusCode int) (message string, description string) {
		if isSuccess(statusCode) {
			output := domain.DeleteUserResponse{}
			if err := json.Unmarshal(out, &output); err != nil {
				log.Error(ctx, err)
			}
			return fmt.Sprintf("사용자 [%s]를 삭제하였습니다.", output.AccountId), ""
		} else {
			return "사용자를 삭제하는데 실패하였습니다. ", errorText(ctx, out)
		}
	}, internalApi.Admin_CreateOrganization: func(ctx context.Context, out []byte, in []byte, statusCode int) (message string, description string) {
		input := domain.CreateOrganizationRequest{}
		if err := json.Unmarshal(in, &input); err != nil {
			log.Error(ctx, err)
		}
		if isSuccess(statusCode) {
			return fmt.Sprintf("조직 [%s]를 생성하였습니다.", input.Name), ""
		} else {
			return fmt.Sprintf("조직 [%s]을 생성하는데 실패하였습니다.", input.Name), errorText(ctx, out)
		}
	}, internalApi.Admin_DeleteOrganization: func(ctx context.Context, out []byte, in []byte, statusCode int) (message string, description string) {
		if isSuccess(statusCode) {
			output := domain.DeleteOrganizationResponse{}
			if err := json.Unmarshal(out, &output); err != nil {
				log.Error(ctx, err)
			}
			return fmt.Sprintf("조직 [ID:%s]를 삭제하였습니다.", output.ID), ""
		} else {
			return "조직을 삭제하는데 실패하였습니다. ", errorText(ctx, out)
		}
	}, internalApi.CreateAppServeApp: func(ctx context.Context, out []byte, in []byte, statusCode int) (message string, description string) {
		input := domain.CreateAppServeAppRequest{}
		if err := json.Unmarshal(in, &input); err != nil {
			log.Error(ctx, err)
		}
		if isSuccess(statusCode) {
			return fmt.Sprintf("앱서빙 [%s]를 생성하였습니다.", input.Name), ""
		} else {
			return fmt.Sprintf("앱서빙 [%s]을 생성하는데 실패하였습니다.", input.Name), errorText(ctx, out)
		}
	}, internalApi.Admin_CreateStackTemplate: func(ctx context.Context, out []byte, in []byte, statusCode int) (message string, description string) {
		input := domain.CreateStackTemplateRequest{}
		if err := json.Unmarshal(in, &input); err != nil {
			log.Error(ctx, err)
		}
		if isSuccess(statusCode) {
			return fmt.Sprintf("스택 템플릿 [%s]를 생성하였습니다.", input.Name), ""
		} else {
			return fmt.Sprintf("스택 템플릿 [%s]을 생성하는데 실패하였습니다.", input.Name), errorText(ctx, out)
		}
	}, internalApi.Admin_CreateUser: func(ctx context.Context, out []byte, in []byte, statusCode int) (message string, description string) {
		input := domain.CreateUserRequest{}
		if err := json.Unmarshal(in, &input); err != nil {
			log.Error(ctx, err)
		}
		if isSuccess(statusCode) {
			return fmt.Sprintf("어드민 [%s]를 생성하였습니다.", input.Name), ""
		} else {
			return fmt.Sprintf("어드민 [%s]을 생성하는데 실패하였습니다.", input.Name), errorText(ctx, out)
		}
	}, internalApi.Admin_CreatePolicyTemplate: func(ctx context.Context, out []byte, in []byte, statusCode int) (message string, description string) {
		input := domain.CreatePolicyTemplateRequest{}
		if err := json.Unmarshal(in, &input); err != nil {
			log.Error(ctx, err)
		}
		if isSuccess(statusCode) {
			return fmt.Sprintf("폴리시템플릿 [%s]를 생성하였습니다.", input.TemplateName), ""
		} else {
			return fmt.Sprintf("폴리시템플릿 [%s]을 생성하는데 실패하였습니다.", input.TemplateName), errorText(ctx, out)
		}
	}, internalApi.CreateSystemNotificationRule: func(ctx context.Context, out []byte, in []byte, statusCode int) (message string, description string) {
		input := domain.CreateSystemNotificationRuleRequest{}
		if err := json.Unmarshal(in, &input); err != nil {
			log.Error(ctx, err)
		}
		if isSuccess(statusCode) {
			return fmt.Sprintf("시스템알림설정 [%s]를 생성하였습니다.", input.Name), ""
		} else {
			return fmt.Sprintf("시스템알림설정 [%s]을 생성하는데 실패하였습니다.", input.Name), errorText(ctx, out)
		}
	}, internalApi.DeleteSystemNotificationRule: func(ctx context.Context, out []byte, in []byte, statusCode int) (message string, description string) {
		if isSuccess(statusCode) {
			output := domain.DeleteSystemNotificationRuleResponse{}
			if err := json.Unmarshal(out, &output); err != nil {
				log.Error(ctx, err)
			}
			return fmt.Sprintf("시스템알림설정 [%s]를 삭제하였습니다.", output.Name), ""
		} else {
			return "시스템알림설정을 삭제하는데 실패하였습니다. ", errorText(ctx, out)
		}
	}, internalApi.CreatePolicyTemplate: func(ctx context.Context, out []byte, in []byte, statusCode int) (message string, description string) {
		input := domain.CreatePolicyTemplateRequest{}
		if err := json.Unmarshal(in, &input); err != nil {
			log.Error(ctx, err)
		}
		if isSuccess(statusCode) {
			return fmt.Sprintf("정책템플릿 [%s]를 생성하였습니다.", input.TemplateName), ""
		} else {
			return fmt.Sprintf("정책템플릿 [%s]을 생성하는데 실패하였습니다.", input.TemplateName), errorText(ctx, out)
		}
	}, internalApi.CreatePolicy: func(ctx context.Context, out []byte, in []byte, statusCode int) (message string, description string) {
		input := domain.CreatePolicyRequest{}
		if err := json.Unmarshal(in, &input); err != nil {
			log.Error(ctx, err)
		}
		if isSuccess(statusCode) {
			return fmt.Sprintf("정책 [%s]를 생성하였습니다.", input.PolicyName), ""
		} else {
			return fmt.Sprintf("정책 [%s]을 생성하는데 실패하였습니다.", input.PolicyName), errorText(ctx, out)
		}
	},
}

func errorText(ctx context.Context, out []byte) string {
	var e httpErrors.RestError
	if err := json.NewDecoder(bytes.NewBuffer(out)).Decode(&e); err != nil {
		log.Error(ctx, err)
		return ""
	}
	return e.Text()
}

func isSuccess(statusCode int) bool {
	if statusCode >= 200 && statusCode < 300 {
		return true
	}
	return false
}
