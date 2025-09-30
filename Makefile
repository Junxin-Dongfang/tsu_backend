MAIN_DB_URL="postgres://tsu_user:tsu_test@localhost:5432/tsu_db?sslmode=disable"

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

.PHONY: help swagger-gen swagger-admin dev-up dev-down dev-logs generate-models install-sqlboiler dev-rebuild clean sqlboiler install-swag

help:
	@echo "Available commands:"
	@echo "  swagger-gen      - Generate admin service swagger docs"
	@echo "  generate-models  - Generate database models using SQLBoiler"
	@echo "  dev-up          - Start development environment"
	@echo "  dev-down        - Stop development environment"
	@echo "  dev-logs        - Show logs from all services"
	@echo "  dev-rebuild     - Rebuild and restart development environment"
	@echo "  clean           - Clean up Docker resources"
	@echo "  migrate-create  - Create a new migration file"
	@echo "  migrate-up      - Apply all new migrations"
	@echo "  migrate-down    - Rollback the last migration"
	@echo "  sqlboiler       - Generate entity models using SQLBoiler"

# 安装 swag 工具
install-swag:
	go install github.com/swaggo/swag/cmd/swag@latest

# 安装 SQLBoiler 工具
install-sqlboiler:
	go install github.com/aarondl/sqlboiler/v4@latest
	go install github.com/aarondl/sqlboiler/v4/drivers/sqlboiler-psql@latest

# 生成数据库模型
generate-models: install-sqlboiler
	PATH="$(shell go env GOPATH)/bin:$$PATH" sqlboiler psql --config sqlboiler.toml
	@echo "✅ 实体模型生成完成"

# 生成 admin 服务的 swagger 文档
swagger-admin: install-swag
	swag init -g internal/modules/admin/http_handle.go -o ./docs --parseDependency --parseInternal	

# 生成所有 swagger 文档
swagger-gen: swagger-admin

# 启动开发环境
dev-up:
	docker network create tsu-network 2>/dev/null || true
	docker-compose -f deployments/docker-compose/docker-compose-main.local.yml up -d

# 停止开发环境
dev-down:
	docker-compose -f deployments/docker-compose/docker-compose-main.local.yml down

# 查看日志
dev-logs:
	docker-compose -f deployments/docker-compose/docker-compose-main.local.yml logs -f

# 重新构建并启动
dev-rebuild:
	docker-compose -f deployments/docker-compose/docker-compose-main.local.yml up -d --build

# 清理
clean:
	docker-compose -f deployments/docker-compose/docker-compose-main.local.yml down -v
	docker system prune -f

# 用sqlboiler生成entitys
sqlboiler: install-sqlboiler
	PATH="$(shell go env GOPATH)/bin:$$PATH" sqlboiler psql --config sqlboiler.toml
	@echo "✅ 实体模型生成完成"