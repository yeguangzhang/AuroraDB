package conn

import (
	"AuroraDB/backend/model"
	"database/sql"
)

type (
	Executor interface {
		SetConn(db *sql.DB)
		SelectOne(db, sql string) any
		SelectList(db, sql string) []map[string]interface{}
		SelectPage(db, tb string, param *model.TableDataParams) (*model.TableData, error)
		InsertOne(db, sql string) any
		InsertMany(db, sql string) any
		Update(db, sql string) any
		Delete(db, sql string) any
	}
)
