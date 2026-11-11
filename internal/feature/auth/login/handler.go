package login

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func NewGoogleConf(clientID, clientSecret, redirectURL string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}
}

type Handler struct {
	conf *oauth2.Config
	svc  *Service
}

func NewHandler(conf *oauth2.Config, svc *Service) *Handler {
	return &Handler{conf: conf, svc: svc}
}

func (h *Handler) Register(r gin.IRouter) {
	r.GET("/auth/google/login", h.Login)
	r.GET("/auth/google/callback", h.Callback)
}

func (h *Handler) Login(c *gin.Context) {
	state, err := randomState()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot create state"})
		return
	}
	// На проде Secure:true и домен; для localhost оставляем как есть.
	c.SetCookie("oauth_state", state, 300, "/", "", false, true)
	c.Redirect(http.StatusFound, h.conf.AuthCodeURL(state))
}

func (h *Handler) Callback(c *gin.Context) {
	wantState, err := c.Cookie("oauth_state")
	if err != nil || wantState == "" || c.Query("state") != wantState {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid oauth state"})
		return
	}

	st, err := h.svc.Login(c.Request.Context(), c.Query("code"))
	if err != nil {
		log.Printf("login failed: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "login failed", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token_type":    "Bearer",
		"access_token":  st.AccessToken,
		"expires_in":    st.ExpiresIn,
		"refresh_token": st.RefreshToken,
		"user_id":       st.UserID,
		"email":         st.Email,
		"name":          st.Name,
	})
}

func randomState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
