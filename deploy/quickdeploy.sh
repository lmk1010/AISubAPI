#!/usr/bin/env bash
# =============================================================================
#  Sub2API — Commercial One-Click Deployment Script
# =============================================================================
#  Usage:
#    curl -fsSL https://raw.githubusercontent.com/Wei-Shaw/sub2api/main/deploy/quickdeploy.sh | sudo bash
#    — or —
#    sudo bash quickdeploy.sh [--port 8080] [--domain api.example.com] [--with-caddy]
#
#  What it does:
#    1. Detects OS / arch, installs Docker + Compose if missing
#    2. Tunes kernel for high-connection workloads (BBR, file descriptors, etc.)
#    3. Ensures swap exists on low-memory VPS (< 2GB)
#    4. Opens firewall ports automatically (ufw / firewalld)
#    5. Creates isolated project directory /opt/sub2api
#    6. Generates ALL secrets (PG / Redis / JWT / TOTP / Admin) randomly
#    7. Launches PostgreSQL 18 + Redis 8 + Sub2API via Docker Compose
#    8. Waits for health checks to pass
#    9. Prints a clean "Delivery Card" with all credentials
#
#  Optimized for overseas VPS (DigitalOcean, Vultr, Hetzner, AWS Lightsail, etc.)
#  Designed for commercial delivery — zero manual config required.
# =============================================================================

set -euo pipefail

# ── Tunables (overridable via flags) ─────────────────────────────────────────
INSTALL_DIR="/opt/sub2api"
SERVER_PORT=8080
DOMAIN=""
WITH_CADDY=false
IMAGE_TAG="latest"
TZ="Asia/Shanghai"
SKIP_DOCKER_INSTALL=false
SKIP_TUNING=false
# ── End tunables ─────────────────────────────────────────────────────────────

# ── Parse CLI flags ──────────────────────────────────────────────────────────
while [[ $# -gt 0 ]]; do
  case "$1" in
    --port)        SERVER_PORT="$2"; shift 2 ;;
    --domain)      DOMAIN="$2"; shift 2 ;;
    --dir)         INSTALL_DIR="$2"; shift 2 ;;
    --image-tag)   IMAGE_TAG="$2"; shift 2 ;;
    --tz)          TZ="$2"; shift 2 ;;
    --with-caddy)  WITH_CADDY=true; shift ;;
    --skip-docker) SKIP_DOCKER_INSTALL=true; shift ;;
    --skip-tuning) SKIP_TUNING=true; shift ;;
    -h|--help)
      echo "Usage: sudo bash quickdeploy.sh [options]"
      echo ""
      echo "  --port <port>       Sub2API listen port (default: 8080)"
      echo "  --domain <fqdn>     Domain for Caddy HTTPS (auto-enables --with-caddy)"
      echo "  --with-caddy        Add Caddy reverse proxy with auto-TLS"
      echo "  --dir <path>        Installation directory (default: /opt/sub2api)"
      echo "  --image-tag <tag>   Docker image tag (default: latest)"
      echo "  --tz <timezone>     App timezone (default: Asia/Shanghai)"
      echo "  --skip-docker       Skip Docker installation check"
      echo "  --skip-tuning       Skip kernel/swap/firewall tuning"
      echo ""
      echo "Examples:"
      echo "  sudo bash quickdeploy.sh --domain api.example.com"
      echo "  sudo bash quickdeploy.sh --port 3000 --tz UTC"
      exit 0
      ;;
    *) echo "Unknown option: $1"; exit 1 ;;
  esac
done

# If domain is given, enable caddy automatically
[[ -n "$DOMAIN" ]] && WITH_CADDY=true

# ── Colors ───────────────────────────────────────────────────────────────────
R='\033[0;31m'; G='\033[0;32m'; Y='\033[1;33m'; B='\033[0;34m'; C='\033[0;36m'; W='\033[1;37m'; N='\033[0m'
info()    { echo -e "${B}▸${N} $*"; }
ok()      { echo -e "${G}✓${N} $*"; }
warn()    { echo -e "${Y}⚠${N} $*"; }
err()     { echo -e "${R}✗${N} $*"; }
fatal()   { err "$*"; exit 1; }
banner()  { echo -e "\n${C}═══════════════════════════════════════════════════════${N}"; echo -e "${W}  $*${N}"; echo -e "${C}═══════════════════════════════════════════════════════${N}\n"; }

