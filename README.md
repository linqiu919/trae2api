# Trae2API

Trae：Trae是字节跳动开发的一个类似Cursor的编码辅助工具，目前处于免费试用期，提供了gpt-4o和claude3.5模型的免费使用。

Trae2API：这是一个将 Trae API 转换为 OpenAI API 格式的封装服务。

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

下载Trae客户端，登录账号，向AI发起一次对话


* Mac:
执行以下脚本可以一键获取需要的环境变量值
```bash
curl -fsSL https://gist.githubusercontent.com/IM594/b0bbcdc3eb6a5e849f5e306246781a48/raw/get_trae_tokens.sh | bash
```
* Windows:
  * 安装`everything`工具
  * 搜索`storge.json`文件获取RefreshToken
  * 搜索`main.log`文件获取ClientID、UserID
  * 搜索`ai_1_stdout.log`文件获取AppID

服务启动后，可以通过以下接口访问：

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
