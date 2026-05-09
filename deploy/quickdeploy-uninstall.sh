#!/usr/bin/env bash
# =============================================================================
#  Sub2API — Uninstall / Clean Removal Script
# =============================================================================
#  Usage:  sudo bash quickdeploy-uninstall.sh [--dir /opt/sub2api] [--keep-data]
# =============================================================================

set -euo pipefail

INSTALL_DIR="/opt/sub2api"
KEEP_DATA=false

while [[ $# -gt 0 ]]; do
  case "$1" in
    --dir)        INSTALL_DIR="$2"; shift 2 ;;
    --keep-data)  KEEP_DATA=true; shift ;;
    *) echo "Unknown option: $1"; exit 1 ;;
  esac
done

R='\033[0;31m'; G='\033[0;32m'; Y='\033[1;33m'; W='\033[1;37m'; N='\033[0m'

echo ""
echo -e "${W}Sub2API Uninstaller${N}"
echo -e "────────────────────────────────"
echo -e "Install dir: ${Y}${INSTALL_DIR}${N}"
echo -e "Keep data:   ${Y}${KEEP_DATA}${N}"
echo ""

[[ "$(id -u)" -ne 0 ]] && { echo -e "${R}Please run as root.${N}"; exit 1; }

if [[ ! -d "$INSTALL_DIR" ]]; then
  echo -e "${Y}Directory ${INSTALL_DIR} not found — nothing to do.${N}"
  exit 0
fi

read -rp "⚠️  This will stop and remove all Sub2API containers. Continue? (y/N): " CONFIRM
[[ ! "$CONFIRM" =~ ^[Yy]$ ]] && { echo "Cancelled."; exit 0; }

cd "$INSTALL_DIR"

echo -e "\n${W}Stopping containers...${N}"
docker compose down --timeout 30 2>/dev/null || true

if [[ "$KEEP_DATA" == "false" ]]; then
  read -rp "🗑️  Also delete ALL data (database, redis, uploads)? (y/N): " DEL_DATA
  if [[ "$DEL_DATA" =~ ^[Yy]$ ]]; then
    echo -e "${R}Removing volumes...${N}"
    docker compose down -v --timeout 30 2>/dev/null || true
    echo -e "${G}Volumes removed.${N}"
  fi
fi

echo -e "${W}Removing project files...${N}"
rm -f docker-compose.yml Caddyfile .credentials
echo -e "${G}✓ Sub2API uninstalled.${N}\n"