# ── Preflight checks ────────────────────────────────────────────────────────
banner "Sub2API — Commercial Deployment"

[[ "$(id -u)" -ne 0 ]] && fatal "Please run as root:  sudo bash quickdeploy.sh"

# Detect OS
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) fatal "Unsupported architecture: $ARCH" ;;
esac

# Detect distro info
DISTRO="unknown"
DISTRO_VERSION=""
if [[ -f /etc/os-release ]]; then
  . /etc/os-release
  DISTRO="$ID"
  DISTRO_VERSION="${VERSION_ID:-}"
fi

TOTAL_RAM_MB=$(awk '/MemTotal/ {printf "%d", $2/1024}' /proc/meminfo 2>/dev/null || echo "0")
CPU_CORES=$(nproc 2>/dev/null || echo "1")

info "System: ${OS}/${ARCH} | ${DISTRO} ${DISTRO_VERSION} | RAM: ${TOTAL_RAM_MB}MB | CPU: ${CPU_CORES} cores"

# ── Step 0: System tuning (swap + kernel + firewall) ─────────────────────────
if [[ "$SKIP_TUNING" != "true" ]]; then
  banner "System Optimization"

  # ── 0a: Swap — ensure at least 1GB on low-memory VPS ──────────────────────
  setup_swap() {
    local SWAP_SIZE_MB=1024
    local CURRENT_SWAP_MB
    CURRENT_SWAP_MB=$(free -m | awk '/Swap:/ {print $2}')

    if [[ "$TOTAL_RAM_MB" -le 2048 && "$CURRENT_SWAP_MB" -lt 512 ]]; then
      info "Low memory detected (${TOTAL_RAM_MB}MB). Creating ${SWAP_SIZE_MB}MB swap..."
      if [[ ! -f /swapfile ]]; then
        dd if=/dev/zero of=/swapfile bs=1M count=${SWAP_SIZE_MB} status=none
        chmod 600 /swapfile
        mkswap /swapfile >/dev/null
        swapon /swapfile
        # Persist across reboots
        if ! grep -q '/swapfile' /etc/fstab; then
          echo '/swapfile none swap sw 0 0' >> /etc/fstab
        fi
        ok "Swap created: ${SWAP_SIZE_MB}MB"
      else
        swapon /swapfile 2>/dev/null || true
        ok "Swap already exists, ensured active"
      fi
      # Lower swappiness for better performance
      sysctl -w vm.swappiness=10 >/dev/null 2>&1
    else
      ok "Memory sufficient (${TOTAL_RAM_MB}MB) or swap already configured (${CURRENT_SWAP_MB}MB)"
    fi
  }

  # ── 0b: Kernel tuning — high connections + BBR ─────────────────────────────
  tune_kernel() {
    local SYSCTL_FILE="/etc/sysctl.d/99-sub2api.conf"

    if [[ -f "$SYSCTL_FILE" ]]; then
      ok "Kernel tuning already applied"
      return
    fi

    info "Applying kernel optimizations..."
    cat > "$SYSCTL_FILE" <<'SYSCTL'
# Sub2API — Kernel tuning for high-connection API gateway
# TCP BBR congestion control (better throughput on overseas links)
net.core.default_qdisc = fq
net.ipv4.tcp_congestion_control = bbr

# Connection tracking & backlog
net.core.somaxconn = 65535
net.core.netdev_max_backlog = 65535
net.ipv4.tcp_max_syn_backlog = 65535

# File descriptor limits
fs.file-max = 1048576
fs.inotify.max_user_instances = 8192
fs.inotify.max_user_watches = 524288

# TCP keepalive (detect dead connections faster)
net.ipv4.tcp_keepalive_time = 600
net.ipv4.tcp_keepalive_intvl = 30
net.ipv4.tcp_keepalive_probes = 10

# TCP performance
net.ipv4.tcp_fin_timeout = 30
net.ipv4.tcp_tw_reuse = 1
net.ipv4.tcp_max_tw_buckets = 65535
net.ipv4.ip_local_port_range = 1024 65535
net.ipv4.tcp_fastopen = 3
net.ipv4.tcp_mtu_probing = 1
net.ipv4.tcp_slow_start_after_idle = 0

# Memory buffers
net.core.rmem_max = 16777216
net.core.wmem_max = 16777216
net.ipv4.tcp_rmem = 4096 212992 16777216
net.ipv4.tcp_wmem = 4096 212992 16777216

# VM
vm.swappiness = 10
vm.overcommit_memory = 1
SYSCTL
    sysctl --system >/dev/null 2>&1
    ok "Kernel optimized (BBR enabled, connection limits raised)"
  }

  # ── 0c: File descriptor limits for systemd services ────────────────────────
  tune_limits() {
    local LIMITS_FILE="/etc/security/limits.d/99-sub2api.conf"
    if [[ ! -f "$LIMITS_FILE" ]]; then
      cat > "$LIMITS_FILE" <<'LIMITS'
* soft nofile 1048576
* hard nofile 1048576
root soft nofile 1048576
root hard nofile 1048576
LIMITS
      ok "File descriptor limits raised (1M)"
    fi
  }

  # ── 0d: Firewall — open required ports ─────────────────────────────────────
  configure_firewall() {
    local PORTS_TO_OPEN=("${SERVER_PORT}/tcp")
    if [[ "$WITH_CADDY" == "true" ]]; then
      PORTS_TO_OPEN=("80/tcp" "443/tcp" "443/udp")
    fi

    # ufw (Ubuntu/Debian)
    if command -v ufw &>/dev/null && ufw status | grep -q "active"; then
      for port in "${PORTS_TO_OPEN[@]}"; do
        ufw allow "$port" >/dev/null 2>&1
      done
      ok "UFW: opened ports ${PORTS_TO_OPEN[*]}"
      return
    fi

    # firewalld (CentOS/RHEL/Rocky)
    if command -v firewall-cmd &>/dev/null && systemctl is-active firewalld &>/dev/null; then
      for port in "${PORTS_TO_OPEN[@]}"; do
        firewall-cmd --permanent --add-port="$port" >/dev/null 2>&1
      done
      firewall-cmd --reload >/dev/null 2>&1
      ok "Firewalld: opened ports ${PORTS_TO_OPEN[*]}"
      return
    fi

    # iptables fallback
    if command -v iptables &>/dev/null; then
      for port in "${PORTS_TO_OPEN[@]}"; do
        local p="${port%/*}"
        local proto="${port#*/}"
        iptables -C INPUT -p "$proto" --dport "$p" -j ACCEPT 2>/dev/null \
          || iptables -I INPUT -p "$proto" --dport "$p" -j ACCEPT 2>/dev/null
      done
      ok "iptables: opened ports ${PORTS_TO_OPEN[*]}"
      return
    fi

    warn "No firewall detected — ensure ports ${PORTS_TO_OPEN[*]} are accessible"
  }

  setup_swap
  tune_kernel
  tune_limits
  configure_firewall
