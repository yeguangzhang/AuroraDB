package model // TableStats 表统计信息
type TableStats struct {
	TableName   string `json:"tableName"`
	RecordCount int    `json:"recordCount"`
}

// DatabaseStats 数据库统计信息
type DatabaseStats struct {
	TotalRecords int          `json:"totalRecords"`
	TableStats   []TableStats `json:"tableStats"`
}

// TableDataParams 表格数据参数
type TableDataParams struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
}

// TableData 表格数据结构
type TableData struct {
	Columns []Column `json:"columns"`
	Data    [][]any  `json:"data"`
	Total   int64    `json:"total"`
}

// Column 表格列定义
type Column struct {
	Title string `json:"title"` // 列标题
	Key   string `json:"key"`   // 列键名
}

// ColumnInfo 表字段信息
type ColumnInfo struct {
	Name     string `json:"name"`     // 字段名
	Type     string `json:"type"`     // 字段类型
	Nullable string `json:"nullable"` // 是否可为空
	Default  string `json:"default"`  // 默认值
	Comment  string `json:"comment"`  // 字段注释
}
