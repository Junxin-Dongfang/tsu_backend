package test

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

// SetupTestDB 设置测试数据库连接
func SetupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	// 从环境变量获取数据库配置
	host := getEnv("TEST_DB_HOST", "localhost")
	port := getEnv("TEST_DB_PORT", "5432")
	user := getEnv("TEST_DB_USER", "postgres")
	password := getEnv("TEST_DB_PASSWORD", "postgres")
	dbname := getEnv("TEST_DB_NAME", "tsu_db")

	// 构建连接字符串
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	// 连接数据库
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Skipf("无法连接测试数据库: %v", err)
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		db.Close()
		t.Skipf("无法ping测试数据库: %v", err)
	}

	// 开启事务(用于测试隔离)
	// 注意: 这里不开启事务,因为某些测试需要提交
	// 测试结束后会清理数据

	return db
}

// TeardownTestDB 清理测试数据库
func TeardownTestDB(t *testing.T, db *sql.DB) {
	t.Helper()

	if db != nil {
		// 清理测试数据
		// 注意: 这里可以添加清理逻辑,删除测试创建的数据
		// 为了简单起见,我们依赖数据库的隔离性

		db.Close()
	}
}

// getEnv 获取环境变量,如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// TruncateTable 清空表数据
func TruncateTable(t *testing.T, db *sql.DB, schema, table string) {
	t.Helper()

	query := fmt.Sprintf("TRUNCATE TABLE %s.%s CASCADE", schema, table)
	_, err := db.Exec(query)
	if err != nil {
		t.Fatalf("清空表失败 %s.%s: %v", schema, table, err)
	}
}

// BeginTestTransaction 开启测试事务
func BeginTestTransaction(t *testing.T, db *sql.DB) *sql.Tx {
	t.Helper()

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("开启测试事务失败: %v", err)
	}

	return tx
}

// RollbackTestTransaction 回滚测试事务
func RollbackTestTransaction(t *testing.T, tx *sql.Tx) {
	t.Helper()

	if tx != nil {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			t.Logf("回滚测试事务失败: %v", err)
		}
	}
}
