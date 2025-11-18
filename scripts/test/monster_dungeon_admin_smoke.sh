#!/usr/bin/env bash
set -euo pipefail

# Basic configuration
BASE_URL=${ADMIN_BASE_URL:-http://localhost/api/v1}
USERNAME=${ADMIN_USERNAME:-root}
PASSWORD=${ADMIN_PASSWORD:-password}
MONSTER_ID=${ADMIN_MONSTER_ID:-}
DUNGEON_ID=${ADMIN_DUNGEON_ID:-}
ROOM_SEQUENCE=${ADMIN_DUNGEON_ROOMS:-}

if [[ -z "${MONSTER_ID}" ]]; then
  echo "请通过 ADMIN_MONSTER_ID 指定用于验证的怪物 ID" >&2
  exit 1
fi

if [[ -z "${DUNGEON_ID}" ]]; then
  echo "请通过 ADMIN_DUNGEON_ID 指定用于验证的地城 ID" >&2
  exit 1
fi

IFS=',' read -r -a ROOM_IDS <<< "${ROOM_SEQUENCE}"
if [[ ${#ROOM_IDS[@]} -lt 2 ]]; then
  cat <<'EOF' >&2
请在 ADMIN_DUNGEON_ROOMS 中至少提供两个房间 ID，格式例如:
  export ADMIN_DUNGEON_ROOMS="room-id-1,room-id-2"
EOF
  exit 1
fi

login_payload=$(cat <<EOF
{"identifier":"${USERNAME}","password":"${PASSWORD}"}
EOF
)

echo ">>> 登录 ${BASE_URL}"
login_resp=$(curl --noproxy '*' -s -H 'Content-Type: application/json' -d "${login_payload}" "${BASE_URL}/admin/auth/login")
TOKEN=$(python3 - <<'PY' "${login_resp}"
import json, sys
data = json.loads(sys.argv[1])
print(data["data"]["session_token"])
PY
)
AUTH_HEADER=("Authorization: Bearer ${TOKEN}")

request() {
  local method=$1
  local path=$2
  local expect_status=$3
  local body=${4-}

  local tmp
  tmp=$(mktemp)
  if [[ -n "${body}" ]]; then
    status=$(curl --noproxy '*' -s -o "${tmp}" -w '%{http_code}' -X "${method}" -H "${AUTH_HEADER[0]}" -H 'Content-Type: application/json' -d "${body}" "${BASE_URL}${path}")
  else
    status=$(curl --noproxy '*' -s -o "${tmp}" -w '%{http_code}' -X "${method}" -H "${AUTH_HEADER[0]}" "${BASE_URL}${path}")
  fi

  if [[ "${status}" != "${expect_status}" ]]; then
    echo "请求 ${method} ${path} 期望状态码 ${expect_status}，实际 ${status}" >&2
    cat "${tmp}" >&2
    exit 1
  fi

  cat "${tmp}"
  rm -f "${tmp}"
}

echo ">>> 校验怪物详情聚合"
monster_detail=$(request GET "/admin/monsters/${MONSTER_ID}" 200)
python3 - <<'PY' "${monster_detail}"
import json, sys
data = json.loads(sys.argv[1])
detail = data["data"]
for field in ("skills", "drops", "tags"):
    if field not in detail:
        raise SystemExit(f"怪物详情缺少 {field} 字段")
print("怪物详情包含聚合字段")
PY

echo ">>> 校验怪物列表聚合字段"
monster_list=$(request GET "/admin/monsters?limit=50" 200)
python3 - <<'PY' "${monster_list}" "${MONSTER_ID}"
import json, sys
payload = json.loads(sys.argv[1])
target = sys.argv[2]
for monster in payload["data"]["list"]:
    if monster["id"] == target:
        for field in ("skills", "drops", "tags"):
            if field not in monster:
                raise SystemExit(f"怪物列表缺少 {field} 字段")
        print("怪物列表包含聚合字段")
        break
else:
    print("提示: 目标怪物未出现在当前分页结果中", file=sys.stderr)
PY

echo ">>> 验证 order_by 白名单"
request GET "/admin/monsters?order_by=hack" 400 >/dev/null

echo ">>> 验证 JSON 字段校验信息"
invalid_json='{"damage_resistances":{"FIRE_RESIST":1.5}}'
invalid_resp=$(request PUT "/admin/monsters/${MONSTER_ID}" 400 "${invalid_json}")
python3 - <<'PY' "${invalid_resp}"
import json, sys
resp = json.loads(sys.argv[1])
msg = resp.get("message", "")
if "抗性" not in msg:
    raise SystemExit(f"期望看到 \"抗性\" 错误提示，但实际为: {msg}")
print("JSON 校验错误提示:", msg)
PY

echo ">>> 验证地城房间排序校验"
invalid_sequence=$(cat <<EOF
{"room_sequence":[{"room_id":"${ROOM_IDS[0]}","sort":1},{"room_id":"${ROOM_IDS[1]}","sort":3}]}
EOF
)
sequence_resp=$(request PUT "/admin/dungeons/${DUNGEON_ID}" 400 "${invalid_sequence}")
python3 - <<'PY' "${sequence_resp}"
import json, sys
resp = json.loads(sys.argv[1])
msg = resp.get("message", "")
if "房间" not in msg and "排序" not in msg:
    raise SystemExit(f"期望看到房间排序错误提示，但实际为: {msg}")
print("房间校验错误提示:", msg)
PY

echo ">>> 所有冒烟场景通过"
