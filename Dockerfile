# 构建阶段
FROM --platform=$BUILDPLATFORM golang:1.23-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制 go mod 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 设置构建参数
ARG TARGETARCH
ARG TARGETOS

# 构建应用
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o main .

# 运行阶段
FROM --platform=$TARGETPLATFORM alpine:latest

# 安装 CA 证书
RUN apk --no-cache add ca-certificates

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/main .

# 创建templates目录并复制模板文件
COPY --from=builder /app/templates /app/templates

# 暴露端口
EXPOSE 17080

# 运行应用
CMD ["./main"]