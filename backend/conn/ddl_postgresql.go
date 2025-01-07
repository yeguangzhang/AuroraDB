package conn

import (
	"AuroraDB/backend/model"
	"database/sql"
	"fmt"
)

type PostgresqlHandler struct {
	conn *sql.DB
}

func (m *PostgresqlHandler) SetConn(conn *sql.DB) {
	m.conn = conn
}
func (m *PostgresqlHandler) ShowDBs() ([]string, error) {
	query := "SELECT datname FROM pg_database WHERE datistemplate = false"
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
func (m *PostgresqlHandler) ShowTables(dbname string) ([]string, error) {
	query := `
			SELECT table_name 
			FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_catalog = $1
		`
	rows, err := m.conn.Query(query, dbname)
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
func (m *PostgresqlHandler) ShowTableStructure(db, table string) ([]*model.ColumnInfo, error) {
	query := `
			SELECT 
				column_name as name,
				data_type as type,
				is_nullable as nullable,
				column_default as default_value,
				col_description((table_schema || '.' || table_name)::regclass::oid, ordinal_position) as comment
			FROM information_schema.columns
			WHERE table_schema = $1 AND table_name = $2
			ORDER BY ordinal_position
		`
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
func (m *PostgresqlHandler) DropTable(db, table string) error {
	return nil
}

func (m *PostgresqlHandler) UseDB(db string) error {
	_, err := m.conn.Exec(fmt.Sprintf("USE `%s`", db))
	if err != nil {
		return fmt.Errorf("exec error: %v", err)
	}
	return nil
}

func (m *PostgresqlHandler) GetTableStats(db string) (*model.DatabaseStats, error) {
	return nil, nil
}
