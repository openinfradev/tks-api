package ses

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	awsSes "github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/openinfradev/tks-api/pkg/log"
	"github.com/spf13/viper"
)

var Client *awsSes.Client

func init() {
	if viper.GetString("AWS_ACCESS_KEY_ID") != "" || viper.GetString("AWS_SECRET_ACCESS_KEY") != "" {
		log.Warn("aws secret is used on env. Be aware of security")

	}
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("ap-northeast-2"))
	//cfg, err := config.NewEnvConfig()
	if err != nil {
		log.Fatalf("aws configuration error, " + err.Error())
	}

	Client = awsSes.NewFromConfig(cfg)
}
func SendEmailForVerityIdentity(client *awsSes.Client, targetEmailAddress string, code string) error {
	subject := "[TKS] Code: [" + code + "] – Email Address Verification Request"
	body := "Dear TKS Customer,\n\nWe have received a request to authorize this email address for use with TKS." +
		"If you requested this verification, please type the authorization code below.\n" +
		"Authorization Code: " + code + "\n\nIf you did not request this verification, please ignore this email." +
		"Thank you for using TKS.\n\nSincerely,\nTKS Team"

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
		Source: aws.String("cho4036@gmail.com"),
	}

	if _, err := client.SendEmail(context.Background(), input); err != nil {
		log.Errorf("failed to send email, %v", err)
		return err
	}

	return nil
}

func SendEmailForTemporaryPassword(client *awsSes.Client, targetEmailAddress string, randomPassword string) error {
	subject := "[TKS] 비밀번호 초기화"
	body := "임시 비밀번호: " + randomPassword + "\n\n"

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
		Source: aws.String("cho4036@gmail.com"),
	}

	if _, err := client.SendEmail(context.Background(), input); err != nil {
		log.Errorf("failed to send email, %v", err)
		return err
	}

	return nil
}
