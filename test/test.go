package test

import (
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/XShareGrid/cap/ss/db"
)

func InitMySQL() {
	dbcfg := &db.MySQLClientConfig{
		Driver:   "mysql",
		Host:     os.Getenv("TEST_MYSQL_ADDR"),
		Port:     os.Getenv("TEST_MYSQL_PORT"),
		Database: os.Getenv("TEST_MYSQL_DB"),
		User:     os.Getenv("TEST_MYSQL_USER"),
		Password: os.Getenv("TEST_MYSQL_PWD"),
	}

	err := db.InitDB(dbcfg, 100)
	if err != nil {
		panic(err)
	}
}

func DisplayObject(obj interface{}) {
	js, err := json.MarshalIndent(obj, "", "    ")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%s\n", js)
}

// IsInTests 判断是否在单元测试中
func IsInTests() bool {
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "-test.") {
			return true
		}
	}
	return false
}
