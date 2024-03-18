package mail

import (
	"bytes"
	"context"
	"crypto/tls"
	"embed"
	"errors"
	"fmt"
	"io"
	"net"
	"net/smtp"
	"strings"
	"time"

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
	SendMail(ctx context.Context) error
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

func (s *SmtpMailer) SendMail(ctx context.Context) error {
	s.client.SetHeader("From", s.message.From)
	s.client.SetHeader("To", s.message.To)
	s.client.SetHeader("Subject", s.message.Subject)
	s.client.SetBody("text/html", s.message.Body)

	d := NewDialer(s.Host, s.Port, s.Username, s.Password)
	if err := d.DialAndSend(s.client); err != nil {
		log.Errorf(ctx, "failed to send email, %v", err)
		return err
	}

	return nil
}

func Initialize(ctx context.Context) error {
	mailProvider = viper.GetString("mail-provider")
	if mailProvider != "smtp" {
		mailProvider = "aws"
	}

	switch mailProvider {
	case "aws":
		if err := initialize(ctx); err != nil {
			log.Errorf(ctx, "aws config initialize error, %v", err)
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
			log.Error(ctx, "smtp-host is not set")
			return fmt.Errorf("smtp-host is not set")
		}
		if port == 0 {
			log.Error(ctx, "smtp-port is not set")
			return fmt.Errorf("smtp-port is not set")
		}
		if username == "" {
			log.Error(ctx, "smtp-username is not set")
			return fmt.Errorf("smtp-username is not set")
		}
		if password == "" {
			log.Error(ctx, "smtp-password is not set")
			return fmt.Errorf("smtp-password is not set")
		}
		if from == "" {
			log.Error(ctx, "smtp-from-email is not set")
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
		log.Infof(context.TODO(), "aws ses mailer, %v", mailer)

	case "smtp":
		mailer = NewSmtpMailer(m)
		log.Infof(context.TODO(), "smtp mailer, %v", mailer)
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

// smtp 25 NonTLS Support
type unencryptedAuth struct {
	smtp.Auth
}

func (a unencryptedAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	s := *server
	s.TLS = true
	return a.Auth.Start(&s)
}

// type, function override
var (
	netDialTimeout = net.DialTimeout
	tlsClient      = tls.Client
	smtpNewClient  = func(conn net.Conn, host string) (smtpClient, error) {
		return smtp.NewClient(conn, host)
	}
)

type Dialer struct {
	Host      string
	Port      int
	Username  string
	Password  string
	Auth      smtp.Auth
	SSL       bool
	TLSConfig *tls.Config
	LocalName string
}

func NewDialer(host string, port int, username, password string) *Dialer {
	return &Dialer{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		SSL:      port == 465,
	}
}

func (d *Dialer) DialAndSend(m ...*gomail.Message) error {
	s, err := d.Dial()
	if err != nil {
		return err
	}
	defer s.Close()

	return gomail.Send(s, m...)
}

func addr(host string, port int) string {
	return fmt.Sprintf("%s:%d", host, port)
}

func (d *Dialer) Dial() (gomail.SendCloser, error) {
	conn, err := netDialTimeout("tcp", addr(d.Host, d.Port), 10*time.Second)
	if err != nil {
		return nil, err
	}

	if d.SSL {
		conn = tlsClient(conn, d.tlsConfig())
	}

	c, err := smtpNewClient(conn, d.Host)
	if err != nil {
		return nil, err
	}

	if d.LocalName != "" {
		if err := c.Hello(d.LocalName); err != nil {
			return nil, err
		}
	}

	if !d.SSL {
		if ok, _ := c.Extension("STARTTLS"); ok {
			if err := c.StartTLS(d.tlsConfig()); err != nil {
				c.Close()
				return nil, err
			}
		}
	}

	if d.Auth == nil && d.Username != "" {
		if ok, auths := c.Extension("AUTH"); ok {
			if strings.Contains(auths, "CRAM-MD5") {
				d.Auth = smtp.CRAMMD5Auth(d.Username, d.Password)
			} else if strings.Contains(auths, "LOGIN") &&
				!strings.Contains(auths, "PLAIN") {
				d.Auth = &loginAuth{
					username: d.Username,
					password: d.Password,
					host:     d.Host,
				}
			} else {
				// NonTLS SMTP 25 support
				d.Auth = unencryptedAuth{
					smtp.PlainAuth(
						"",
						d.Username,
						d.Password,
						d.Host,
					),
				}
				//d.Auth = smtp.PlainAuth("", d.Username, d.Password, d.Host)
			}
		}
	}

	if d.Auth != nil {
		if err = c.Auth(d.Auth); err != nil {
			c.Close()
			return nil, err
		}
	}

	return &smtpSender{c, d}, nil
}

func (d *Dialer) tlsConfig() *tls.Config {
	if d.TLSConfig == nil {
		return &tls.Config{ServerName: d.Host}
	}
	return d.TLSConfig
}

type smtpClient interface {
	Hello(string) error
	Extension(string) (bool, string)
	StartTLS(*tls.Config) error
	Auth(smtp.Auth) error
	Mail(string) error
	Rcpt(string) error
	Data() (io.WriteCloser, error)
	Quit() error
	Close() error
}

type loginAuth struct {
	username string
	password string
	host     string
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	if !server.TLS {
		advertised := false
		for _, mechanism := range server.Auth {
			if mechanism == "LOGIN" {
				advertised = true
				break
			}
		}
		if !advertised {
			return "", nil, errors.New("gomail: unencrypted connection")
		}
	}
	if server.Name != a.host {
		return "", nil, errors.New("gomail: wrong host name")
	}
	return "LOGIN", nil, nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if !more {
		return nil, nil
	}

	switch {
	case bytes.Equal(fromServer, []byte("Username:")):
		return []byte(a.username), nil
	case bytes.Equal(fromServer, []byte("Password:")):
		return []byte(a.password), nil
	default:
		return nil, fmt.Errorf("gomail: unexpected server challenge: %s", fromServer)
	}
}

type smtpSender struct {
	smtpClient
	d *Dialer
}

func (c *smtpSender) Send(from string, to []string, msg io.WriterTo) error {
	if err := c.Mail(from); err != nil {
		if err == io.EOF {
			// This is probably due to a timeout, so reconnect and try again.
			sc, derr := c.d.Dial()
			if derr == nil {
				if s, ok := sc.(*smtpSender); ok {
					*c = *s
					return c.Send(from, to, msg)
				}
			}
		}
		return err
	}

	for _, addr := range to {
		if err := c.Rcpt(addr); err != nil {
			return err
		}
	}

	w, err := c.Data()
	if err != nil {
		return err
	}

	if _, err = msg.WriteTo(w); err != nil {
		w.Close()
		return err
	}

	return w.Close()
}

func (c *smtpSender) Close() error {
	return c.Quit()
}
