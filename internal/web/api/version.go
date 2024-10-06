package api

import (
	"github.com/gin-gonic/gin"
	"github.com/ixugo/goweb/internal/core/version"
	"github.com/ixugo/goweb/pkg/web"
)

type Version struct {
	ver *version.Core
}

func registerVersion(r gin.IRouter, uc *Usecase, handler ...gin.HandlerFunc) {
	verEngine := Version{ver: uc.Version}

	ver := r.Group("/version", handler...)
	ver.GET("", web.WarpH(verEngine.getVersion))
}

func (v Version) getVersion(_ *gin.Context, _ *struct{}) (any, error) {
	return gin.H{"msg": "test"}, nil
}
