package main

import (
	"fmt"
	"testing"

	"github.com/XShareGrid/cap/msg/errors"
	"github.com/XShareGrid/cap/msg/errors/example/ue"
)

func Test_grpcFindSomething(t *testing.T) {
	err := grpcFindSomething(0)
	if err != nil {
		userErr := errors.Wrap(err)
		// 打印error dump
		fmt.Println("---------------------------0 error dump--------------------------")
		userErr.DumpErrors().Log()
		// ==> outputs: [0] sql: no rows in result set
		fmt.Println("---------------------------0 log--------------------------")
		userErr.Log()
		// ==> outputs: sql: no rows in result set
		fmt.Println("--------------------------0 stack trace---------------------------")
		userErr.PrintStackTrace()
		fmt.Println("-------------------------0 translate----------------------------")
		fmt.Println(userErr.TrError(ue.Lang_zhCN))
		// ==> outputs: 会话过期，请重新登陆 <nil>
	}

	err = grpcFindSomething(1)
	if err != nil {
		userErr := errors.Wrap(err)
		// 打印error dump
		fmt.Println("---------------------------1 error dump--------------------------")
		userErr.DumpErrors().Log()
		// ==> outputs:
		// [0] business layer error 0
		// triggered by:
		// [1] db error 0: 1, 2
		// triggered by:
		// [2] sql: no rows in result set
		fmt.Println("---------------------------1 log--------------------------")
		userErr.Log()
		// ==> outputs: business layer error 0
		fmt.Println("--------------------------1 stack trace---------------------------")
		userErr.PrintStackTrace()
		fmt.Println("-------------------------1 translate----------------------------")
		fmt.Println(userErr.TrError(ue.Lang_zhCN))
		// ==> outputs: 错误二 1, 123 <nil>
	}

	err = grpcFindSomething(2)
	if err != nil {
		userErr := errors.Wrap(err)
		// 打印error dump
		fmt.Println("---------------------------2 error dump--------------------------")
		userErr.DumpErrors().Log()
		// ==> outputs:
		// [0] business layer error 1
		fmt.Println("---------------------------2 log--------------------------")
		userErr.Log()
		// ==> outputs: business layer error 1
		fmt.Println("--------------------------2 stack trace---------------------------")
		userErr.PrintStackTrace()
		fmt.Println("-------------------------2 translate----------------------------")
		fmt.Println(userErr.TrError(ue.Lang_zhCN))
		// ==> outputs: 用户名或密码错误 <nil>
	}

	err = grpcFindSomething(3)
	if err != nil {
		userErr := errors.Wrap(err)
		// 打印error dump
		fmt.Println("---------------------------3 error dump--------------------------")
		userErr.DumpErrors().Log()
		// ==> outputs:
		// [0] golang error
		fmt.Println("---------------------------3 log--------------------------")
		userErr.Log()
		// ==> outputs: golang error
		fmt.Println("--------------------------3 stack trace---------------------------")
		userErr.PrintStackTrace()
		fmt.Println("-------------------------3 translate----------------------------")
		fmt.Println(userErr.TrError(ue.Lang_zhCN))
		// ==> 会话过期，请重新登陆 <nil>
	}
}
