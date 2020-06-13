package db

import (
	"config"
	"database/sql"
	"fmt"
	"github.com/astaxie/beego/logs"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

const (
	USERNAME = "root"
	PASSWORD = "553fb4cfaa"
	DATABASE = "chess_stat"
)

func InitDB() {
	var err error
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s", USERNAME, PASSWORD, config.Get().Connect.Mysql, DATABASE)
	g_mysql, err = sql.Open("mysql", dsn)
	if err != nil {
		logs.Error("Open mysql failed, err: %v\n", err)
		return
	}
	g_mysql.SetConnMaxLifetime(100 * time.Second) //最大连接周期，超过时间的连接就close
	g_mysql.SetMaxOpenConns(100)                  //设置最大连接数
	g_mysql.SetMaxIdleConns(16)                   //设置闲置连接数
	err = g_mysql.Ping()
	if err != nil {
		logs.Error("Connect mysql failed, err: %v\n", err)
	}
}

func GetMySql() *sql.DB {
	return g_mysql
}

var g_mysql *sql.DB
