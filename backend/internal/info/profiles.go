package info

import (
	"net/http"
	"strings"

	"github.com/limon4ik-black/AI-buddy/backend/internal/contracts"
	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/httpapi"
	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/postgres"
)

func (i *Service) getProfile(r *http.Request) (contracts.Profile, error) {
	return postgres.One[contracts.Profile](r.Context(), i.DB, "SELECT * FROM profiles WHERE user_id=$1", httpapi.User(r).Subject)
}

func (i *Service) profile(w http.ResponseWriter, r *http.Request) error {
	p, e := i.getProfile(r)
	if e != nil {
		return e
	}
	httpapi.JSON(w, 200, p)
	return nil
}

func (i *Service) updateProfile(w http.ResponseWriter, r *http.Request) error {
	v, e := httpapi.Decode[struct {
		FullName string `json:"fullName"`
		Position string `json:"position"`
	}](r)
	if e != nil {
		return e
	}
	v.FullName = strings.TrimSpace(v.FullName)
	v.Position = strings.TrimSpace(v.Position)
	if v.FullName == "" || len(v.FullName) > 200 || v.Position == "" || len(v.Position) > 200 {
		return httpapi.Invalid("Укажите имя и должность длиной до 200 байт")
	}
	_, e = i.DB.Exec(r.Context(), "UPDATE profiles SET full_name=$2,position=$3 WHERE user_id=$1", httpapi.User(r).Subject, v.FullName, v.Position)
	if e != nil {
		return e
	}
	return i.profile(w, r)
}
