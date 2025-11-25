# 使用官方 Go 镜像
FROM golang:1.24-alpine

# 安装必要工具（可选）
RUN apk add --no-cache git bash

# 设置工作目录
WORKDIR /app

# 复制 Go Modules 文件并下载依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制整个后端代码
COPY . .

# 编译 Go 程序
RUN go build -o server ./main.go

# 容器暴露端口
EXPOSE 3080

# 启动程序
CMD ["./server"]
