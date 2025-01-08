package t

import (
	"fmt"
	"runtime"
	"time"

	"github.com/XShareGrid/cap/ss/db"
)

func init() {
	if runtime.GOOS == "windows" {
		dbcfg := &db.MySQLClientConfig{
			Driver:   "mysql",
			Host:     "127.0.0.1",
			Port:     "3306",
			Database: "txtest",
			User:     "root",
			Password: "125801",
		}

		err := db.InitDB(dbcfg, 100)
		if err != nil {
			panic(err)
		}
		err = db.InitGDB(dbcfg, 100)
		if err != nil {
			panic(err)
		}
		db.TxTimout = 5 * time.Second
	}
}

func testSQLTxTimeout() {
	ctx, err := db.GetSQLContextWithTimout()
	if err != nil {
		panic(err)
	}
	tx, _ := db.GetSQLTxFromContext(ctx)
	r, err := tx.Exec("UPDATE box SET serial_num = '1' WHERE serial_num = '2';")
	if err != nil {
		panic(err)
	}
	fmt.Println(r.RowsAffected())
	time.Sleep(time.Minute)
}

func testSQLTxPanic() {
	ctx, err := db.GetSQLContextWithTimout()
	if err != nil {
		panic(err)
	}
	defer db.CleanContext(ctx, nil)

	tx, _ := db.GetSQLTxFromContext(ctx)
	r, err := tx.Exec("UPDATE box SET serial_num = '1' WHERE serial_num = '2';")
	if err != nil {
		panic(err)
	}
	fmt.Println(r.RowsAffected())
	panic("stop here")
	time.Sleep(time.Minute)
}

func testSQLTxTimeoutWithClean() {
	ctx, err := db.GetSQLContextWithTimout()
	if err != nil {
		panic(err)
	}
	tx, _ := db.GetSQLTxFromContext(ctx)
	r, err := tx.Exec("UPDATE box SET serial_num = '1' WHERE serial_num = '2';")
	if err != nil {
		panic(err)
	}
	fmt.Println(r.RowsAffected())
	db.CleanContext(ctx, nil)
	time.Sleep(time.Minute)
}