fi

# ── Step 1: Install Docker if needed ────────────────────────────────────────
install_docker() {
  if command -v docker &>/dev/null && docker compose version &>/dev/null; then
    ok "Docker + Compose already installed ($(docker --version | awk '{print $3}' | tr -d ','))"
    return
  fi

  if [[ "$SKIP_DOCKER_INSTALL" == "true" ]]; then
    fatal "Docker not found and --skip-docker specified"
  fi

  banner "Installing Docker"

  if [[ -f /etc/os-release ]]; then
    . /etc/os-release
    case "$ID" in
      ubuntu|debian)
        apt-get update -qq
        apt-get install -y -qq ca-certificates curl gnupg lsb-release >/dev/null 2>&1
        install -m 0755 -d /etc/apt/keyrings
        curl -fsSL "https://download.docker.com/linux/${ID}/gpg" | gpg --dearmor -o /etc/apt/keyrings/docker.gpg 2>/dev/null
        chmod a+r /etc/apt/keyrings/docker.gpg
        echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/${ID} $(lsb_release -cs) stable" > /etc/apt/sources.list.d/docker.list
        apt-get update -qq
        apt-get install -y -qq docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin >/dev/null 2>&1
        ;;
      centos|rhel|rocky|almalinux|fedora)
        yum install -y -q yum-utils >/dev/null 2>&1
        yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo >/dev/null 2>&1
        yum install -y -q docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin >/dev/null 2>&1
        ;;
      *)
        warn "Unknown distro '${ID}', trying get.docker.com..."
        curl -fsSL https://get.docker.com | sh
        ;;
    esac
  elif [[ "$OS" == "darwin" ]]; then
    fatal "macOS detected — please install Docker Desktop manually: https://docker.com/products/docker-desktop"
  else
    warn "Unknown OS, trying get.docker.com..."
    curl -fsSL https://get.docker.com | sh
  fi

  # Enable and start Docker
  systemctl enable --now docker >/dev/null 2>&1 || true

  # Configure Docker daemon for production
  mkdir -p /etc/docker
  if [[ ! -f /etc/docker/daemon.json ]]; then
    cat > /etc/docker/daemon.json <<'DAEMON'
{
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "20m",
    "max-file": "5"
  },
  "storage-driver": "overlay2",
  "default-ulimits": {
    "nofile": { "Name": "nofile", "Soft": 100000, "Hard": 100000 }
  },
  "live-restore": true
}
DAEMON
    systemctl restart docker >/dev/null 2>&1 || true
  fi

  ok "Docker installed & configured for production"
}

