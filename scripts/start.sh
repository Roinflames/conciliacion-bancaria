#!/bin/bash
set -e

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
BACKEND="$ROOT/backend"

export APP_PROFILE="${APP_PROFILE:-juridico}"
export PORT="${PORT:-3032}"
export JWT_SECRET="${JWT_SECRET:-dev-secret-change-in-production}"
export DB_PATH="${DB_PATH:-$BACKEND/conciliacion.db}"

source ~/.env_discord 2>/dev/null || true

notify_discord() {
  [ -z "$DISCORD_WEBHOOK_URL" ] && return
  curl -s -X POST "$DISCORD_WEBHOOK_URL" \
    -H "Content-Type: application/json" \
    -d "{\"content\": \"$1\"}" > /dev/null
}

echo "Perfil: $APP_PROFILE | Puerto: $PORT"

cd "$BACKEND"

if go build -o conciliacion-bancaria .; then
  notify_discord "🚀 **conciliacion-bancaria** [$APP_PROFILE] desplegado en http://localhost:$PORT"
  ./conciliacion-bancaria
else
  notify_discord "❌ **conciliacion-bancaria** [$APP_PROFILE] falló al compilar"
  exit 1
fi
