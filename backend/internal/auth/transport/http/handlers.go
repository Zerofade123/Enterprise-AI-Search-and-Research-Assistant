package http

import (
	"net/http"

	"github.com/Zerofade123/Enterprise-AI-Search-and-Research-Assistant/backend/internal/auth/service"
	platformErrors "github.com/Zerofade123/Enterprise-AI-Search-and-Research-Assistant/backend/internal/platform/errors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handler struct {
	auth   *service.AuthService
	logger *zap.Logger
}

func NewHandler(auth *service.AuthService, logger *zap.Logger) *Handler {
	return &Handler{auth: auth, logger: logger}
}

type registerRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type authResponse struct {
	User         userResponse `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int64        `json:"expires_in"`
}

type userResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

func (h *Handler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeError(c, platformErrors.Wrap("http.Register", platformErrors.CodeValidation, "invalid request", err))
		return
	}

	user, tokens, err := h.auth.Register(c.Request.Context(), service.RegisterInput{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		h.writeError(c, err)
		return
	}

	c.JSON(http.StatusCreated, authResponse{
		User: userResponse{
			ID:        user.ID.String(),
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
		},
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    int64(tokens.ExpiresIn.Seconds()),
	})
}

func (h *Handler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeError(c, platformErrors.Wrap("http.Login", platformErrors.CodeValidation, "invalid request", err))
		return
	}

	user, tokens, err := h.auth.Login(c.Request.Context(), service.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	}, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		h.writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, authResponse{
		User: userResponse{
			ID:        user.ID.String(),
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
		},
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    int64(tokens.ExpiresIn.Seconds()),
	})
}

func (h *Handler) Refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeError(c, platformErrors.Wrap("http.Refresh", platformErrors.CodeValidation, "invalid request", err))
		return
	}

	pair, err := h.auth.Refresh(c.Request.Context(), req.RefreshToken, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		h.writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  pair.AccessToken,
		"refresh_token": pair.RefreshToken,
		"expires_in":    int64(pair.ExpiresIn.Seconds()),
	})
}

func (h *Handler) Logout(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeError(c, platformErrors.Wrap("http.Logout", platformErrors.CodeValidation, "invalid request", err))
		return
	}

	if err := h.auth.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		h.writeError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) writeError(c *gin.Context, err error) {
	appErr, ok := err.(*platformErrors.AppError)
	if !ok {
		h.logger.Error("unhandled error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	switch appErr.Code {
	case platformErrors.CodeValidation:
		c.JSON(http.StatusBadRequest, gin.H{"error": appErr.Message})
	case platformErrors.CodeUnauthorized:
		c.JSON(http.StatusUnauthorized, gin.H{"error": appErr.Message})
	case platformErrors.CodeForbidden:
		c.JSON(http.StatusForbidden, gin.H{"error": appErr.Message})
	case platformErrors.CodeNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": appErr.Message})
	case platformErrors.CodeConflict:
		c.JSON(http.StatusConflict, gin.H{"error": appErr.Message})
	default:
		h.logger.Error("internal error", zap.Error(appErr))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	}
}