install_docker

# ── Step 2: Generate all secrets ─────────────────────────────────────────────
banner "Generating Secure Credentials"

gen_secret()  { openssl rand -hex 32; }
gen_passwd()  { openssl rand -base64 18 | tr -d '/+=' | head -c 24; }

POSTGRES_PASSWORD="$(gen_secret)"
REDIS_PASSWORD="$(gen_secret)"
JWT_SECRET="$(gen_secret)"
TOTP_ENCRYPTION_KEY="$(gen_secret)"
ADMIN_PASSWORD="$(gen_passwd)"
ADMIN_EMAIL="admin@sub2api.local"

ok "All secrets generated randomly"

# ── Step 3: Create project directory & files ─────────────────────────────────
banner "Creating Project: ${INSTALL_DIR}"

mkdir -p "${INSTALL_DIR}"
cd "${INSTALL_DIR}"

# ── Auto-size PostgreSQL for this VPS ────────────────────────────────────────
# Rule of thumb: shared_buffers = 25% RAM, effective_cache = 75% RAM
if [[ "$TOTAL_RAM_MB" -le 1024 ]]; then
  PG_SHARED_BUFFERS="128MB"
  PG_EFFECTIVE_CACHE="512MB"
  PG_MAINT_MEM="64MB"
  DB_MAX_CONNS=50
  REDIS_MAXMEM="128mb"
elif [[ "$TOTAL_RAM_MB" -le 2048 ]]; then
  PG_SHARED_BUFFERS="256MB"
  PG_EFFECTIVE_CACHE="1GB"
  PG_MAINT_MEM="128MB"
  DB_MAX_CONNS=100
  REDIS_MAXMEM="256mb"
elif [[ "$TOTAL_RAM_MB" -le 4096 ]]; then
  PG_SHARED_BUFFERS="512MB"
  PG_EFFECTIVE_CACHE="2GB"
  PG_MAINT_MEM="256MB"
  DB_MAX_CONNS=100
  REDIS_MAXMEM="512mb"
elif [[ "$TOTAL_RAM_MB" -le 8192 ]]; then
  PG_SHARED_BUFFERS="1GB"
  PG_EFFECTIVE_CACHE="4GB"
  PG_MAINT_MEM="512MB"
  DB_MAX_CONNS=200
  REDIS_MAXMEM="1gb"
else
  PG_SHARED_BUFFERS="2GB"
  PG_EFFECTIVE_CACHE="6GB"
  PG_MAINT_MEM="1GB"
  DB_MAX_CONNS=256
  REDIS_MAXMEM="2gb"
fi

info "Auto-sized for ${TOTAL_RAM_MB}MB RAM: PG shared_buffers=${PG_SHARED_BUFFERS}, Redis maxmem=${REDIS_MAXMEM}"

# ── docker-compose.yml ───────────────────────────────────────────────────────
info "Writing docker-compose.yml..."

