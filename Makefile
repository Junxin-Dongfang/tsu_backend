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

.PHONY: help swagger-gen swagger-admin swagger-game dev-up dev-down

help:
	@echo "Available commands:"
	@echo "  swagger-gen    - Generate all swagger docs"
	@echo "  swagger-admin  - Generate admin service swagger docs"
	@echo "  swagger-game   - Generate game service swagger docs"
	@echo "  dev-up         - Start development environment"
	@echo "  dev-down       - Stop development environment"

# 安装 swag 工具
install-swag:
	go install github.com/swaggo/swag/cmd/swag@latest

# 生成 admin 服务的 swagger 文档
swagger-admin: install-swag
	swag init -g internal/modules/admin/http_handle.go -o ./docs/admin

# 生成 game 服务的 swagger 文档（如果有的话）
swagger-game: install-swag
	swag init -g internal/modules/game/http_handle.go -o ./docs/game

# 生成所有 swagger 文档
swagger-gen: swagger-admin swagger-game

# 启动开发环境
dev-up:
	docker-compose -f deployments/docker-compose/docker-compose-main.local.yml up -d

# 停止开发环境
dev-down:
	docker-compose -f deployments/docker-compose/docker-compose-main.local.yml down

# 重新构建并启动
dev-rebuild:
	docker-compose -f deployments/docker-compose/docker-compose-main.local.yml up -d --build