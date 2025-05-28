<div align="center">

![new-api](/web/public/logo.png)

# New API

🍥 Next Generation LLM Gateway and AI Asset Management System

<a href="https://trendshift.io/repositories/8227" target="_blank"><img src="https://trendshift.io/api/badge/repositories/8227" alt="Calcium-Ion%2Fnew-api | Trendshift" style="width: 250px; height: 55px;" width="250" height="55"/></a>

<p align="center">
  <a href="https://raw.githubusercontent.com/Calcium-Ion/new-api/main/LICENSE">
    <img src="https://img.shields.io/github/license/Calcium-Ion/new-api?color=brightgreen" alt="license">
  </a>
  <a href="https://github.com/Calcium-Ion/new-api/releases/latest">
    <img src="https://img.shields.io/github/v/release/Calcium-Ion/new-api?color=brightgreen&include_prereleases" alt="release">
  </a>
  <a href="https://github.com/users/Calcium-Ion/packages/container/package/new-api">
    <img src="https://img.shields.io/badge/docker-ghcr.io-blue" alt="docker">
  </a>
  <a href="https://hub.docker.com/r/CalciumIon/new-api">
    <img src="https://img.shields.io/badge/docker-dockerHub-blue" alt="docker">
  </a>
  <a href="https://goreportcard.com/report/github.com/Calcium-Ion/new-api">
    <img src="https://goreportcard.com/badge/github.com/Calcium-Ion/new-api" alt="GoReportCard">
  </a>
</p>
</div>

## 📝 Project Description

