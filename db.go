package sapi

import (
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
)

var engine *xorm.Engine

func InitDb(user, pass, host, dbname string, done chan struct{}) error {
	var err error
	var dbAddress string

	dbAddress = fmt.Sprintf("%s:%s@tcp(%s:3306)/%s", user, pass, host, dbname)
	engine, err = xorm.NewEngine("mysql", dbAddress)
	if err != nil {
		return err
	}

	//keep db connection alive
	go func() {
		Seconds := time.NewTimer(time.Second * 1800)
		for {
			select {
			case <-Seconds.C:
				engine.Ping()
				Seconds.Reset(time.Second * 1800)
			case <-done:
				return
			}
		}
	}()

	return nil
}

func DB() *xorm.Engine {
	return engine
}

func Truncate(tables []string) error {
	for _, table := range tables {
		if _, err := engine.Exec(fmt.Sprintf("truncate table %s", table)); err != nil {
			return err
		}
	}
	return nil
}
