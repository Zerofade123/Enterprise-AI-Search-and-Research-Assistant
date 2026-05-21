package http

import (
	"strconv"

	"github.com/Zerofade123/Enterprise-AI-Search-and-Research-Assistant/backend/internal/auth/service"
	"github.com/Zerofade123/Enterprise-AI-Search-and-Research-Assistant/backend/internal/platform/config"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Server struct {
	router *gin.Engine
	port   int
}

func NewServer(auth *service.AuthService, logger *zap.Logger) *Server {
	r := gin.New()
	r.Use(gin.Recovery())

	h := NewHandler(auth, logger)

	r.GET("/health", h.Health)

	v1 := r.Group("/api/v1")
	{
		v1.POST("/auth/register", h.Register)
		v1.POST("/auth/login", h.Login)
		v1.POST("/auth/refresh", h.Refresh)
		v1.POST("/auth/logout", h.Logout)
	}

	return &Server{router: r, port: 8080}
}

func (s *Server) Router() *gin.Engine {
	return s.router
}

func (s *Server) Port() string {
	return strconv.Itoa(s.port)
}
