package httpapi

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

type Endpoint func(http.ResponseWriter, *http.Request) error

func JSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if status != 204 {
		_ = json.NewEncoder(w).Encode(value)
	}
}

func Decode[T any](r *http.Request) (T, error) {
	var v T
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		return v, Problem(415, "unsupported_media_type", "Ожидается application/json")
	}
	d := json.NewDecoder(io.LimitReader(r.Body, 1048577))
	d.DisallowUnknownFields()
	if err := d.Decode(&v); err != nil {
		return v, Invalid("Некорректное тело запроса")
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return v, Invalid("Ожидается один JSON-объект")
	}
	return v, nil
}
