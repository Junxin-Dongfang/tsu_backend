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