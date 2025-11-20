#!/usr/bin/env bash
set -euo pipefail

SERVICE=${SERVICE:-admin}
BASE_URL=${BASE_URL:-http://localhost:80}
ITERATIONS=${ITERATIONS:-20}
THRESHOLD_MS=${THRESHOLD_MS:-250}
PROM_FILE=${PROM_FILE:-tmp/metrics/login_benchmark.prom}
TMP_DIR=$(dirname "$PROM_FILE")
mkdir -p "$TMP_DIR"

case "$SERVICE" in
  admin)
    IDENTIFIER=${ADMIN_USERNAME:-root}
    PASSWORD=${ADMIN_PASSWORD:-password}
    LOGIN_PATH="/api/v1/admin/auth/login"
    ;;
  game)
    IDENTIFIER=${GAME_USERNAME:-}
    PASSWORD=${GAME_PASSWORD:-}
    LOGIN_PATH="/api/v1/game/auth/login"
    ;;
  *)
    echo "[login-benchmark] Unsupported SERVICE=$SERVICE (use admin|game)" >&2
    exit 1
    ;;
 esac

if [[ -z "$IDENTIFIER" || -z "$PASSWORD" ]]; then
  echo "[login-benchmark] Missing credentials for SERVICE=$SERVICE" >&2
  exit 1
fi

payload() {
  printf '{"identifier":"%s","password":"%s"}' "$IDENTIFIER" "$PASSWORD"
}

login_once() {
  local resp
  resp=$(curl -sS -X POST "$BASE_URL$LOGIN_PATH" \
    -H 'Content-Type: application/json' \
    --data "$(payload)")
  python3 - <<'PY' "$resp"
import json, sys
try:
    data=json.loads(sys.argv[1])
    token=data['data']['session_token']
    if not token:
        raise ValueError('empty token')
    print(token)
except Exception as exc:
    raise SystemExit(f"Failed to parse login response: {exc}: {sys.argv[1][:200]}")
PY
}

SESSION_TOKEN=$(login_once)

benchmark_login() {
  curl -sS -o /dev/null -w '%{time_total}' -X POST "$BASE_URL$LOGIN_PATH" \
    -H 'Content-Type: application/json' \
    -H "Authorization: Bearer $SESSION_TOKEN" \
    -H "X-Session-Token: $SESSION_TOKEN" \
    -H "Cookie: ory_kratos_session=$SESSION_TOKEN" \
    --data "$(payload)"
}

DURATIONS=()
for ((i=1;i<=ITERATIONS;i++)); do
  DURATIONS+=("$(benchmark_login)")
done

STATS=$(python3 - "${DURATIONS[@]}" <<'PY'
import sys
values=[float(arg)*1000 for arg in sys.argv[1:] if arg]
if not values:
    raise SystemExit('no samples collected')
values.sort()
def quantile(data, q):
    pos=(len(data)-1)*q
    low=int(pos)
    high=min(low+1, len(data)-1)
    weight=pos-low
    return data[low]*(1-weight)+data[high]*weight
p50=quantile(values,0.5)
p95=quantile(values,0.95)
p99=quantile(values,0.99)
avg=sum(values)/len(values)
print(f"{p50:.2f} {p95:.2f} {p99:.2f} {avg:.2f}")
PY
)
read -r P50 P95 P99 AVG <<<"$STATS"

cat <<'EOM'
============================================
Login Benchmark
============================================
EOM
printf "Service       : %s\n" "$SERVICE"
printf "Base URL      : %s\n" "$BASE_URL"
printf "Iterations   : %s\n" "$ITERATIONS"
printf "p50 (ms)     : %s\n" "$P50"
printf "p95 (ms)     : %s\n" "$P95"
printf "p99 (ms)     : %s\n" "$P99"
printf "avg (ms)     : %s\n" "$AVG"
printf "Threshold(ms): %s\n" "$THRESHOLD_MS"

awk -v svc="$SERVICE" -v p50="$P50" -v p95="$P95" -v p99="$P99" -v avg="$AVG" -v runs="$ITERATIONS" -v thr="$THRESHOLD_MS" 'BEGIN {
  printf("tsu_login_benchmark_duration_milliseconds{service=\"%s\",quantile=\"p50\"} %s\n", svc, p50);
  printf("tsu_login_benchmark_duration_milliseconds{service=\"%s\",quantile=\"p95\"} %s\n", svc, p95);
  printf("tsu_login_benchmark_duration_milliseconds{service=\"%s\",quantile=\"p99\"} %s\n", svc, p99);
  printf("tsu_login_benchmark_duration_milliseconds{service=\"%s\",quantile=\"avg\"} %s\n", svc, avg);
  printf("tsu_login_benchmark_runs_total{service=\"%s\"} %s\n", svc, runs);
  printf("tsu_login_benchmark_threshold_milliseconds{service=\"%s\"} %s\n", svc, thr);
}' > "$PROM_FILE"

echo "[login-benchmark] metrics written to $PROM_FILE"

awk -v p95="$P95" -v thr="$THRESHOLD_MS" 'BEGIN { if (p95+0 > thr+0) exit 1 }'
