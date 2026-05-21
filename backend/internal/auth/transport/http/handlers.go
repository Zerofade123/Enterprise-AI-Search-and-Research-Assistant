package http

import (
	"net/http"
	"strings"

	"github.com/Zerofade123/Enterprise-AI-Search-and-Research-Assistant/backend/internal/auth/service"
	platformErrors "github.com/Zerofade123/Enterprise-AI-Search-and-Research-Assistant/backend/internal/platform/errors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handler struct {
	auth   *service.AuthService
	logger *zap.Logger
}

func NewHandler(auth *service.AuthService, logger *zap.Logger) *Handler { return &Handler{auth: auth, logger: logger} }

type registerRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required"`
	FirstName string `json:"first_name" binding:"required,min=1,max=100"`
	LastName  string `json:"last_name" binding:"required,min=1,max=100"`
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type authResponse struct {
	Success bool        `json:"success"`
	Data    authData    `json:"data"`
	Meta    responseMeta `json:"meta"`
}

type authData struct {
	User         userResponse `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int64        `json:"expires_in"`
}

type responseMeta struct {
	RequestID string `json:"request_id"`
}

type userResponse struct { ID, Email, FirstName, LastName string }

func (h *Handler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil { h.writeError(c, platformErrors.Wrap("http.Register", platformErrors.CodeValidation, "invalid request", err)); return }
	user, tokens, err := h.auth.Register(c.Request.Context(), service.RegisterInput{Email: strings.TrimSpace(req.Email), Password: req.Password, FirstName: strings.TrimSpace(req.FirstName), LastName: strings.TrimSpace(req.LastName)}, c.ClientIP(), c.Request.UserAgent())
	if err != nil { h.writeError(c, err); return }
	c.JSON(http.StatusCreated, authResponse{Success: true, Data: authData{User: userResponse{ID: user.ID.String(), Email: user.Email, FirstName: user.FirstName, LastName: user.LastName}, AccessToken: tokens.AccessToken, RefreshToken: tokens.RefreshToken, ExpiresIn: int64(tokens.ExpiresIn.Seconds())}, Meta: responseMeta{RequestID: c.GetString("request_id")}})
}

func (h *Handler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil { h.writeError(c, platformErrors.Wrap("http.Login", platformErrors.CodeValidation, "invalid request", err)); return }
	user, tokens, err := h.auth.Login(c.Request.Context(), service.LoginInput{Email: strings.TrimSpace(req.Email), Password: req.Password}, c.ClientIP(), c.Request.UserAgent())
	if err != nil { h.writeError(c, err); return }
	c.JSON(http.StatusOK, authResponse{Success: true, Data: authData{User: userResponse{ID: user.ID.String(), Email: user.Email, FirstName: user.FirstName, LastName: user.LastName}, AccessToken: tokens.AccessToken, RefreshToken: tokens.RefreshToken, ExpiresIn: int64(tokens.ExpiresIn.Seconds())}, Meta: responseMeta{RequestID: c.GetString("request_id")}})
}

func (h *Handler) Refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil { h.writeError(c, platformErrors.Wrap("http.Refresh", platformErrors.CodeValidation, "invalid request", err)); return }
	pair, err := h.auth.Refresh(c.Request.Context(), req.RefreshToken, c.ClientIP(), c.Request.UserAgent())
	if err != nil { h.writeError(c, err); return }
	c.JSON(http.StatusOK, authResponse{Success: true, Data: authData{AccessToken: pair.AccessToken, RefreshToken: pair.RefreshToken, ExpiresIn: int64(pair.ExpiresIn.Seconds())}, Meta: responseMeta{RequestID: c.GetString("request_id")}})
}

func (h *Handler) Logout(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil { h.writeError(c, platformErrors.Wrap("http.Logout", platformErrors.CodeValidation, "invalid request", err)); return }
	if err := h.auth.Logout(c.Request.Context(), req.RefreshToken); err != nil { h.writeError(c, err); return }
	c.Status(http.StatusNoContent)
}

func (h *Handler) Health(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) }

func (h *Handler) writeError(c *gin.Context, err error) {
	appErr, ok := err.(*platformErrors.AppError)
	if !ok { h.logger.Error("unhandled error", zap.Error(err)); c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "internal error"}); return }
	switch appErr.Code {
	case platformErrors.CodeValidation:
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": appErr.Message})
	case platformErrors.CodeUnauthorized:
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": appErr.Message})
	case platformErrors.CodeForbidden:
		c.JSON(http.StatusForbidden, gin.H{"success": false, "error": appErr.Message})
	case platformErrors.CodeNotFound:
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": appErr.Message})
	case platformErrors.CodeConflict:
		c.JSON(http.StatusConflict, gin.H{"success": false, "error": appErr.Message})
	default:
		h.logger.Error("internal error", zap.Error(appErr))
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "internal error"})
	}
}
