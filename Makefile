MAIN_DB_URL="postgres://postgres:postgres@localhost:5432/tsu_db?sslmode=disable"

.PHONY: migrate-create migrate-up migrate-down

# 创建一个新的迁移文件
migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir ./migrations -seq $$name

# 应用所有新的迁移
migrate-up:
	migrate -database $(MAIN_DB_URL) -path ./migrations up

# 回滚最后一个迁移
migrate-down:
	migrate -database $(MAIN_DB_URL) -path ./migrations down 1

.PHONY: help swagger-gen swagger-admin dev-up dev-down dev-logs generate-models install-sqlboiler dev-rebuild clean sqlboiler install-swag proto generate install-protoc deploy prod-up prod-down prod-logs prod-build

help:
	@echo "Available commands:"
	@echo "  Code Generation:"
	@echo "    proto            - Generate Protobuf code"
	@echo "    generate-entity  - Generate database models using SQLBoiler"
	@echo "    generate-errors  - Generate frontend error code enums (TypeScript)"
	@echo "    generate         - Generate all code (proto + entity)"
	@echo "    swagger-gen      - Generate admin service swagger docs"
	@echo ""
	@echo "  Development:"
	@echo "    dev-up           - Start development environment"
	@echo "    dev-down         - Stop development environment"
	@echo "    dev-logs         - Show logs from all services"
	@echo "    dev-rebuild      - Rebuild and restart development environment"
	@echo ""
	@echo "  Production Deployment (Layered - Recommended):"
	@echo "    deploy-prod-step1            - Step 1: Deploy infrastructure (PostgreSQL, Redis, etc.)"
	@echo "    deploy-prod-step2            - Step 2: Deploy Ory services (Kratos, Keto, Oathkeeper)"
	@echo "    deploy-prod-step3            - Step 3: Deploy main service (Admin Server)"
	@echo "    deploy-prod-step4            - Step 4: Deploy Nginx (Reverse proxy)"
	@echo "    deploy-prod-all              - Deploy all steps automatically"
	@echo "    deploy-prod-all-interactive  - Deploy all steps with confirmation prompts"
	@echo "    import-game-config-prod      - Import game config to production database"
	@echo ""
	@echo "  Production Deployment (Legacy):"
	@echo "    deploy           - Deploy source code to server via SSH"
	@echo "    prod-build       - Build production Docker images locally"
	@echo "    prod-up          - Start production environment locally"
	@echo "    prod-down        - Stop production environment"
	@echo "    prod-logs        - Show production logs"
	@echo "    build-push       - Build and push image to registry"
	@echo "    deploy-image     - Deploy image from registry to server"
	@echo "    deploy-full      - Full pipeline: build + push + deploy"
	@echo ""
	@echo "  Database:"
	@echo "    migrate-create   - Create a new migration file"
	@echo "    migrate-up       - Apply all new migrations"
	@echo "    migrate-down     - Rollback the last migration"
	@echo ""
	@echo "  Utilities:"
	@echo "    clean            - Clean up Docker resources"

# 安装 protoc 编译器
install-protoc:
	@echo "检查 protoc 是否已安装..."
	@which protoc > /dev/null || (echo "❌ protoc 未安装,请运行: brew install protobuf (macOS) 或访问 https://grpc.io/docs/protoc-installation/" && exit 1)
	@echo "✅ protoc 已安装: $$(protoc --version)"
	@echo "安装 Go protobuf 插件..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@echo "✅ protoc-gen-go 已安装"

