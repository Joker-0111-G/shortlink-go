# ---- 1. 构建阶段 ----
# 使用官方的 Go alpine 镜像作为构建环境
FROM golang:1.24-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制 go.mod 和 go.sum 文件并下载依赖
# 这一步可以利用 Docker 的缓存机制，如果依赖没变，下次构建会更快
COPY go.mod go.sum ./
# 设置Go模块代理为阿里云的镜像
RUN go env -w GOPROXY=https://mirrors.aliyun.com/goproxy/,direct
RUN go mod download

# 复制所有源代码
COPY . .

# 编译Go应用
# CGO_ENABLED=0: 禁用CGO，以实现静态编译
# -o /main: 指定输出文件为 /main
# -ldflags="-s -w": 减小最终二进制文件的大小
RUN CGO_ENABLED=0 GOOS=linux go build -o /main -ldflags="-s -w" ./cmd/main.go


# ---- 2. 运行阶段 ----
# 使用一个非常小的 alpine 基础镜像
FROM alpine:latest

# 设置工作目录
WORKDIR /app

# 从构建阶段复制编译好的二进制可执行文件
COPY --from=builder /main .

# 复制配置文件目录
COPY ./configs ./configs

# 复制前端文件目录
COPY ./frontend ./frontend

# 暴露容器的 8080 端口
EXPOSE 8080

# 容器启动时执行的命令
CMD ["/app/main"]