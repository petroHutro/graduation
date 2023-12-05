package notification

import (
	"graduation/internal/config"
	"graduation/internal/storage"
)

type Notification struct {
	storage storage.Storage
	conf    *config.SMTP
}

func Init(st storage.Storage, conf *config.SMTP) *Notification {
	return &Notification{
		storage: st,
		conf:    conf,
	}
}
