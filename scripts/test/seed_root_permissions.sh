#!/usr/bin/env bash
set -euo pipefail

# 为本地/测试环境补种 root 管理员的 Keto 权限，避免因权限缺失导致 admin API 403。
# 依赖：docker、正在运行的 tsu_keto_service 容器。

KETO_CONTAINER="${KETO_CONTAINER:-tsu_keto_service}"
ADMIN_USER_ID="${ADMIN_USER_ID:-daf99445-61cc-4b24-9973-17eb79a53318}"
READ_REMOTE="${READ_REMOTE:-tsu_keto_service:4466}"
WRITE_REMOTE="${WRITE_REMOTE:-tsu_keto_service:4467}"

if ! docker ps --format '{{.Names}}' | grep -qx "$KETO_CONTAINER"; then
  echo "Keto container \"$KETO_CONTAINER\" not running" >&2
  exit 1
fi

TMP_FILE="/tmp/root_perms.json"
docker exec -i "$KETO_CONTAINER" sh -c "cat > $TMP_FILE" <<EOF
[
  {"namespace":"roles","object":"admin","relation":"member","subject_id":"users:${ADMIN_USER_ID}"},
  {"namespace":"permissions","object":"user:read","relation":"granted","subject_id":"users:${ADMIN_USER_ID}"},
  {"namespace":"permissions","object":"user:create","relation":"granted","subject_id":"users:${ADMIN_USER_ID}"},
  {"namespace":"permissions","object":"user:update","relation":"granted","subject_id":"users:${ADMIN_USER_ID}"},
  {"namespace":"permissions","object":"user:delete","relation":"granted","subject_id":"users:${ADMIN_USER_ID}"},
  {"namespace":"permissions","object":"user:ban","relation":"granted","subject_id":"users:${ADMIN_USER_ID}"},
  {"namespace":"permissions","object":"role:read","relation":"granted","subject_id":"users:${ADMIN_USER_ID}"},
  {"namespace":"permissions","object":"role:create","relation":"granted","subject_id":"users:${ADMIN_USER_ID}"},
  {"namespace":"permissions","object":"role:update","relation":"granted","subject_id":"users:${ADMIN_USER_ID}"},
  {"namespace":"permissions","object":"role:delete","relation":"granted","subject_id":"users:${ADMIN_USER_ID}"},
  {"namespace":"permissions","object":"role:assign","relation":"granted","subject_id":"users:${ADMIN_USER_ID}"},
  {"namespace":"permissions","object":"permission:read","relation":"granted","subject_id":"users:${ADMIN_USER_ID}"},
  {"namespace":"permissions","object":"permission:assign","relation":"granted","subject_id":"users:${ADMIN_USER_ID}"},
  {"namespace":"permissions","object":"permission:grant_user","relation":"granted","subject_id":"users:${ADMIN_USER_ID}"},
  {"namespace":"permissions","object":"system:config","relation":"granted","subject_id":"users:${ADMIN_USER_ID}"},
  {"namespace":"permissions","object":"system:monitor","relation":"granted","subject_id":"users:${ADMIN_USER_ID}"},
  {"namespace":"permissions","object":"hero:manage","relation":"granted","subject_id":"users:${ADMIN_USER_ID}"},
  {"namespace":"permissions","object":"skill:manage","relation":"granted","subject_id":"users:${ADMIN_USER_ID}"},
  {"namespace":"permissions","object":"class:manage","relation":"granted","subject_id":"users:${ADMIN_USER_ID}"},
  {"namespace":"permissions","object":"team:read","relation":"granted","subject_id":"users:${ADMIN_USER_ID}"},
  {"namespace":"permissions","object":"team:moderate","relation":"granted","subject_id":"users:${ADMIN_USER_ID}"},
  {"namespace":"permissions","object":"tools:grant_item","relation":"granted","subject_id":"users:${ADMIN_USER_ID}"},
  {"namespace":"permissions","object":"world-drop:manage-items","relation":"granted","subject_id":"users:${ADMIN_USER_ID}"}
]
EOF

docker exec "$KETO_CONTAINER" keto relation-tuple create "$TMP_FILE" \
  --insecure-disable-transport-security --write-remote "$WRITE_REMOTE" >/dev/null

echo "✅ seeded permissions for user ${ADMIN_USER_ID} into Keto (${KETO_CONTAINER})"
