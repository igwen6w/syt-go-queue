package database

import (
	"context"
	"fmt"
	_ "github.com/go-sql-driver/mysql" // MySQL 驱动
	"github.com/jmoiron/sqlx"
	"regexp"
	"strings"
)

type ValuationRecord struct {
	ID              int64  `db:"id"`
	Status          string `db:"status"`
	UserMessage     string `db:"user_message"`
	SysMessage      string `db:"sys_message"`
	Report          string `db:"report"`
	FailedTimes     int    `db:"failed_times"`
	FailedInfo      string `db:"failed_info"`
	Progress        string `db:"progress"`
	ProgressInfo    string `db:"progress_info"`
	CurrentTaskNode int    `db:"current_task_node"`
	CallbackURL     string `db:"callback_url"`
}

type Database struct {
	db *sqlx.DB
}

func NewDatabase(db *sqlx.DB) *Database {
	return &Database{db: db}
}

// validateTableName 验证表名是否合法，防止SQL注入
func validateTableName(tableName string) error {
	// 只允许字母、数字、下划线和特定前缀
	validTableNameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	if !validTableNameRegex.MatchString(tableName) {
		return fmt.Errorf("invalid table name: %s", tableName)
	}

	// 防止使用保留关键字作为表名
	reservedKeywords := []string{"select", "from", "where", "insert", "update", "delete", "drop", "alter", "table"}
	tableNameLower := strings.ToLower(tableName)
	for _, keyword := range reservedKeywords {
		if tableNameLower == keyword {
			return fmt.Errorf("table name cannot be a reserved keyword: %s", tableName)
		}
	}

	return nil
}

// validateFieldName 验证字段名是否合法，防止SQL注入
func validateFieldName(fieldName string) error {
	// 只允许字母、数字、下划线
	validFieldNameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	if !validFieldNameRegex.MatchString(fieldName) {
		return fmt.Errorf("invalid field name: %s", fieldName)
	}

	// 防止使用保留关键字作为字段名
	reservedKeywords := []string{"select", "from", "where", "insert", "update", "delete", "drop", "alter", "table",
		"and", "or", "not", "like", "in", "between", "is", "null", "true", "false"}
	fieldNameLower := strings.ToLower(fieldName)
	for _, keyword := range reservedKeywords {
		if fieldNameLower == keyword {
			return fmt.Errorf("field name cannot be a reserved keyword: %s", fieldName)
		}
	}

	return nil
}

// GetValuationRecord 获取评估记录
func (d *Database) GetValuationRecord(ctx context.Context, tableName string, id int64) (*ValuationRecord, error) {
	// 验证表名
	if err := validateTableName(tableName); err != nil {
		return nil, err
	}

	query := fmt.Sprintf(`
        SELECT id, status, user_message, sys_message, report,
               failed_times, failed_info, progress, progress_info, current_task_node, callback_url
        FROM %s WHERE id = ?`, tableName)

	var record ValuationRecord
	err := d.db.GetContext(ctx, &record, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get valuation record: %w", err)
	}
	return &record, nil
}

// UpdateStatus 更新状态
func (d *Database) UpdateStatus(ctx context.Context, tableName string, id int64, status string) error {
	// 验证表名
	if err := validateTableName(tableName); err != nil {
		return err
	}

	query := fmt.Sprintf("UPDATE %s SET status = ? WHERE id = ?", tableName)
	_, err := d.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}
	return nil
}

// UpdateFailedInfo 更新失败信息
func (d *Database) UpdateFailedInfo(ctx context.Context, tableName string, id int64, failedInfo string, failedTimes int) error {
	// 验证表名
	if err := validateTableName(tableName); err != nil {
		return err
	}

	query := fmt.Sprintf("UPDATE %s SET failed_info = ?, failed_times = ? WHERE id = ?", tableName)
	_, err := d.db.ExecContext(ctx, query, failedInfo, failedTimes, id)
	if err != nil {
		return fmt.Errorf("failed to update failed info: %w", err)
	}
	return nil
}

// UpdateRecord 更新记录
func (d *Database) UpdateRecord(ctx context.Context, tableName string, id int64, updates map[string]interface{}) error {
	// 验证表名
	if err := validateTableName(tableName); err != nil {
		return err
	}

	// 验证字段名
	for field := range updates {
		if err := validateFieldName(field); err != nil {
			return err
		}
	}

	// 构建 UPDATE 语句
	var setClauses []string
	var args []interface{}

	for field, value := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = ?", field))
		args = append(args, value)
	}

	query := fmt.Sprintf("UPDATE %s SET %s WHERE id = ?",
		tableName,
		strings.Join(setClauses, ", "))

	args = append(args, id)

	_, err := d.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update record: %w", err)
	}
	return nil
}
