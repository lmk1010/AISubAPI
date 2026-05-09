---
description: How to deploy Sub2API for a customer (commercial one-click deployment on overseas VPS)
---

# Sub2API Commercial Deployment Workflow

## Overview

Sub2API 商业交付部署使用 `deploy/quickdeploy.sh` 一键脚本，面向客户售卖场景。
脚本会自动完成：系统调优 → Docker 安装 → 密码生成 → PG/Redis/App 启动 → 输出交付凭证。

所有数据库和 Redis **不暴露公网端口**，仅 Docker 内网通信。密码全部随机生成。

---

## 文件清单

| 文件 | 用途 |
|------|------|
| `deploy/quickdeploy.sh` | 一键部署脚本（786 行） |
| `deploy/quickdeploy-uninstall.sh` | 卸载/清理脚本 |
| `deploy/docker-compose.yml` | 内部开发用 compose（不是商业交付用的） |
| `deploy/docker-deploy.sh` | 内部准备脚本（不是商业交付用的） |

> **注意**：`quickdeploy.sh` 是商业交付脚本，与 `docker-deploy.sh`/`install.sh` 是完全不同的东西。

---

## 部署步骤

### 场景 1：最简部署（HTTP，端口 8080）

SSH 登录客户服务器后执行：

```bash
curl -fsSL https://raw.githubusercontent.com/Wei-Shaw/sub2api/main/deploy/quickdeploy.sh | sudo bash
```

或者先下载再执行：

```bash
wget -O quickdeploy.sh https://raw.githubusercontent.com/Wei-Shaw/sub2api/main/deploy/quickdeploy.sh
sudo bash quickdeploy.sh
```

### 场景 2：带域名 + HTTPS（推荐）

> 前提：域名 A 记录已指向服务器 IP

```bash
sudo bash quickdeploy.sh --domain api.customer.com
```

自动启用 Caddy 反代 + Let's Encrypt 自动 HTTPS。

### 场景 3：自定义端口

```bash
sudo bash quickdeploy.sh --port 3000
```

### 场景 4：全部自定义

```bash
sudo bash quickdeploy.sh \
  --domain api.customer.com \
  --port 8080 \
  --dir /data/sub2api \
  --tz UTC \
  --image-tag v1.2.3
```

---

## 脚本参数

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `--port <port>` | `8080` | Sub2API 监听端口 |
| `--domain <fqdn>` | 空 | 域名，自动启用 Caddy HTTPS |
| `--with-caddy` | `false` | 手动启用 Caddy（无域名时用） |
| `--dir <path>` | `/opt/sub2api` | 安装目录 |
| `--image-tag <tag>` | `latest` | Docker 镜像版本 |
| `--tz <timezone>` | `Asia/Shanghai` | 应用时区 |
| `--skip-docker` | `false` | 跳过 Docker 安装 |
| `--skip-tuning` | `false` | 跳过系统调优（swap/内核/防火墙） |

---

## 脚本执行流程（9 步）

1. **系统检测** — OS / arch / RAM / CPU cores / distro
2. **Swap** — ≤2GB RAM 自动建 1GB swap
3. **内核调优** — BBR、somaxconn 65535、tcp_fastopen、file-max 1M
4. **文件描述符** — nofile 提到 1048576
5. **防火墙** — 自动检测 ufw / firewalld / iptables 放行端口
6. **Docker 安装** — 支持 Ubuntu/Debian/CentOS/RHEL/Rocky/Alma/Fedora，配置 daemon.json
7. **密码生成 + 资源自适应** — 根据 RAM 自动算 PG shared_buffers / Redis maxmemory
8. **启动** — docker compose up -d，等健康检查通过
9. **交付** — 输出 Delivery Card（Dashboard URL、Admin 账密、基础设施密钥、服务器信息）

---

## RAM 自适应配置

| RAM | PG shared_buffers | PG effective_cache | Redis maxmemory | DB 连接池 |
|-----|-------------------|--------------------|-----------------|----------|
| ≤1GB | 128MB | 512MB | 128mb | 50 |
| ≤2GB | 256MB | 1GB | 256mb | 100 |
| ≤4GB | 512MB | 2GB | 512mb | 100 |
| ≤8GB | 1GB | 4GB | 1gb | 200 |
| >8GB | 2GB | 6GB | 2gb | 256 |

---

## 部署后文件结构

```
/opt/sub2api/
├── docker-compose.yml    # 生成的 compose 文件
├── .credentials          # 所有凭证（mode 600，仅 root 可读）
└── Caddyfile             # 仅 --domain 时生成
```

---

## 部署后常用操作

```bash
# 查看日志
cd /opt/sub2api && docker compose logs -f

# 重启
cd /opt/sub2api && docker compose restart

# 停止
cd /opt/sub2api && docker compose down

# 更新到最新版
cd /opt/sub2api && docker compose pull && docker compose up -d

# 查看状态
cd /opt/sub2api && docker compose ps

# 查看凭证
sudo cat /opt/sub2api/.credentials
```

---

## 卸载

```bash
sudo bash /path/to/quickdeploy-uninstall.sh
# 可选参数：
#   --dir /opt/sub2api   指定目录
#   --keep-data          保留数据库数据
```

---

## 安全要点

- PG / Redis **仅 Docker 内网**，不暴露公网端口
- 所有密码 `openssl rand` 生成，64 位 hex
- `.credentials` 文件 mode 600
- Docker 日志自动滚动（20MB × 5 files）
- 部署后建议：SSH 改端口 + 禁密码登录、Admin 开 2FA、配域名走 HTTPS

---

## 故障排查

```bash
# 容器没起来
docker compose -f /opt/sub2api/docker-compose.yml ps
docker compose -f /opt/sub2api/docker-compose.yml logs sub2api

# PG 连接失败
docker compose -f /opt/sub2api/docker-compose.yml logs postgres

# Redis 连接失败
docker compose -f /opt/sub2api/docker-compose.yml logs redis

# 端口被占用
ss -tlnp | grep :8080

# 防火墙没开
ufw status          # Ubuntu
firewall-cmd --list-all  # CentOS
```
