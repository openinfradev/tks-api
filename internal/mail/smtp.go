package mail

import (
	"embed"
	"fmt"
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

func New(m *MessageInfo) Mailer {
	var mailer Mailer

	switch mailProvider {
	case "aws":
		mailer = NewAwsMailer(m)
		log.Infof("aws ses mailer, %v", mailer)

	case "smtp":
		mailer = NewSmtpMailer(m)
		log.Infof("smtp mailer, %v", mailer)
	}

	return mailer
}

func NewSmtpMailer(m *MessageInfo) *SmtpMailer {
	mailer := &SmtpMailer{
		client:   goClient,
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		message:  m,
	}

	return mailer
}
