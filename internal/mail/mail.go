package mail

// pnkofqqiobcynulu

// адрес почтового сервера — smtp.yandex.ru;
// защита соединения — SSL;
// порт — 465;

import (
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/gomail.v2"
)

const (
	smtpServer   = "smtp.yandex.ru"
	smtpPort     = 465
	smtpUsername = "event.ne@yandex.ru"
	smtpPassword = "pnkofqqiobcynulu"

	from = "event.ne@yandex.ru"
)

type Mail struct {
	Con *gomail.Dialer
}

func Init() (*Mail, error) {
	mail := Mail{
		Con: gomail.NewDialer(smtpServer, smtpPort, smtpUsername, smtpPassword),
	}
	if err := mail.CheckConnection(); err != nil {
		return nil, fmt.Errorf("cannot connect: %w", err)
	}

	return &mail, nil
}

func (m *Mail) CheckConnection() error {
	sender, err := m.Con.Dial()
	if err != nil {
		return fmt.Errorf("cannot connect to SMTP server: %w", err)
	}

	if err := sender.Close(); err != nil {
		return fmt.Errorf("cannot to close connection: %w", err)
	}
	return nil
}

func (m *Mail) Send(to, body string, urls []string) error {
	message := gomail.NewMessage()
	message.SetAddressHeader("From", from, "Containerum Go Course")
	message.SetAddressHeader("To", to, "")
	message.SetHeader("Subject", "You are successfully registered!")
	for i, image := range urls {
		cid := "image" + strconv.Itoa(i)
		body = strings.Replace(body, image, "cid:"+cid, 1)
		message.Embed(image, gomail.Rename(cid))
	}
	message.SetBody("text/html", body)

	fmt.Println(body)
	// if err := m.Con.DialAndSend(message); err != nil {
	// 	return fmt.Errorf("failed to send mail: %w", err)
	// }

	return nil
}
