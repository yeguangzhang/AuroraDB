package conn

import (
	"AuroraDB/backend/model"
	"database/sql"
	_ "github.com/taosdata/driver-go/v3/taosRestful"
)

type TdEngineExecutor struct {
	conn *sql.DB
}

func (e *TdEngineExecutor) SetConn(current *sql.DB) {
	e.conn = current
}

func (e *TdEngineExecutor) SelectOne(db, sql string) any {

	return nil
}
func (e *TdEngineExecutor) SelectList(db, sql string) []map[string]interface{} {

	return nil
}
func (e *TdEngineExecutor) InsertOne(db, sql string) any {
	return nil
}
func (e *TdEngineExecutor) InsertMany(db, sql string) any {
	return nil
}
func (e *TdEngineExecutor) Update(db, sql string) any {
	return nil
}
func (e *TdEngineExecutor) Delete(db, sql string) any {
	return nil
}
func (e *TdEngineExecutor) SelectPage(db, table string, params *model.TableDataParams) (*model.TableData, error) {
	var tableData model.TableData
	return &tableData, nil
}
