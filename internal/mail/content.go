package mail

import (
	"bytes"
	"html/template"

	"github.com/openinfradev/tks-api/pkg/log"
)

func MakeVerityIdentityMessage(to, code string) (*MessageInfo, error) {
	subject := "[TKS] [인증번호:" + code + "] 인증번호가 발급되었습니다."

	tmpl, err := template.ParseFS(templateFS, "contents/authcode.html")
	if err != nil {
		log.Errorf("failed to parse template, %v", err)
		return nil, err
	}

	data := map[string]string{"AuthCode": code}

	var tpl bytes.Buffer
	if err := tmpl.Execute(&tpl, data); err != nil {
		log.Errorf("failed to execute template, %v", err)
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

func MakeTemporaryPasswordMessage(to, randomPassword string) (*MessageInfo, error) {
	subject := "[TKS] 임시 비밀번호가 발급되었습니다."

	tmpl, err := template.ParseFS(templateFS, "contents/temporary_password.html")
	if err != nil {
		log.Errorf("failed to parse template, %v", err)
		return nil, err
	}

	data := map[string]string{"TemporaryPassword": randomPassword}

	var tpl bytes.Buffer
	if err := tmpl.Execute(&tpl, data); err != nil {
		log.Errorf("failed to execute template, %v", err)
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

func MakeGeneratingOrganizationMessage(
	organizationId string, organizationName string,
	to string, userAccountId string, randomPassword string) (*MessageInfo, error) {
	subject := "[TKS] 조직이 생성되었습니다."

	tmpl, err := template.ParseFS(templateFS, "contents/organization_creation.html")
	if err != nil {
		log.Errorf("failed to parse template, %v", err)
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
		log.Errorf("failed to execute template, %v", err)
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
