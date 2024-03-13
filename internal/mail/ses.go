package mail

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	awsSes "github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/openinfradev/tks-api/pkg/log"
	"github.com/spf13/viper"
)

var awsClient *awsSes.Client

type AwsMailer struct {
	client  *awsSes.Client
	message *MessageInfo
}

func (a *AwsMailer) SendMail(ctx context.Context) error {
	input := &awsSes.SendEmailInput{
		Destination: &types.Destination{
			ToAddresses: []string{a.message.To},
		},
		Message: &types.Message{
			Subject: &types.Content{
				Data: aws.String(a.message.Subject),
			},
			Body: &types.Body{
				Html: &types.Content{
					Data: aws.String(a.message.Body),
				},
			},
		},
		Source: aws.String(a.message.From),
	}

	if _, err := a.client.SendEmail(context.Background(), input); err != nil {
		log.Errorf(ctx, "failed to send email, %v", err)
		return err
	}

	return nil
}

func initialize(ctx context.Context) error {
	if viper.GetString("aws-access-key-id") != "" || viper.GetString("aws-secret-access-key") != "" {
		log.Warn(ctx, "aws access key information is used on env. Be aware of security")
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

	awsClient = awsSes.NewFromConfig(cfg)
	return nil
}

func NewAwsMailer(m *MessageInfo) *AwsMailer {
	mailer := &AwsMailer{
		client:  awsClient,
		message: m,
	}

	return mailer
}
