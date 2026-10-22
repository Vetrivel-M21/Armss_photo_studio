#!/usr/bin/env bash
# ==============================================================================
# ARMSS Photo Studio -- Production Docker Deployment Script for Ubuntu Server
# ==============================================================================
set -euo pipefail

echo "========================================================"
echo "    ARMSS Photo Studio Docker Production Deployment     "
echo "========================================================"

# -- 1. Check Docker Installation ---------------------------------------------
if ! command -v docker &> /dev/null; then
    echo "[!] Docker not found. Installing Docker Engine..."
    sudo apt-get update -qq
    sudo apt-get install -y ca-certificates curl gnupg lsb-release
    sudo mkdir -p /etc/apt/keyrings
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
    echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
    sudo apt-get update -qq
    sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin
    sudo systemctl enable --now docker
    sudo usermod -aG docker "$USER"
    echo "[?] Docker installed successfully."
fi

# -- 2. Verify docker compose plugin ------------------------------------------
if ! docker compose version &> /dev/null; then
    echo "[!] 'docker compose' plugin not found. Installing..."
    sudo apt-get install -y docker-compose-plugin
fi
echo "[?] $(docker compose version)"

# -- 3. Check / generate .env file --------------------------------------------
if [ ! -f ".env" ]; then
    echo "[i] .env file not found. Creating from .env.example..."
    cp .env.example .env

    # Auto-generate a secure random session secret
    RANDOM_SESSION=$(head -c 48 /dev/urandom | base64 | tr -dc 'a-zA-Z0-9' | head -c 48)
    sed -i "s/change-me-to-a-long-random-string-2026/${RANDOM_SESSION}/" .env

    echo "[?] .env generated with a secure random SESSION_SECRET."
    echo ""
    echo "  +--------------------------------------------------------------+"
    echo "  ¦  IMPORTANT: Before going live, update these in .env:         ¦"
    echo "  ¦   * MYSQL_PASSWORD      -- strong DB password                 ¦"
    echo "  ¦   * MINIO_ROOT_PASSWORD -- strong MinIO password              ¦"
    echo "  ¦   * DOMAIN              -- your actual domain name            ¦"
    echo "  ¦   * PUBLIC_BASE_URL     -- full public URL                    ¦"
    echo "  ¦   * ANTHROPIC_API_KEY or GEMINI_API_KEY (for AI features)    ¦"
    echo "  ¦   * FLUX_API_KEY        (for image generation)               ¦"
    echo "  +--------------------------------------------------------------+"
    echo ""
else
    echo "[?] .env file found."
fi

# -- 4. Read gateway port ------------------------------------------------------
GATEWAY_PORT=$(grep -E '^GATEWAY_PORT=' .env | cut -d '=' -f2 | tr -d ' ') || true
if [ -z "${GATEWAY_PORT:-}" ]; then
    GATEWAY_PORT="8070"
fi
echo "[i] Gateway Port: $GATEWAY_PORT"

# -- 5. Check port availability ------------------------------------------------
if ss -tlnp 2>/dev/null | grep -q ":${GATEWAY_PORT} "; then
    echo "[?] Warning: Port ${GATEWAY_PORT} appears to be in use."
    echo "    Change GATEWAY_PORT in .env to a free port and re-run."
fi

# -- 6. Tear down old containers, rebuild and start ---------------------------
echo "[i] Stopping any existing containers..."
docker compose down --remove-orphans || true

echo "[i] Building images (--no-cache to pick up all code changes)..."
docker compose build --no-cache

echo "[i] Starting all containers..."
docker compose up -d

# -- 7. Wait for MySQL to be healthy ------------------------------------------
echo ""
echo "[i] Waiting for MySQL to become healthy..."
MAX_WAIT=90
ELAPSED=0
INTERVAL=5

until docker compose ps mysql 2>/dev/null | grep -q "healthy"; do
    if [ "$ELAPSED" -ge "$MAX_WAIT" ]; then
        echo ""
        echo "[?] MySQL did not become healthy within ${MAX_WAIT}s."
        docker compose logs --tail=30 mysql
        exit 1
    fi
    printf "."
    sleep "$INTERVAL"
    ELAPSED=$((ELAPSED + INTERVAL))
done
echo ""
echo "[?] MySQL healthy after ${ELAPSED}s."

# -- 8. Wait for MinIO to be healthy ------------------------------------------
echo "[i] Waiting for MinIO to become healthy..."
ELAPSED=0

until docker compose ps minio 2>/dev/null | grep -q "healthy"; do
    if [ "$ELAPSED" -ge 60 ]; then
        echo ""
        echo "[?] MinIO did not become healthy within 60s."
        docker compose logs --tail=30 minio
        exit 1
    fi
    printf "."
    sleep "$INTERVAL"
    ELAPSED=$((ELAPSED + INTERVAL))
