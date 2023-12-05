package ticket

import (
	"graduation/internal/config"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type TicketToken struct {
	secretKey string
}

type TicketClaims struct {
	UserID  int       `json:"userID"`
	EventID int       `json:"eventID"`
	Exp     time.Time `json:"exp"`
	jwt.RegisteredClaims
}

func Init(conf *config.TicketKey) *TicketToken {
	return &TicketToken{secretKey: conf.TicketSecretKey}
}