CADDY_SERVICE=""
CADDY_VOLUME=""
CADDY_NETWORK_DEP=""
if [[ "$WITH_CADDY" == "true" ]]; then
  CADDY_SERVICE="
  # =========================================================================
  # Caddy — Auto-TLS Reverse Proxy
  # =========================================================================
  caddy:
    image: caddy:2-alpine
    container_name: sub2api-caddy
    restart: unless-stopped
    ports:
      - '0.0.0.0:80:80'
      - '0.0.0.0:443:443'
      - '0.0.0.0:443:443/udp'
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile:ro
      - caddy_data:/data
      - caddy_config:/config
    networks:
      - sub2api-net
    depends_on:
      sub2api:
        condition: service_healthy"

  CADDY_VOLUME="
  caddy_data:
    driver: local
  caddy_config:
    driver: local"
fi

# Determine sub2api port binding
if [[ "$WITH_CADDY" == "true" ]]; then
  # Only expose on internal network, caddy handles public
  PORT_BINDING="127.0.0.1:${SERVER_PORT}:8080"
else
  PORT_BINDING="0.0.0.0:${SERVER_PORT}:8080"
fi

cat > docker-compose.yml <<DEOF
# =============================================================================
# Sub2API — Production Stack (auto-generated by quickdeploy.sh)
# =============================================================================
services:
  # =========================================================================
  # Sub2API Application
  # =========================================================================
  sub2api:
    image: weishaw/sub2api:${IMAGE_TAG}
    container_name: sub2api
    restart: unless-stopped
    ulimits:
      nofile: { soft: 100000, hard: 100000 }
    ports:
      - "${PORT_BINDING}"
    volumes:
      - sub2api_data:/app/data
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    networks:
      - sub2api-net
    environment:
      - AUTO_SETUP=true
      - SERVER_HOST=0.0.0.0
      - SERVER_PORT=8080
      - SERVER_MODE=release
      - SERVER_H2C_ENABLED=true
      - RUN_MODE=standard
      - TZ=${TZ}
      # ── Database ──
      - DATABASE_HOST=postgres
      - DATABASE_PORT=5432
      - DATABASE_USER=sub2api
      - DATABASE_PASSWORD=${POSTGRES_PASSWORD}
      - DATABASE_DBNAME=sub2api
      - DATABASE_SSLMODE=disable
      - DATABASE_MAX_OPEN_CONNS=${DB_MAX_CONNS}
      - DATABASE_MAX_IDLE_CONNS=$((DB_MAX_CONNS / 2))
      - DATABASE_CONN_MAX_LIFETIME_MINUTES=30
      - DATABASE_CONN_MAX_IDLE_TIME_MINUTES=5
      # ── Redis ──
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - REDIS_PASSWORD=${REDIS_PASSWORD}
      - REDIS_DB=0
      - REDIS_POOL_SIZE=1024
      - REDIS_MIN_IDLE_CONNS=10
      - REDIS_ENABLE_TLS=false
      # ── Auth ──
      - ADMIN_EMAIL=${ADMIN_EMAIL}
      - ADMIN_PASSWORD=${ADMIN_PASSWORD}
      - JWT_SECRET=${JWT_SECRET}
      - JWT_EXPIRE_HOUR=24
      - TOTP_ENCRYPTION_KEY=${TOTP_ENCRYPTION_KEY}
      # ── Logging ──
      - LOG_LEVEL=info
      - LOG_FORMAT=json
      - LOG_OUTPUT_TO_STDOUT=true
      - LOG_OUTPUT_TO_FILE=true
      # ── Dashboard ──
      - DASHBOARD_AGGREGATION_ENABLED=true
      - DASHBOARD_AGGREGATION_INTERVAL_SECONDS=60
      # ── Ops ──
      - OPS_ENABLED=true
    healthcheck:
      test: ["CMD", "wget", "-q", "-T", "5", "-O", "/dev/null", "http://localhost:8080/health"]
      interval: 15s
      timeout: 10s
      retries: 5
      start_period: 30s

  # =========================================================================
  # PostgreSQL 18
  # =========================================================================
  postgres:
    image: postgres:18-alpine
    container_name: sub2api-postgres
    restart: unless-stopped
    ulimits:
      nofile: { soft: 100000, hard: 100000 }
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - sub2api-net
    environment:
      - PGDATA=/var/lib/postgresql/data
      - POSTGRES_USER=sub2api
      - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
      - POSTGRES_DB=sub2api
      - TZ=${TZ}
    command:
      - "postgres"
      - "-c" 
      - "max_connections=256"
      - "-c"
      - "shared_buffers=${PG_SHARED_BUFFERS}"
      - "-c"
      - "effective_cache_size=${PG_EFFECTIVE_CACHE}"
      - "-c"
      - "work_mem=4MB"
      - "-c"
      - "maintenance_work_mem=${PG_MAINT_MEM}"
      - "-c"
      - "wal_buffers=16MB"
      - "-c"
      - "checkpoint_completion_target=0.9"
      - "-c"
      - "random_page_cost=1.1"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U sub2api -d sub2api"]
      interval: 10s
      timeout: 5s
      retries: 5
      start_period: 10s
    # DB only on internal network — never exposed to host
