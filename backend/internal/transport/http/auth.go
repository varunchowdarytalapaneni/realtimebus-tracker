package http

import (
	"errors"
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"

	"bus-tracker/internal/domain"
	"bus-tracker/internal/service"
)

const claimsKey = "claims"

func RequireAuth(auth *service.AuthService, roles ...domain.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		tok := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
		if tok == "" {
			tok = c.Query("token")
		}
		claims, err := auth.ParseToken(tok)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		if len(roles) > 0 && !slices.Contains(roles, claims.Role) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.Set(claimsKey, claims)
		c.Next()
	}
}

func ClaimsFrom(c *gin.Context) domain.Claims {
	v, _ := c.Get(claimsKey)
	claims, _ := v.(domain.Claims)
	return claims
}

func (h *Handler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email and password are required"})
		return
	}
	res, err := h.auth.Login(c.Request.Context(), req.Email, req.Password)
	switch {
	case errors.Is(err, domain.ErrInvalidCredentials):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
	case err != nil:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	default:
		c.JSON(http.StatusOK, gin.H{"token": res.Token, "user": res.User, "bus_id": res.BusID})
	}
}
