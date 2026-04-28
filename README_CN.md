# 只狼云 Sub2API 二开版

<div align="center">

[![Go](https://img.shields.io/badge/Go-1.25.7-00ADD8.svg)](https://golang.org/)
[![Vue](https://img.shields.io/badge/Vue-3.4+-4FC08D.svg)](https://vuejs.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-336791.svg)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-7+-DC382D.svg)](https://redis.io/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED.svg)](https://www.docker.com/)

**基于 Sub2API 的二次开发版本，面向 AI API 中转站运营场景增强**

[English](README.md) | 中文 | [日本語](README_JA.md)

</div>

---

## 在线体验

体验地址：**[https://sekirocloud.site:8443/](https://sekirocloud.site:8443/)**

> 这是只狼云线上站点。账号、额度、充值、模型和可用能力以站点实际配置为准。

---

## 项目定位

本仓库是基于开源 Sub2API 的二次开发版本，不是官方原版 README 所描述的纯通用发行版。

二开目标：

- 更适合自建 AI API 中转站运营
- 强化用户侧充值、兑换、签到、模型展示和一键配置体验
- 强化后台账号管理、额度阈值、模型发现、使用记录和运营排障能力
- 保留 Sub2API 原有网关、账号池、API Key、计费、调度、支付等核心能力

---

## 与官方版本的主要差异

| 能力 | 官方 Sub2API | 只狼云二开版 |
|------|--------------|--------------|
| API 网关与账号池 | 支持 | 保留并持续兼容 |
| 用户 API Key 分发 | 支持 | 保留 |
| Token 用量与成本统计 | 支持 | 保留，并优化使用记录展示 |
| 支付/充值/兑换 | 支持 | 保留，并强化用户侧兑换体验 |
| 每日签到 | 无独立运营功能 | 新增：每天签到随机发放 `$10-$29` 余额 |
| 签到记录 | 无 | 新增：记录在 `user_daily_checkins`，展示到客户端“兑换”菜单的“最近活动” |
| 签到数据语义 | - | 不写入 `redeem_codes`，避免污染兑换码数据 |
| OpenAI 出站协议 | 默认逻辑为主 | 新增账号级出站协议开关：Responses / Chat Completions 可选 |
| OpenAI 兼容上游 | 依赖默认转换 | 可按账号选择原生 `/v1/chat/completions` 出站，提升第三方上游兼容性 |
| Base URL + Key 添加账号 | 基础配置 | 新增自动拉取上游模型；失败时展示错误且允许手动添加 |
| Quota 阈值 | 基础额度逻辑 | 新增 5h / 7d 剩余额度百分比阈值，支持提前停调度 |
| OpenAI OAuth 额度阈值 | 基础显示/保存 | 修复并支持阈值显示、保存和回显 |
| 使用记录 | 基础表格 | 修复 Token / 成本图标对齐，管理员默认展示 User-Agent |
| 侧边栏体验 | 基础导航 | 修复侧边栏滚动后跳顶问题 |
| 一键配置助手 | 原始脚本入口 | 新增/保留只狼云一键配置助手：`sekiro-install.sh` / `sekiro-install.ps1` |
| 移动端体验 | 基础适配 | 优化部分移动端留白、弹层和入口体验 |
| 管理端签到菜单 | - | 不单独开后台菜单，避免后台冗余；用户活动直接在客户端展示 |

---

## 当前二开功能清单

### 1. API 中转核心能力

- 多上游账号管理
- API Key 创建、禁用、额度和并发控制
- 用户级、账号级调度
- Token 级用量统计和成本计算
- 支持 OpenAI / Anthropic / Gemini / Claude Code / Codex 等兼容场景
- 支持分组、倍率、模型白名单等运营配置

### 2. OpenAI 兼容增强

针对不同 OpenAI-compatible 上游能力不一致的问题，二开版新增账号级出站协议控制：

- 默认关闭：保持原有兼容逻辑，不影响历史账号
- 开启后可选：
  - `responses`：继续走 Responses API 逻辑
  - `chat_completions`：用户请求 `/v1/chat/completions` 时，原生转发到上游 `/v1/chat/completions`

适合处理部分上游只支持 Chat Completions、不支持 Responses API 的情况。

### 3. Base URL + Key 模型发现

添加账号时，如果使用 Base URL + API Key：

- 自动请求上游模型列表
- 自动带出可用模型
- 获取失败时展示明确错误
- 允许继续手动选择或添加模型

### 4. 额度阈值与调度保护

新增 Codex / OpenAI OAuth 相关额度阈值字段：

- 5h 剩余额度百分比阈值
- 7d 剩余额度百分比阈值
- 空值或 0 表示不提前停，用到耗尽后等待重置
- 配置后可在额度即将耗尽前停止调度，减少上游报错和用户侧失败

### 5. 每日签到奖励

用户侧顶部入口支持每日签到：

- 签到按钮位于公告图标左侧
- 签到前显示“签到”
- 签到后显示复选标记图标
- 每个用户每天只能签到一次
- 每次随机发放 `$10-$29` 余额
- 签到记录写入 `user_daily_checkins`
- 签到记录展示在客户端“兑换”菜单的“最近活动”
- 不写入 `redeem_codes`，兑换码表只保留真实兑换语义

### 6. 用户侧兑换与最近活动

客户端“兑换”页面除了兑换码兑换，还展示最近活动：

- 兑换码兑换记录
- 每日签到奖励记录
- 活动金额使用 `$` 风格展示，例如 `+$20`

### 7. 只狼云一键配置助手

提供面向用户的一键配置入口，降低 Claude Code / Codex 等客户端接入成本：

- Linux / macOS 脚本：`/sekiro-install.sh`
- Windows PowerShell 脚本：`/sekiro-install.ps1`
- 旧入口 `/sekiro-codex.sh`、`/sekiro-codex.ps1` 保留兼容

### 8. 后台体验优化

- 管理员使用记录默认展示 User-Agent
- Token 明细和成本明细图标对齐
- 侧边栏滚动位置保持，不再切换菜单后跳顶
- 账号编辑页保留并回显二开配置项
- 移动端部分页面留白和弹层体验优化

---

## 技术栈

| 组件 | 技术 |
|------|------|
| 后端 | Go 1.25.7, Gin, Ent |
| 前端 | Vue 3.4+, Vite 5+, TailwindCSS |
| 数据库 | PostgreSQL 15+ |
| 缓存/队列 | Redis 7+ |
| 部署 | Docker / Docker Compose / 二进制 |

---

## 重要数据表说明

| 表 | 用途 |
|----|------|
| `users` | 用户、余额、状态等 |
| `api_keys` | 用户 API Key |
| `accounts` | 上游账号配置 |
| `redeem_codes` | 真实兑换码、兑换记录 |
| `user_daily_checkins` | 每日签到奖励记录 |

注意：签到奖励不应写入 `redeem_codes`。二开版已经将签到记录独立到 `user_daily_checkins`，并在用户侧最近活动中合并展示。

---

## Nginx 反向代理注意事项

通过 Nginx 反向代理并配合 Codex CLI / Claude Code 等客户端使用时，建议在 Nginx 的 `http` 块中开启：

```nginx
underscores_in_headers on;
```

Nginx 默认可能丢弃包含下划线的请求头，例如 `session_id`，会影响多账号场景下的粘性会话。

---

## 部署说明

### Docker Compose

推荐生产环境使用 Docker Compose 部署 PostgreSQL、Redis 和 Sub2API 服务。

```bash
# 克隆仓库
git clone https://github.com/Wei-Shaw/sub2api.git
cd sub2api/deploy

# 准备环境变量
cp .env.example .env

# 编辑 .env，至少配置数据库密码、JWT_SECRET、TOTP_ENCRYPTION_KEY
nano .env

# 启动
docker compose up -d

# 查看状态
docker compose ps

# 查看日志
docker compose logs -f sub2api
```

### 源码编译

```bash
# 前端构建
cd frontend
pnpm install
pnpm build

# 后端编译，注意必须带 embed 标签
cd ../backend
go build -tags embed -o sub2api ./cmd/server
```

`-tags embed` 会把前端产物嵌入到后端二进制。不带该参数时，线上 `/`、`/login` 等页面可能无法访问。

---

## 配置建议

### 安全配置

- 生产环境必须使用 HTTPS
- 为 `JWT_SECRET`、`TOTP_ENCRYPTION_KEY`、数据库密码生成高强度随机值
- 不要将 API Key、上游账号 Cookie、OAuth Token 提交到仓库
- 管理后台建议限制访问来源或接入额外认证/WAF

### 运营配置

- 根据账号类型配置模型白名单
- 对高价值账号配置 5h / 7d 剩余额度阈值
- 对只支持 Chat Completions 的上游账号开启原生 Chat Completions 出站
- 定期查看使用记录、错误日志和账号健康状态

---

## 项目结构

```text
sub2api/
├── backend/                  # Go 后端服务
│   ├── cmd/server/           # 应用入口
│   ├── internal/             # 内部模块
│   │   ├── config/           # 配置管理
│   │   ├── handler/          # HTTP 处理器
│   │   ├── repository/       # 数据访问
│   │   ├── service/          # 业务逻辑
│   │   └── gateway/          # API 网关核心
│   ├── migrations/           # 数据库迁移
│   └── resources/            # 静态资源与模型价格等
│
├── frontend/                 # Vue 3 前端
│   └── src/
│       ├── api/              # API 调用
│       ├── stores/           # 状态管理
│       ├── views/            # 页面组件
│       └── components/       # 通用组件
│
└── deploy/                   # 部署文件
    ├── docker-compose.yml
    ├── .env.example
    └── config.example.yaml
```

---

## 免责声明

本项目为基于 Sub2API 的二次开发版本，仅供技术学习、研究与自建服务使用。使用者需要自行确认上游服务条款、账号安全、数据安全和合规风险。

因使用本项目造成的账户封禁、服务中断、数据丢失、费用损失或其他问题，项目维护者不承担责任。

---

## 许可证

本项目延续上游项目许可证，基于 [GNU 宽通用公共许可证 v3.0](LICENSE)（或更高版本）授权。