${CADDY_SERVICE}
  # =========================================================================
  # Redis 8
  # =========================================================================
  redis:
    image: redis:8-alpine
    container_name: sub2api-redis
    restart: unless-stopped
    ulimits:
      nofile: { soft: 100000, hard: 100000 }
    volumes:
      - redis_data:/data
    command: >
      sh -c 'redis-server
        --save 60 1
        --appendonly yes
        --appendfsync everysec
        --maxmemory ${REDIS_MAXMEM}
        --maxmemory-policy allkeys-lru
        --requirepass "${REDIS_PASSWORD}"'
    environment:
      - TZ=${TZ}
      - REDISCLI_AUTH=${REDIS_PASSWORD}
    networks:
      - sub2api-net
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5
      start_period: 5s
    # Redis only on internal network — never exposed to host

volumes:
  sub2api_data:
    driver: local
  postgres_data:
    driver: local
  redis_data:
    driver: local${CADDY_VOLUME}

networks:
  sub2api-net:
    driver: bridge
DEOF

ok "docker-compose.yml created"

# ── Caddyfile (if --with-caddy) ─────────────────────────────────────────────
if [[ "$WITH_CADDY" == "true" && -n "$DOMAIN" ]]; then
  info "Writing Caddyfile for ${DOMAIN}..."
  cat > Caddyfile <<CEOF
${DOMAIN} {
    reverse_proxy sub2api:8080
    encode gzip zstd

    header {
        -Server
        X-Content-Type-Options nosniff
        X-Frame-Options DENY
        Referrer-Policy strict-origin-when-cross-origin
    }

    log {
        output file /data/access.log {
            roll_size 50mb
            roll_keep 5
        }
    }
}
CEOF
  ok "Caddyfile created (domain: ${DOMAIN})"
fi

# ── Save credentials to file (600 perms) ────────────────────────────────────
cat > .credentials <<CRED
# ═══════════════════════════════════════════════════════════════════
# Sub2API Production Credentials
# Generated: $(date -u '+%Y-%m-%d %H:%M:%S UTC')
# ⚠️  KEEP THIS FILE SECURE — contains all production secrets
# ═══════════════════════════════════════════════════════════════════

# ── Access ──
DASHBOARD_URL=http://\${PUBLIC_IP:-localhost}:${SERVER_PORT}
API_ENDPOINT=http://\${PUBLIC_IP:-localhost}:${SERVER_PORT}/v1
SERVER_PORT=${SERVER_PORT}

# ── Admin Account ──
ADMIN_EMAIL=${ADMIN_EMAIL}
ADMIN_PASSWORD=${ADMIN_PASSWORD}

# ── PostgreSQL ──
POSTGRES_HOST=localhost
POSTGRES_PORT=5432  # internal only, not exposed
POSTGRES_USER=sub2api
POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
POSTGRES_DB=sub2api

# ── Redis ──
REDIS_HOST=localhost
REDIS_PORT=6379  # internal only, not exposed
REDIS_PASSWORD=${REDIS_PASSWORD}

# ── Security Keys ──
JWT_SECRET=${JWT_SECRET}
TOTP_ENCRYPTION_KEY=${TOTP_ENCRYPTION_KEY}

# ── Docker ──
IMAGE=weishaw/sub2api:${IMAGE_TAG}
INSTALL_DIR=${INSTALL_DIR}
CRED
chmod 600 .credentials
ok "Credentials saved to ${INSTALL_DIR}/.credentials (mode 600)"

# ── Step 4: Pull & Start ────────────────────────────────────────────────────
banner "Pulling Images & Starting Services"

