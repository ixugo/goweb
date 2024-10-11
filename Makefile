# Makefile 使用文档
# https://www.gnu.org/software/make/manual/html_node/index.html

# include .envrc
SHELL = /bin/bash

# ==================================================================================== #
# HELPERS
# ==================================================================================== #

## help: print this help message
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/n] ' && read ans && [ $${ans:-N} = y ]

.PHONY: title
title:
	@echo -e "\033[34m$(content)\033[0m"

# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## init: 安装开发环境
init:
	go install github.com/google/wire/cmd/wire@latest
	go install github.com/divan/expvarmon@latest
	go install github.com/rakyll/hey@latest
	go install mvdan.cc/gofumpt@latest

## wire: 生成依赖注入代码
wire:
	go mod tidy
	go get github.com/google/wire/cmd/wire@latest
	go generate ./...
	go mod tidy

## expva/http: 监听网络请求指标
expva/http:
	expvarmon --ports=":9999" -i 1s -vars="version,request,requests,responses,goroutines,errors,panics,mem:memstats.Alloc"

## expva/db: 监听数据库连接指标
expva/db:
	expvarmon --ports=":9999" -i 5s -vars="databse.MaxOpenConnections,databse.OpenConnections,database.InUse,databse.Idle"

# 发起 100 次请求，每次并发 50
# hey -n 100 -c 50 http://localhost:9999/healthcheck


# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #

## audit: 检查代码依赖/格式化/测试
.PHONY: audit
audit:
	@make title content='Formatting code...'
	gofumpt -l -w .
	@make title content='Vetting code...'
	go vet ./...
	@make title content='Running tests...'
	go test -race -vet=off ./...

## vendor: 整理并下载依赖
.PHONY: vendor
vendor:
	@make title content='Tidying and verifying module dependencies...'
	go mod tidy && go mod verify
	@make title content='Vendoring dependencies...'
	go mod vendor

# ==================================================================================== #
# BUILD
# ==================================================================================== #

# 版本号规则说明
# 1. 版本号使用 Git tag，格式为 v1.0.0。
# 2. 如果当前提交没有 tag，找到最近的 tag，计算从该 tag 到当前提交的提交次数。例如，最近的 tag 为 v1.0.1，当前提交距离它有 10 次提交，则版本号为 v1.0.11（v1.0.1 + 10 次提交）。
# 3. 如果没有任何 tag，则默认版本号为 v0.0.0，后续提交次数作为版本号的次版本号。

# Get the current module name
MODULE_NAME := $(shell pwd | awk -F "/" '{print $$NF}')
# Get the latest commit hash and date
HASH_AND_DATE := $(shell git log -n1 --pretty=format:"%h-%cd" --date=format:%y%m%d | awk '{print $1}')
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)

# 如果想仅支持注释标签，可以去掉 --tags，否则会包含轻量标签
RECENT_TAG := $(shell git describe --tags --abbrev=0  2>&1 | grep -v "fatal" || echo "v0.0.0")

ifeq ($(RECENT_TAG),v0.0.0)
	COMMITS := $(shell git rev-list --count HEAD)
else
	COMMITS := $(shell git rev-list --count $(RECENT_TAG)..HEAD)
endif


test:
	@echo "hello"
	@echo ">>>${RECENT_TAG}"

# 从版本字符串中提取主版本号、次版本号和修订号
GIT_VERSION_MAJOR := $(shell echo $(RECENT_TAG) | cut -d. -f1 | sed 's/v//')
GIT_VERSION_MINOR := $(shell echo $(RECENT_TAG) | cut -d. -f2)
GIT_VERSION_PATCH := $(shell echo $(RECENT_TAG) | cut -d. -f3)
FINAL_PATCH := $(shell echo $(GIT_VERSION_PATCH) + $(COMMITS) | bc)
VERSION := v$(GIT_VERSION_MAJOR).$(GIT_VERSION_MINOR).$(FINAL_PATCH)


# .PHONY: build
# BUILD_LOCAL_DIR := ./build/local
# build:
# 	@echo -n 'Building local...'
# 	@rm -rf BUILD_LOCAL_DIR
# 	@go build -ldflags="-s -w -X main.build=local -X main.buildVersion=$(VERSION)" -o=$(BUILD_LOCAL_DIR)/app ./cmd/server
# 	@tar -czf $(BUILD_LOCAL_DIR)/$(MODULE_NAME)-$(VERSION)-$(HASH_AND_DATE).tar.gz $(BUILD_LOCAL_DIR)/app
# 	@echo 'OK'


## build/clean: 清理构建缓存目录
.PHONY: build/clean
build/clean:
	@rm -rf ./build/*


## build/linux: 构建 linux 应用
.PHONY: build/linux
BUILD_LINUX_AMD64_DIR := ./build/linux_amd64
build/linux:
	@echo -n 'Building linux...'
	@rm -rf BUILD_LINUX_AMD64_DIR
	@GOOS=linux GOARCH=amd64 go build \
		-trimpath \
		-ldflags="-s -w -X main.buildVersion=$(VERSION) -X main.gitBranch=$(BRANCH_NAME) -X main.gitHash=$(HASH_AND_DATE) -X main.buildTimeAt=$(date +%s) -X main.release=true" \
		-o=$(BUILD_LINUX_AMD64_DIR)/bin ./cmd/server
	@echo 'OK'

## build/windows: 构建 windows 应用
.PHONY: build/windows
BUILD_WINDOWS_AMD64_DIR := ./build/windows_amd64
build/windows:
	@echo -n 'Building windows...'
	@rm -rf BUILD_WINDOWS_AMD64_DIR
	@GOOS=windows GOARCH=amd64 go build \
		-trimpath \
		-ldflags="-s -w -X main.buildVersion=$(VERSION) -X main.gitBranch=$(BRANCH_NAME) -X main.gitHash=$(HASH_AND_DATE) -X main.buildTimeAt=$(date +%s) -X main.release=true" \
		-o=$(BUILD_WINDOWS_AMD64_DIR)/bin ./cmd/server
	@echo 'OK'



## info: 查看构建版本相关信息
.PHONY: info
info:
	@echo "dir: $(MODULE_NAME)"
	@echo "version: $(VERSION)"
	@echo "branch $(BRANCH)"
	@echo "hash: $(HASH_AND_DATE)"
	@echo "support $$(go tool dist list | grep amd64 | grep linux)"


IMAGE_NAME := $(MODULE_NAME):latest

docker/build:
	@docker buildx build --force-rm=true --platform linux/amd64 -t $(IMAGE_NAME) .

docker/save:
	@docker save -o $(MODULE_NAME)_$(VERSION).tar $(IMAGE_NAME)

docker/push:
	@docker push $(IMAGE_NAME)

# ==================================================================================== #
# PRODUCTION
# ==================================================================================== #

PRODUCTION_HOST = remoteHost

## release/push: 发布产品到服务器，仅上传文件
# 中小项目可以引入 CI/CD，也可以通过命令快速发布到测试服务器上。
release/push:
	@scp build/linux_amd64/bin $(PRODUCTION_HOST):/home/app/$(MODULE_NAME)
	@echo "push Successed"
