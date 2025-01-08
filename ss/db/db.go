package db

import (
	"context"
	"fmt"
	"log"
	"time"
)

//sqlx
type ctxkey int

var sqlCancelFuncCtxKey ctxkey = 100
var gormCancelFuncCtxKey ctxkey = 101

const sqlDatabaseCtxKey ctxkey = 0
const sqlTxCtxKey ctxkey = 1

const gormDatabaseCtxKey ctxkey = 2
const gormTxCtxKey ctxkey = 3

// GetContextWithTimout 创建context, 同时包含sqlx和gorm的连接, 只能使用CleanContext来清理
func GetContextWithTimout(timeout ...time.Duration) (context.Context, error) {
	to := TxTimout
	if len(timeout) > 0 {
		to = timeout[0]
	}
	// 创建tx
	tx1 := GORMDB.Begin()
	if tx1.Error != nil {
		return nil, tx1.Error
	}
	ctx := NewGDatabaseContext(context.Background(), GORMDB)
	ctx = NewGTxContext(ctx, tx1)

	tx2, err := DB.Beginx()
	if err != nil {
		tx1.Rollback()
		return nil, err
	}

	ctx = NewSQLDatabaseContext(ctx, DB)
	ctx = NewSQLTxContext(ctx, tx2)
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, to)
	go func() {
		<-ctx.Done()
		err := ctx.Err()
		// 超时处理
		if err == context.DeadlineExceeded {
			log.Println("tx timeout, cleaned")
			CleanContext(ctx, err)
		}
	}()

	ctx = context.WithValue(ctx, sqlCancelFuncCtxKey, cancel)
	return ctx, nil
}

func GetContextWithTimoutSaveToContext(ctx context.Context, timeout ...time.Duration) (context.Context, error) {
	to := TxTimout
	if len(timeout) > 0 {
		to = timeout[0]
	}
	// 创建tx
	tx1 := GORMDB.Begin()
	if tx1.Error != nil {
		return nil, tx1.Error
	}
	ctx = NewGDatabaseContext(ctx, GORMDB)
	ctx = NewGTxContext(ctx, tx1)

	tx2, err := DB.Beginx()
	if err != nil {
		tx1.Rollback()
		return nil, err
	}

	ctx = NewSQLDatabaseContext(ctx, DB)
	ctx = NewSQLTxContext(ctx, tx2)
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, to)
	go func() {
		<-ctx.Done()
		err := ctx.Err()
		// 超时处理
		if err == context.DeadlineExceeded {
			log.Println("tx timeout, cleaned")
			CleanContext(ctx, err)
		}
	}()

	ctx = context.WithValue(ctx, sqlCancelFuncCtxKey, cancel)
	return ctx, nil
}

// CleanContext 统一的Clean接口，支持所有类型的ctx
func CleanContext(ctx context.Context, err error) error {
	if r := recover(); r != nil {
		log.Println("recovered from panic")
		err = fmt.Errorf("recovered from panic [%v]", r)
	}
	var sqlErr, gErr error
	_, ok := GetSQLTxFromContext(ctx)
	if ok {
		sqlErr = CleanSQLContext(ctx, err)
	}
	_, ok = GetGTxFromContext(ctx)
	if ok {
		gErr = CleanGContext(ctx, err)
	}
	if sqlErr != nil || gErr != nil {
		return fmt.Errorf("%v, %v", sqlErr, gErr)
	}
	return err
}

// CleanContextByERR 统一的Clean接口，支持所有类型的ctx
func CleanContextByERR(ctx context.Context, err error) error {
	ctxErr := CleanContext(ctx, err)
	if ctxErr != nil {
		if err != nil {
			return fmt.Errorf(err.Error()+";"+CommitDataFailure, ctxErr.Error())
		}
		return fmt.Errorf(CommitDataFailure, ctxErr.Error())
	}
	return err
}
