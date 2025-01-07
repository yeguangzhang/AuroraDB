package conn

import (
	"AuroraDB/backend/model"
	"database/sql"
	"fmt"
	"strings"
)

type PostgresqlExecutor struct {
	conn *sql.DB
}

func (e *PostgresqlExecutor) SetConn(c *sql.DB) {
	e.conn = c
}
func (e *PostgresqlExecutor) SelectOne(db, sql string) any {
	return nil
}
func (e *PostgresqlExecutor) SelectList(db, sql string) []map[string]interface{} {
	return nil
}
func (e *PostgresqlExecutor) InsertOne(db, sql string) any {
	return nil
}
func (e *PostgresqlExecutor) InsertMany(db, sql string) any {
	return nil
}
func (e *PostgresqlExecutor) Update(db, sql string) any {
	return nil
}
func (e *PostgresqlExecutor) Delete(db, sql string) any {
	return nil
}

func (e *PostgresqlExecutor) SelectPage(db, table string, params *model.TableDataParams) (*model.TableData, error) {
	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
	err := e.conn.QueryRow(countQuery).Scan(&total)
	var columns []model.Column
	var columnNames []string
	columnsQuery := `
			SELECT column_name 
			FROM information_schema.columns 
			WHERE  table_name = $2 
			ORDER BY ordinal_position
		`
	rows, err := e.conn.Query(columnsQuery, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var columnName string
		if err := rows.Scan(&columnName); err != nil {
			return nil, err
		}
		columns = append(columns, model.Column{Title: columnName, Key: columnName})
		columnNames = append(columnNames, fmt.Sprintf(`"%s"`, columnName))
	}
	offset := (params.Page - 1) * params.PageSize
	dataQuery := fmt.Sprintf("SELECT %s FROM %s LIMIT %d OFFSET %d",
		strings.Join(columnNames, ", "),
		table,
		params.PageSize,
		offset,
	)
	rows, err2 := e.conn.Query(dataQuery)
	if err2 != nil {
		return nil, fmt.Errorf("data query error: %v", err)
	}
	defer rows.Close()
	// 4. 处理数据行
	var data [][]any
	for rows.Next() {
		values := make([]any, len(columns))
		scanArgs := make([]any, len(columns))
		for i := range values {
			scanArgs[i] = &values[i]
		}

		if err := rows.Scan(scanArgs...); err != nil {
			return nil, fmt.Errorf("scan error: %v", err)
		}

		// 处理 NULL 值和类型转换
		row := make([]any, len(columns))
		for i, v := range values {
			if v == nil {
				row[i] = nil
			} else {
				switch v := v.(type) {
				case []byte:
					row[i] = string(v)
				default:
					row[i] = v
				}
			}
		}
		data = append(data, row)
	}
	tableData := &model.TableData{
		Columns: columns,
		Data:    data,
		Total:   total,
	}

	return tableData, nil
}
