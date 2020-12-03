package email

import (
	"github.com/go-gomail/gomail"
	"github.com/sparrowhawktech/go-lib/util"
	"strings"
)

type SenderConfig struct {
	Sender     *string `json:"sender"`
	SenderName *string `json:"senderName"`
	SmtpUser   *string `json:"smtpUser"`
	SmtpPass   *string `json:"smtpPass"`
	Host       *string `json:"host"`
	Port       *int    `json:"port"`
	Bcc        *string `json:"bcc"`
}

type Sender struct {
	config *SenderConfig
}

func (o *Sender) Send(message MessageData) {
	m := gomail.NewMessage()
	m.SetBody("text/html", string(message.Body))

	recipient := *message.Recipient

	headers := map[string][]string{
		"From":         {m.FormatAddress(*o.config.Sender, *o.config.SenderName)},
		"To":           strings.Split(recipient, ";"),
		"Subject":      {*message.Subject},
		"Content-Type": {*message.ContentType},
	}
	if o.config.Bcc != nil {
		headers["Bcc"] = []string{*o.config.Bcc}
	}
	m.SetHeaders(headers)

	util.Logger("info").Printf("Sending email: %v", headers)

	d := gomail.NewDialer(*o.config.Host, *o.config.Port, *o.config.SmtpUser, *o.config.SmtpPass)
	err := d.DialAndSend(m)
	util.CheckErr(err)
}

func NewSender(config SenderConfig) *Sender {
	return &Sender{config: &config}
}
