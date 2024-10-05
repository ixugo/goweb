package api

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/ixugo/goweb/pkg/web"
)

func setupRouter(router *gin.Engine, uc *Usecase) {
	router.Use(
		web.Mertics(),
		web.Logger(slog.Default(), uc.Conf.Server.Debug, func(path string) bool {
			return false
		}, nil),
	)

	router.GET("/health", web.WarpH(getHealth))
}

func getHealth(c *gin.Context, _ *struct{}) (any, error) {
	return gin.H{"msg": "OK"}, nil
}
