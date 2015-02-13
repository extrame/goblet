package smtp

import (
	"crypto/tls"
	"fmt"
	toml "github.com/extrame/go-toml-config"
	"github.com/extrame/smtpoverttl"
	"io"
	"net/smtp"
	"path/filepath"
	"text/template"
)

var Daemon = new(_SmtpSender)
var StandardHeader = `To:{{ $.Receiver }}
From: {{ $.Sender}}
`

type _SmtpSender struct {
	Root      *string
	Server    *string
	User      *string
	Pwd       *string
	Ttl       *bool
	Port      *int
	Templates map[string]*template.Template
}

func (s *_SmtpSender) Init() error {
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
	s.Pwd = toml.String("mail.password", "")
	s.Ttl = toml.Bool("mail.ttl", false)
	s.Port = toml.Int("mail.port", 465)
	*s.Root = filepath.FromSlash(*s.Root)
	return
}

func SendTo(template_name string, receivers []string, args map[string]interface{}) (err error) {
	return Daemon.SendTo(template_name, receivers, args)
}

func (s *_SmtpSender) SendTo(template_name string, receivers []string, args map[string]interface{}) (err error) {

	var template *template.Template
	var ok bool

	if template, ok = s.Templates[template_name]; !ok {
		if template, err = template.ParseFiles(filepath.Join(*s.Root, template_name)); err == nil {
			s.Templates[template_name] = template
		} else {
			return
		}
	}

	var c client
	if *s.Ttl {
		var config tls.Config
		config.ServerName = *s.Server
		if c, err = smtpoverttl.DialTTL(fmt.Sprintf("%s:%d", *s.Server, *s.Port), &config); err == nil {
			return s.sendMail(c, template, receivers, args)
		}
	} else {
		if c, err = smtp.Dial(*s.Server); err == nil {
			return s.sendMail(c, template, receivers, args)
		}
	}
	return
}

func (s *_SmtpSender) sendMail(c client, mail_body *template.Template, receivers []string, args map[string]interface{}) (err error) {
	var standard_header_template *template.Template

	if standard_header_template, err = template.New("standard_header").Parse(StandardHeader); err == nil {
		if err = c.Auth(smtpoverttl.PlainAuth("", *s.User, *s.Pwd, *s.Server)); err == nil {
			for _, receiver := range receivers {
				if err = c.Mail(*s.User); err == nil {
					if err = c.Rcpt(receiver); err == nil {
						// Send the email body.
						var wc io.WriteCloser
						if wc, err = c.Data(); err == nil {
							defer wc.Close()
							if err = standard_header_template.Execute(wc, map[string]string{"Receiver": receiver, "Sender": *s.User}); err != nil {
								return
							}
							if err = mail_body.Execute(wc, args); err != nil {
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
