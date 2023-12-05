package mail

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
