package gomail

import (
	"bm_binus/internal/config"
	"bm_binus/pkg/util/general"
	"errors"
	"strconv"

	"github.com/sirupsen/logrus"
	"gopkg.in/gomail.v2"
)

func SendMail(recipient, subject, bodyHtml string) error {
	if bodyHtml == "" {
		return errors.New("error parsing body html")
	}

	mailer := gomail.NewMessage()
	mailer.SetHeader("From", config.Get().Gomail.AuthEmail, config.Get().Gomail.SenderName)
	mailer.SetHeader("To", recipient)
	mailer.SetHeader("Subject", subject)
	mailer.SetBody("text/plain", general.ParseTemplateEmailToPlainText(bodyHtml))
	mailer.AddAlternative("text/html", bodyHtml)

	portMail, _ := strconv.Atoi(config.Get().Gomail.SmtpPort)
	dialer := gomail.NewDialer(
		config.Get().Gomail.SmtpHost,
		portMail,
		config.Get().Gomail.AuthEmail,
		config.Get().Gomail.AuthPassword,
	)
	dialer.SSL = true

	err := dialer.DialAndSend(mailer)
	if err != nil {
		return err
	}

	logrus.Info("Mail sent!")
	return nil
}
