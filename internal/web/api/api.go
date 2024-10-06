package api

import (
	"expvar"
	"log/slog"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ixugo/goweb/pkg/web"
)

var startRuntime = time.Now()

func setupRouter(r *gin.Engine, uc *Usecase) {
	r.Use(
		web.Mertics(),
		web.Logger(slog.Default(), uc.Conf.Server.Debug, func(path string) bool {
			return false
		}, nil),
	)

	auth := web.AuthMiddleware(uc.Conf.Server.HTTP.JwtSecret)
	r.GET("/health", web.WarpH(getHealth))

	registerVersion(r, uc, auth)
}

type getHealthOutput struct {
	Version   string    `json:"version"`
	StartAt   time.Time `json:"start_at"`
	GitBranch string    `json:"git_branch"`
	GitHash   string    `json:"git_hash"`
}

func getHealth(_ *gin.Context, _ *struct{}) (getHealthOutput, error) {
	return getHealthOutput{
		Version:   strings.Trim(expvar.Get("version").String(), `"`),
		GitBranch: strings.Trim(expvar.Get("git_branch").String(), `"`),
		GitHash:   strings.Trim(expvar.Get("git_hash").String(), `"`),
		StartAt:   startRuntime,
	}, nil
}
