package runtime

import (
	"github.com/wwqdrh/datamanager/internal/pgwatcher"
	"gorm.io/gorm"
)

type (
	IStructHandler interface {
		GetTables() []*Table
		GetFields(string) []*Fields
		GetFieldName(string, string) string // 通过表名，字段id 获取字段名
		GetTableName(string) string         // 通过表id获取表名字
	}

	Table struct {
		TableID   string `json:"table_id"`
		TableName string `json:"table_name"`
		IsListen  bool   `json:"is_listen"`
	}

	Fields struct {
		FieldID   string `json:"field_id"`
		FieldName string `json:"field_name"`
	}
)

type TablePolicy struct {
	Table        interface{}
	MinLogNum    int
	Outdate      int
	RelaField    string
	Relations    string
	SenseFields  []string
	IgnoreFields []string
}

type RuntimeConfig struct {
	Outdate       int                     // 保存的记录时间
	MinLogNum     int                     // 最少保留的日志数
	TempLogTable  string                  // 临时日志名字
	PerReadNum    int                     // 一次读取多少条
	ReadPolicy    string                  // 读取的策略
	WritePolicy   string                  // 保存的策略
	LogDataPath   string                  // 保存的存储记录位置
	DB            *gorm.DB                // 可以外部调用者自己传递
	Handler       IStructHandler          // 表字段的映射
	RegisterTable []pgwatcher.TablePolicy // 初始化注册的静态监听的表
}

type RuntimeConfigOpt = func(*RuntimeConfig)

func NewRuntimeConfig(opts ...RuntimeConfigOpt) *RuntimeConfig {
	conf := &RuntimeConfig{}
	for _, opt := range opts {
		opt(conf)
	}

	if conf.Outdate <= 0 {
		conf.Outdate = 10
	}
	if conf.MinLogNum <= 0 {
		conf.MinLogNum = 10
	}
	if conf.TempLogTable == "" {
		conf.TempLogTable = "_action_log"
	}
	if conf.PerReadNum <= 0 {
		conf.PerReadNum = 1000
	}
	if conf.ReadPolicy == "" {
		conf.ReadPolicy = "trigger"
	}
	if conf.WritePolicy == "" {
		conf.WritePolicy = "leveldb"
	}
	if conf.LogDataPath == "" {
		conf.LogDataPath = "./version"
	}

	return conf
}