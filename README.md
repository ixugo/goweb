这是一个适合中小项目的 Web 框架模板

使用此模型开始项目，必须清楚了解业务领域，做好领域拆分，根据业务功能拆分成不同的业务core。

将这些解耦的 core 组合起来，达成业务需求。

## 目录说明


```bash
.
├── cmd				可执行程序
│   └── server
├── configs			配置文件
├── docs			设计文档/用户文档
├── internal		私有业务
│   ├── conf		配置模型
│   ├── core		业务领域
│   │   └── version	实际业务
│   │       └── store
│   │           └── versiondb 数据库操作
│   ├── data		数据库初始化
│   └── web
│       └── api 	RESTful API
└── pkg				依赖库
```


## 项目说明

1. 程序启动强依赖的组件，发生异常时主动 panic，尽快崩溃尽快解决错误。

2. core 为业务领域，包含领域模型，领域业务功能

3. store 为数据库操作模块，需要依赖模型，此处依赖反转 core，避免每一层都定义模型。

4. api 层的入参/出参，可以正向依赖 core 层定义模型，参数模型以 `Input/Output` 来简单区分入参出数。

## Makefile

执行 `make` 或 `make help` 来获取更多帮助

在编写 makefile 时，应主动在命令上面增加注释，以 `## <命令>: <描述>` 格式书写，具体参数 Makefile 文件已有命令。其目的是 `make help` 时提供更多信息。

makefile 中提供了一些默认的操作便于快速编写

`make confirm` 用于确认下一步

`make title content=标题`  用于重点突出输出标题

`make info` 获取构建版本相关信息

**makefile 构建的版本号规则说明**

1. 版本号使用 Git tag，格式为 v1.0.0。

2. 如果当前提交没有 tag，找到最近的 tag，计算从该 tag 到当前提交的提交次数。例如，最近的 tag 为 v1.0.1，当前提交距离它有 10 次提交，则版本号为 v1.0.11（v1.0.1 + 10 次提交）。

3. 如果没有任何 tag，则默认版本号为 v0.0.0，后续提交次数作为版本号的次版本号。

## 快速开始

业务说明:

假设我们要做一个版本管理的业务，curd 步骤如下:

在 「internal」-「core」 创建 「version」 目录，创建「model.go」写入领域模型，该模型为数据库表结构映射。

创建「core.go」 写入如下内容

```go
package version

import (
	"fmt"
	"strings"
)

// Storer 依赖反转的数据持久化接口
type Storer interface {
	First(*Version) error
	Add(*Version) error
}

// Core 业务对象
type Core struct {
	Storer    Storer
}

// NewCore 创建业务对象
func NewCore(store Storer) *Core {
	return &Core{
		Storer: store,
	}
}

// IsAutoMigrate 是否需要进行表迁移
// 判断硬编码在代码中的数据库表版本号，与数据库存储的版本号做对比
func (c *Core) IsAutoMigrate(currentVer, remark string) bool {
	var ver Version
	if err := c.Storer.First(&ver); err != nil {
		isMigrate := true
		c.IsMigrate = &isMigrate
		return isMigrate
	}
	isMigrate := compareVersionFunc(currentVer, ver.Version, func(a, b string) bool {
		return a > b
	})
	c.IsMigrate = &isMigrate
	return isMigrate
}

func compareVersionFunc(a, b string, f func(a, b string) bool) bool {
	s1 := versionToStr(a)
	s2 := versionToStr(b)
	if len(s1) != len(s2) {
		return true
	}
	return f(s1, s2)
}

func versionToStr(str string) string {
	var result strings.Builder
	arr := strings.Split(str, ".")
	for _, item := range arr {
		if idx := strings.Index(item, "-"); idx != -1 {
			item = item[0:idx]
		}
		result.WriteString(fmt.Sprintf("%03s", item))
	}
	return result.String()
}
```

创建 「store/versiondb」 目录，创建「db.go」 文件写入

```go
type DB struct {
	db *gorm.DB
}

func NewDB(db *gorm.DB) DB {
	return DB{db: db}
}

// AutoMigrate 表迁移
func (d DB) AutoMigrate(ok bool) DB {
	if !ok {
		return d
	}
	if err := d.db.AutoMigrate(
		new(version.Version),
	); err != nil {
		panic(err)
	}
	return d
}

func (d DB) First(v *version.Version) error {
	return d.db.Order("id DESC").First(v).Error
}

func (d DB) Add(v *version.Version) error {
	return d.db.Create(v).Error
}
```

在 API 层做依赖注入，对 「web/api/provider.go」 写入函数，往 Usecase 中注入业务对象

```go
var ProviderSet = wire.NewSet(
	wire.Struct(new(Usecase), "*"),
	NewHTTPHandler,
	NewVersion,
)

func NewVersion(db *gorm.DB) *version.Core {
	vdb := versiondb.NewDB(db)
	core := version.NewCore(vdb)
	isOK := core.IsAutoMigrate(dbVersion, dbRemark)
	vdb.AutoMigrate(isOK)
	if isOK {
		slog.Info("更新数据库表结构")
		if err := core.RecordVersion(dbVersion, dbRemark); err != nil {
			slog.Error("RecordVersion", "err", err)
		}
	}
	return core
}
```

在 API 层新建「version.go」文件，写入

```go
package api

import (
	"github.com/gin-gonic/gin"
	"github.com/ixugo/goweb/internal/core/version"
	"github.com/ixugo/goweb/pkg/web"
)

// version 业务函数命名空间
type Version struct {
	ver *version.Core
}

// registerVersion 向路由注册业务接口
func registerVersion(r gin.IRouter, uc *Usecase, handler ...gin.HandlerFunc) {
	verEngine := Version{ver: uc.Version}

	ver := r.Group("/version", handler...)
	ver.GET("", web.WarpH(verEngine.getVersion))
}

func (v Version) getVersion(_ *gin.Context, _ *struct{}) (any, error) {
	return gin.H{"msg": "test"}, nil
}
```
