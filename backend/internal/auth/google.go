package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/config"
	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/httpapi"
	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/idgen"
)

type googleProfile struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	EmailVerified bool   `json:"email_verified"`
}

func googleEnv() (id, secret, redirect string) {
	return config.Env("GOOGLE_CLIENT_ID", ""), config.Env("GOOGLE_CLIENT_SECRET", ""), config.Env("GOOGLE_REDIRECT_URL", "")
}

// googleConf builds the OAuth config from environment; nil means disabled.
func googleConf() *oauth2.Config {
	id, secret, redirect := googleEnv()
	if id == "" || secret == "" || redirect == "" {
		return nil
	}
	return &oauth2.Config{
		ClientID:     id,
		ClientSecret: secret,
		RedirectURL:  redirect,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}
}

func (a *Service) googleLogin(w http.ResponseWriter, r *http.Request) error {
	if a.Google == nil {
		return httpapi.Problem(404, "google_disabled", "Вход через Google отключен")
	}
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return err
	}
	state := base64.RawURLEncoding.EncodeToString(buf)
	// SameSite=Lax: state-cookie должна доехать обратно через редирект с Google,
	// Strict- cookie при таком переходе браузер не отправит.
	http.SetCookie(w, &http.Cookie{
		Name: "buddy_oauth_state", Value: state, Path: "/api/v1/auth/google", MaxAge: 300,
		HttpOnly: true, Secure: a.SecureCookie, SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, a.Google.AuthCodeURL(state), http.StatusFound)
	return nil
}

func (a *Service) googleCallback(w http.ResponseWriter, r *http.Request) error {
	if a.Google == nil {
		return httpapi.Problem(404, "google_disabled", "Вход через Google отключен")
	}
	cookie, err := r.Cookie("buddy_oauth_state")
	if err != nil || cookie.Value == "" || cookie.Value != r.URL.Query().Get("state") {
		return httpapi.Problem(400, "invalid_state", "Недействительная попытка входа")
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	tok, err := a.Google.Exchange(ctx, r.URL.Query().Get("code"))
	if err != nil {
		return httpapi.Problem(502, "google_exchange", "Не удалось подтвердить вход через Google")
	}
	profile, err := a.fetchGoogleProfile(ctx, a.Google, tok)
	if err != nil {
		return httpapi.Problem(502, "google_profile", "Не удалось получить профиль Google")
	}
	if profile.Email == "" {
		return httpapi.Problem(400, "google_no_email", "У аккаунта Google недоступен email — разрешите доступ к email при входе")
	}

	userID, err := a.googleUser(r.Context(), profile)
	if err != nil {
		return err
	}
	if err = a.startSession(w, r, userID); err != nil {
		return err
	}
	// Фронт при загрузке сам делает cookie-refresh и входит без токена в URL.
	origin := strings.TrimRight(config.Env("APP_ORIGIN", "http://localhost:8080"), "/")
	http.Redirect(w, r, origin+"/", http.StatusFound)
	return nil
}

// googleUser resolves the profile to a local account: by google_sub, then by
// email (attaching Google to the existing account), or a fresh passwordless one.
// ВАЖНО: их register не подтверждает email — линк по совпадению почты
// доверяет неподтверждённому адресу (см. TODO).
func (a *Service) googleUser(ctx context.Context, p *googleProfile) (string, error) {
	email := strings.ToLower(p.Email)

	var id string
	err := a.DB.QueryRow(ctx, "SELECT id FROM users WHERE google_sub=$1", p.Sub).Scan(&id)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return "", err
	}

	err = a.DB.QueryRow(ctx, "UPDATE users SET google_sub=$1, email_verified=true WHERE email=$2 RETURNING id", p.Sub, email).Scan(&id)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return "", err
	}
	if !p.EmailVerified {
		// Линковка только по подтвержденной гуглом почте.
		return "", httpapi.Problem(409, "email_unverified", "Подтвердите email в аккаунте Google и войдите снова")
	}

	uid := idgen.New()
	// Google подтверждает почту сам: созданный юзер сразу verified.
	if _, err = a.DB.Exec(ctx, "INSERT INTO users(id,email,google_sub,email_verified) VALUES($1,$2,$3,$4)", uid, email, p.Sub, p.EmailVerified); err != nil {
		return "", fmt.Errorf("insert google user: %w", err)
	}
	return uid, nil
}

func (a *Service) fetchGoogleProfile(ctx context.Context, conf *oauth2.Config, tok *oauth2.Token) (*googleProfile, error) {
	userInfo := a.GoogleUserInfo
	if userInfo == "" {
		userInfo = "https://openidconnect.googleapis.com/v1/userinfo"
	}
	resp, err := conf.Client(ctx, tok).Get(userInfo)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google userinfo: %s", resp.Status)
	}
	var p googleProfile
	if err = json.NewDecoder(io.LimitReader(resp.Body, 1<<16)).Decode(&p); err != nil {
		return nil, err
	}
	if p.Sub == "" {
		return nil, errors.New("google profile: empty sub")
	}
	return &p, nil
}
