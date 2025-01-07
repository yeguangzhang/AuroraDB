package conn

import (
	"AuroraDB/backend/model"
	"database/sql"
	"fmt"
	"strings"
)

type MysqlExecutor struct {
	conn *sql.DB
}

func (e *MysqlExecutor) SetConn(current *sql.DB) {
	e.conn = current
}

func (e *MysqlExecutor) SelectOne(db, sql string) any {
	return nil
}
func (e *MysqlExecutor) SelectList(db, sql string) []map[string]interface{} {
	return nil
}
func (e *MysqlExecutor) InsertOne(db, sql string) any {
	return nil
}
func (e *MysqlExecutor) InsertMany(db, sql string) any {
	return nil
}
func (e *MysqlExecutor) Update(db, sql string) any {
	return nil
}
func (e *MysqlExecutor) Delete(db, sql string) any {
	return nil
}

func (e *MysqlExecutor) SelectPage(db, table string, params *model.TableDataParams) (*model.TableData, error) {
	var total int64
	countQuery := fmt.Sprintf("select count(*) from %s.%s", db, table)
	fmt.Println(countQuery)
	err := e.conn.QueryRow(countQuery).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("count query error: %v", err)
	}
	columnsQuery := fmt.Sprintf("SHOW COLUMNS FROM %s.%s", db, table)
	rows, err := e.conn.Query(columnsQuery)
	var columns []model.Column
	var columnNames []string
	for rows.Next() {
		var field, typ, null, key, defaultVal, extra sql.NullString
		if err := rows.Scan(&field, &typ, &null, &key, &defaultVal, &extra); err != nil {
			return nil, err
		}
		if field.Valid {
			columns = append(columns, model.Column{Title: field.String, Key: field.String})
			columnNames = append(columnNames, fmt.Sprintf("`%s`", field.String))
		}
	}
	// 3. 获取分页数据
	offset := (params.Page - 1) * params.PageSize
	dataQuery := fmt.Sprintf("SELECT %s FROM %s.%s LIMIT %d OFFSET %d",
		strings.Join(columnNames, ", "),
		db, table,
		params.PageSize,
		offset,
	)
	fmt.Println(dataQuery)
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
