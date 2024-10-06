package main

import (
	"expvar"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/google/wire"
	"github.com/ixugo/goweb/internal/conf"
	"github.com/ixugo/goweb/pkg/logger"
	"github.com/ixugo/goweb/pkg/server"
	"github.com/ixugo/goweb/pkg/system"
)

var (
	buildVersion = "0.0.1" // 构建版本号
	gitBranch    = "dev"   // git 分支
	gitHash      = "debug" // git 提交点哈希值
	release      string    // 发布模式 true/false
	buildTime    string    // 构建时间戳
)

var providerSet = wire.NewSet(GetBuildRelease) // nolint

func GetBuildRelease() bool {
	v, _ := strconv.ParseBool(release)
	return v
}

func main() {
	// 初始化配置
	var bc conf.Bootstrap
	if err := conf.SetupConfig(&bc); err != nil {
		panic(err)
	}
	// 初始化日志
	logDir := filepath.Join(system.GetCWD(), bc.Log.Dir)
	log, clean := logger.SetupSlog(logger.Config{
		Dir:          logDir,                            // 日志地址
		Debug:        !GetBuildRelease(),                // 服务级别Debug/Release
		MaxAge:       bc.Log.MaxAge.Duration(),          // 日志存储时间
		RotationTime: bc.Log.RotationTime.Duration(),    // 循环时间
		RotationSize: bc.Log.RotationSize * 1024 * 1024, // 循环大小
		Level:        bc.Log.Level,                      // 日志级别
	})
	{
		expvar.NewString("version").Set(buildVersion)
		expvar.NewString("git_branch").Set(gitBranch)
		expvar.NewString("git_hash").Set(gitHash)
		expvar.NewString("build_time").Set(buildTime)
		expvar.Publish("timestamp", expvar.Func(func() any {
			return time.Now().Format(time.DateTime)
		}))
	}

	handler, cleanUp, err := wireApp(&bc, log)
	if err != nil {
		slog.Error("程序构建失败", "err", err)
		panic(err)
	}
	defer cleanUp()

	svc := server.New(handler,
		server.Port(strconv.Itoa(bc.Server.HTTP.Port)),
		server.ReadTimeout(60*time.Minute),
		server.WriteTimeout(60*time.Minute),
	)
	go svc.Start()
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		slog.Info(`<-interrupt`, "signal", s.String())
	case err := <-svc.Notify():
		slog.Error(`<-server.Notify()`, "err", err)
	}
	if err := svc.Shutdown(); err != nil {
		slog.Error(`server.Shutdown()`, "err", err)
	}

	defer clean()
}