> [!NOTE]  
> This is an open-source project developed based on [One API](https://github.com/songquanpeng/veloera)

> [!IMPORTANT]  
> - Users must comply with OpenAI's [Terms of Use](https://openai.com/policies/terms-of-use) and relevant laws and regulations. Not to be used for illegal purposes.
> - This project is for personal learning only. Stability is not guaranteed, and no technical support is provided.

## ✨ Key Features

1. 🎨 New UI interface (some interfaces pending update)
2. 🌍 Multi-language support (work in progress)
3. 🎨 Added [Midjourney-Proxy(Plus)](https://github.com/novicezk/midjourney-proxy) interface support, [Integration Guide](Midjourney.md)
4. 💰 Online recharge support, configurable in system settings:
    - [x] EasyPay
5. 🔍 Query usage quota by key:
    - Works with [neko-api-key-tool](https://github.com/Calcium-Ion/neko-api-key-tool)
6. 📑 Configurable items per page in pagination
7. 🔄 Compatible with original One API database (veloera.db)
8. 💵 Support per-request model pricing, configurable in System Settings - Operation Settings
9. ⚖️ Support channel **weighted random** selection
10. 📈 Data dashboard (console)
11. 🔒 Configurable model access per token
12. 🤖 Telegram authorization login support:
    1. System Settings - Configure Login Registration - Allow Telegram Login
    2. Send /setdomain command to [@Botfather](https://t.me/botfather)
    3. Select your bot, then enter http(s)://your-website/login
    4. Telegram Bot name is the bot username without @
13. 🎵 Added [Suno API](https://github.com/Suno-API/Suno-API) interface support, [Integration Guide](Suno.md)
14. 🔄 Support for Rerank models, compatible with Cohere and Jina, can integrate with Dify, [Integration Guide](Rerank.md)
15. ⚡ **[OpenAI Realtime API](https://platform.openai.com/docs/guides/realtime/integration)** - Support for OpenAI's Realtime API, including Azure channels
16. 🧠 Support for setting reasoning effort through model name suffix:
    - Add suffix `-high` to set high reasoning effort (e.g., `o3-mini-high`)
    - Add suffix `-medium` to set medium reasoning effort
    - Add suffix `-low` to set low reasoning effort
17. 🔄 Thinking to content option `thinking_to_content` in `Channel->Edit->Channel Extra Settings`, default is `false`, when `true`, the `reasoning_content` of the thinking content will be converted to `<think>` tags and concatenated to the content returned.
18. 🔄 Think tag to reasoning option `think_tag_to_reasoning` converts `<think>` or `<thinking>` tagged content back to `reasoning_content`.
19. 🔄 Model rate limit, support setting total request limit and successful request limit in `System Settings->Rate Limit Settings`
20. 💰 Cache billing support, when enabled can charge a configurable ratio for cache hits:
    1. Set `Prompt Cache Ratio` in `System Settings -> Operation Settings`
    2. Set `Prompt Cache Ratio` in channel settings, range 0-1 (e.g., 0.5 means 50% charge on cache hits)
    3. Supported channels:
        - [x] OpenAI
        - [x] Azure 
        - [x] DeepSeek
        - [ ] Claude

## Model Support
This version additionally supports:
1. Third-party model **gpts** (gpt-4-gizmo-*)
2. [Midjourney-Proxy(Plus)](https://github.com/novicezk/midjourney-proxy) interface, [Integration Guide](Midjourney.md)
3. Custom channels with full API URL support
4. [Suno API](https://github.com/Suno-API/Suno-API) interface, [Integration Guide](Suno.md)
5. Rerank models, supporting [Cohere](https://cohere.ai/) and [Jina](https://jina.ai/), [Integration Guide](Rerank.md)
6. Dify

You can add custom models gpt-4-gizmo-* in channels. These are third-party models and cannot be called with official OpenAI keys.

## Additional Configurations Beyond One API
- `GENERATE_DEFAULT_TOKEN`: Generate initial token for new users, default `false`
- `STREAMING_TIMEOUT`: Set streaming response timeout, default 60 seconds
- `DIFY_DEBUG`: Output workflow and node info to client for Dify channel, default `true`
- `FORCE_STREAM_OPTION`: Override client stream_options parameter, default `true`
- `GET_MEDIA_TOKEN`: Calculate image tokens, default `true`
- `GET_MEDIA_TOKEN_NOT_STREAM`: Calculate image tokens in non-stream mode, default `true`
- `UPDATE_TASK`: Update async tasks (Midjourney, Suno), default `true`
- `GEMINI_MODEL_MAP`: Specify Gemini model versions (v1/v1beta), format: "model:version", comma-separated
- `COHERE_SAFETY_SETTING`: Cohere model [safety settings](https://docs.cohere.com/docs/safety-modes#overview), options: `NONE`, `CONTEXTUAL`, `STRICT`, default `NONE`
- `GEMINI_VISION_MAX_IMAGE_NUM`: Gemini model maximum image number, default `16`, set to `-1` to disable
- `MAX_FILE_DOWNLOAD_MB`: Maximum file download size in MB, default `20`
- `CRYPTO_SECRET`: Encryption key for encrypting database content
- `AZURE_DEFAULT_API_VERSION`: Azure channel default API version, if not specified in channel settings, use this version, default `2024-12-01-preview`
- `NOTIFICATION_LIMIT_DURATION_MINUTE`: Duration of notification limit in minutes, default `10`
- `NOTIFY_LIMIT_COUNT`: Maximum number of user notifications in the specified duration, default `2`

## Deployment

> [!TIP]
> Latest Docker image: `calciumion/new-api:latest`  
> Default account: root, password: 123456

### Multi-Server Deployment
- Must set `SESSION_SECRET` environment variable, otherwise login state will not be consistent across multiple servers.
- If using a public Redis, must set `CRYPTO_SECRET` environment variable, otherwise Redis content will not be able to be obtained in multi-server deployment.

### Requirements
- Local database (default): SQLite (Docker deployment must mount `/data` directory)
- Remote database: MySQL >= 5.7.8, PgSQL >= 9.6

### Deployment with BT Panel
Install BT Panel (**version 9.2.0** or above) from [BT Panel Official Website](https://www.bt.cn/new/download.html), choose the stable version script to download and install.  
After installation, log in to BT Panel and click Docker in the menu bar. First-time access will prompt to install Docker service. Click Install Now and follow the prompts to complete installation.  
After installation, find **New-API** in the app store, click install, configure basic options to complete installation.  
[Pictorial Guide](BT.md)

### Docker Deployment

### Using Docker Compose (Recommended)
```shell
# Clone project
git clone https://github.com/Calcium-Ion/new-api.git
cd new-api
# Edit docker-compose.yml as needed
# nano docker-compose.yml
# vim docker-compose.yml
# Start
docker-compose up -d
```

#### Update Version
```shell
docker-compose pull
docker-compose up -d
```

### Direct Docker Image Usage
```shell
# SQLite deployment:
docker run --name new-api -d --restart always -p 3000:3000 -e TZ=Asia/Shanghai -v /home/ubuntu/data/new-api:/data calciumion/new-api:latest

# MySQL deployment (add -e SQL_DSN="root:123456@tcp(localhost:3306)/oneapi"), modify database connection parameters as needed
# Example:
docker run --name new-api -d --restart always -p 3000:3000 -e SQL_DSN="root:123456@tcp(localhost:3306)/oneapi" -e TZ=Asia/Shanghai -v /home/ubuntu/data/new-api:/data calciumion/new-api:latest
```

#### Update Version
```shell
# Pull the latest image
docker pull calciumion/new-api:latest
# Stop and remove the old container
docker stop new-api
docker rm new-api
# Run the new container with the same parameters as before
docker run --name new-api -d --restart always -p 3000:3000 -e TZ=Asia/Shanghai -v /home/ubuntu/data/new-api:/data calciumion/new-api:latest
```

Alternatively, you can use Watchtower for automatic updates (not recommended, may cause database incompatibility):
```shell
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock containrrr/watchtower -cR
```

## Channel Retry
Channel retry is implemented, configurable in `Settings->Operation Settings->General Settings`. **Cache recommended**.  
If retry is enabled, the system will automatically use the next priority channel for the same request after a failed request.

### Cache Configuration
1. `REDIS_CONN_STRING`: Use Redis as cache
    + Example: `REDIS_CONN_STRING=redis://default:redispw@localhost:49153`
2. `MEMORY_CACHE_ENABLED`: Enable memory cache, default `false`
    + Example: `MEMORY_CACHE_ENABLED=true`

### Why Some Errors Don't Retry
Error codes 400, 504, 524 won't retry
### To Enable Retry for 400
In `Channel->Edit`, set `Status Code Override` to:
```json
{
  "400": "500"
}
```

## Integration Guides
- [Midjourney Integration](Midjourney.md)
- [Suno Integration](Suno.md)

## Related Projects
- [One API](https://github.com/songquanpeng/veloera): Original project
- [Midjourney-Proxy](https://github.com/novicezk/midjourney-proxy): Midjourney interface support
- [chatnio](https://github.com/Deeptrain-Community/chatnio): Next-gen AI B/C solution
- [neko-api-key-tool](https://github.com/Calcium-Ion/neko-api-key-tool): Query usage quota by key

## 🌟 Star History

[![Star History Chart](https://api.star-history.com/svg?repos=Calcium-Ion/new-api&type=Date)](https://star-history.com/#Calcium-Ion/new-api&Date)