docker compose pull --quiet 2>/dev/null || docker compose pull
docker compose up -d

# ── Step 5: Wait for healthy ─────────────────────────────────────────────────
banner "Waiting for Services to Be Ready"

MAX_WAIT=120
ELAPSED=0
INTERVAL=5

while [[ $ELAPSED -lt $MAX_WAIT ]]; do
  STATUS=$(docker inspect --format='{{.State.Health.Status}}' sub2api 2>/dev/null || echo "starting")
  if [[ "$STATUS" == "healthy" ]]; then
    ok "Sub2API is healthy!"
    break
  fi
  echo -ne "\r  ⏳ Waiting... (${ELAPSED}s / ${MAX_WAIT}s) — status: ${STATUS}  "
  sleep $INTERVAL
  ELAPSED=$((ELAPSED + INTERVAL))
done

echo ""

if [[ $ELAPSED -ge $MAX_WAIT ]]; then
  warn "Health check didn't pass within ${MAX_WAIT}s — service might still be starting."
  warn "Check logs: docker compose -f ${INSTALL_DIR}/docker-compose.yml logs -f sub2api"
fi

# ── Step 6: Detect public IP & region ────────────────────────────────────────
PUBLIC_IP=$(curl -s --connect-timeout 5 https://api.ipify.org 2>/dev/null \
         || curl -s --connect-timeout 5 https://ifconfig.me 2>/dev/null \
         || curl -s --connect-timeout 5 https://ipinfo.io/ip 2>/dev/null \
         || echo "<your-server-ip>")

# Try to detect server region for the delivery card
SERVER_REGION=$(curl -s --connect-timeout 3 "https://ipinfo.io/${PUBLIC_IP}/json" 2>/dev/null | grep -o '"city":"[^"]*"' | head -1 | cut -d'"' -f4)
SERVER_COUNTRY=$(curl -s --connect-timeout 3 "https://ipinfo.io/${PUBLIC_IP}/json" 2>/dev/null | grep -o '"country":"[^"]*"' | head -1 | cut -d'"' -f4)
SERVER_LOCATION="${SERVER_REGION:+${SERVER_REGION}, }${SERVER_COUNTRY:-Unknown}"

if [[ -n "$DOMAIN" ]]; then
  ACCESS_URL="https://${DOMAIN}"
elif [[ "$SERVER_PORT" == "80" ]]; then
  ACCESS_URL="http://${PUBLIC_IP}"
else
  ACCESS_URL="http://${PUBLIC_IP}:${SERVER_PORT}"
fi

API_BASE="${ACCESS_URL}/v1"

# ── Step 7: Delivery Card ────────────────────────────────────────────────────
DEPLOY_TIME=$(date -u '+%Y-%m-%d %H:%M UTC')

