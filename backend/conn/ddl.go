package conn

import (
	"AuroraDB/backend/model"
	"database/sql"
)

type DDLHandler interface {
	SetConn(db *sql.DB)
	ShowDBs() ([]string, error)
	ShowTables(dbname string) ([]string, error)
	ShowTableStructure(db, table string) ([]*model.ColumnInfo, error)
	DropTable(db, table string) error
	UseDB(dbname string) error
	GetTableStats(db string) (*model.DatabaseStats, error)
}
