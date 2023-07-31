package ses

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"html/template"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	awsSes "github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/openinfradev/tks-api/pkg/log"
	"github.com/spf13/viper"
)

//go:embed contents/*.html
var templateFS embed.FS

var Client *awsSes.Client

const (
	senderEmailAddress = "tks-dev@sktelecom.com"
)

func Initialize() error {
	if viper.GetString("aws-access-key-id") != "" || viper.GetString("aws-secret-access-key") != "" {
		log.Warn("aws access key information is used on env. Be aware of security")
	}
	if viper.GetString("aws-access-key-id") != "" {
		err := os.Setenv("AWS_ACCESS_KEY_ID", viper.GetString("aws-access-key-id"))
		if err != nil {
			return err
		}
	}
	if viper.GetString("aws-secret-access-key") != "" {
		err := os.Setenv("AWS_SECRET_ACCESS_KEY", viper.GetString("aws-secret-access-key"))
		if err != nil {
			return err
		}
	}
	if viper.GetString("aws-region") != "" {
		err := os.Setenv("AWS_REGION", viper.GetString("aws-region"))
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("aws region is not set")
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("ap-northeast-2"))
	if err != nil {
		return err
	}

	Client = awsSes.NewFromConfig(cfg)
	return nil
}
func SendEmailForVerityIdentity(client *awsSes.Client, targetEmailAddress string, code string) error {
	subject := "[TKS] [인증번호:" + code + "] 인증번호가 발급되었습니다."

	tmpl, err := template.ParseFS(templateFS, "contents/authcode.html")
	if err != nil {
		log.Errorf("failed to parse template, %v", err)
		return err
	}

	type TemplateData struct {
		AuthCode string
	}

	data := TemplateData{
		AuthCode: code,
	}

	var tpl bytes.Buffer
	if err := tmpl.Execute(&tpl, data); err != nil {
		log.Errorf("failed to execute template, %v", err)
		return err
	}

	body := tpl.String()

	err = sendEmail(client, targetEmailAddress, subject, body)
	if err != nil {
		return err
	}
	return nil
}

func SendEmailForTemporaryPassword(client *awsSes.Client, targetEmailAddress string, randomPassword string) error {
	subject := "[TKS] 임시 비밀번호가 발급되었습니다."

	tmpl, err := template.ParseFS(templateFS, "contents/temporary_password.html")
	if err != nil {
		log.Errorf("failed to parse template, %v", err)
		return err
	}

	type TemplateData struct {
		TemporaryPassword string
	}

	data := TemplateData{
		TemporaryPassword: randomPassword,
	}

	var tpl bytes.Buffer
	if err := tmpl.Execute(&tpl, data); err != nil {
		log.Errorf("failed to execute template, %v", err)
		return err
	}

	body := tpl.String()

	err = sendEmail(client, targetEmailAddress, subject, body)
	if err != nil {
		return err
	}
	return nil
}

func SendEmailForGeneratingOrganization(client *awsSes.Client, organizationId string, organizationName string,
	targetEmailAddress string, userAccountId string, randomPassword string) error {
	subject := "[TKS] 조직이 생성되었습니다."

	tmpl, err := template.ParseFS(templateFS, "contents/organization_creation.html")
	if err != nil {
		log.Errorf("failed to parse template, %v", err)
		return err
	}

	type TemplateData struct {
		OrganizationId   string
		Id               string
		Password         string
		OrganizationName string
		AdminName        string
	}

	data := TemplateData{
		OrganizationId:   organizationId,
		Id:               userAccountId,
		Password:         randomPassword,
		OrganizationName: organizationName,
		AdminName:        userAccountId,
	}

	var tpl bytes.Buffer
	if err := tmpl.Execute(&tpl, data); err != nil {
		log.Errorf("failed to execute template, %v", err)
		return err
	}
	body := tpl.String()

	err = sendEmail(client, targetEmailAddress, subject, body)
	if err != nil {
		return err
	}
	return nil
}

func sendEmail(client *awsSes.Client, targetEmailAddress string, subject string, htmlBody string) error {
	input := &awsSes.SendEmailInput{
		Destination: &types.Destination{
			ToAddresses: []string{targetEmailAddress},
		},
		Message: &types.Message{
			Body: &types.Body{
				Html: &types.Content{
					Data: aws.String(htmlBody),
				},
			},
			Subject: &types.Content{
				Data: aws.String(subject),
			},
		},
		Source: aws.String(senderEmailAddress),
	}

	if _, err := client.SendEmail(context.Background(), input); err != nil {
		log.Errorf("failed to send email, %v", err)
		return err
	}

	return nil
}
