package refresh

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(r gin.IRouter) {
	r.POST("/auth/refresh", h.Refresh)
}

func (h *Handler) Refresh(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "refresh_token is required"})
		return
	}

	st, err := h.svc.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, ErrInvalid) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "refresh failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token_type":    "Bearer",
		"access_token":  st.AccessToken,
		"expires_in":    st.ExpiresIn,
		"refresh_token": st.RefreshToken,
		"user_id":       st.UserID,
	})
}
