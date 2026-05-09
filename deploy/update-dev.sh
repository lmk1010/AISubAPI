#!/usr/bin/env bash
# =============================================================================
# Sub2API dev-custom 快速更新部署脚本
# 用法: bash deploy/update-dev.sh
# 说明: 从 dev-custom 分支拉取最新代码, 构建镜像, 重启服务
# =============================================================================
set -euo pipefail

# ── 配置 ──────────────────────────────────────────────────────────────────────
SERVER_HOST="104.129.51.171"
SERVER_PORT="22080"
SERVER_USER="root"
SERVER_PASS="1SOWkkP0Ullh357fI2"
BRANCH="dev-custom"
REPO_URL="https://github.com/lmk1010/AISubAPI.git"
IMAGE_TAG="aisubapi:dev-custom"
COMPOSE_DIR="/opt/aisubapi"
TMP_DIR="/tmp/AISubAPI"

# ── 颜色 ──────────────────────────────────────────────────────────────────────
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; CYAN='\033[0;36m'; NC='\033[0m'
info()  { echo -e "${CYAN}[INFO]${NC} $*"; }
ok()    { echo -e "${GREEN}[OK]${NC} $*"; }
warn()  { echo -e "${YELLOW}[WARN]${NC} $*"; }
err()   { echo -e "${RED}[ERROR]${NC} $*"; exit 1; }

# ── SSH 辅助 ──────────────────────────────────────────────────────────────────
SSH_CMD="sshpass -p '${SERVER_PASS}' ssh -o StrictHostKeyChecking=no -o ServerAliveInterval=15 -p ${SERVER_PORT} ${SERVER_USER}@${SERVER_HOST}"

remote() {
  eval "${SSH_CMD} \"$*\""
}

# ── 检查依赖 ──────────────────────────────────────────────────────────────────
command -v sshpass >/dev/null 2>&1 || err "需要安装 sshpass: brew install sshpass 或 brew install esolitos/ipa/sshpass"

# ── Step 1: Git push 本地代码 ─────────────────────────────────────────────────
info "Step 1/5: 推送本地代码到 ${BRANCH}..."
git push origin "${BRANCH}" 2>/dev/null && ok "代码已推送" || warn "推送跳过（可能已是最新）"

# ── Step 2: 远程克隆 ─────────────────────────────────────────────────────────
info "Step 2/5: 远程克隆 ${BRANCH} 分支..."
remote "rm -rf ${TMP_DIR} && git clone --depth 1 -b ${BRANCH} ${REPO_URL} ${TMP_DIR}" || err "克隆失败"
ok "克隆完成"

# ── Step 3: Docker build (no-cache) ──────────────────────────────────────────
info "Step 3/5: 构建 Docker 镜像 ${IMAGE_TAG}（约 3-5 分钟）..."
remote "docker build --no-cache -f ${TMP_DIR}/deploy/Dockerfile -t ${IMAGE_TAG} ${TMP_DIR}" || err "构建失败"
ok "镜像构建完成"

# ── Step 4: 重启服务 ─────────────────────────────────────────────────────────
info "Step 4/5: 重启服务..."
remote "cd ${COMPOSE_DIR} && docker compose down && docker compose up -d" || err "重启失败"
ok "服务已重启"

# ── Step 5: 健康检查 ─────────────────────────────────────────────────────────
info "Step 5/5: 等待健康检查..."
for i in $(seq 1 12); do
  sleep 5
  STATUS=$(remote "docker inspect sub2api --format '{{.State.Health.Status}}' 2>/dev/null" || echo "unknown")
  if [ "${STATUS}" = "healthy" ]; then
    ok "所有服务 healthy!"
    echo ""
    remote "docker ps --format 'table {{.Names}}\t{{.Status}}' | grep sub2api"
    echo ""
    ok "✅ 部署完成! 访问 http://${SERVER_HOST}"
    # 清理临时文件
    remote "rm -rf ${TMP_DIR}" 2>/dev/null || true
    exit 0
  fi
  info "  等待中... (${i}/12) 状态: ${STATUS}"
done

err "健康检查超时，请手动检查: docker logs sub2api"
