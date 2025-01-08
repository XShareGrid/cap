package main

import (
	"database/sql"
	"fmt"

	"github.com/XShareGrid/cap/msg/errors"
	"github.com/XShareGrid/cap/msg/errors/example/ue"
)

func wrapError() {

}

// Layer DB

// 预定义

var ErrLayerDB0 = errors.New("db error 0: %d, %s")
var ErrLayerDB1 = errors.New("db error 1")

func findSomethingFromDB(p int) error {
	if p == 0 {
		// 直接返回底层的错误
		err := sql.ErrNoRows
		return errors.Wrap(err)
	} else if p == 1 {
		// 底层错误触发了新的错误
		err := sql.ErrNoRows
		return errors.Wrap(err).Triggers(ErrLayerDB0).FillDebugArgs(1, "2").Log()
	} else if p == 2 {
		// 包装一个原生错误
		return errors.Wrap(ErrLayerDB1)
	} else if p == 3 {
		// 直接返回一个原生的错误
		return fmt.Errorf("golang error")
	}
	return nil
}

// Layer Business

// 预定义

var ErrLayerBiz0 = errors.New("business layer error 0")
var ErrLayerBiz1 = errors.New("business layer error 1")

func findSomething(p int) error {
	err := findSomethingFromDB(p)
	if errors.Match(err, ErrLayerDB0) {
		// 下层错误触发了一个新的错误
		return errors.Wrap(err).Triggers(ErrLayerBiz0).Log()
	} else if errors.Match(err, ErrLayerDB1) {
		// 返回一个新的错误
		return errors.Wrap(ErrLayerBiz1)
	}
	// 未处理，直接返回
	return err
}

// Layer interface

func grpcFindSomething(p int) error {
	err := findSomething(p)
	// 判断错误
	if errors.Match(err, ErrLayerBiz0) {
		// 将下层错误附加用户错误的ID与参数
		return errors.Wrap(err).FillIDAndArgs(ue.ERR_TEST_2, 1, "123").Log()
	} else if errors.Match(err, ErrLayerBiz1) {
		// 将下层错误附加用户错误的ID
		return errors.Wrap(err).FillIDAndArgs(ue.ERR_INVALID_USER_NAME_PASSWORD).Log()
	}
	return errors.Wrap(err).FillIDAndArgs(ue.ERR_SESSION_EXPIRED).Log()
}
