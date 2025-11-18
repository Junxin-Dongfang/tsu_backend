#!/bin/bash

# Lightweight performance smoke test for team APIs.
# Measures: team creation, permission-protected GET, warehouse query.

set -euo pipefail

BASE_URL="${GAME_API:-http://localhost/api/v1/game}"
EMAIL="${TEAM_PERF_EMAIL:-root@example.com}"
PASSWORD="${TEAM_PERF_PASSWORD:-password}"
ITERATIONS="${TEAM_PERF_ITERS:-5}"

require_cmd() {
  command -v "$1" >/dev/null || { echo "Missing dependency: $1" >&2; exit 1; }
}

require_cmd curl
require_cmd jq

log() {
  echo "[$(date +%H:%M:%S)] $1"
}

login() {
  local resp
  resp=$(curl -s -X POST "$BASE_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"identifier":"'"$EMAIL"'","password":"'"$PASSWORD"'"}')
  SESSION_TOKEN=$(echo "$resp" | jq -r '.data.session_token')
  USER_ID=$(echo "$resp" | jq -r '.data.user_id')
  if [[ -z "$SESSION_TOKEN" || "$SESSION_TOKEN" == "null" ]]; then
    echo "Failed to login as $EMAIL" >&2
    exit 1
  fi
  log "Logged in as $EMAIL ($USER_ID)"
}

fetch_basic_class() {
  local resp
  resp=$(curl -s -H "Authorization: Bearer $SESSION_TOKEN" "$BASE_URL/classes/basic")
  CLASS_ID=$(echo "$resp" | jq -r '.data[0].id')
  if [[ -z "$CLASS_ID" || "$CLASS_ID" == "null" ]]; then
    echo "Failed to fetch basic class" >&2
    exit 1
  fi
}

create_hero() {
  local name="$1"
  local resp
  resp=$(curl -s -X POST "$BASE_URL/heroes" \
    -H "Authorization: Bearer $SESSION_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"hero_name":"'"$name"'","class_id":"'"$CLASS_ID"'"}')
  HERO_ID=$(echo "$resp" | jq -r '.data.id')
  if [[ -z "$HERO_ID" || "$HERO_ID" == "null" ]]; then
    echo "Failed to create hero" >&2
    echo "$resp"
    exit 1
  fi
  log "Created hero $HERO_ID ($name)"
}

create_team_base() {
  local resp
  resp=$(curl -s -X POST "$BASE_URL/teams" \
    -H "Authorization: Bearer $SESSION_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"hero_id":"'"$HERO_ID"'","team_name":"PerfBase","description":"Perf baseline team"}')
  BASE_TEAM_ID=$(echo "$resp" | jq -r '.data.id')
  if [[ -z "$BASE_TEAM_ID" || "$BASE_TEAM_ID" == "null" ]]; then
    echo "Failed to create baseline team" >&2
    exit 1
  fi
  log "Baseline team: $BASE_TEAM_ID"
}

disband_team() {
  local team_id="$1"
  curl -s -X POST "$BASE_URL/teams/$team_id/disband?hero_id=$HERO_ID" \
    -H "Authorization: Bearer $SESSION_TOKEN" >/dev/null
}

measure_create_disband() {
  log "Team creation test ($ITERATIONS iterations)"
  local total_ms=0
  for i in $(seq 1 "$ITERATIONS"); do
    local tmp
    tmp=$(mktemp)
    local duration
    duration=$(curl -s -o "$tmp" -w "%{time_total}" -X POST "$BASE_URL/teams" \
      -H "Authorization: Bearer $SESSION_TOKEN" \
      -H "Content-Type: application/json" \
      -d '{"hero_id":"'"$HERO_ID"'","team_name":"Perf_'$i'_'$(date +%s)'","description":"perf"}')
    local team_id
    team_id=$(cat "$tmp" | jq -r '.data.id')
    rm -f "$tmp"
    [[ -n "$team_id" && "$team_id" != "null" ]] || continue
    disband_team "$team_id"
    total_ms=$(awk "BEGIN{print $total_ms + ($duration * 1000)}")
  done
  local avg
  avg=$(awk "BEGIN{printf \"%.2f\", $total_ms / $ITERATIONS}")
  log "Average creation latency: ${avg} ms"
}

measure_team_get() {
  log "Permission-protected GET /teams/{id}"
  local total_ms=0
  for _ in $(seq 1 "$ITERATIONS"); do
    local duration
    duration=$(curl -s -o /dev/null -w "%{time_total}" \
      -H "Authorization: Bearer $SESSION_TOKEN" \
      "$BASE_URL/teams/$BASE_TEAM_ID?hero_id=$HERO_ID")
    total_ms=$(awk "BEGIN{print $total_ms + ($duration * 1000)}")
  done
  local avg
  avg=$(awk "BEGIN{printf \"%.2f\", $total_ms / $ITERATIONS}")
  log "Average GET /teams latency: ${avg} ms"
}

measure_warehouse_get() {
  log "GET /teams/{id}/warehouse"
  local total_ms=0
  for _ in $(seq 1 "$ITERATIONS"); do
    local duration
    duration=$(curl -s -o /dev/null -w "%{time_total}" \
      -H "Authorization: Bearer $SESSION_TOKEN" \
      "$BASE_URL/teams/$BASE_TEAM_ID/warehouse?hero_id=$HERO_ID")
    total_ms=$(awk "BEGIN{print $total_ms + ($duration * 1000)}")
  done
  local avg
  avg=$(awk "BEGIN{printf \"%.2f\", $total_ms / $ITERATIONS}")
  log "Average warehouse latency: ${avg} ms"
}

cleanup() {
  disband_team "$BASE_TEAM_ID"
  log "Cleanup complete"
}

login
fetch_basic_class
create_hero "PerfHero_$(date +%s)"
create_team_base

measure_create_disband
measure_team_get
measure_warehouse_get

cleanup
