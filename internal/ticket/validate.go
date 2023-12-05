package ticket

import (
	"errors"
	"graduation/internal/entity"

	"github.com/golang-jwt/jwt/v4"
)

func (t *TicketToken) Validate(tick *entity.Ticket) error {
	token, err := jwt.ParseWithClaims(tick.Token, &TicketClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(t.secretKey), nil
	})

	if err != nil {
		return err
	}

	claims, ok := token.Claims.(*TicketClaims)
	if !ok || !token.Valid {
		return errors.New("invalid token")
	}

	tick.UserID = claims.UserID
	tick.EventID = claims.EventID

	return nil
}
