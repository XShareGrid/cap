package main

import (
	"context"
	"fmt"

	"github.com/XShareGrid/cap/msg/errors"
	"github.com/XShareGrid/cap/msg/errors/handle"
)

// 预定义错误
var ErrInternal1 = errors.New("internal error 1")
var ErrInternal2 = errors.New("internal error 2 [%s]")
var ErrInternal3 = errors.New("internal error 3")

func init() {
	// 注册处理器
	handle.Register(ErrInternal1, func(ctx context.Context, err *errors.UserError) *errors.UserError {
		return err.FillIDAndArgs(ERR_INTERNAL_1)
	})

	handle.Register(ErrInternal2, func(ctx context.Context, err *errors.UserError) *errors.UserError {
		return err.FillIDAndArgs(ERR_INTERNAL_2, err.DegbugArg(0))
	})

}

func main() {
	// 一个普通错误处理
	err1 := ErrInternal1
	fmt.Println(handle.Handle(context.Background(), err1).TrError(Lang_zhCN))
	// 一个带参数的错误处理
	err2 := errors.Wrap(ErrInternal2).FillDebugArgs("test args")
	fmt.Println(handle.Handle(context.Background(), err2).TrError(Lang_zhCN))
	//
	err3 := errors.Wrap(ErrInternal3)
	fmt.Println(handle.Handle(context.Background(), err3).TrError(Lang_zhCN))
}
