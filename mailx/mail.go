package mailx

import (
	"errors"
	"github.com/jordan-wright/email"
	"github.com/madlabx/pkgx/viperx"
	"net/smtp"
	"strings"
	"sync"
)

type MailType struct {
	addr      string
	body      string
	title     string
	from      string
	recv_list []string
	mail_to   string
	mail_from string
	auth      smtp.Auth
	nodeName  string
}

type MailClass int

const (
	MailClassReport MailClass = 1
)

type NodeNamePrefix string

const (
	NodeNamePrefixPub NodeNamePrefix = "p"
	NodeNamePrefixDev NodeNamePrefix = "d"
)

type loginAuth struct {
	username, password string
}

type MailContext struct {
	UserFrom      string
	UserAuth      string
	Password      string
	SendToPublish []string
	SendToDev     []string
	Identify      string
	SmtpPort      string
	SmtpHost      string
}

var mc *MailContext

func (mc *MailContext) getSendTo(c MailClass, n string) []string {
	r := strings.Index(n, string(NodeNamePrefixPub))
	if c == MailClassReport && r >= 0 {
		return mc.SendToPublish
	} else {
		return mc.SendToDev
	}
}

func InitMailContext(userFrom, userAuth, password, identify, smtpHost, smtpPort string, sendToPublish, sendToDev []string) *MailContext {
	mc = &MailContext{
		UserFrom:      userFrom,
		UserAuth:      userAuth,
		Password:      password,
		SendToPublish: sendToPublish,
		SendToDev:     sendToDev,
		Identify:      identify,
		SmtpPort:      smtpPort,
		SmtpHost:      smtpHost,
	}

	return mc
}

var once sync.Once

func newEmail() *email.Email {
	if mc == nil {
		return nil
	}
	e := email.NewEmail()
	e.From = mc.UserFrom

	return e
}

func SendMailHtml(title, body string) error {
	// 创建邮件内容
	e := newEmail()
	if e == nil {
		return errors.New("meed call InitMailContext")
	}

	nodeName := viperx.GetString("tradeNodeName", "")
	if len(nodeName) != 0 {
		nodeName = "[" + nodeName + "]"
	}

	if strings.Index(nodeName, "pub") >= 0 {
		e.To = mc.SendToPublish
	} else {
		e.To = mc.SendToDev
	}

	e.Subject = nodeName + title

	e.HTML = []byte(body)
	return e.Send(mc.SmtpHost+":"+mc.SmtpPort, smtp.PlainAuth("", mc.UserAuth, mc.Password, mc.SmtpHost))
}

func SendMailHtmlWithAttach(nt MailClass, title, body string, attachFiles []string) error {
	// 创建邮件内容
	e := newEmail()
	e.From = UserFrom

	nodeName := viperx.GetString("tradeNodeName", "")
	if len(nodeName) != 0 {
		nodeName = "[" + nodeName + "]"
	}

	e.To = getSendTo(nt, nodeName)

	e.Subject = nodeName + title
	for _, file := range attachFiles {
		e.AttachFile(file)
	}
	e.HTML = []byte(body)
	return e.Send(mc.SmtpHost+":"+mc.SmtpPort, smtp.PlainAuth("", mc.UserAuth, mc.Password, mc.SmtpHost))
}
