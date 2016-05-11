package smtp

import (
	"net/mail"
	"testing"
	"text/template"
)

func TestSendMail(t *testing.T) {
	daemon := new(_SmtpSender)
	daemon.Port = new(int)
	daemon.Ttl = new(bool)
	daemon.User = new(string)
	daemon.Pwd = new(string)
	daemon.Root = new(string)
	daemon.Server = new(string)
	daemon.Templates = make(map[string]*template.Template)
	*daemon.Root = "/Users/liuming/Documents/workspace/datajia/mail"
	*daemon.User = "no-reply@shuhaidata.com"
	*daemon.Pwd = "SHuhai0806"
	*daemon.Port = 465
	*daemon.Server = "smtp.exmail.qq.com"
	*daemon.Ttl = true
	daemon.SendTo("default.html", "测试用", []mail.Address{mail.Address{"刘铭", "extrafliu@gmail.com"}}, map[string]interface{}{"tst": "tets"})
}
