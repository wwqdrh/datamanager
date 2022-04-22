package datamanager

import (
	"context"
	"errors"
	"os"

	"gorm.io/gorm"

	"github.com/ohko/hst"
	"github.com/wwqdrh/datamanager/app"
	"github.com/wwqdrh/datamanager/internal/datautil"
	"github.com/wwqdrh/datamanager/internal/driver"
	"github.com/wwqdrh/datamanager/runtime"
)

var R = runtime.Runtime
var Datamanager = new(dataManager) // 单例

type dataManager struct {
	ctx    context.Context
	cancel context.CancelFunc
}

// 初始化配置 但是不启动后台的读取与写入数据变更策略
// 可能会传入静态的表
func Initial(handler *hst.HST, opts ...runtime.RuntimeConfigOpt) *hst.HST {
	R.SetConfig(func() (*runtime.RuntimeConfig, error) {
		conf := &runtime.RuntimeConfig{}
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
		if conf.Handler == nil {
			conf.Handler = DefaultHandler
		}
		return conf, nil
	})

	return app.RegisterApi(handler)
}

// 启动后台服务，处理数据变更
// 初始化Initialize传入的表 需要初始化触发器
func (d *dataManager) Start(db *gorm.DB, ctx context.Context) error {
	if d.ctx != nil {
		return errors.New("已经启动过")
	}
	// 监听的数据库db
	if err := R.SetDB(func() (*driver.PostgresDriver, error) {
		if db != nil {
			R.GetConfig().DB = db
		}
		if R.GetConfig().DB == nil {
			return nil, errors.New("未传入db")
		}
		return new(driver.PostgresDriver).InitWithDB(R.GetConfig().DB), nil
	}); err != nil {
		return err
	}

	// leveldb
	if err := R.SetLogDB(func() (*driver.LevelDBDriver, error) {
		logPath := R.GetConfig().LogDataPath
		if err := os.MkdirAll(logPath, 0777); err != nil {
			return nil, err
		} // TODO: 需要显示传递文件路径否则重启日志不存在
		driver, err := driver.NewLevelDBDriver(logPath)
		if err != nil {
			return nil, err
		}
		return driver, err
	}); err != nil {
		return err
	}

	// 后台任务队列
	if err := R.SetBackQueue(func() (*datautil.Queue, error) {
		return datautil.NewQueue(), nil
	}); err != nil {
		return err
	}

	// 注册静态表
	for _, item := range R.GetConfig().RegisterTable {
		R.GetWatcher().Register(&item)
	}

	// 一个协程读一个协程写
	d.ctx, d.cancel = context.WithCancel(ctx)
	go func() {
		ch := R.GetWatcher().ListenAll()
		saver := R.GetLogSave()
		for {
			select {
			case item := <-ch:
				saver.Write(item)
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

func (d *dataManager) Stop() {
	d.cancel()
}