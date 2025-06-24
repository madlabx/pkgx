package mailx

import (
	"net/smtp"
	"strings"

	"github.com/jordan-wright/email"
	"github.com/madlabx/pkgx/viperx"
)

type MailType int

const (
	MailClassReport MailType = iota
	MailClassAlarm
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
	UserFrom       string
	UserAuth       string
	Password       string
	MailListReport []string
	MailListAlarm  []string
	MailListDev    []string
	Identify       string
	SmtpPort       string
	SmtpHost       string
}

var mc *MailContext = nil

func (mc *MailContext) GetSendTo(c MailType, n string) []string {
	r := strings.Index(n, string(NodeNamePrefixPub))
	if r == -1 {
		return mc.MailListDev
	}

	switch c {
	case MailClassReport:
		return mc.MailListReport
	case MailClassAlarm:
		return mc.MailListAlarm
	default:
		return mc.MailListAlarm
	}
}

func NewMailContext(a MailContext) *MailContext {
	*mc = a
	return mc
}

func InitMailContext(userFrom, userAuth, password, identify, smtpHost, smtpPort string, mailListReport, mailListAlarm, mailListDev []string) *MailContext {
	mc = &MailContext{
		UserFrom:       userFrom,
		UserAuth:       userAuth,
		Password:       password,
		MailListAlarm:  mailListAlarm,
		MailListReport: mailListReport,
		MailListDev:    mailListDev,
		Identify:       identify,
		SmtpPort:       smtpPort,
		SmtpHost:       smtpHost,
	}

	return mc
}

func newEmail() *email.Email {
	if mc == nil {
		panic("need InitMailContext")
	}
	e := email.NewEmail()
	e.From = mc.UserFrom

	return e
}

func SendMailHtml(t MailType, title, body string) error {
	return SendMailHtmlWithAttach(t, title, body, nil)
}

func SendMailHtmlWithAttach(nt MailType, title, body string, attachFiles []string) error {
	// 创建邮件内容
	e := newEmail()

	nodeName := viperx.GetString("tradeNodeName", "")
	if len(nodeName) != 0 {
		nodeName = "[" + nodeName + "]"
	}

	e.To = mc.GetSendTo(nt, nodeName)

	e.Subject = nodeName + title
	for _, file := range attachFiles {
		e.AttachFile(file)
	}
	e.HTML = []byte(body)
	return e.Send(mc.SmtpHost+":"+mc.SmtpPort, smtp.PlainAuth(mc.Identify, mc.UserAuth, mc.Password, mc.SmtpHost))
}
