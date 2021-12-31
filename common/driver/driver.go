package driver

import "datamanager/pkg/plugger/postgres"

func InitDriver(targetDB *postgres.PostgresConfig) (errs []error) {
	// 目标库 postgres
	if err := InitPostgresDriver(targetDB); err != nil {
		errs = append(errs, err)
	}
	// 配置库 sqlite3
	if err := InitSqliteDriver(); err != nil {
		errs = append(errs, err)
	}
	// 日志存储库
	if err := InitLevelDBDriver(); err != nil {
		errs = append(errs, err)
	}
	return
}
