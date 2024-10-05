package data

import (
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/glebarez/sqlite"
	"github.com/google/wire"
	"github.com/ixugo/goweb/internal/conf"
	"github.com/ixugo/goweb/pkg/orm"
	"github.com/ixugo/goweb/pkg/system"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(SetupDB)

// SetupDB 初始化数据存储
func SetupDB(c *conf.Bootstrap, l *slog.Logger, release bool) (*gorm.DB, error) {
	cfg := c.Data.Database
	dial, isSQLite := getDialector(filepath.Join(system.GetCWD(), cfg.Dsn))
	if isSQLite {
		cfg.MaxIdleConns = 1
		cfg.MaxOpenConns = 1
	}
	// 使用postgres驱动打开数据库连接
	// pg := postgres.Open(cfg.Dsn)
	// 使用orm包初始化数据库连接
	db, err := orm.New(true, dial, orm.Config{
		MaxIdleConns:    int(cfg.MaxIdleConns),
		MaxOpenConns:    int(cfg.MaxOpenConns),
		ConnMaxLifetime: cfg.ConnMaxLifetime.Duration(),
		SlowThreshold:   cfg.SlowThreshold.Duration(),
	}, orm.NewLogger(l, !release, cfg.SlowThreshold.Duration()))
	return db, err
}

// getDialector 返回 dial 和 是否 sqlite
func getDialector(dsn string) (gorm.Dialector, bool) {
	if strings.HasPrefix(dsn, "postgres") {
		return postgres.New(postgres.Config{
			DriverName: "pgx",
			DSN:        dsn,
		}), false
	}
	return sqlite.Open(dsn), true
}
