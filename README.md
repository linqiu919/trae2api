# Trae2API

Trae：Trae是字节跳动开发的一个类似Cursor的编码辅助工具，目前提供了免费和付费两种使用模式，提供了gpt-4o和claude3.5等模型的免费使用。

Trae2API：这是一个将 Trae API 转换为 OpenAI API 格式的封装服务。

# 注意
* 项目仅供技术交流学习，不可用于其他用途
* 本项目已经停止更新，并不接受Issue
* 项目不支持Trae- v1.3.0之后的新模型
* 所有调用均受到Trae本身的限额和排队限制

## 功能特点

- 支持获取模型列表
- 支持发起对话
- 支持流式输出
- 支持 Docker 部署
- 动态配置环境变量

## Docker 部署指南

### 1. 准备环境变量
首先复制环境变量示例文件并修改：
```bash
cp .env.example .env
```

编辑 `.env` 文件，填入你的配置：
```env
# 必需的配置项
APP_ID=your_app_id_here                # Trae API 应用 ID
CLIENT_ID=your_client_id_here          # OAuth Client ID
REFRESH_TOKEN=your_refresh_token_here  # OAuth Refresh Token
USER_ID=your_user_id_here              # User ID
AUTH_TOKEN=your_auth_token_here        # API 访问鉴权 Token
```

### 2. 构建 Docker 镜像
```bash
docker build -t trae2api .
```

### 3. 运行容器
使用环境变量文件运行（推荐）：
```bash
docker run -d \
  --name trae2api \
  -p 17080:17080 \
  --env-file .env \
  --restart always \
  trae2api
```

或者直接设置环境变量运行：
```bash
docker run -d \
  --name trae2api \
  -p 17080:17080 \
  -e APP_ID="your_app_id" \
  -e IDE_TOKEN="your_ide_token" \
  -e AUTH_TOKEN="your_auth_token" \
  --restart always \
  trae2api
```

### 4. 查看容器状态
```bash
# 查看容器运行状态
docker ps

# 查看容器日志
docker logs trae2api

# 查看容器详细信息
docker inspect trae2api
```

## API 使用说明


### 环境变量如何获取

参考文档链接获取。


### 获取模型列表
```http
GET http://localhost:17080/v1/models
Authorization: Bearer your_auth_token
```

### 发起对话
```http
POST http://localhost:17080/v1/chat/completions
Authorization: Bearer your_auth_token
Content-Type: application/json

{
  "model": "gpt-4o",
  "messages": [
    {
      "role": "user",
      "content": "你好"
    }
  ],
  "stream": false
}
```

## 环境变量说明

### 必需配置
- `APP_ID`: Trae AppID
- `CLIENT_ID`: Trae ClientID
- `REFRESH_TOKEN`: Trae RefreshToken
- `USER_ID`: Trae UserID
- `AUTH_TOKEN`: API 访问鉴权 Token

### 可选配置
- `BASE_URL`: Trae API 基础 URL（默认：https://a0ai-api-sg.byteintlapi.com）
- `IDE_VERSION`: IDE 版本号（默认：1.0.2）
- `AUTH_ENABLED`: 是否启用 API 鉴权（默认：true）
- `REDIS_CONN_STRING`： Redis 连接字符串。示例：`redis://default:<password>@<addr>:<port>`，可用来缓存`REFRESH_TOKEN`
- `REFRESH_TOKEN_CACHE_ENABLED`: 是否启用`refresh_token`缓存（默认：false）。需配置环境变量`REDIS_CONN_STRING`,配置此项后刷新`refresh_token`时会缓存进`Redis`,容器重启后优先使用`Redis`中的`refresh_token` 

## 常见问题

1. 如果遇到权限问题，请检查 AUTH_TOKEN 是否正确设置
2. 如果需要更新配置，可以修改 .env 文件后重启容器：
```bash
docker restart trae2api
```

## 注意事项

- 请妥善保管你的 APP_ID 和 IDE_TOKEN
- 建议在生产环境中启用 AUTH_ENABLED
- 确保 17080 端口未被其他服务占用
- 建议设置容器自动重启策略 (--restart always)
