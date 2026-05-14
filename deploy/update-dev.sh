#!/usr/bin/env bash
# =============================================================================
# AISubAPI / Sub2API production app-only deploy script
#
# Usage:
#   DEPLOY_SSH_PASSWORD='***' bash deploy/update-dev.sh
#
# This script intentionally deploys only the sub2api application container:
#   - builds a new application image on the production host
#   - retags it to the compose image tag after build succeeds
#   - runs: docker compose up -d --no-deps --force-recreate sub2api
#
# It does not run docker compose down, and it does not restart PostgreSQL,
# Redis, Nginx, nginx-ui, or any other middleware.
# =============================================================================
set -euo pipefail

SERVER_HOST="${SERVER_HOST:-104.129.51.171}"
SERVER_PORT="${SERVER_PORT:-22080}"
SERVER_USER="${SERVER_USER:-root}"
BRANCH="${BRANCH:-dev-custom}"
REPO_URL="${REPO_URL:-https://github.com/lmk1010/AISubAPI.git}"
IMAGE_TAG="${IMAGE_TAG:-aisubapi:dev-custom}"
COMPOSE_DIR="${COMPOSE_DIR:-/opt/aisubapi}"
SERVICE_NAME="${SERVICE_NAME:-sub2api}"
TMP_DIR="${TMP_DIR:-/tmp/AISubAPI-deploy}"
DOCKERFILE_PATH="${DOCKERFILE_PATH:-deploy/Dockerfile}"
NO_CACHE="${NO_CACHE:-0}"
PUSH_FIRST="${PUSH_FIRST:-1}"
HEALTH_RETRIES="${HEALTH_RETRIES:-18}"
HEALTH_SLEEP_SECONDS="${HEALTH_SLEEP_SECONDS:-5}"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

info() { echo -e "${CYAN}[INFO]${NC} $*"; }
ok() { echo -e "${GREEN}[OK]${NC} $*"; }
warn() { echo -e "${YELLOW}[WARN]${NC} $*"; }
err() { echo -e "${RED}[ERROR]${NC} $*" >&2; exit 1; }

usage() {
  cat <<EOF
Usage:
  DEPLOY_SSH_PASSWORD='***' bash deploy/update-dev.sh [options]

Options:
  --host <host>          SSH host, default: ${SERVER_HOST}
  --port <port>          SSH port, default: ${SERVER_PORT}
  --user <user>          SSH user, default: ${SERVER_USER}
  --branch <branch>      Git branch, default: ${BRANCH}
  --compose-dir <path>   Remote compose dir, default: ${COMPOSE_DIR}
  --image-tag <tag>      Compose image tag, default: ${IMAGE_TAG}
  --service <name>       Compose service name, default: ${SERVICE_NAME}
  --no-cache             Build with docker --no-cache
  --no-push              Do not push local branch before remote build
  -h, --help             Show this help

Environment:
  DEPLOY_SSH_PASSWORD    Optional SSH password for sshpass. If unset, normal SSH auth is used.
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --host) SERVER_HOST="$2"; shift 2 ;;
    --port) SERVER_PORT="$2"; shift 2 ;;
    --user) SERVER_USER="$2"; shift 2 ;;
    --branch) BRANCH="$2"; shift 2 ;;
    --compose-dir) COMPOSE_DIR="$2"; shift 2 ;;
    --image-tag) IMAGE_TAG="$2"; shift 2 ;;
    --service) SERVICE_NAME="$2"; shift 2 ;;
    --no-cache) NO_CACHE=1; shift ;;
    --no-push) PUSH_FIRST=0; shift ;;
    -h|--help) usage; exit 0 ;;
    *) err "Unknown option: $1" ;;
  esac
done

SSH_TARGET="${SERVER_USER}@${SERVER_HOST}"
SSH_OPTS=(
  -o StrictHostKeyChecking=no
  -o ServerAliveInterval=15
  -o ServerAliveCountMax=4
  -p "${SERVER_PORT}"
)

if [[ -n "${DEPLOY_SSH_PASSWORD:-}" ]]; then
  command -v sshpass >/dev/null 2>&1 || err "sshpass is required when DEPLOY_SSH_PASSWORD is set"
  export SSHPASS="${DEPLOY_SSH_PASSWORD}"
  SSH_BIN=(sshpass -e ssh)
