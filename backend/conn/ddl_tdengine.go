package conn

import (
	"AuroraDB/backend/model"
	"database/sql"
	"fmt"
)

type TDHandler struct {
	conn *sql.DB
}

func (m *TDHandler) SetConn(conn *sql.DB) {
	m.conn = conn
}
func (m *TDHandler) ShowDBs() ([]string, error) {
	return nil, nil
}
func (m *TDHandler) ShowTables(dbname string) ([]string, error) {
	return nil, nil

}
func (m *TDHandler) ShowTableStructure(db, table string) ([]*model.ColumnInfo, error) {
	return nil, nil
}
func (m *TDHandler) DropTable(db, table string) error {
	return nil
}

func (m *TDHandler) UseDB(db string) error {
	_, err := m.conn.Exec(fmt.Sprintf("USE `%s`", db))
	if err != nil {
		return fmt.Errorf("exec error: %v", err)
	}
	return nil
}

func (m *TDHandler) GetTableStats(db string) (*model.DatabaseStats, error) {
	var stats model.DatabaseStats
	query := fmt.Sprintf(`
			select stable_name,count(1) from information_schema.ins_tables 
			where type='CHILD_TABLE'and db_name='%s' 
			group by stable_name
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
