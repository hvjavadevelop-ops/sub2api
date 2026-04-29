# SekiroCloud Sub2API 二开版

<div align="center">

[![Go](https://img.shields.io/badge/Go-1.26.2-00ADD8.svg)](https://golang.org/)
[![Vue](https://img.shields.io/badge/Vue-3.4+-4FC08D.svg)](https://vuejs.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-336791.svg)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-7+-DC382D.svg)](https://redis.io/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED.svg)](https://www.docker.com/)

**云栖小铺 / SekiroCloud 使用的 Sub2API 定制版本**

[English](README.md) | 中文

</div>

---

## 在线体验

体验地址：**https://sekirocloud.site:8443/**

本仓库是基于官方 Sub2API v0.1.120 合并维护的二开版本，用于 AI API 中转站、兑换码/余额分发、模型广场、渠道状态监控和运营后台管理。

> 本项目不是 Sub2API 官方仓库。官方项目请参考 Wei-Shaw/sub2api。本仓库主要记录 SekiroCloud 线上使用的定制功能与部署代码。

## 项目定位

Sub2API 本身是一个 AI API 网关平台，用于把上游 AI 账号、订阅额度、API Key 或 OAuth 能力统一转成 OpenAI/Anthropic 等兼容接口，并提供用户、余额、并发、分组、渠道、用量和支付相关管理能力。

本二开版在官方能力基础上，重点补齐了运营场景：

- 用户侧更容易理解的模型广场与一键配置入口
- 可运营的每日签到奖励和后台签到管理
- 管理员发放“首次使用才开始计时”的限时权益
- 针对 Codex / Claude Code / Cursor 等使用场景的兼容与展示优化
- 更安全的站点设置初始化，避免升级/重启覆盖品牌配置

## 当前版本

- 上游基线：Sub2API v0.1.120
- 线上体验：`https://sekirocloud.site:8443/`
- 当前生产镜像标签参考：`sub2api-mobilefix:0.1.120-production-merged-custom-404fix-versionfix2`

## 二开功能概览

### 1. 每日签到

- 用户顶部栏日历图标签到入口，替换原语言切换位置。
- 已签到后按钮不置灰，再点提示“今天已经签到过了，明天再来”。
- 奖励范围由后台配置，当前线上默认 `$10-$29`/`$10-$30` 区间按配置读取。
- 签到奖励直接进入用户余额，并记录到独立签到表。
- 不写入 `redeem_codes`，避免污染兑换码业务语义。
- 客户端兑换/近期记录可展示签到奖励流水。

### 2. 签到管理后台

- 独立后台页面：`/admin/daily-checkins`。
- 支持查看签到记录、记录 ID、用户信息、奖励金额、签到时间。
- 支持按邮箱、用户名、数字用户 ID 搜索。
- 页面内置“签到奖励配置”，可直接维护 `daily_checkin_reward_min` / `daily_checkin_reward_max`。
- 后端列表查询同步支持 numeric user id 与邮箱/用户名模糊搜索。

### 3. 限时权益

用于管理员给用户发放临时余额或临时并发，适合试用、活动、人工补偿等场景。

- 用户管理每行直接显示“限时权益”入口，不只藏在更多菜单里。
- 权益初始状态为“待激活”。
- 用户第一次通过 API Key 发起真实 API 请求时才激活并开始倒计时。
- 到期后后台自动扣回对应余额/并发，并写入历史记录。
- 支持状态：`pending`、`active`、`expired`、`cancelled`。
- 后端独立表：`timed_user_grants`。
- API Key 鉴权链路内激活权益，激活后刷新用户/Key 缓存，确保当前请求即可看到新额度。

### 4. 模型广场

- 用户侧模型广场展示主流模型与示例。
- 支持后台“模型配置”按厂商分类管理用户侧展示内容。
- Codex 下载入口只保留官方页面：`https://developers.openai.com/codex/app`。
- 不再托管本站安装包，旧 `/downloads/codex/...` 路径返回真实 404，避免误消耗服务器出站流量。
- 联系二维码独立卡片展示，图片按比例 `object-contain`，不裁剪。

### 5. 自定义菜单外链打开方式

- 后台自定义菜单支持 `open_mode`。
- 默认 `new_tab` 新窗口打开，避免外部网站因 `X-Frame-Options` / CSP 禁止 iframe 导致白屏或拒绝连接。
- 显式选择 `iframe` 时仍保留内嵌页面能力，并提供新窗口打开提示。

### 6. OpenAI / Codex / Responses 兼容增强

- 支持 OpenAI Responses 与 Chat Completions 相关转发兼容。
- 支持账号级 OpenAI outbound protocol 切换。
- 支持 Base URL 模型发现与模型列表同步能力。
- 保留 Codex 官方下载/配置入口。
- 面向 Claude Code / Codex / Cursor 等客户端使用场景做了 UI 与接口适配。

### 7. 渠道状态与监控展示

- 用户侧渠道状态页 `/monitor`。
- 顶部状态区展示系统运行时长文案：`系统已正常运行 156天 X分钟X秒`。
- 保留渠道监控、可用性、延迟、历史状态展示。

### 8. 运营与后台体验优化

- 管理员给当前登录管理员自己充值/扣款后，顶部余额立即刷新。
- 使用表格对齐、默认 User-Agent、配额阈值等运营细节优化。
- 后台模型配置、签到管理、限时权益均有明确入口。
- 站点设置初始化改为“只补缺失 key”，不会覆盖已有品牌配置、Logo、自定义菜单、签到配置。

### 9. 一键配置脚本

保留面向用户的安装/配置辅助脚本入口，例如：

- `sekiro-install.sh`
- `sekiro-install.ps1`
- `sekiro-codex.sh`
- `sekiro-codex.ps1`

具体以线上部署暴露的脚本为准。

## 与官方版主要差异

| 模块 | 官方 Sub2API | SekiroCloud 二开版 |
| --- | --- | --- |
| 每日签到 | 无独立运营闭环 | 用户签到、奖励入余额、后台管理、奖励配置 |
| 签到记录 | 无 | 独立表记录，不污染兑换码 |
| 限时权益 | 无 | 管理员发放，首次 API 请求激活，到期自动扣回 |
| 模型广场 | 基础能力/或无定制 | 用户侧展示、后台模型配置、Codex 官方入口、联系卡片 |
| 自定义菜单 | iframe 为主 | 支持新窗口/iframe 打开方式 |
| Codex 下载 | 官方逻辑 | 只保留官方页面，不托管安装包 |
| 设置初始化 | 可能覆盖默认项 | 只补缺失 key，保护品牌与运营配置 |
| 渠道状态 | 基础监控 | 增加运营运行时长展示 |
| 管理体验 | 官方后台 | 增加限时权益、签到管理、充值后余额刷新等 |

## 技术栈

- Backend：Go 1.26.2、Gin、Ent、PostgreSQL、Redis
- Frontend：Vue 3、TypeScript、Vite、TailwindCSS、Pinia
- 部署：Docker / Docker Compose
- 数据库：PostgreSQL
- 缓存：Redis

## 本地开发

### 前端

```bash
cd frontend
corepack pnpm install
corepack pnpm exec vue-tsc --noEmit
corepack pnpm build
```

### 后端

```bash
cd backend
GOCACHE=$PWD/.cache/go-build \
GOMODCACHE=$PWD/.cache/go-mod \
GOPROXY=https://goproxy.cn,direct \
go test ./internal/handler/admin ./internal/server ./cmd/server ./internal/service ./internal/server/middleware -count=1
```

### 构建嵌入前端的后端二进制

先构建前端，再构建后端，并启用 `embed` tag：

```bash
cd frontend
corepack pnpm build

cd ../backend
GOCACHE=$PWD/.cache/go-build \
GOMODCACHE=$PWD/.cache/go-mod \
GOPROXY=https://goproxy.cn,direct \
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
go build -tags embed \
  -ldflags "-X main.Version=0.1.120 -X main.BuildType=source" \
  -o main ./cmd/server
```

注意：生产容器真实入口通常是 `/app/sub2api`，替换镜像二进制时不要误放到 `/app/main`。

## 部署注意事项

- 部署前备份 `settings` 表中的品牌、Logo、自定义菜单、签到配置。
- 不要用旧镜像覆盖当前线上二开功能。
- 更新 Compose 时只替换 `services.sub2api.image`，不要误改 Postgres/Redis 镜像。
- 前端改动必须先 `pnpm build`，再用 `go build -tags embed` 打入后端二进制。
- 旧 Codex 本站安装包路径应保持 404：`/downloads/codex/...`。
- 版本号需要通过 ldflags 注入，否则前端/公开 settings 可能仍显示旧版本。

## 重要数据表/配置

- `settings`：站点配置、品牌、Logo、自定义菜单、签到奖励配置等。
- `user_daily_checkins`：每日签到记录。
- `timed_user_grants`：限时权益记录。
- `redeem_codes`：兑换码业务，不用于签到奖励。

## License

本仓库继承上游 Sub2API 的开源许可；二开部分用于 SekiroCloud 线上运营场景。
