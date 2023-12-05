package mail

import (
	"fmt"

	"gopkg.in/gomail.v2"
)

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
