package server

import (
	"avito-autumn-2025/internal/http/handlers"
	"avito-autumn-2025/internal/logger"
	"avito-autumn-2025/internal/service/pull_request"
	"avito-autumn-2025/internal/service/team"
	"avito-autumn-2025/internal/service/user"
	"avito-autumn-2025/internal/storage/postgres"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	router *gin.Engine
	server *http.Server
	db     *pgxpool.Pool
	log    logger.Logger
}

func NewServer(db *pgxpool.Pool, log logger.Logger) *Server {
	router := gin.Default()

	return &Server{
		router: router,
		server: &http.Server{Handler: router},
		db:     db,
		log:    log,
	}
}

func (s *Server) SetupRoutes() {
	userStorage := postgres.NewUserStorage(s.db, s.log)
	teamStorage := postgres.NewTeamStorage(s.db, s.log)
	prStorage := postgres.NewPullRequestStorage(s.db, s.log)

	userSvc := user.NewUserService(&userStorage, s.log)
	teamSvc := team.NewTeamService(&teamStorage, s.log)
	prSvc := pull_request.NewPullRequestService(&prStorage, &userStorage, &teamStorage, s.log)

	userHandler := handlers.NewUserHandler(&userSvc, s.log)
	teamHandler := handlers.NewTeamHandler(&teamSvc, s.log)
	prHandler := handlers.NewPullRequestHandler(prSvc, s.log)

	api := s.router.Group("/api/v1")
	{
		api.POST("/users", userHandler.CreateUser)
		api.GET("/users/:id", userHandler.GetUserByID)

		api.POST("/team/add", teamHandler.PostTeamAdd)
		api.GET("/team/:teamName", teamHandler.GetTeamTeamName)

		api.POST("/pull-request/create", prHandler.PostPullRequestCreate)
		api.POST("/pull-request/merge", prHandler.PostPullRequestMerge)
		api.POST("/pull-request/reassign", prHandler.PostPullRequestReassign)
		api.GET("/users/get-review", prHandler.GetUsersGetReview)
		api.GET("/statistics", prHandler.GetReviewStatistics)
	}
}

func (s *Server) Run(addr string) error {
	s.log.Info("Starting HTTP server", "address", addr)
	s.server.Addr = addr
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.log.Info("Shutting down HTTP server")
	return s.server.Shutdown(ctx)
}

func (s *Server) GetRouter() *gin.Engine {
	return s.router
}
