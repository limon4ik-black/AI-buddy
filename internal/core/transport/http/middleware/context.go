// Package middleware содержит HTTP-мидлвари transport-слоя.
package middleware

import (
	"context"
)

// User - данные об авторизованном пользователе, доступные в контексте.
type User struct {
	ID int64
}

// ctxKey - приватный тип ключа контекста, снаружи подделать нельзя.
type ctxKey struct{}

// NewContext кладет пользователя в контекст.
func NewContext(ctx context.Context, u User) context.Context {
	return context.WithValue(ctx, ctxKey{}, u)
}

// FromContext достает пользователя из контекста.
// ok = false означает, что мидлваря аутентификации не отработала.
func FromContext(ctx context.Context) (User, bool) {
	u, ok := ctx.Value(ctxKey{}).(User)
	return u, ok
}
