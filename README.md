<p align="center">
    <img src="./logo.png#gh-light-mode-only" alt="Goyave Logo" width="550"/>
    <img src="./logo_dark.png#gh-dark-mode-only" alt="Goyave Logo" width="550"/>
</p>

<p align="center">
    <a href="https://github.com/ixugo/goweb/releases"><img src="https://img.shields.io/github/v/release/ixugo/goweb?include_prereleases" alt="Version"/></a>
    <a href="https://github.com/ixugo/goweb/blob/master/LICENSE.txt"><img src="https://img.shields.io/dub/l/vibe-d.svg" alt="License"/></a>
	<a href="https://gin-gonic.com"><img width=30px src="https://avatars.githubusercontent.com/u/7894478?s=48&v=4" alt="GIN"/></a>
    <a href="https://gorm.io"><img width=70px src="https://gorm.io/gorm.svg" alt="GORM"/></a>

</p>

[English](./README.md) | [简体中文](./README_zh.md)

# Enterprise REST API Template

This is a complete CRUD solution focused on REST API.

The goal of Goweb is to:

+ Provide a clean architecture suitable for small and medium-sized projects.
+ Provide a modular structure for quickly starting a project, focusing on business development.
+ Simplify projects, making development more efficient and enjoyable.

If this aligns with your needs, you can start your project based on this template. This project will continually provide documentation to guide effective usage.

Supports [code generation](github.com/ixugo/gowebx).

## References

[Google API Design Guide](https://google-cloud.gitbook.io/api-design-guide)

## Directory Structure

```bash
.
├── cmd                     Executable program
│   └── server
├── configs                 Configuration files
├── docs                    Design/User documentation
├── internal                Private business
│   ├── conf                Configuration models
│   ├── core                Business domain
│   │   └── version         Actual business
│   │       └── store
│   │           └── versiondb Database operations
│   ├── data                Database initialization
│   └── web
│       └── api             RESTful API
└── pkg                     Dependencies
```

## Project Description

1. Components strongly relied upon by the program will trigger a panic on error, so that issues are resolved as quickly as possible.

2. The core directory represents the business domain, containing domain models and domain business functions.

3. The store is the database operation module, dependent on models with dependency inversion towards the core, avoiding the need to define models at each layer.

4. Input/output parameters in the API layer may directly depend on models defined in the core layer, with input and output models distinguished by appending `Input/Output` to the model names.

## Makefile

For Windows systems, please use the Git Bash terminal to run the Makefile instead of the default cmd/powershell terminal, as issues may arise.

Use `make` or `make help` to get more help.

When writing a Makefile, add comments above each command in the format `## <command>: <description>` for readability, with available parameters provided in the Makefile. The goal is to make `make help` output more informative.

Some default operations are provided in the Makefile to assist with rapid development.

`make confirm` confirms the next step.

`make title content=Title` highlights a title in the output.

`make info` fetches build version information.

**Versioning Rules in the Makefile**

1. Git tags are used for versioning, in the format v1.0.0.

2. If the current commit lacks a tag, the closest tag is found, and the number of commits from that tag is calculated. For example, if the latest tag is v1.0.1, and there have been 10 commits since, the version number becomes v1.0.11 (v1.0.1 + 10 commits).

3. If there are no tags, the default version is v0.0.0, with the minor version incremented based on the number of commits.

## Quick Start

Example business logic:

Assume we want to implement version management. The CRUD steps are as follows:

Under "internal" - "core," create the "version" directory, then create `model.go` and define the domain model representing the database table structure.

Create `core.go` and add the following content:

```go
package version

import (
	"fmt"
	"strings"
)

// Storer Interface for dependency inversion in data persistence.
type Storer interface {
	First(*Version) error
	Add(*Version) error
}

// Core Business object
type Core struct {
	Storer    Storer
}

// NewCore Creates a business object.
func NewCore(store Storer) *Core {
	return &Core{
		Storer: store,
	}
}

// IsAutoMigrate Checks if table migration is required
// Compares the hard-coded database table version with the stored version.
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

Under "store/versiondb," create the `db.go` file with the following content:

```go
type DB struct {
	db *gorm.DB
}

func NewDB(db *gorm.DB) DB {
	return DB{db: db}
}

// AutoMigrate Table migration.
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

In the API layer, inject dependencies by adding a function in `web/api/provider.go` to inject the business object into Usecase:

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
		slog.Info("Updating database schema")
		if err := core.RecordVersion(dbVersion, dbRemark); err != nil {
			slog.Error("RecordVersion", "err", err)
		}
	}
	return core
}
```

Create a new `version.go` file in the API layer with the following content:

```go
// VersionAPI Namespace for version business functions.
type VersionAPI struct {
	ver *version.Core
}

