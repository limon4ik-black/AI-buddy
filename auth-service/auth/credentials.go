package auth

import (
	"net/http"
	"net/mail"
	"strings"

	"github.com/limon4ik-black/AI-buddy/shared/platform/httpapi"
)

func validEmail(email string) error {
	address, e := mail.ParseAddress(email)
	if e != nil || address.Address != email || len(email) > 254 {
		return httpapi.Invalid("Укажите корректный email")
	}
	return nil
}

func (a *Service) readCredentials(r *http.Request) (credentials, error) {
	c, e := httpapi.Decode[credentials](r)
	c.Email = strings.ToLower(strings.TrimSpace(c.Email))
	if e != nil {
		return c, e
	}
	if e = validEmail(c.Email); e != nil {
		return c, e
	}
	if len(c.Password) < 10 || len(c.Password) > 72 {
		return c, httpapi.Invalid("Пароль должен содержать от 10 до 72 байт")
	}
	return c, nil
}
