package auth

import (
	"net/http"
	"net/mail"
	"strings"

	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/httpapi"
)

func (a *Service) readCredentials(r *http.Request) (credentials, error) {
	c, e := httpapi.Decode[credentials](r)
	c.Email = strings.ToLower(strings.TrimSpace(c.Email))
	if e != nil {
		return c, e
	}
	address, e := mail.ParseAddress(c.Email)
	if e != nil || address.Address != c.Email || len(c.Email) > 254 {
		return c, httpapi.Invalid("Укажите корректный email")
	}
	if len(c.Password) < 10 || len(c.Password) > 72 {
		return c, httpapi.Invalid("Пароль должен содержать от 10 до 72 байт")
	}
	return c, nil
}
