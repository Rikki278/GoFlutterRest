package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/acidsoft/gorestteach/internal/config"
	"github.com/acidsoft/gorestteach/internal/handler"
	"github.com/acidsoft/gorestteach/internal/jwt"
	"github.com/acidsoft/gorestteach/internal/middleware"
	"github.com/acidsoft/gorestteach/internal/repository"
	"github.com/acidsoft/gorestteach/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// Server encapsulates the HTTP server and all dependencies.
type Server struct {
	httpServer *http.Server
	router     *gin.Engine
	cfg        *config.Config
}

// New wires all dependencies and registers all routes.
func New(cfg *config.Config, db *gorm.DB) *Server {
	gin.SetMode(cfg.Server.Mode)

	router := gin.New()
	router.Use(middleware.Recovery())
	router.Use(middleware.ErrorHandler())

	// ─── Dependency injection (manual DI — clear for teaching) ───────────────
	jwtService := jwt.NewService(&cfg.JWT)

	userRepo := repository.NewUserRepository(db)
	postRepo := repository.NewPostRepository(db)
	imageRepo := repository.NewImageRepository(db)
	tokenRepo := repository.NewRefreshTokenRepository(db)

	authUC := usecase.NewAuthUseCase(userRepo, tokenRepo, jwtService, &cfg.JWT)
	userUC := usecase.NewUserUseCase(userRepo, imageRepo, &cfg.Upload)
	postUC := usecase.NewPostUseCase(postRepo, imageRepo, &cfg.Upload)

	authH := handler.NewAuthHandler(authUC)
	userH := handler.NewUserHandler(userUC)
	postH := handler.NewPostHandler(postUC)
	imageH := handler.NewImageHandler(imageRepo)

	authMiddleware := middleware.Auth(jwtService)

	// ─── Routes ──────────────────────────────────────────────────────────────
	router.GET("/health", handler.HealthCheck)

	v1 := router.Group("/api/v1")
	{
		// Auth — public
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authH.Register)
			auth.POST("/login", authH.Login)
			auth.POST("/refresh", authH.Refresh)
			auth.POST("/logout", authMiddleware, authH.Logout)
		}

		// Images — public (images are served by their UUID, not sensitive)
		v1.GET("/images/:id", imageH.GetImage)

		// Protected routes
		protected := v1.Group("/", authMiddleware)
		{
			users := protected.Group("/users")
			{
				users.GET("/me", userH.GetMe)
				users.PUT("/me", userH.UpdateMe)
				users.POST("/me/avatar", userH.UploadAvatar)
				users.GET("/:id", userH.GetUser)
			}

			posts := protected.Group("/posts")
			{
				posts.POST("", postH.Create)
				posts.GET("", postH.List)
				posts.GET("/:id", postH.GetByID)
				posts.PUT("/:id", postH.Update)
				posts.DELETE("/:id", postH.Delete)
				posts.POST("/:id/image", postH.AttachImage)
			}
		}
	}

	return &Server{
		cfg:    cfg,
		router: router,
		httpServer: &http.Server{
			Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
			Handler:      router,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}
}

// Start begins listening for incoming HTTP requests.
func (s *Server) Start() error {
	log.Info().Msgf("Server listening on http://localhost%s", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}