# 生成 Protobuf 代码
proto: install-protoc
	@echo "🔄 生成 Protobuf 代码..."
	@mkdir -p internal/pb/common internal/pb/auth internal/pb/admin internal/pb/game
	@echo "  - 生成 common 包..."
	@protoc --go_out=. --go_opt=paths=source_relative proto/common/*.proto 2>/dev/null && \
		mv proto/common/*.pb.go internal/pb/common/ || echo "⚠️  proto/common/ 目录不存在,跳过"
	@echo "  - 生成 auth 包..."
	@protoc --go_out=. --go_opt=paths=source_relative proto/auth/*.proto 2>/dev/null && \
		mv proto/auth/*.pb.go internal/pb/auth/ || echo "⚠️  proto/auth/ 目录不存在,跳过"
	@echo "  - 生成 admin 包..."
	@protoc --go_out=. --go_opt=paths=source_relative proto/admin/*.proto 2>/dev/null && \
		mv proto/admin/*.pb.go internal/pb/admin/ || echo "⚠️  proto/admin/ 目录不存在,跳过"
	@echo "  - 生成 game 包..."
	@protoc --go_out=. --go_opt=paths=source_relative proto/game/*.proto 2>/dev/null && \
		mv proto/game/*.pb.go internal/pb/game/ || echo "⚠️  proto/game/ 目录不存在,跳过"
	@echo "✅ Protobuf 代码生成完成"

# 生成前端错误码枚举
generate-errors:
	@echo "🔄 生成前端错误码枚举..."
	@go run cmd/generate-error-codes/main.go -output ./generated/frontend -format all
	@echo "✅ 错误码枚举生成完成"

# 生成所有代码
generate: proto generate-entity
	@echo "✅ 所有代码生成完成"

# 安装 swag 工具
install-swag:
	go install github.com/swaggo/swag/cmd/swag@latest

# 安装 SQLBoiler 工具
install-sqlboiler:
	go install github.com/aarondl/sqlboiler/v4@latest
	go install github.com/aarondl/sqlboiler/v4/drivers/sqlboiler-psql@latest

# 生成数据库模型 - 所有 schema
generate-entity: install-sqlboiler
	@echo "🔄 生成 auth schema 模型..."
	PATH="$(shell go env GOPATH)/bin:$$PATH" sqlboiler psql --config sqlboiler.auth.toml
	@echo "🔄 生成 game_config schema 模型..."
	PATH="$(shell go env GOPATH)/bin:$$PATH" sqlboiler psql --config sqlboiler.game_config.toml
	@echo "🔄 生成 game_runtime schema 模型..."
	PATH="$(shell go env GOPATH)/bin:$$PATH" sqlboiler psql --config sqlboiler.game_runtime.toml
	@echo "✅ 所有 schema 的实体模型生成完成"

# 生成 admin 服务的 swagger 文档
swagger-admin: install-swag
	@echo "🔄 生成 Admin Server Swagger 文档..."
	swag init -g cmd/admin-server/main.go -o ./docs/admin \
		--parseDependency --parseInternal \
		--exclude internal/modules/game
	@echo "✅ Admin Swagger 文档生成完成: docs/admin/"

# 生成 game 服务的 swagger 文档
swagger-game: install-swag
	@echo "🔄 生成 Game Server Swagger 文档..."
	swag init -g cmd/game-server/main.go -o ./docs/game \
		--parseDependency --parseInternal \
		--exclude internal/modules/admin
	@echo "✅ Game Swagger 文档生成完成: docs/game/"

# 生成所有 swagger 文档
swagger-gen: swagger-admin swagger-game
	@echo "✅ 所有 Swagger 文档生成完成"

# 启动开发环境
dev-up:
	docker network create tsu-network 2>/dev/null || true
	@echo "🚀 启动 Ory 服务 (Kratos, Keto, Oathkeeper)..."
	docker-compose -f deployments/docker-compose/docker-compose-ory.local.yml up -d
	@echo "⏳ 等待 Ory 服务就绪..."
	sleep 10
	@echo "🚀 启动主服务 (Admin, Game)..."
	docker-compose -f deployments/docker-compose/docker-compose-main.local.yml up -d
	@echo "🚀 启动 Nginx..."
	docker-compose -f deployments/docker-compose/docker-compose-nginx.local.yml up -d
	@echo "✅ 所有服务已启动"
	@echo ""
	@echo "📋 访问地址:"
	@echo "  - 统一 Swagger 入口: http://localhost/swagger"
	@echo "  - Admin Swagger:      http://localhost/admin/swagger/index.html"
	@echo "  - Game Swagger:       http://localhost/game/swagger/index.html"

# 停止开发环境
dev-down:
	docker-compose -f deployments/docker-compose/docker-compose-nginx.local.yml down
	docker-compose -f deployments/docker-compose/docker-compose-main.local.yml down
	docker-compose -f deployments/docker-compose/docker-compose-ory.local.yml down

# 查看日志
dev-logs:
	docker-compose -f deployments/docker-compose/docker-compose-main.local.yml logs -f

# 查看所有服务日志
dev-logs-all:
	docker-compose -f deployments/docker-compose/docker-compose-ory.local.yml logs -f & \
	docker-compose -f deployments/docker-compose/docker-compose-main.local.yml logs -f & \
	docker-compose -f deployments/docker-compose/docker-compose-nginx.local.yml logs -f

# 重新构建并启动
dev-rebuild:
	docker-compose -f deployments/docker-compose/docker-compose-main.local.yml up -d --build

# 清理
clean:
	docker-compose -f deployments/docker-compose/docker-compose-nginx.local.yml down -v
	docker-compose -f deployments/docker-compose/docker-compose-main.local.yml down -v
	docker-compose -f deployments/docker-compose/docker-compose-ory.local.yml down -v
	docker system prune -f

# ==========================================
# 生产环境分步部署（推荐方案）
# ==========================================

# 步骤 1: 部署基础设施（PostgreSQL、Redis、NATS、Consul）
deploy-prod-step1:
	@echo "🚀 步骤 1: 部署基础设施..."
	@chmod +x scripts/deploy-prod-step1-infra.sh
	@./scripts/deploy-prod-step1-infra.sh

# 步骤 2: 部署 Ory 服务（Kratos、Keto、Oathkeeper）
deploy-prod-step2:
	@echo "🚀 步骤 2: 部署 Ory 服务..."
	@chmod +x scripts/deploy-prod-step2-ory.sh
	@./scripts/deploy-prod-step2-ory.sh

# 步骤 3: 部署主服务（Admin Server + 数据库迁移）
deploy-prod-step3:
	@echo "🚀 步骤 3: 部署主服务..."
	@chmod +x scripts/deploy-prod-step3-app.sh
	@./scripts/deploy-prod-step3-app.sh

# 步骤 4: 部署 Nginx（反向代理）
deploy-prod-step4:
	@echo "🚀 步骤 4: 部署 Nginx..."
	@chmod +x scripts/deploy-prod-step4-nginx.sh
	@./scripts/deploy-prod-step4-nginx.sh

# 一键部署所有步骤（自动模式）
deploy-prod-all:
	@echo "🎯 一键部署所有步骤..."
	@chmod +x scripts/deploy-prod-all.sh
	@./scripts/deploy-prod-all.sh --auto

# 一键部署所有步骤（交互模式）
deploy-prod-all-interactive:
	@echo "🎯 一键部署所有步骤（交互模式）..."
	@chmod +x scripts/deploy-prod-all.sh
	@./scripts/deploy-prod-all.sh

# 导入游戏配置到生产服务器
import-game-config-prod:
	@echo "📦 导入游戏配置到生产服务器..."
	@chmod +x scripts/import-game-config-prod.sh
	@./scripts/import-game-config-prod.sh