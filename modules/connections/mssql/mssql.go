package mssql

import (
	"database/sql"
	"sync"
	"goAdmin/modules/config"
	"net/url"
	"fmt"
	"goAdmin/modules/connections/performer"
	_ "github.com/denisenkom/go-mssqldb"
)

type Mssql struct {
	SqlDBmap map[string]*sql.DB
	Once     sync.Once
}

var DB = Mssql{
	SqlDBmap: map[string]*sql.DB{},
}

func GetMssqlDB() *Mssql {
	return &DB
}

func (db *Mssql) Query(query string, args ...interface{}) ([]map[string]interface{}, *sql.Rows) {
	return performer.Query(db.SqlDBmap["default"], query, args...)
}

func (db *Mssql) Exec(query string, args ...interface{}) sql.Result {
	return performer.Exec(db.SqlDBmap["default"], query, args...)
}

func (db *Mssql) InitDB(cfglist map[string]config.Database) {
	db.Once.Do(func() {
		var (
			err      error
			SqlDB   *sql.DB
		)

		for conn, cfg := range cfglist {

			u := &url.URL{
				Scheme:   "sqlserver",
				User:     url.UserPassword(cfg.USER, cfg.PWD),
				Host:     fmt.Sprintf("%s:%s", cfg.IP, cfg.PORT),
			}

			SqlDB, err = sql.Open("sqlserver", u.String())

			if err != nil {
				SqlDB.Close()
				panic(err.Error())
			} else {
				// 设置数据库最大连接 减少timewait 正式环境调大
				SqlDB.SetMaxIdleConns(cfg.MAX_IDLE_CON) // 连接池连接数 = mysql最大连接数/2
				SqlDB.SetMaxOpenConns(cfg.MAX_OPEN_CON) // 最大打开连接 = mysql最大连接数

				db.SqlDBmap[conn] = SqlDB
			}
		}
	})
}