done
echo ""
echo "[?] MinIO healthy after ${ELAPSED}s."

# -- 9. Wait for backend healthcheck to pass -----------------------------------
echo "[i] Waiting for backend to become healthy..."
MAX_WAIT=120
ELAPSED=0

until docker compose ps backend 2>/dev/null | grep -q "healthy"; do
    if [ "$ELAPSED" -ge "$MAX_WAIT" ]; then
        echo ""
        echo "[?] Backend did not become healthy within ${MAX_WAIT}s."
        echo "    Check logs with: docker compose logs backend"
        docker compose logs --tail=60 backend
        exit 1
    fi
    printf "."
    sleep "$INTERVAL"
    ELAPSED=$((ELAPSED + INTERVAL))
done
echo ""
echo "[?] Backend healthy after ${ELAPSED}s."

# -- 10. Wait for frontend to start --------------------------------------------
echo "[i] Waiting for frontend container..."
ELAPSED=0
until docker compose ps frontend 2>/dev/null | grep -qE "running|Up"; do
    if [ "$ELAPSED" -ge 60 ]; then
        echo "[?] Frontend did not start within 60s."
        docker compose logs --tail=40 frontend
        exit 1
    fi
    printf "."
    sleep "$INTERVAL"
    ELAPSED=$((ELAPSED + INTERVAL))
done
echo ""
echo "[?] Frontend running."

# -- 11. Smoke test: backend /healthz directly --------------------------------
echo "[i] Smoke-testing backend /healthz..."
if docker compose exec -T backend wget -qO- "http://127.0.0.1:8080/healthz" 2>/dev/null | grep -qi '"status"'; then
    echo "[?] Backend /healthz: OK"
else
    echo "[?] Backend /healthz did not return expected JSON. Check: docker compose logs backend"
fi

# -- 12. Smoke test: backend /readyz (MySQL + MinIO check) ---------------------
echo "[i] Smoke-testing backend /readyz (DB + storage)..."
READYZ_RESPONSE=$(docker compose exec -T backend wget -qO- "http://127.0.0.1:8080/readyz" 2>/dev/null || echo "")
if echo "$READYZ_RESPONSE" | grep -qi '"status"'; then
    echo "[?] Backend /readyz: OK -- MySQL and object store reachable."
else
    echo "[?] Backend /readyz check failed. Response: ${READYZ_RESPONSE}"
    echo "    Check: docker compose logs backend"
fi

# -- 13. Smoke test: gateway /healthz through nginx ----------------------------
echo "[i] Smoke-testing /healthz through nginx gateway..."
if curl -sf "http://localhost:${GATEWAY_PORT}/healthz" | grep -qi '"status"'; then
    echo "[?] Gateway /healthz: OK"
else
    echo "[?] Gateway /healthz check failed. Check: docker compose logs nginx"
fi

# -- 14. Final status ----------------------------------------------------------
echo ""
echo "========================================================"
echo "               Container Status                         "
echo "========================================================"
docker compose ps

echo ""
echo "========================================================"
echo "      ARMSS Photo Studio Deployment Successful!         "
echo "========================================================"
SERVER_IP=$(hostname -I | awk '{print $1}')
echo ""
echo "  ? App (public):      http://${SERVER_IP}:${GATEWAY_PORT}"
echo "  ? Liveness probe:    http://${SERVER_IP}:${GATEWAY_PORT}/healthz"
echo "  ? Readiness probe:   http://${SERVER_IP}:${GATEWAY_PORT}/readyz"
echo "  ? MinIO S3 console:  http://${SERVER_IP}:${GATEWAY_PORT}/minio-console/"
echo ""
echo "  ? Health endpoints explained:"
echo "    /healthz  -- liveness:  process is up (no DB/storage check)"
echo "    /readyz   -- readiness: MySQL AND object store are reachable"
echo ""
echo "  ? Daily maintenance commands:"
echo "    docker compose logs -f               # stream all logs"
echo "    docker compose logs -f backend       # backend only"
echo "    docker compose logs -f frontend      # frontend only"
echo "    docker compose logs -f nginx         # gateway only"
echo "    docker compose logs -f minio         # MinIO storage"
echo "    docker compose restart               # restart all services"
echo "    docker compose down                  # stop all"
echo "    docker compose up -d --build backend # rebuild backend only"
echo "    docker compose up -d --build frontend # rebuild frontend only"
echo "    git pull && docker compose up -d --build # full update"
echo "========================================================"
