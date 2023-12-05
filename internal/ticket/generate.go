package ticket

import (
	"graduation/internal/entity"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func (t *TicketToken) Generate(tick *entity.Ticket) error {
	claims := TicketClaims{
		UserID:  tick.UserID,
		EventID: tick.EventID,
		Exp:     time.Now().Add(time.Hour * time.Duration(tick.Exp)),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(tick.Exp))),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(t.secretKey))
	if err != nil {
		return err
	}

	tick.Token = signedToken

	return nil
}
