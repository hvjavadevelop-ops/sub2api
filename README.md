# SekiroCloud Sub2API Custom Edition

<div align="center">

[![Go](https://img.shields.io/badge/Go-1.26.2-00ADD8.svg)](https://golang.org/)
[![Vue](https://img.shields.io/badge/Vue-3.4+-4FC08D.svg)](https://vuejs.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-336791.svg)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-7+-DC382D.svg)](https://redis.io/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED.svg)](https://www.docker.com/)

**Customized Sub2API build used by 云栖小铺 / SekiroCloud**

English | [中文](README_CN.md)

</div>

---

## Live Site

Experience URL: **https://sekirocloud.site:8443/**

This repository tracks the customized SekiroCloud build based on upstream Sub2API v0.1.120. It is used for AI API relay operations, redeem-code/balance distribution, model showcase, channel monitoring, and admin-side operations.

> This is not the official Sub2API repository. For the upstream project, see Wei-Shaw/sub2api. This fork documents the production customizations used by SekiroCloud.

## Positioning

Sub2API is an AI API gateway platform that converts upstream AI accounts, subscription quotas, API Keys, or OAuth capabilities into OpenAI/Anthropic-compatible APIs. It also provides user, balance, concurrency, group, channel, usage, and payment management.

This custom edition focuses on operational features:

- A clearer model showcase and one-click setup entry for customers
- Daily check-in rewards with an admin management page
- Limited-time grants that start counting only on first API usage
- Compatibility and UI polish for Codex / Claude Code / Cursor workflows
- Safer settings initialization that does not overwrite site branding during upgrades/restarts

## Current Version

- Upstream baseline: Sub2API v0.1.120
- Live site: `https://sekirocloud.site:8443/`
- Production image reference: `sub2api-mobilefix:0.1.120-production-merged-custom-404fix-versionfix2`

## Custom Features

### Daily Check-in

- Calendar icon entry in the top bar, replacing the original language-switcher position.
- The button is not grayed out after check-in; clicking again shows “今天已经签到过了，明天再来”.
- Reward range is configurable through admin settings; display uses `$` style.
- Rewards are credited directly to user balance and recorded in a dedicated check-in table.
- Rewards do not create fake `redeem_codes`, keeping redeem-code semantics clean.

### Admin Check-in Management

- Dedicated admin page: `/admin/daily-checkins`.
- Shows check-in record ID, user information, reward amount, and check-in time.
- Supports search by email, username, or numeric user ID.
- Includes inline reward configuration for `daily_checkin_reward_min` and `daily_checkin_reward_max`.

### Limited-Time Grants

For trials, campaigns, or manual compensation, admins can issue temporary balance or concurrency grants.

- A visible “限时权益” button is available on each user row in `/admin/users`.
- New grants start as `pending`.
- The first real API-Key request activates pending grants and starts the timer.
- Expired grants are deducted automatically and recorded in history.
- Supported statuses: `pending`, `active`, `expired`, `cancelled`.
- Backed by a dedicated `timed_user_grants` table.
- Activation happens in API-Key authentication and refreshes auth/user cache so the current request can see the new quota.

### Model Square

- User-facing model showcase with model examples.
- Admin-side model catalog grouped by provider.
- Codex download entry links only to the official page: `https://developers.openai.com/codex/app`.
- Local mirrored installers are not hosted; stale `/downloads/codex/...` paths return real 404.
- Contact QR code is displayed in a standalone card without cropping.

### Custom Menu Open Mode

- Custom menu items support `open_mode`.
- Default is `new_tab`, avoiding iframe failures caused by `X-Frame-Options` or CSP `frame-ancestors`.
- Explicit `iframe` mode remains available, with a new-window fallback.

### OpenAI / Codex / Responses Compatibility

- OpenAI Responses and Chat Completions compatibility improvements.
- Per-account OpenAI outbound protocol switch.
- Base URL model discovery and model-list sync.
- Codex official setup/download entry retained.
- UI/API flow adjusted for Claude Code / Codex / Cursor-style clients.

### Channel Status

- User route: `/monitor`.
- Top status area displays the operational uptime text, e.g. `系统已正常运行 156天 X分钟X秒`.
- Keeps channel availability, latency, and history views.

### Operational Polish

- If an admin updates their own balance, the top-header balance refreshes immediately.
- Usage-table alignment, User-Agent defaults, quota threshold details, and sidebar scrolling are polished.
- Settings initialization inserts missing keys only and does not overwrite existing branding, logo, custom menus, or check-in configuration.

### One-click Setup Scripts

Depending on deployment configuration, the live service can expose helper scripts such as:

- `sekiro-install.sh`
- `sekiro-install.ps1`
- `sekiro-codex.sh`
- `sekiro-codex.ps1`

## Differences from Upstream

| Area | Upstream Sub2API | SekiroCloud Custom Edition |
| --- | --- | --- |
| Daily check-in | Not an operational loop | User check-in, balance reward, admin records/config |
| Check-in records | N/A | Dedicated table; no redeem-code pollution |
| Limited-time grants | N/A | Admin-issued, first API request activates, auto-deduct on expiry |
| Model Square | Basic/upstream behavior | Provider catalog, Codex official link, contact card |
| Custom menu | iframe-oriented | `new_tab` / `iframe` open mode |
| Codex installers | Upstream/default behavior | Official page only; stale local installer paths return 404 |
| Settings initialization | Default insertion may overwrite | Insert missing keys only, preserve production branding |
| Channel status | Basic monitor | Adds operational uptime copy |
| Admin UX | Upstream admin panel | Check-in management, grants, balance refresh, visible entries |

## Stack

- Backend: Go 1.26.2, Gin, Ent, PostgreSQL, Redis
- Frontend: Vue 3, TypeScript, Vite, TailwindCSS, Pinia
- Deployment: Docker / Docker Compose

## Development

### Frontend

```bash
cd frontend
corepack pnpm install
corepack pnpm exec vue-tsc --noEmit
corepack pnpm build
```

### Backend

```bash
cd backend
GOCACHE=$PWD/.cache/go-build \
GOMODCACHE=$PWD/.cache/go-mod \
GOPROXY=https://goproxy.cn,direct \
go test ./internal/handler/admin ./internal/server ./cmd/server ./internal/service ./internal/server/middleware -count=1
```

### Embedded Binary Build

Build the frontend first, then build the backend with the `embed` tag:

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

Production containers normally execute `/app/sub2api`; do not replace `/app/main` by mistake.

## Deployment Notes

- Back up branding/settings before deployment: site name, subtitle, logo, custom menu, and check-in reward config.
- Do not deploy an old image over the current custom feature set.
- When editing Compose, update only `services.sub2api.image`; do not touch Postgres/Redis images.
- Frontend changes must be embedded through `pnpm build` + `go build -tags embed`.
- Stale Codex installer paths under `/downloads/codex/...` should keep returning 404.
- Inject version ldflags during release builds, otherwise public settings/sidebar may show an old version.

## Important Tables / Settings

- `settings`: site branding, logo, custom menus, check-in reward config, and other options.
- `user_daily_checkins`: daily check-in records.
- `timed_user_grants`: limited-time grant records.
- `redeem_codes`: redeem-code business only; not used for check-in rewards.

## License

This fork inherits the upstream Sub2API license. Custom changes are maintained for SekiroCloud production operations.
