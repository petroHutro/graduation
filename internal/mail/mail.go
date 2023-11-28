package mail

// pnkofqqiobcynulu

// адрес почтового сервера — smtp.yandex.ru;
// защита соединения — SSL;
// порт — 465;

import (
	"fmt"
	"graduation/internal/config"

	"gopkg.in/gomail.v2"
)

type mailData struct {
	from string
}

type Mail struct {
	Con *gomail.Dialer
	mailData
}

func Init(conf *config.SMTP) (*Mail, error) {
	mail := Mail{
		Con:      gomail.NewDialer(conf.SMTPServer, conf.SMTPPort, conf.SMTPUsername, conf.SMTPPassword),
		mailData: mailData{from: conf.From},
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
	message.SetAddressHeader("From", m.from, "EVENT.NE")
	message.SetAddressHeader("To", to, "")
	message.SetHeader("Subject", "Event in 3 hours")
	// for i, image := range urls {
	// 	cid := "image" + strconv.Itoa(i)
	// 	body = strings.Replace(body, image, "cid:"+cid, 1)
	// 	message.Embed(image, gomail.Rename(cid))
	// 	message.EmbedURL(image, gomail.Rename(cid))
	// }
	message.SetBody("text/html", body)

	if err := m.Con.DialAndSend(message); err != nil {
		return fmt.Errorf("failed to send mail: %w", err)
	}

	return nil
}
