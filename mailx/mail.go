package mailx

import (
	"errors"
	"fmt"
	"github.com/jordan-wright/email"
	"github.com/madlabx/pkgx/viperx"
	"net/smtp"
	"strings"
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

var (
	SendToPublish = []string{"chukaiyan@foxmail.com", "jonathanwang.wzz@hotmail.com"}
	SendToDev     = []string{"zhongzi253@hotmail.com"}
)

func getSendTo(c MailClass, n string) []string {
	r := strings.Index(n, string(NodeNamePrefixPub))
	if c == MailClassReport && r >= 0 {
		return SendToPublish
	} else {
		return SendToDev
	}
}

type loginAuth struct {
	username, password string
}

const (
	UserFrom = "trader_reporter<77264952@qq.com>"
	UserAuth = "77264952@qq.com"
	Password = "ddyynbvevqqtbicg"
)

func (l loginAuth) Start(server *smtp.ServerInfo) (proto string, toServer []byte, err error) {
	return "LOGIN", []byte(l.username), nil
}

func (l loginAuth) Next(fromServer []byte, more bool) (toServer []byte, err error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(l.username), nil
		case "Password:":
			return []byte(l.password), nil
		default:
			return nil, errors.New("unknown field: " + string(fromServer))
		}
	}
	return nil, nil
}

func getLoginAuth(name, pw string) smtp.Auth {
	return &loginAuth{
		username: name,
		password: pw,
	}
}
func NewMail() *MailType {
	auth := getLoginAuth(
		UserAuth,
		Password,
	)

	nodeName := viperx.GetString("tradeNodeName", "")
	if len(nodeName) != 0 {
		nodeName = "[" + nodeName + "]"
	}

	mailTo := strings.Join(SendToPublish, `;`)
	return &MailType{
		addr:      "smtp.qq.com:587",
		recv_list: SendToPublish,
		mail_to:   mailTo,
		mail_from: UserFrom,
		auth:      auth,
		body:      "",
		nodeName:  nodeName,
	}
}

func (h *MailType) WriteMailBody(format string, args ...interface{}) {
	h.body += fmt.Sprintf(format, args...)
}

func (h *MailType) WriteMailTitle(title string) {
	//title = fmt.Sprintf("有效场次(%d)_最大(%.2f)", g_sum_infor.count, g_sum_infor.max)
	h.title = title
}

func (h *MailType) Send(title, body string) error {
	h.body += body
	h.title = title
	return smtp.SendMail(
		h.addr,
		h.auth,
		UserAuth,
		h.recv_list,
		[]byte("To: "+h.mail_to+",\r\nFrom: "+h.mail_from+",\r\nSubject: "+h.title+"\r\n\r\n"+h.body+"\r\n"),
		//[]byte(sub+content),
	)
}

func SendMail(title, body string) error {
	h := NewMail()
	h.body += body
	h.title = h.nodeName + title
	return smtp.SendMail(
		h.addr,
		h.auth,
		UserAuth,
		h.recv_list,
		[]byte("To: "+h.mail_to+",\r\nFrom: "+h.mail_from+",\r\nSubject: "+h.title+"\r\n\r\n"+h.body+"\r\n"),
		//[]byte(sub+content),
	)

}

func SendMailHtml(title, body string) error {
	// 创建邮件内容
	e := email.NewEmail()
	e.From = UserFrom

	nodeName := viperx.GetString("tradeNodeName", "")
	if len(nodeName) != 0 {
		nodeName = "[" + nodeName + "]"
	}

	if strings.Index(nodeName, "pub") >= 0 {
		e.To = SendToPublish
	} else {
		e.To = SendToDev
	}

	e.Subject = nodeName + title

	e.HTML = []byte(body)
	return e.Send("smtp.qq.com:587", smtp.PlainAuth("", UserAuth, Password, "smtp.qq.com"))
}

func SendMailHtmlWithAttach(nt MailClass, title, body string, attachFiles []string) error {
	// 创建邮件内容
	e := email.NewEmail()
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
	return e.Send("smtp.qq.com:587", smtp.PlainAuth("", UserAuth, Password, "smtp.qq.com"))
}
