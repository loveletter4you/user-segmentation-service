package httpserver

import (
	"github.com/gin-gonic/gin"
	"github.com/loveletter4you/user-segmentation-service/config"
	_ "github.com/loveletter4you/user-segmentation-service/docs"
	"github.com/loveletter4you/user-segmentation-service/internal/controllers"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type HttpServer struct {
	router     *gin.Engine
	controller *controllers.Controller
}

func NewServer() *HttpServer {
	return &HttpServer{
		router:     gin.Default(),
		controller: controllers.NewController(),
	}
}

func (server *HttpServer) Routes() {
	server.router.Static("/static", "/static")
	api := server.router.Group("/api")
	api.POST("/segment", server.controller.CreateSegment)
	api.GET("/segments", server.controller.GetSegments)
	api.POST("/user", server.controller.CreateUser)
	api.GET("/users", server.controller.GetUsers)
	api.POST("/user/:id/segments", server.controller.CreateUserSegment)
	api.GET("/user/:id/segments", server.controller.GetUserSegments)
	api.GET("user/:id/report", server.controller.GetUserMonthReport)
}

func (server *HttpServer) StartServer(cfg *config.Config) error {
	if err := server.controller.OpenConnection(cfg); err != nil {
		return err
	}
	server.Routes()
	server.router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	if err := server.router.Run(":" + cfg.Server.Port); err != nil {
		return err
	}
	err := server.controller.CloseConnection()
	return err
}
