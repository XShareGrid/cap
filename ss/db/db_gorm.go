package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jinzhu/gorm"
)

// NewGDatabaseContext create a sql database context
func NewGDatabaseContext(ctx context.Context, db *gorm.DB) context.Context {
	return context.WithValue(ctx, gormDatabaseCtxKey, db)
}

// GetGDatabaseFromContext get sql database from context
func GetGDatabaseFromContext(ctx context.Context) (db *gorm.DB, ok bool) {
	db, ok = ctx.Value(gormDatabaseCtxKey).(*gorm.DB)
	return db, ok
}

// NewGTxContext create sql tx context
func NewGTxContext(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, gormTxCtxKey, tx)
}

// GetGTxFromContext get sql tx from context
func GetGTxFromContext(ctx context.Context) (tx *gorm.DB, ok bool) {
	tx, ok = ctx.Value(gormTxCtxKey).(*gorm.DB)
	return tx, ok
}

// GORMDB database connection
var GORMDB *gorm.DB

// InitGDB 初始化GORMDB
func InitGDB(config *MySQLClientConfig, maxConnection int) error {
	if GORMDB != nil {
		return fmt.Errorf("db is already initialized")
	}
	db, err := gorm.Open(config.DriverName(), config.DataSource())
	if err != nil {
		panic("failed to connect database")
	}
	db.LogMode(true)
	db.SingularTable(true)
	db.DB().SetMaxOpenConns(maxConnection)
	db.DB().SetMaxIdleConns(maxConnection / 10)
	db.DB().SetConnMaxLifetime(10 * time.Minute)
	GORMDB = db
	return nil
}

// DeinitGDB 清除GORMDB
func DeinitGDB() error {
	if DB == nil {
		return fmt.Errorf("db is not initialized")
	}
	DB.Close()
	return nil
}

// GetGContext 获取GORM上下文
func GetGContext() (context.Context, error) {
	tx := GORMDB.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	ctx := NewGDatabaseContext(context.Background(), GORMDB)
	ctx = NewGTxContext(ctx, tx)

	return ctx, nil
}

func GetGContextToContext(ctx context.Context) (context.Context, error) {
	tx := GORMDB.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	ctx = NewGDatabaseContext(ctx, GORMDB)
	ctx = NewGTxContext(ctx, tx)

	return ctx, nil
}

// GetGContextWithTimout 创建事务，超时将自动回滚
// 如果不设置超时时间，将默认使用全局超时参数TxTimout
// 全局超时时间设置TxTimeout变量，实时生效
func GetGContextWithTimout(timeout ...time.Duration) (context.Context, error) {
	to := TxTimout
	if len(timeout) > 0 {
		to = timeout[0]
	}
	ctx, err := GetGContext()
	if err != nil {
		return ctx, err
	}

	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, to)
	go func() {
		<-ctx.Done()
		err := ctx.Err()
		// 超时处理
		if err == context.DeadlineExceeded {
			log.Println("tx timeout, cleaned")
			CleanGContext(ctx, err)
		}
	}()

	ctx = context.WithValue(ctx, gormCancelFuncCtxKey, cancel)
	return ctx, nil
}

func GetGContextWithTimoutToContext(ctx context.Context, timeout ...time.Duration) (context.Context, error) {
	to := TxTimout
	if len(timeout) > 0 {
		to = timeout[0]
	}
	var err error
	ctx, err = GetGContextToContext(ctx)
	if err != nil {
		return ctx, err
	}

	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, to)
	go func() {
		<-ctx.Done()
		err := ctx.Err()
		// 超时处理
		if err == context.DeadlineExceeded {
			log.Println("tx timeout, cleaned")
			CleanGContext(ctx, err)
		}
	}()

	ctx = context.WithValue(ctx, gormCancelFuncCtxKey, cancel)
	return ctx, nil
}

// CleanGContext 完成事务，err != nil时将回滚
// 返回提交/回滚的结果
func CleanGContext(ctx context.Context, err error) error {
	cancel := ctx.Value(gormCancelFuncCtxKey)
	if cancel != nil {
		if cancelFunc, ok := cancel.(context.CancelFunc); ok {
			cancelFunc()
			log.Println("ctx canceled")
		}
	}
	if r := recover(); r != nil {
		log.Println("recovered from panic")
		err = fmt.Errorf("recovered from panic [%v]", r)
	}
	tx, ok := GetGTxFromContext(ctx)
	if !ok {
		return fmt.Errorf("No tx in context")
	}

	tx.Error = nil
	if err != nil {
		return tx.Rollback().Error
	}

	err = tx.Commit().Error
	if err != nil {
		return tx.Rollback().Error
	}

	return nil
}

// CleanGContextByERR 完成事务，err != nil时将回滚
// 返回一个拼接的错误
func CleanGContextByERR(ctx context.Context, err error) error {
	ctxErr := CleanGContext(ctx, err)
	if ctxErr != nil {
		if err != nil {
			return fmt.Errorf(err.Error()+";"+CommitDataFailure, ctxErr.Error())
		}
		return fmt.Errorf(CommitDataFailure, ctxErr.Error())
	}
	return err
}
