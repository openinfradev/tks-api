package ses

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	awsSes "github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/openinfradev/tks-api/pkg/log"
	"github.com/spf13/viper"
	"os"
)

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
	subject := "[TKS][인증번호:" + code + "] – 요청하신 인증번호를 알려드립니다."
	body := "아래의 인증번호를 인증번호 입력창에 입력해 주세요.\n\n" +
		"인증번호: " + code + "\n\n" +
		"TKS를 이용해 주셔서 감사합니다.\nTKS Team 드림"

	input := &awsSes.SendEmailInput{
		Destination: &types.Destination{
			ToAddresses: []string{targetEmailAddress},
		},
		Message: &types.Message{
			Body: &types.Body{
				Text: &types.Content{
					Data: aws.String(body),
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

func SendEmailForTemporaryPassword(client *awsSes.Client, targetEmailAddress string, randomPassword string) error {
	subject := "[TKS] 비밀번호 초기화"
	body := "임시 비밀번호가 발급되었습니다.\n" +
		"로그인 후 비밀번호를 변경하여 사용하십시오.\n\n" +
		"임시 비밀번호: " + randomPassword + "\n\n" +
		"TKS를 이용해 주셔서 감사합니다.\nTKS Team 드림"

	input := &awsSes.SendEmailInput{
		Destination: &types.Destination{
			ToAddresses: []string{targetEmailAddress},
		},
		Message: &types.Message{
			Body: &types.Body{
				Text: &types.Content{
					Data: aws.String(body),
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
