package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/XShareGrid/cap/database/mysql"
	"github.com/XShareGrid/cap/table/demo/tables"
	"github.com/XShareGrid/cap/table/doc"
	cap "github.com/XShareGrid/cap/table/proto/go"
	"github.com/XShareGrid/cap/table/registry"
	"github.com/XShareGrid/cap/table/service"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"

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
	// 3. 启动表格文档服务
	{
		docPort := ":" + viper.GetString("app.docport")
		go doc.ServeHTTP(docPort, dbWrite, registry.GlobalTableRegistry())
		log.Println("Table doc server listen at", docPort)
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
	// 5. 启动gateway
	{
		gRPCPort := ":" + viper.GetString("app.grpcport")
		// 启动gateway
		conn, err := grpc.NewClient(
			gRPCPort,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			log.Fatalln("Failed to dial server:", err)
		}

		gwmux := runtime.NewServeMux(
			runtime.WithMarshalerOption(
				runtime.MIMEWildcard,
				&runtime.JSONPb{
					MarshalOptions: protojson.MarshalOptions{
						UseEnumNumbers:  true,
						EmitUnpopulated: true,
					},
				},
			),
		)
		httpPort := ":" + viper.GetString("app.httpport")
		cap.RegisterTableWServiceHandler(context.Background(), gwmux, conn)
		gwServer := &http.Server{
			Addr:    httpPort,
			Handler: gwmux,
		}
		go func() {
			log.Println("grpc gw listen at", httpPort)

			if err := gwServer.ListenAndServe(); err != nil {
				log.Fatal(err)
			}
		}()

	}

	/******************************************************************************/
	//  hold
	{
		fmt.Println("Demo is running")
		ch := make(chan int, 1)
		<-ch
	}

}
