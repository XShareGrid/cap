package db

import (
	"fmt"
	"os"
	"runtime"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// for windows test
func init() {
	if runtime.GOOS == "windows" {
		dbcfg := &MySQLClientConfig{
			Driver:   "mysql",
			Host:     "energy.itc-demo.xyz",
			Port:     "3306",
			Database: "mts",
			User:     "energy",
			Password: "QQww1234",
		}

		err := InitDB(dbcfg, 100)
		if err != nil {
			panic(err)
		}
	}
}
func GetSQLURL() string {
	return os.Getenv("CLSERPENV_MySQLURL")
}

func GetWinSQLDatabase() string {
	return os.Getenv("CLSERPENV_MySQLDatabase")
}

func GetSQLDriver() string {
	return os.Getenv("CLSERPENV_MySQLDriver")
}

func TestSQLContext(t *testing.T) {
	ctx, err := GetSQLContext()
	if err != nil {
		panic(err)
	}
	defer CleanSQLContext(ctx, err)
}

func SQLContextHaTest() {
	sleepTime := 5
	fmt.Println("create CTX.")
	ctx, err := GetSQLContext()
	if err != nil {
		panic(err)
	}
	fmt.Println("start sleep: ", sleepTime, "秒")
	time.Sleep(time.Duration(sleepTime) * time.Second)
	defer CleanSQLContext(ctx, err)
	fmt.Println("clean CTX.")
	fmt.Println()
}

func BenchmarkGetSQLContext(b *testing.B) {
	for i := 0; i < 1; i++ {
		sleepTime := 3
		SQLContextHaTest()
		fmt.Println("start sleep: ", sleepTime, "秒")
		time.Sleep(time.Duration(sleepTime) * time.Second)
		fmt.Println("-=-=-=-=-=-=-=-=-=-=-=-=-=END=-=-=-=-=-=-=-=-=-=-=-=-=-=-")
		fmt.Println()
	}
}