func NewVersionAPI(ver *version.Core) VersionAPI {
	return VersionAPI{ver: ver}
}
// registerVersion Registers business interface with the router.
func registerVersion(r gin.IRouter, verAPI VersionAPI, handler ...gin.HandlerFunc) {
	ver := r.Group("/version", handler...)
	ver.GET("", web.WarpH(verAPI.getVersion))
}

func (v VersionAPI) getVersion(_ *gin.Context, _ *struct{}) (any, error) {
	return gin.H{"msg": "test"}, nil
}
```

## FAQ

> Why not define models in each layer separately?

This is a trade-off between development efficiency and decoupling, balancing code readability and efficiency.

> Where should API layer parameter models and table mapping models be defined?

Understanding the dependency relationships between layers is crucial. The API directly depends on the core, while the DB layer is inverted to depend on the core. Thus, domain models are defined in the core, and input/output parameter models can also be defined in the core. If they are unused in the core, defining them in the API layer is fine too.

> Why does the API layer directly depend on the core layer rather than an interface?

Interfaces aim to decouple, but in practice, it is more common to replace the API layer than the core layer.

The API only retrieves parameters and returns response parameters, doing the minimum necessary to facilitate the transition from HTTP to GRPC.

Design for the future, but program for the present. Increasing development efficiency now allows for a better approach in the future when needed.

> Why is the DB layer inverted to depend on the core?

Data persistence is not independent; it serves the business. That is, persistence serves the

 business and depends on it.

Through dependency inversion, other databases, such as Redis cache, can be inserted between business operations.

> Why suffix input/output models with `Input/Output`?

Convention is preferable to configuration. Some projects use `Request/Response` as suffixes to standardize parameter names.

Of course, output parameters can also serve as input, and you can define an alias or use them directly.

Frequently, we want clarity on what we’re doing and why. This FAQ aims to offer some insight.

> How to write business plugins for Goweb?

```go
// RegisterVersion Some general business functions are depended upon by other business functions, such as table version control, dictionary, verification code, scheduled tasks, user management, etc.
// Conventionally, write functions in the format Register<Core>, injecting three parameters: gin router, namespace, and middleware.
// Refer to project code for specifics.
func RegisterVersion(r gin.IRouter, verAPI VersionAPI, handler ...gin.HandlerFunc) {
	ver := r.Group("/version", handler...)
	ver.GET("", web.WarpH(verAPI.getVersion))
}
```

## Table Migration

Executing table migration on every program start is too slow.

Therefore, migration control is implemented through the version table, so migration only occurs when the database table version is outdated. Modify the `dbVersion` in api/db.go to control the version number.

## Error Handling

Errors returned from core functions or the API layer should be of type `web.Error`.

In the web.WarpH wrapper, errors are properly logged and returned to the frontend.

```go
type Error struct {
	reason  string   // Reason for the error
	msg     string   // User-readable message
	details []string // Developer-readable error details
}
```

`reason` is a predefined error reason, defined in English and used to differentiate HTTP response status codes.

`msg` is the user-facing error message.

`details` are displayed in developer mode to provide complete error content for debugging purposes.

## Custom Configuration Directory

The default configuration directory is `configs`, located in the same directory as the executable. You can also specify other configuration directories.

`./bin -conf ./configs`

## Main Project Dependencies

+ gin
+ gorm
+ slog / zap
+ wire