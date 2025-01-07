package conn

import (
	"AuroraDB/backend/model"
	"database/sql"
	"fmt"
)

type MysqlHandler struct {
	conn *sql.DB
}

func (m *MysqlHandler) SetConn(conn *sql.DB) {
	m.conn = conn
}
func (m *MysqlHandler) ShowDBs() ([]string, error) {
	query := "SHOW DATABASES"
	rows, err := m.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query error: %v", err)
	}
	defer rows.Close()

	var databases []string
	for rows.Next() {
		var dbName string
		if err := rows.Scan(&dbName); err != nil {
			return nil, fmt.Errorf("scan error: %v", err)
		}
		databases = append(databases, dbName)
	}
	return databases, nil
}
func (m *MysqlHandler) ShowTables(dbname string) ([]string, error) {
	query := fmt.Sprintf("SHOW TABLES FROM `%s`", dbname)
	rows, err := m.conn.Query(query)

	if err != nil {
		return nil, fmt.Errorf("query error: %v", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("scan error: %v", err)
		}
		tables = append(tables, tableName)
	}
	return tables, nil

}
func (m *MysqlHandler) ShowTableStructure(db, table string) ([]*model.ColumnInfo, error) {
	query := `	SELECT 
				COLUMN_NAME as name,
				COLUMN_TYPE as type,
				IS_NULLABLE as nullable,
				COLUMN_DEFAULT as default_value,
				COLUMN_COMMENT as comment
			FROM INFORMATION_SCHEMA.COLUMNS 
			WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
			ORDER BY ORDINAL_POSITION`
	rows, err := m.conn.Query(query, db, table)
	if err != nil {
		return nil, err
	}
	var columns []*model.ColumnInfo
	for rows.Next() {
		var col model.ColumnInfo
		var defaultVal, comment sql.NullString
		err := rows.Scan(&col.Name, &col.Type, &col.Nullable, &defaultVal, &comment)
		if err != nil {
			return nil, err
		}

		if defaultVal.Valid {
			col.Default = defaultVal.String
		}
		if comment.Valid {
			col.Comment = comment.String
		}

		columns = append(columns, &col)
	}

	return columns, nil
}
func (m *MysqlHandler) DropTable(db, table string) error {
	return nil
}
func (m *MysqlHandler) UseDB(db string) error {
	sqlStr := fmt.Sprintf("USE `%s`", db)
	fmt.Println(sqlStr)
	_, err := m.conn.Exec(sqlStr)
	if err != nil {
		return fmt.Errorf("exec error: %v", err)
	}
	var currentDB string
	err = m.conn.QueryRow("SELECT DATABASE()").Scan(&currentDB)
	if err != nil {
		return fmt.Errorf("exec error: %v", err)
	}
	if currentDB != db {
		fmt.Println("current db is", currentDB)
		_, _ = m.conn.Exec(sqlStr)
	}
	return nil
}

func (m *MysqlHandler) GetTableStats(db string) (*model.DatabaseStats, error) {
	var stats model.DatabaseStats
	query := fmt.Sprintf(`
			SELECT 
				table_name,
				table_rows
			FROM 
				information_schema.tables 
			WHERE 
				table_schema = '%s'
		`, db)
	rows, err := m.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	totalRecords := 0
	for rows.Next() {
		var stat model.TableStats
		var count int
		if err := rows.Scan(&stat.TableName, &count); err != nil {
			return nil, err
		}
		stat.RecordCount = count
		totalRecords += count
		stats.TableStats = append(stats.TableStats, stat)
	}

	stats.TotalRecords = totalRecords
	return &stats, nil
}
