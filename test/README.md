# Test Directory Overview

This directory只保留当前维护的内容：

- `integration/`：Go 集成/回归测试（包含新的 `smoke` 冒烟包）。
- `internal/apitest/`：配置加载、HTTP 客户端、DTO、数据隔离与夹具工厂。
- `fixtures/`：静态测试资源（例如 `sample-upload.txt` 占位上传文件）。
- `tools/export-swagger-matrix/`：将 `docs/admin` + `docs/game` Swagger 展平成 CSV。
- `matrix/`：已生成的端点覆盖矩阵（`make test-matrix` 自动写入）。

## Running Tests

```bash
# Local developer smoke
 BASE_URL=http://localhost:80 ADMIN_USERNAME=root ADMIN_PASSWORD=admin make test-smoke

 # Plain go test (no JUnit)
 BASE_URL=http://localhost:80 go test ./test/integration/...
```

`make test-smoke` 会自动安装 `gotestsum` 并输出 JUnit（默认 `test/results/junit/api-smoke.xml`）。

## Environment Files

Sample env files live in `configs/testing/`:

- `api-tests.local.env`：本地开发默认值。
- `api-tests.testprod.env.example`：测试/生产共享模板，由 CI 注入真实 HOST/Secret。

关键变量：`BASE_URL`、`ADMIN_USERNAME/ADMIN_PASSWORD`、`PLAYER_USERNAME_PREFIX/PLAYER_EMAIL_SUFFIX`、`GAME_USERNAME/GAME_PASSWORD`（可选，提供已有玩家账号则跳过注册）、`SMOKE_JUNIT_FILE`。

## E2E 测试套件

- `integration/e2e/player_flows_test.go`：
  1. 注册→登录→查职业→创建英雄→查看英雄。
  2. 创建团队→更新团队信息。
  3. 查看团队仓库→解散团队→验证访问受限。
  4. 属性加点失败分支（经验不足）。
- `integration/e2e/admin_flows_test.go`：管理员登录→用户列表→用户详情。

运行全部 e2e：

```bash
BASE_URL=http://localhost:80 ADMIN_USERNAME=root ADMIN_PASSWORD=admin go test ./test/integration/e2e -count=1
```

## Coverage Matrix

```
make test-matrix
cat test/matrix/swagger_matrix.csv | head
```

The CSV contains `service,method,path,summary,tags,auth_required,notes` columns。运行 `make test-matrix` 会保留既有 notes 并按最新 Swagger（当前 291 个 operation）更新。

## 模块化回归

目录 `test/integration/modules/` 用于针对具体模块的正/负例验证。例如：
- `heroes_test.go`：未认证加点返回 401、非法属性代码触发 `CodeResourceNotFound`。
- `team_members_test.go`：未登录创建团队、成员邀请权限、队长查看仓库。
- `team_join_test.go`：未认证申请 401、审批不存在请求 404。
- `team_invitation_test.go`：伪造 invitation 接受 404、普通成员审批 403。
- `team_dungeon_test.go`：未认证/成员选择副本失败。
- `team_leave_test.go`：校验队长无法离队。
- `team_warehouse_test.go`：仓库分金币的未认证/成员权限限制。
- `admin_permissions_test.go` / `admin_user_detail_test.go`：后台权限、用户详情的 auth/unauth 场景。
- `equipment_test.go`：装备接口的未授权与参数校验。

运行：

```bash
BASE_URL=http://localhost:80 ADMIN_USERNAME=root ADMIN_PASSWORD=password go test ./test/integration/modules -count=1
```

若本地 Keto 数据丢失导致 admin API 出现 403（缺少 system:config 等权限），先执行：

```bash
KETO_CONTAINER=tsu_keto_service ADMIN_USER_ID=daf99445-61cc-4b24-9973-17eb79a53318 scripts/test/seed_root_permissions.sh
```
然后重新运行测试。

## Fixture 工厂与数据隔离

```go
cfg := apitest.LoadConfig()
factory := apitest.NewFixtureFactory(cfg)
heroReq := factory.BuildCreateHeroRequest("class-warrior-001", "")
username, email, password := factory.UniquePlayerCredentials("case01")
teamReq := factory.BuildCreateTeamRequest("hero-uuid")
```

- `RunID`（时间戳）保证相同 CI 任务中的资源唯一。
- `PLAYER_USERNAME_PREFIX` 与 `PLAYER_EMAIL_SUFFIX` 允许区分环境。
- `SampleUploadPath()` 返回 `test/fixtures/sample-upload.txt`，用于所有需要文件上传的接口。
