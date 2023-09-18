package mail

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"

	"github.com/openinfradev/tks-api/pkg/log"
	"github.com/spf13/viper"
	"gopkg.in/gomail.v2"
)

var (
	//go:embed contents/*.html
	templateFS                     embed.FS
	goClient                       *gomail.Message
	mailProvider                   string
	host, username, password, from string
	port                           int
)

type Mailer interface {
	SetMessage(m *MessageInfo)
	SendMail() error
}

type MessageInfo struct {
	From    string
	To      string
	Subject string
	Body    string
}

type SmtpMailer struct {
	client   *gomail.Message
	Host     string
	Port     int
	Username string
	Password string

	message *MessageInfo
}

func (s *SmtpMailer) SetMessage(m *MessageInfo) {
	s.message = m
}

func (s *SmtpMailer) SendMail() error {
	s.client.SetHeader("From", s.message.From)
	s.client.SetHeader("To", s.message.To)
	s.client.SetHeader("Subject", s.message.Subject)
	s.client.SetBody("text/html", s.message.Body)

	d := gomail.NewDialer(s.Host, s.Port, s.Username, s.Password)

	if err := d.DialAndSend(s.client); err != nil {
		log.Errorf("failed to send email, %v", err)
		return err
	}

	return nil
}

func Initialize() error {
	mailProvider = viper.GetString("mail-provider")
	if mailProvider != "smtp" {
		mailProvider = "aws"
	}

	switch mailProvider {
	case "aws":
		if err := initialize(); err != nil {
			log.Errorf("aws config initialize error, %v", err)
			return err
		}
		from = "tks-dev@sktelecom.com"
	case "smtp":
		host = viper.GetString("smtp-host")
		port = viper.GetInt("smtp-port")
		username = viper.GetString("smtp-username")
		password = viper.GetString("smtp-password")
		from = viper.GetString("smtp-from-email")

		if host == "" {
			log.Error("smtp-host is not set")
			return fmt.Errorf("smtp-host is not set")
		}
		if port == 0 {
			log.Error("smtp-port is not set")
			return fmt.Errorf("smtp-port is not set")
		}
		if username == "" {
			log.Error("smtp-username is not set")
			return fmt.Errorf("smtp-username is not set")
		}
		if password == "" {
			log.Error("smtp-password is not set")
			return fmt.Errorf("smtp-password is not set")
		}
		if from == "" {
			log.Error("smtp-from-email is not set")
			return fmt.Errorf("smtp-from-email is not set")
		}

		goClient = gomail.NewMessage()
	}

	return nil
}

func New() Mailer {
	var mailer Mailer

	switch mailProvider {
	case "aws":
		mailer = NewAwsMailer()
		log.Infof("aws ses mailer, %v", mailer)

	case "smtp":
		mailer = NewSmtpMailer()
		log.Infof("smtp mailer, %v", mailer)
	}

	return mailer
}

func NewSmtpMailer() *SmtpMailer {
	mailer := &SmtpMailer{
		client:   goClient,
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
	}

	return mailer
}

func MakeVerityIdentityMessage(mailer Mailer, to, code string) error {
	subject := "[TKS] [인증번호:" + code + "] 인증번호가 발급되었습니다."

	tmpl, err := template.ParseFS(templateFS, "contents/authcode.html")
	if err != nil {
		log.Errorf("failed to parse template, %v", err)
		return err
	}

	data := map[string]string{"AuthCode": code}

	var tpl bytes.Buffer
	if err := tmpl.Execute(&tpl, data); err != nil {
		log.Errorf("failed to execute template, %v", err)
		return err
	}

	mailer.SetMessage(
		&MessageInfo{from, to, subject, tpl.String()},
	)

	return nil
}

func MakeTemporaryPasswordMessage(mailer Mailer, to, randomPassword string) error {
	subject := "[TKS] 임시 비밀번호가 발급되었습니다."

	tmpl, err := template.ParseFS(templateFS, "contents/temporary_password.html")
	if err != nil {
		log.Errorf("failed to parse template, %v", err)
		return err
	}

	data := map[string]string{"TemporaryPassword": randomPassword}

	var tpl bytes.Buffer
	if err := tmpl.Execute(&tpl, data); err != nil {
		log.Errorf("failed to execute template, %v", err)
		return err
	}

	mailer.SetMessage(
		&MessageInfo{from, to, subject, tpl.String()},
	)

	return nil
}

func MakeGeneratingOrganizationMessage(mailer Mailer, organizationId string, organizationName string,
	to string, userAccountId string, randomPassword string) error {
	subject := "[TKS] 조직이 생성되었습니다."

	tmpl, err := template.ParseFS(templateFS, "contents/organization_creation.html")
	if err != nil {
		log.Errorf("failed to parse template, %v", err)
		return err
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
		return err
	}

	mailer.SetMessage(
		&MessageInfo{from, to, subject, tpl.String()},
	)

	return nil
}
