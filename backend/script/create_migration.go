//go:build ignore
// +build ignore

package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql/schema"
	"github.com/chaitin/WhaleHire/backend/config"
	"github.com/chaitin/WhaleHire/backend/db"
	_ "github.com/chaitin/WhaleHire/backend/db/runtime"
	_ "github.com/lib/pq"
)

// getMigrationDir 获取migration目录的绝对路径
func getMigrationDir() (string, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("failed to get current file path")
	}

	// 获取当前脚本文件所在目录
	scriptDir := filepath.Dir(filename)
	// 获取backend目录
	backendDir := filepath.Dir(scriptDir)
	// 构建migration目录路径
	migrationDir := filepath.Join(backendDir, "migration")

	return migrationDir, nil
}

// formatSQL 格式化SQL语句，提高可读性
func formatSQL(sql string) string {
	// 移除多余的空白字符
	sql = regexp.MustCompile(`\s+`).ReplaceAllString(sql, " ")
	sql = strings.TrimSpace(sql)

	// 在关键字前后添加换行
	keywords := []string{
		"CREATE TABLE", "ALTER TABLE", "DROP TABLE",
		"CREATE INDEX", "DROP INDEX",
		"CREATE UNIQUE INDEX",
		"ADD CONSTRAINT", "DROP CONSTRAINT",
	}

	for _, keyword := range keywords {
		sql = regexp.MustCompile(`(?i)`+regexp.QuoteMeta(keyword)).ReplaceAllString(sql, "\n"+keyword)
	}

	// 格式化表列定义
	sql = formatTableColumns(sql)

	// 格式化约束
	sql = formatConstraints(sql)

	// 在分号后添加换行
	sql = regexp.MustCompile(`;\s*`).ReplaceAllString(sql, ";\n\n")

	// 清理多余的换行
	sql = regexp.MustCompile(`\n{3,}`).ReplaceAllString(sql, "\n\n")
	sql = strings.TrimSpace(sql)

	return sql
}

// formatTableColumns 格式化表列定义
func formatTableColumns(sql string) string {
	// 在左括号后和右括号前添加换行和缩进
	sql = regexp.MustCompile(`\(\s*`).ReplaceAllString(sql, " (\n    ")
	sql = regexp.MustCompile(`\s*\)`).ReplaceAllString(sql, "\n)")

	// 在逗号后添加换行和缩进（仅在括号内）
	lines := strings.Split(sql, "\n")
	for i, line := range lines {
		if strings.Contains(line, ",") && (strings.Contains(line, "    ") || strings.TrimSpace(line) != line) {
			line = regexp.MustCompile(`,\s*`).ReplaceAllString(line, ",\n    ")
			lines[i] = line
		}
	}

	return strings.Join(lines, "\n")
}

// formatConstraints 格式化约束定义
func formatConstraints(sql string) string {
	// 在约束关键字前添加换行和缩进
	constraintKeywords := []string{
		"PRIMARY KEY", "FOREIGN KEY", "UNIQUE", "CHECK",
		"NOT NULL", "DEFAULT",
	}

	for _, keyword := range constraintKeywords {
		pattern := `(?i)\s+` + regexp.QuoteMeta(keyword)
		sql = regexp.MustCompile(pattern).ReplaceAllString(sql, " "+keyword)
	}

	return sql
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run create_migration.go <migration_name>")
	}

	migrationName := os.Args[1]

	// 获取下一个迁移版本号
	nextVersion, err := getNextMigrationVersion()
	if err != nil {
		log.Fatalf("failed to get next migration version: %v", err)
	}

	// 加载配置
	cfg, err := config.Init()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	ctx := context.Background()

	// 连接数据库
	client, err := db.Open("postgres", cfg.Database.Master)
	if err != nil {
		log.Fatalf("failed opening connection: %v", err)
	}
	defer client.Close()

	// 生成迁移文件
	if err := generateMigrationFiles(ctx, client, nextVersion, migrationName); err != nil {
		log.Fatalf("failed to generate migration files: %v", err)
	}

	fmt.Printf("Migration files generated successfully:\n")
	fmt.Printf("  - %06d_%s.up.sql\n", nextVersion, migrationName)
	fmt.Printf("  - %06d_%s.down.sql\n", nextVersion, migrationName)
	fmt.Printf("\nTo apply the migration, run app\n")
}

// getNextMigrationVersion 获取下一个迁移版本号
func getNextMigrationVersion() (int, error) {
	migrationDir, err := getMigrationDir()
	if err != nil {
		return 0, fmt.Errorf("failed to get migration directory: %w", err)
	}

	files, err := os.ReadDir(migrationDir)
	if err != nil {
		return 1, nil // 如果目录不存在，从1开始
	}

	maxVersion := 0
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		fileName := file.Name()
		if strings.HasSuffix(fileName, ".up.sql") {
			parts := strings.Split(fileName, "_")
			if len(parts) > 0 {
				if version, err := strconv.Atoi(parts[0]); err == nil {
					if version > maxVersion {
						maxVersion = version
					}
				}
			}
		}
	}

	return maxVersion + 1, nil
}

// generateMigrationFiles 生成迁移文件
func generateMigrationFiles(ctx context.Context, client *db.Client, version int, name string) error {
	migrationDir, err := getMigrationDir()
	if err != nil {
		return fmt.Errorf("failed to get migration directory: %w", err)
	}

	if err := os.MkdirAll(migrationDir, 0755); err != nil {
		return fmt.Errorf("failed to create migration directory: %w", err)
	}

	// 生成 up 文件
	upFileName := fmt.Sprintf("%06d_%s.up.sql", version, name)
	upFilePath := filepath.Join(migrationDir, upFileName)
	upFile, err := os.Create(upFilePath)
	if err != nil {
		return fmt.Errorf("failed to create up file: %w", err)
	}
	defer upFile.Close()

	// 将schema写入缓冲区
	var buf bytes.Buffer
	if err := client.Schema.WriteTo(ctx, &buf, schema.WithDialect("postgres")); err != nil {
		return fmt.Errorf("failed to write schema to buffer: %w", err)
	}

	// 格式化SQL并写入文件
	formattedSQL := formatSQL(buf.String())
	if _, err := upFile.WriteString(formattedSQL); err != nil {
		return fmt.Errorf("failed to write formatted SQL to up file: %w", err)
	}

	// 生成 down 文件（空文件，需要手动填写回滚逻辑）
	downFileName := fmt.Sprintf("%06d_%s.down.sql", version, name)
	downFilePath := filepath.Join(migrationDir, downFileName)
	downFile, err := os.Create(downFilePath)
	if err != nil {
		return fmt.Errorf("failed to create down file: %w", err)
	}
	defer downFile.Close()

	// 写入down文件的注释
	downContent := fmt.Sprintf("-- Migration: %s\n-- Created: %s\n-- TODO: Add rollback SQL statements here\n",
		name, time.Now().Format("2006-01-02 15:04:05"))
	if _, err := downFile.WriteString(downContent); err != nil {
		return fmt.Errorf("failed to write down file content: %w", err)
	}

	return nil
}
