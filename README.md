Go语言、Echo框架、MySQL、Redis、依赖倒置、分层架构、前后端分离、WSL2中的Docker环境、.sql 数据库脚本以及 yaml 配置文件。

让我们从头开始，一步一步构建整个应用。

1. 最终项目架构
首先，这是我们将要创建的完整项目目录结构。它包含了后端、前端、配置和数据库脚本。

Bash

shortlink-go/
├── cmd/
│   └── main.go              # ➡️ Go程序入口, 负责依赖注入和启动服务
├── configs/
│   └── config.yaml          # ➡️ 应用配置文件 (数据库, Redis, 服务端口等)
├── frontend/
│   └── index.html           # ➡️ 前端页面 (HTML, CSS, JavaScript)
├── internal/
│   ├── controller/          # ➡️ 控制层 (处理HTTP请求)
│   │   ├── link_controller.go
│   │   └── request.go
│   ├── service/             # ➡️ 业务逻辑层 (核心业务处理)
│   │   ├── interface.go
│   │   └── link_service.go
│   ├── repository/          # ➡️ 数据持久层 (与数据库交互)
│   │   ├── interface.go
│   │   └── link_repository.go
│   └── model/               # ➡️ 数据模型 (对应数据库表结构)
│       └── link.go
├── pkg/
│   └── util/                # ➡️ 工具函数 (如Base62编码)
│       └── base62.go
├── sql/
│   └── schema.sql           # ➡️ 数据库初始化脚本
├── go.mod                     # Go模块文件 (会自动生成)
├── go.sum                     # (会自动生成)
└── docker-compose.yml         # ➡️ Docker编排文件 (启动MySQL和Redis)




. 运行与测试
现在，所有文件都已就位，让我们来启动并运行整个项目。

安装依赖: 在项目根目录的WSL2终端中，运行 go mod tidy，它会自动下载所有需要的Go库（echo, gorm, redis, viper等）。

启动基础设施:

Bash

docker compose up -d
等待命令执行完毕，MySQL和Redis容器将在后台运行。

初始化数据库:
使用 docker exec 命令在 MySQL 容器内执行我们的 .sql 脚本。

Bash

docker exec -i shortlink_mysql mysql -uuser -ppassword shortlink_db < sql/schema.sql
(这条命令的意思是：在 shortlink_mysql 容器内，以用户 user 密码 password 登录 shortlink_db 数据库，并执行 sql/schema.sql 文件的内容。)

启动Go后端服务:

Bash

go run cmd/main.go
如果一切顺利，你将看到服务成功启动的日志。

打开前端页面:
在你的Windows主机上，用文件浏览器找到项目目录下的 frontend/index.html 文件，直接用 Chrome 或 Firefox 等浏览器打开它。

测试功能:

在输入框中粘贴一个长链接，例如 https://github.com/labstack/echo。

点击“缩短”按钮。

下方的结果区域会显示出生成的短链接，如 http://localhost:8080/1。

点击这个新生成的链接，浏览器应该会跳转到 Echo 的 GitHub 页面。


优化
链接有效期：可以为链接设置有效期，默认1小时，也可以自定义（或设为永久）。

自动销毁：链接到期后会由后台任务自动进行“软删除”，使其失效。

优先复用：当创建新链接时，系统会优先查找并重新启用一个最旧的、已被销毁的短链接码，而不是无限地生成新码。

前端显示：历史记录中会显示实时更新的剩余有效时间。

智能排序：历史记录会按照“即将过期”的顺序优先显示在最上方。
为现有数据库添加字段 (必须)
请在您的WSL终端运行以下两条命令，为 links 表添加 expires_at 和 deleted_at 字段：
docker exec -i shortlink_mysql mysql -uuser -ppassword shortlink_db -e "ALTER TABLE links ADD COLUMN expires_at DATETIME(3) NULL AFTER short_code;"
docker exec -i shortlink_mysql mysql -uuser -ppassword shortlink_db -e "ALTER TABLE links ADD COLUMN deleted_at DATETIME(3) NULL AFTER expires_at, ADD INDEX idx_links_deleted_at (deleted_at);"

有效避免数据库中出现重复的长链接，并允许用户“续期”他们已有的短链接。如果需要缩短的链接在之前进行操作过，就是之前缩短过，并且没有过期，直接返回之前的链接，并进行时间上的重置刷新