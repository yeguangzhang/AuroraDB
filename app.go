package main

import (
	"AuroraDB/backend/config"
	"AuroraDB/backend/conn"
	"AuroraDB/backend/model"
	"AuroraDB/backend/ssh"
	"context"
	"database/sql"
	"fmt"
	"net"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/taosdata/driver-go/v3/taosRestful"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx        context.Context
	tunnels    map[string]*ssh.Tunnel
	dbConns    map[string]*sql.DB
	executors  map[string]conn.Executor
	handlers   map[string]conn.DDLHandler
	current_db string
}

// DBConfig 数据库配置结构体
type DBConfig struct {
	Type     string `json:"type"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		tunnels:   make(map[string]*ssh.Tunnel),
		dbConns:   make(map[string]*sql.DB),
		handlers:  make(map[string]conn.DDLHandler),
		executors: make(map[string]conn.Executor),
	}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// LoadConfigurations 加载所有配置
func (a *App) LoadConfigurations() ([]config.DBConfig, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}
	return cfg.Connections, nil
}

// SaveConfiguration 保存配置
func (a *App) SaveConfiguration(dbConfig config.DBConfig) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	// 查找是否存在同名配置
	found := false
	for i, conn := range cfg.Connections {
		if conn.Name == dbConfig.Name {
			cfg.Connections[i] = dbConfig
			found = true
			break
		}
	}

	if !found {
		cfg.Connections = append(cfg.Connections, dbConfig)
	}

	return config.SaveConfig(cfg)
}

// DeleteConfiguration 删除配置
func (a *App) DeleteConfiguration(name string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	newConnections := make([]config.DBConfig, 0)
	for _, conn := range cfg.Connections {
		if conn.Name != name {
			newConnections = append(newConnections, conn)
		}
	}
	cfg.Connections = newConnections

	return config.SaveConfig(cfg)
}

// TestConnection 测试连接（包含 SSH 隧道）
func (a *App) TestConnection(dbConfig config.DBConfig) error {
	if dbConfig.UseSSH {
		// 创建 SSH 隧道
		tunnel, err := ssh.NewTunnel(
			dbConfig.SSH.Host,
			dbConfig.SSH.Port,
			dbConfig.SSH.Username,
			dbConfig.SSH.Password,
			dbConfig.SSH.PrivateKey,
			dbConfig.SSH.Passphrase,
			dbConfig.Host,
			dbConfig.Port,
		)
		if err != nil {
			return fmt.Errorf("create SSH tunnel error: %v", err)
		}

		// 启动隧道并获取本地端口
		localAddr, err := tunnel.Start()
		if err != nil {
			return fmt.Errorf("start SSH tunnel error: %v", err)
		}

		// 保存隧道实例以便后续使用
		a.tunnels[dbConfig.Name] = tunnel

		// 使用本地地址替换原始数据库地址
		dbConfig.Host = "127.0.0.1"
		_, dbConfig.Port = parseAddr(localAddr)
	}

	// 测试数据库连接
	// ... 实现数据库连接测试逻辑 ...

	return nil
}

func (a *App) ConnectDatabase(configName string) error {

	dbConfig, err2 := config.GetConfig(configName)
	if err2 != nil {
		return err2
	}
	var dsn string
	host := dbConfig.Host
	port := dbConfig.Port
	// 如果使用 SSH 隧道
	if dbConfig.UseSSH {
		tunnel, err := ssh.NewTunnel(
			dbConfig.SSH.Host,
			dbConfig.SSH.Port,
			dbConfig.SSH.Username,
			dbConfig.SSH.Password,
			dbConfig.SSH.PrivateKey,
			dbConfig.SSH.Passphrase,
			dbConfig.Host,
			dbConfig.Port,
		)
		if err != nil {
			runtime.EventsEmit(a.ctx, "db:"+dbConfig.Name+":error", err.Error())
			return fmt.Errorf("create SSH tunnel error: %v", err)
		}

		// 启动隧道
		localAddr, err := tunnel.Start()
		if err != nil {
			runtime.EventsEmit(a.ctx, "db:"+dbConfig.Name+":error", err.Error())
			return fmt.Errorf("start SSH tunnel error: %v", err)
		}

		a.tunnels[dbConfig.Name] = tunnel
		host, port = parseAddr(localAddr)
	}
	var db *sql.DB
	var err error
	var executor conn.Executor
	var handler conn.DDLHandler
	// 构建连接字符串
	switch dbConfig.Type {
	case "mysql":
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
			dbConfig.Username,
			dbConfig.Password,
			host,
			port,
			dbConfig.Database,
		)
		db, err = sql.Open(dbConfig.Type, dsn)
		executor = &conn.MysqlExecutor{}
		executor.SetConn(db)
		handler = &conn.MysqlHandler{}
		handler.SetConn(db)
	case "postgresql":
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			host,
			port,
			dbConfig.Username,
			dbConfig.Password,
			dbConfig.Database,
		)
		db, err = sql.Open(dbConfig.Type, dsn)
		executor = &conn.PostgresqlExecutor{}
		executor.SetConn(db)
		handler = &conn.PostgresqlHandler{}
		handler.SetConn(db)
	case "tdengine":
		dsn = fmt.Sprintf("%s:%s@http(%s:%d)/%s",
			dbConfig.Username,
			dbConfig.Password,
			host,
			port,
			dbConfig.Database,
		)
		// 建立数据库连接
		db, err = sql.Open("taosRestful", dsn)
		executor = &conn.TdEngineExecutor{}
		executor.SetConn(db)
		handler = &conn.TDHandler{}
		handler.SetConn(db)
	default:
		return fmt.Errorf("unsupported database type: %s", dbConfig.Type)
	}

	if err != nil {
		runtime.EventsEmit(a.ctx, "db:"+dbConfig.Name+":error", err.Error())
		return err
	}
	// 测试连接
	if err := db.Ping(); err != nil {
		_ = db.Close()
		executor = nil
		handler = nil
		runtime.EventsEmit(a.ctx, "db:"+dbConfig.Name+":error", err.Error())
		return err
	}
	a.executors[dbConfig.Name] = executor
	a.handlers[dbConfig.Name] = handler

	// 建立数据库连接成功后
	if dbConfig.Database != "" {
		a.current_db = dbConfig.Database
		// 假设我们有一个方法来获取所有表名
		tables, err := a.GetTables(dbConfig.Name, dbConfig.Database)
		if err != nil {
			runtime.EventsEmit(a.ctx, "db:"+dbConfig.Name+":error", err.Error())
			return err
		}
		runtime.EventsEmit(a.ctx, "db:"+dbConfig.Name+":tables", tables)
	} else {
		// 发送数据库名
		runtime.EventsEmit(a.ctx, "db:"+dbConfig.Name+":database", dbConfig.Database)
	}
	return nil
}

// 关闭数据库连接
func (a *App) DisconnectDatabase(name string) error {
	// 关闭数据库连接
	if db, ok := a.dbConns[name]; ok {
		db.Close()
		delete(a.dbConns, name)
	}

	// 关闭 SSH 隧道
	if tunnel, ok := a.tunnels[name]; ok {
		tunnel.Stop()
		delete(a.tunnels, name)
	}

	runtime.EventsEmit(a.ctx, "db:"+name+":status", "disconnected")
	return nil
}

// 执行查询
func (a *App) ExecuteQuery(name string, query string) error {
	db, ok := a.dbConns[name]
	if !ok {
		return fmt.Errorf("database connection not found: %s", name)
	}

	rows, err := db.Query(query)
	if err != nil {
		runtime.EventsEmit(a.ctx, "db:"+name+":error", err.Error())
		return err
	}
	defer rows.Close()

	// 获取列信息
	columns, err := rows.Columns()
	if err != nil {
		runtime.EventsEmit(a.ctx, "db:"+name+":error", err.Error())
		return err
	}

	// 准备结果
	var results []map[string]interface{}
	for rows.Next() {
		// 创建一个切片，用于存储每一行的数据
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		// 扫描当前行
		if err := rows.Scan(valuePtrs...); err != nil {
			runtime.EventsEmit(a.ctx, "db:"+name+":error", err.Error())
			return err
		}

		// 将当前行转换为 map
		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if val != nil {
				row[col] = val
			}
		}
		results = append(results, row)
	}

	// 发送查询结果
	runtime.EventsEmit(a.ctx, "db:"+name+":queryResult", results)
	return nil
}

func parseAddr(addr string) (string, int) {
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return "", 0
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return "", 0
	}
	return host, port
}

// GetDBs 获取数据库列表
func (a *App) GetDatabases(connectionName string) ([]string, error) {
	handler := a.handlers[connectionName]
	if handler == nil {
		return nil, fmt.Errorf("数据库连接未找到：%s", connectionName)
	}

	return handler.ShowDBs()
}

// GetTables 获取数据表列表
func (a *App) GetTables(connectionName, dbName string) ([]string, error) {
	handler := a.handlers[connectionName]
	if handler == nil {
		return nil, fmt.Errorf("数据库连接未找到：%s", connectionName)
	}
	a.current_db = dbName
	return handler.ShowTables(dbName)

}

// UseDatabase 切换数据库
func (a *App) UseDatabase(connectionName, dbName string) error {
	handler := a.handlers[connectionName]
	if handler == nil {
		return fmt.Errorf("数据库连接未找到：%s", connectionName)
	}
	a.current_db = dbName
	return handler.UseDB(dbName)

}

// getDBConfig 获取数据库配置
func (a *App) getDBConfig(connectionName string) (*config.DBConfig, error) {
	configs, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}

	for _, cfg := range configs.Connections {
		if cfg.Name == connectionName {
			return &cfg, nil
		}
	}

	return nil, fmt.Errorf("configuration not found for connection: %s", connectionName)
}

// GetTableStats 获取数据库表统计信息
func (a *App) GetTableStats(connectionName, dbName string) (*model.DatabaseStats, error) {
	handler := a.handlers[connectionName]
	if handler == nil {
		return nil, fmt.Errorf("数据库连接未找到：%s", connectionName)
	}

	return handler.GetTableStats(dbName)

}

// GetTableData 获取表数据（带分页）
func (a *App) GetTableData(connectionName string, tableName string, params model.TableDataParams) (*model.TableData, error) {
	// 使用 fmt.Sprintf 将所有信息组合成一个字符串
	logMessage := fmt.Sprintf("GetTableData: connectionName=%s, tableName=%s, params=%+v", connectionName, tableName, params)
	runtime.LogDebug(a.ctx, logMessage)

	executor := a.executors[connectionName]
	runtime.LogDebug(a.ctx, fmt.Sprintf("executors:%+v", a.executors))
	if executor == nil {
		return nil, fmt.Errorf("数据库连接未找到：%s", connectionName)
	}

	return executor.SelectPage(a.current_db, tableName, &params)

}

// GetTableStructure 获取表结构
func (a *App) GetTableStructure(connectionName string, dbName string, tableName string) ([]*model.ColumnInfo, error) {
	handler := a.handlers[connectionName]
	if handler == nil {
		return nil, fmt.Errorf("数据库连接未找到：%s", connectionName)
	}
	return handler.ShowTableStructure(dbName, tableName)
}

// formatSize 格式化文件大小
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
