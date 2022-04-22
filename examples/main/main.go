package main

import (
	"context"
	"log"

	"github.com/wwqdrh/datamanager"
	"github.com/wwqdrh/datamanager/runtime"

	"time"

	system_model "github.com/wwqdrh/datamanager/examples/main/system"

	"github.com/wwqdrh/datamanager/examples/main/base"

	"github.com/ohko/hst"
)

type backend struct{}

// Start ...
func Start(h *hst.HST, middlewares ...hst.HandlerFunc) *hst.HST {
	// datamanager.InitService(system_model.DB())
	// datamanager.SetCustom()

	go func() {
		ss := base.Services()
		for _, s := range ss {
			if err := s.Register(); err != nil {
				panic(s.Name() + " register: " + err.Error())
			}
		}
	}()

	// 启动一个清理过期日志的go程
	go func() {
		for {
			system_model.DBServiceLog.ClearTimeout()
			time.Sleep(time.Hour)
		}
	}()
	// h = app.RegisterApi(h, middlewares...)

	return h
}

type Company struct {
	ID      int `gorm:"PRIMARY_KEY;AUTO_INCREMENT"`
	Name    string
	Age     int
	Address string
	Salary  int
}

type CompanyRela struct {
	ID        int `gorm:"PRIMARY_KEY;AUTO_INCREMENT"`
	CompanyID int
	Salary    int
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// system_model.InitDB("postgres", "host=172.18.3.9 user=postgres password=postgres dbname=hui_test port=5435 sslmode=disable TimeZone=Asia/Shanghai")
	system_model.InitDB("postgres", "host=localhost user=postgres password=123456 dbname=hui_dm_test port=5432 sslmode=disable TimeZone=Asia/Shanghai")

	h := hst.New(nil)
	Start(h)

	datamanager.Initial(h, func(rc *runtime.RuntimeConfig) {
		rc.LogDataPath = "./version"
		rc.MinLogNum = 10
		rc.ReadPolicy = "notify"
		rc.WritePolicy = "leveldb"
		rc.DB = system_model.DB()
		rc.RegisterTable = []runtime.TablePolicy{
			{
				Table:     Company{},
				RelaField: "id",
				Relations: "company_rela.company_id",
			}, {
				Table:     &CompanyRela{},
				RelaField: "company_id",
				Relations: "company.id",
			},
		}
	})
	datamanager.Start(nil, ctx)
	err := h.ListenHTTP(":8090")
	log.Println("exit:", err)
}