echo ""
echo -e "${C}╔═══════════════════════════════════════════════════════════════════╗${N}"
echo -e "${C}║${N}${W}              🚀 Sub2API — Deployment Complete                     ${N}${C}║${N}"
echo -e "${C}║${N}              ${B}${DEPLOY_TIME}${N}                                   ${C}║${N}"
echo -e "${C}╠═══════════════════════════════════════════════════════════════════╣${N}"
echo -e "${C}║${N}                                                                   ${C}║${N}"
echo -e "${C}║${N}  ${W}🌐 Access${N}                                                      ${C}║${N}"
echo -e "${C}║${N}  ───────────────────────────────────────────                      ${C}║${N}"
echo -e "${C}║${N}  Dashboard:     ${G}${ACCESS_URL}${N}"
echo -e "${C}║${N}  API Endpoint:  ${G}${API_BASE}${N}"
echo -e "${C}║${N}  Server IP:     ${PUBLIC_IP}  (${SERVER_LOCATION})"
echo -e "${C}║${N}                                                                   ${C}║${N}"
echo -e "${C}╠═══════════════════════════════════════════════════════════════════╣${N}"
echo -e "${C}║${N}  ${W}🔑 Admin Login${N}                                                  ${C}║${N}"
echo -e "${C}║${N}  ───────────────────────────────────────────                      ${C}║${N}"
echo -e "${C}║${N}  Email:          ${Y}${ADMIN_EMAIL}${N}"
echo -e "${C}║${N}  Password:       ${Y}${ADMIN_PASSWORD}${N}"
echo -e "${C}║${N}                                                                   ${C}║${N}"
echo -e "${C}╠═══════════════════════════════════════════════════════════════════╣${N}"
echo -e "${C}║${N}  ${W}🛡️  Infrastructure Secrets${N}                                      ${C}║${N}"
echo -e "${C}║${N}  ───────────────────────────────────────────                      ${C}║${N}"
echo -e "${C}║${N}  PostgreSQL:     sub2api / ${Y}${POSTGRES_PASSWORD:0:20}...${N}"
echo -e "${C}║${N}  Redis:          ${Y}${REDIS_PASSWORD:0:20}...${N}"
echo -e "${C}║${N}  JWT Secret:     ${Y}${JWT_SECRET:0:20}...${N}"
echo -e "${C}║${N}  TOTP Key:       ${Y}${TOTP_ENCRYPTION_KEY:0:20}...${N}"
echo -e "${C}║${N}                                                                   ${C}║${N}"
echo -e "${C}╠═══════════════════════════════════════════════════════════════════╣${N}"
echo -e "${C}║${N}  ${W}📂 File Locations${N}                                               ${C}║${N}"
echo -e "${C}║${N}  ───────────────────────────────────────────                      ${C}║${N}"
echo -e "${C}║${N}  Project:        ${INSTALL_DIR}/"
echo -e "${C}║${N}  Compose:        ${INSTALL_DIR}/docker-compose.yml"
echo -e "${C}║${N}  Credentials:    ${INSTALL_DIR}/.credentials  (mode 600)"
if [[ "$WITH_CADDY" == "true" ]]; then
echo -e "${C}║${N}  Caddyfile:      ${INSTALL_DIR}/Caddyfile"
fi
echo -e "${C}║${N}                                                                   ${C}║${N}"
echo -e "${C}╠═══════════════════════════════════════════════════════════════════╣${N}"
echo -e "${C}║${N}  ${W}⚙️  Server Specs${N}                                                 ${C}║${N}"
echo -e "${C}║${N}  ───────────────────────────────────────────                      ${C}║${N}"
echo -e "${C}║${N}  OS:             ${DISTRO} ${DISTRO_VERSION} (${ARCH})"
echo -e "${C}║${N}  RAM:            ${TOTAL_RAM_MB} MB"
echo -e "${C}║${N}  CPU:            ${CPU_CORES} cores"
echo -e "${C}║${N}  Kernel:         $(uname -r)"
echo -e "${C}║${N}  TCP CC:         BBR (optimized)"
echo -e "${C}║${N}                                                                   ${C}║${N}"
echo -e "${C}╠═══════════════════════════════════════════════════════════════════╣${N}"
echo -e "${C}║${N}  ${W}📋 Common Commands${N}                                              ${C}║${N}"
echo -e "${C}║${N}  ───────────────────────────────────────────                      ${C}║${N}"
echo -e "${C}║${N}  View logs:      cd ${INSTALL_DIR} && docker compose logs -f"
echo -e "${C}║${N}  Restart:        cd ${INSTALL_DIR} && docker compose restart"
echo -e "${C}║${N}  Stop:           cd ${INSTALL_DIR} && docker compose down"
echo -e "${C}║${N}  Update:         cd ${INSTALL_DIR} && docker compose pull && docker compose up -d"
echo -e "${C}║${N}  Status:         cd ${INSTALL_DIR} && docker compose ps"
echo -e "${C}║${N}                                                                   ${C}║${N}"
echo -e "${C}╚═══════════════════════════════════════════════════════════════════╝${N}"
echo ""
echo -e "${G}✓ Full credentials saved to: ${INSTALL_DIR}/.credentials${N}"
echo ""
echo -e "${W}Security Recommendations:${N}"
echo -e "  • Change SSH port and disable password login (use key-based auth)"
echo -e "  • Set up a domain with Caddy for HTTPS:  sudo bash quickdeploy.sh --domain your.domain"
echo -e "  • DB & Redis are internal-only (not exposed to public network)"
echo -e "  • Enable 2FA (TOTP) for admin account after first login"
echo -e "  • Credentials file is root-only readable (mode 600)"
echo ""
echo -e "${R}⚠  Please copy and securely save the credentials above.${N}"
echo -e "${R}   This is the only time they are displayed in full.${N}"
echo ""
