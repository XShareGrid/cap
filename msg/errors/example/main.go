package main

// import "github.com/XShareGrid/cap/msg/errors/example/ue"

import (
	"fmt"

	"github.com/XShareGrid/cap/logger"
	"github.com/XShareGrid/cap/msg/errors"
	"github.com/XShareGrid/cap/msg/errors/example/ue"
)

func main() {
	// 自定义logger输出
	errors.SetLogger(func(fmt string, v ...interface{}) {
		logger.CAP.Error(fmt, v...)
	})

	// 创建一个普通的Error
	err1 := errors.New("错误错误")
	fmt.Println(err1.Error()) // output: 错误错误
	if userErr, ok := err1.(*errors.UserError); ok {
		// 打印调用栈
		userErr.PrintStackTrace()
		// 翻译至目标语言
		fmt.Println(userErr.TrError(ue.Lang_zhCN)) // output: 错误错误 <nil>
	}

	fmt.Println("===========================================================")

	// 创建一个用户错误
	err2 := errors.NewUser(ue.ERR_INVALID_USER_NAME_PASSWORD)
	fmt.Println(err2.Error()) // output: 空
	if userErr, ok := err2.(*errors.UserError); ok {
		// 打印调用栈
		userErr.PrintStackTrace()
		// 翻译至目标语言
		fmt.Println(userErr.TrError(ue.Lang_zhCN)) // output: 用户名或密码错误 <nil>
		fmt.Println(userErr.TrError(ue.Lang_enUS)) // output: Invalid username or password <nil>
	}

	fmt.Println("===========================================================")

	// 创建一个用户错误，带参数
	err3 := errors.NewUser(ue.ERR_TEST_2, 123, "测试")
	fmt.Println(err3.Error()) // output: 空
	if userErr, ok := err3.(*errors.UserError); ok {
		// 打印调用栈
		userErr.PrintStackTrace()
		// 翻译至目标语言
		fmt.Println(userErr.TrError(ue.Lang_zhCN)) // output: 错误二 123, 测试 <nil>
		fmt.Println(userErr.TrError(ue.Lang_enUS)) // output: Error two 123, 测试 <nil>
	}

	fmt.Println("===========================================================")

	// 创建一个带调试字符串的用户错误
	err4 := errors.NewUserD("no rows in result set", ue.ERR_INVALID_USER_NAME_PASSWORD)
	fmt.Println(err4.Error()) // output: no rows in result set
	if userErr, ok := err2.(*errors.UserError); ok {
		// 打印调用栈
		userErr.PrintStackTrace()
		// 翻译至目标语言
		fmt.Println(userErr.TrError(ue.Lang_zhCN)) // output: 用户名或密码错误 <nil>
		fmt.Println(userErr.TrError(ue.Lang_enUS)) // output: Invalid username or password <nil>
	}

	fmt.Println("===========================================================")
}
