package smtp

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"net/mail"
	"net/smtp"
	"path/filepath"
	"text/template"

	toml "github.com/extrame/go-toml-config"
	"github.com/extrame/goblet"
	"github.com/extrame/smtpoverttl"
)

var Daemon = new(_SmtpSender)
var StandardHeader = `To:{{ $.Receiver }}
From: {{ $.Sender}}
Subject: {{ $.Subject }}
MIME-Version: 1.0
Content-Type: text/html; charset=UTF-8
Content-Transfer-Encoding: base64

{{ $.Body }}
`

type _SmtpSender struct {
	Root      *string
	Server    *string
	User      *string
	UserName  *string
	Pwd       *string
	Ttl       *bool
	Port      *int
	Templates map[string]*template.Template
}

func (s *_SmtpSender) Init(server *goblet.Server) error {
	s.Templates = make(map[string]*template.Template)
	return nil
}

type client interface {
	Auth(a smtp.Auth) error
	Mail(from string) error
	Rcpt(to string) error
	Data() (io.WriteCloser, error)
}

func (s *_SmtpSender) ParseConfig() (err error) {
	s.Root = toml.String("mail.root", "./mail")
	s.Server = toml.String("mail.server", "")
	s.User = toml.String("mail.user", "")
	s.UserName = toml.String("mail.user_name", "Sender")
	s.Pwd = toml.String("mail.password", "")
	s.Ttl = toml.Bool("mail.ttl", false)
	s.Port = toml.Int("mail.port", 25)
	*s.Root = filepath.FromSlash(*s.Root)
	return
}

func SendTo(template_name string, subject string, receivers []mail.Address, args map[string]interface{}) (err error) {
	return Daemon.SendTo(template_name, subject, receivers, args)
}

func (s *_SmtpSender) SendTo(template_name string, subject string, receivers []mail.Address, args map[string]interface{}) (err error) {

	var templa *template.Template
	var ok bool

	if templa, ok = s.Templates[template_name]; !ok {
		if templa, err = template.ParseFiles(filepath.Join(*s.Root, template_name)); err == nil {
			s.Templates[template_name] = templa
		} else {
			return
		}
	}

	var c client
	if *s.Ttl {
		var config tls.Config
		config.ServerName = *s.Server
		if c, err = smtpoverttl.DialTTL(fmt.Sprintf("%s:%d", *s.Server, *s.Port), &config); err == nil {
			return s.sendMail(c, subject, templa, receivers, args)
		}
	} else {
		if c, err = smtp.Dial(fmt.Sprintf("%s:%d", *s.Server, *s.Port)); err == nil {
			return s.sendMail(c, subject, templa, receivers, args)
		}
	}
	return
}

func (s *_SmtpSender) sendMail(c client, subject string, mail_body *template.Template, receivers []mail.Address, args map[string]interface{}) (err error) {
	b64 := base64.NewEncoding("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/")

	var standard_header_template *template.Template

	if standard_header_template, err = template.New("standard_header").Parse(StandardHeader); err == nil {
		if err = c.Auth(smtpoverttl.PlainAuth("", *s.User, *s.Pwd, *s.Server)); err == nil {
			for _, receiver := range receivers {
				if err = c.Mail(*s.User); err == nil {
					if err = c.Rcpt(receiver.Address); err == nil {
						// Send the email body.
						var wc io.WriteCloser
						if wc, err = c.Data(); err == nil {
							defer wc.Close()

							from := mail.Address{*s.UserName, *s.User}

							body_writer := new(bytes.Buffer)
							if err = mail_body.Execute(body_writer, args); err != nil {
								return
							}

							if err = standard_header_template.Execute(wc, map[string]string{
								"Receiver": receiver.String(),
								"Sender":   from.String(),
								"Body":     b64.EncodeToString(body_writer.Bytes()),
								"Subject":  fmt.Sprintf("=?UTF-8?B?%s?=", b64.EncodeToString([]byte(subject))),
							}); err != nil {
								return
							}
						}
					}
				}
			}
		}
	}
	return
}
