//go:build wireinject
// +build wireinject

package main

import (
	"log/slog"
	"net/http"

	"github.com/google/wire"
	"github.com/ixugo/goweb/internal/conf"
	"github.com/ixugo/goweb/internal/data"
	"github.com/ixugo/goweb/internal/web/api"
)

func wireApp(bc *conf.Bootstrap, log *slog.Logger) (http.Handler, func(), error) {
	panic(wire.Build(providerSet, data.ProviderSet, api.ProviderSet))
}
