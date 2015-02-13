package smtp

import (
	"fmt"
	toml "github.com/extrame/go-toml-config"
	"github.com/extrame/smtpoverttl"
	"io"
	"log"
	"net/smtp"
	"path/filepath"
	"text/template"
)

var Daemon = new(_SmtpSender)

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
		if c, err = smtpoverttl.DialTTL(fmt.Sprintf("%s:%d", s.Server, s.Port), nil); err == nil {
			s.sendMail(c, template, receivers, args)
		}
	} else {
		if c, err = smtp.Dial(*s.Server); err == nil {
			s.sendMail(c, template, receivers, args)
		}
	}
	return
}

func (s *_SmtpSender) sendMail(c client, mail_body *template.Template, receivers []string, obj interface{}) {
	if err := c.Auth(smtpoverttl.PlainAuth("", *s.User, *s.Pwd, *s.Server)); err == nil {
		for _, receiver := range receivers {
			if err = c.Mail(*s.User); err == nil {
				if err = c.Rcpt(receiver); err == nil {
					// Send the email body.
					if wc, err := c.Data(); err == nil {
						defer wc.Close()
						if err = mail_body.Execute(wc, obj); err != nil {
							log.Println(err)
						}
					}
				}
			}
		}
	}
}
