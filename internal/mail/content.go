package mail

import (
	"bytes"
	"context"
	"html/template"

	"github.com/openinfradev/tks-api/pkg/log"
)

func MakeVerityIdentityMessage(ctx context.Context, to, code string) (*MessageInfo, error) {
	subject := "[TKS] [인증번호:" + code + "] 인증번호가 발급되었습니다."

	tmpl, err := template.ParseFS(templateFS, "contents/authcode.html")
	if err != nil {
		log.Errorf(ctx, "failed to parse template, %v", err)
		return nil, err
	}

	data := map[string]string{"AuthCode": code}

	var tpl bytes.Buffer
	if err := tmpl.Execute(&tpl, data); err != nil {
		log.Errorf(ctx, "failed to execute template, %v", err)
		return nil, err
	}

	m := &MessageInfo{
		From:    from,
		To:      []string{to},
		Subject: subject,
		Body:    tpl.String(),
	}

	return m, nil
}

func MakeTemporaryPasswordMessage(ctx context.Context, to, organizationId, accountId, randomPassword string) (*MessageInfo, error) {
	subject := "[TKS] 임시 비밀번호가 발급되었습니다."

	tmpl, err := template.ParseFS(templateFS, "contents/temporary_password.html")
	if err != nil {
		log.Errorf(ctx, "failed to parse template, %v", err)
		return nil, err
	}

	data := map[string]string{"TemporaryPassword": randomPassword, "OrganizationId": organizationId, "AccountId": accountId}

	var tpl bytes.Buffer
	if err := tmpl.Execute(&tpl, data); err != nil {
		log.Errorf(ctx, "failed to execute template, %v", err)
		return nil, err
	}

	m := &MessageInfo{
		From:    from,
		To:      []string{to},
		Subject: subject,
		Body:    tpl.String(),
	}

	return m, nil
}

func MakeGeneratingOrganizationMessage(
	ctx context.Context,
	organizationId string, organizationName string,
	to string, userAccountId string, randomPassword string) (*MessageInfo, error) {
	subject := "[TKS] 조직이 생성되었습니다."

	tmpl, err := template.ParseFS(templateFS, "contents/organization_creation.html")
	if err != nil {
		log.Errorf(ctx, "failed to parse template, %v", err)
		return nil, err
	}

	data := map[string]string{
		"OrganizationId":   organizationId,
		"Id":               userAccountId,
		"Password":         randomPassword,
		"OrganizationName": organizationName,
		"AdminName":        userAccountId,
	}

	var tpl bytes.Buffer
	if err := tmpl.Execute(&tpl, data); err != nil {
		log.Errorf(ctx, "failed to execute template, %v", err)
		return nil, err
	}

	m := &MessageInfo{
		From:    from,
		To:      []string{to},
		Subject: subject,
		Body:    tpl.String(),
	}

	return m, nil
}

func MakeSystemNotificationMessage(ctx context.Context, organizationId string, title string, to []string) (*MessageInfo, error) {
	subject := "[TKS] 시스템 알림이 발생하였습니다."

	tmpl, err := template.ParseFS(templateFS, "contents/system_notification.html")
	if err != nil {
		log.Errorf(ctx, "failed to parse template, %v", err)
		return nil, err
	}

	data := map[string]string{
		"OrganizationId": organizationId,
	}

	var tpl bytes.Buffer
	if err := tmpl.Execute(&tpl, data); err != nil {
		log.Errorf(ctx, "failed to execute template, %v", err)
		return nil, err
	}

	m := &MessageInfo{
		From:    from,
		To:      to,
		Subject: subject,
		Body:    tpl.String(),
	}

	return m, nil
}