else
  SSH_BIN=(ssh)
fi

remote() {
  "${SSH_BIN[@]}" "${SSH_OPTS[@]}" "${SSH_TARGET}" "$@"
}

remote_bash() {
  remote "bash -lc $(printf '%q' "$1")"
}

if [[ "${PUSH_FIRST}" == "1" ]]; then
  info "Pushing local ${BRANCH} to origin..."
  git push origin "${BRANCH}"
  ok "Git push completed"
else
  warn "Skipping git push because --no-push was supplied"
fi

LOCAL_COMMIT="$(git rev-parse --short=12 "${BRANCH}")"
CANDIDATE_TAG="${IMAGE_TAG}-${LOCAL_COMMIT}"
BUILD_FLAGS=()
if [[ "${NO_CACHE}" == "1" ]]; then
  BUILD_FLAGS+=(--no-cache)
fi

info "Checking remote compose service..."
remote_bash "
  set -euo pipefail
  cd '${COMPOSE_DIR}'
  docker compose config --services | grep -qx '${SERVICE_NAME}'
"
ok "Remote compose service '${SERVICE_NAME}' exists"

info "Cloning ${BRANCH} on production host..."
remote_bash "
  set -euo pipefail
  rm -rf '${TMP_DIR}'
  git clone --depth 1 --branch '${BRANCH}' '${REPO_URL}' '${TMP_DIR}'
"
ok "Remote clone completed"

info "Building candidate image ${CANDIDATE_TAG}..."
remote_bash "
  set -euo pipefail
  docker build ${BUILD_FLAGS[*]} \
    --build-arg COMMIT='${LOCAL_COMMIT}' \
    -f '${TMP_DIR}/${DOCKERFILE_PATH}' \
    -t '${CANDIDATE_TAG}' \
    '${TMP_DIR}'
"
ok "Candidate image built"

info "Deploying only '${SERVICE_NAME}' (no middleware restart)..."
remote_bash "
  set -euo pipefail
  OLD_IMAGE_ID=\$(docker image inspect -f '{{.Id}}' '${IMAGE_TAG}' 2>/dev/null || true)
  docker image tag '${CANDIDATE_TAG}' '${IMAGE_TAG}'
  cd '${COMPOSE_DIR}'
  docker compose up -d --no-deps --force-recreate '${SERVICE_NAME}'
  echo \"\${OLD_IMAGE_ID}\" > /tmp/${SERVICE_NAME}-previous-image-id
"
ok "Application container recreated"

info "Waiting for health check..."
for i in $(seq 1 "${HEALTH_RETRIES}"); do
  sleep "${HEALTH_SLEEP_SECONDS}"
  STATUS="$(remote_bash "docker inspect '${SERVICE_NAME}' --format '{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}' 2>/dev/null" || echo "unknown")"
  if [[ "${STATUS}" == "healthy" || "${STATUS}" == "running" ]]; then
    ok "${SERVICE_NAME} is ${STATUS}"
    remote_bash "
      docker ps --filter 'name=^/${SERVICE_NAME}$' --format 'table {{.Names}}\t{{.Image}}\t{{.Status}}'
      cd '${COMPOSE_DIR}' && docker compose ps '${SERVICE_NAME}'
      rm -rf '${TMP_DIR}'
    "
    ok "Deploy completed. PostgreSQL/Redis/Nginx were not restarted."
    exit 0
  fi
  info "Waiting... (${i}/${HEALTH_RETRIES}) status=${STATUS}"
done

warn "Health check timed out; attempting rollback to previous ${IMAGE_TAG}"
remote_bash "
  set -euo pipefail
  OLD_IMAGE_ID=\$(cat /tmp/${SERVICE_NAME}-previous-image-id 2>/dev/null || true)
  if [[ -n \"\${OLD_IMAGE_ID}\" ]]; then
    docker image tag \"\${OLD_IMAGE_ID}\" '${IMAGE_TAG}'
    cd '${COMPOSE_DIR}'
    docker compose up -d --no-deps --force-recreate '${SERVICE_NAME}'
  fi
"
err "Deploy failed and rollback was attempted. Check: ssh -p ${SERVER_PORT} ${SSH_TARGET} 'docker logs ${SERVICE_NAME} --tail=200'"
