package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/XShareGrid/cap/database/mysql"
	"github.com/XShareGrid/cap/rproxy"
	"github.com/XShareGrid/cap/table/demo/tables"
	"github.com/XShareGrid/cap/table/doc"
	cap "github.com/XShareGrid/cap/table/proto/go"
	"github.com/XShareGrid/cap/table/registry"
	"github.com/XShareGrid/cap/table/service"
	"google.golang.org/grpc"

	// mysql driver
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
)

var dbRead *mysql.DB
var dbWrite *mysql.DB

func main() {
	// 0. Read configuration file
	{
		f := flag.String("f", "etc/app.yml", "config file path")
		flag.Parse()
		log.Println("Configuration file:", *f)
		viper.SetConfigFile(*f)
		if err := viper.ReadInConfig(); err != nil {
			log.Fatal(err)
		}
	}
	/******************************************************************************/
	// 1. 数据库连接
	{
		mysql.DebugLogOn = true
		db, err := mysql.NewDatabase(&mysql.ConnConfig{
			Driver:          "mysql",
			Host:            viper.GetString("mysql.host"),
			Port:            viper.GetString("mysql.port"),
			Database:        viper.GetString("mysql.database"),
			User:            viper.GetString("mysql.user"),
			Password:        viper.GetString("mysql.password"),
			MaxOpenConns:    200,
			MaxIdelConns:    15,
			ConnMaxLifeTime: 5 * time.Second,
			ConnMaxIdelTime: 5 * time.Second,
		})
		if err != nil {
			panic(err)
		}
		// 不用读写分离
		dbRead, dbWrite = db, db
	}
	/******************************************************************************/
	// 2. 注册表格数据
	{
		tables.Init()
	}
	// 启动表格文档服务
	{
		docPort := ":" + viper.GetString("app.docport")
		go doc.ServeHTTP(docPort, dbWrite, registry.GlobalTableRegistry())
		log.Println("Table doc server listen at", docPort)
	}
	/******************************************************************************/
	// 3. 启动gRPC代理
	{

		httpPort := ":" + viper.GetString("app.httpport")
		// GRPC端口
		rproxy.GRPC.CreateOrUpdateRule("/cap.TableWService", "localhost:"+viper.GetString("app.grpcport"), true)
		// HTTP端口
		rproxy.HTTP.CreateOrUpdateRule("/api/v1", "localhost:xxxxx")
		// 对外端口
		go rproxy.StartProxy(httpPort)
		log.Println("Proxy listen at", httpPort)
	}
	/******************************************************************************/
	// 4. 启动表格服务
	{
		gRPCPort := ":" + viper.GetString("app.grpcport")
		lis, err := net.Listen("tcp", gRPCPort)
		if err != nil {
			log.Fatalf("failed to listen at [%s]: %s", gRPCPort, err.Error())
		}
		server := grpc.NewServer()

		cap.RegisterTableWServiceServer(server,
			service.NewTableWService(dbWrite, dbRead, &TestUserInfoProvider{}))
		go func() {
			log.Println("gRPC listen at", gRPCPort)
			if err := server.Serve(lis); err != nil {
				log.Fatal(err)
			}
		}()
	}
	/******************************************************************************/
	// 5. hold
	{
		fmt.Println("Demo is running")
		ch := make(chan int, 1)
		<-ch
	}

}
