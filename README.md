这是一个适合中小项目的 Web 框架模板

使用此模型开始项目，必须清楚了解业务领域，做好领域拆分，根据业务功能拆分成不同的业务core。

将这些解耦的 core 组合起来，达成业务需求。

## 目录说明

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
