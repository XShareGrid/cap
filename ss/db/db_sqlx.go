package db

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
)

// DisplayObject show object
func DisplayObject(obj interface{}) {
	js, err := json.MarshalIndent(obj, "", "    ")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%s\n", js)
}

type MySQLClientConfig struct {
	Driver   string
	Host     string
	Port     string
	Database string
	User     string
	Password string
}

func (cfg MySQLClientConfig) DataSource() string {
	if cfg.Host == "" {
		cfg.Host = "energy.itc-demo.xyz"
	}
	if cfg.Port == "" {
		cfg.Port = "3306"
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=true&allowNativePasswords=true", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database)
}

func (cfg MySQLClientConfig) DriverName() string {
	return cfg.Driver
}

var gRltCfg MySQLClientConfig

func SetRltCfg(rltcfg MySQLClientConfig) {
	gRltCfg = rltcfg
}

func GetRltCfg() MySQLClientConfig {
	return gRltCfg
}

// Set MySQL configure environment
func SetRltCfgToEnv(cfg MySQLClientConfig) {
	os.Setenv("CAPENV_MySQLDriver", cfg.Driver)
	os.Setenv("CAPENV_MySQLHost", cfg.Host)
	os.Setenv("CAPENV_MySQLPort", cfg.Port)
	os.Setenv("CAPENV_MySQLDatabase", cfg.Database)
	os.Setenv("CAPENV_MySQLUser", cfg.User)
	os.Setenv("CAPENV_MySQLPassword", cfg.Password)
}

// Get MySQL configure from environment
func GetRltCfgFromEnv() (cfg MySQLClientConfig) {
	cfg = MySQLClientConfig{
		Driver:   os.Getenv("CAPENV_MySQLDriver"),
		Host:     os.Getenv("CAPENV_MySQLHost"),
		Port:     os.Getenv("CAPENV_MySQLPort"),
		Database: os.Getenv("CAPENV_MySQLDatabase"),
		User:     os.Getenv("CAPENV_MySQLUser"),
		Password: os.Getenv("CAPENV_MySQLPassword"),
	}
	return cfg
}

// Show MySQL configure environment
func ShowRtlCfgEnv() {
	fmt.Printf("CAPENV_MySQLDriver=%s\n", os.Getenv("CAPENV_MySQLDriver"))
	fmt.Printf("CAPENV_MySQLHost=%s\n", os.Getenv("CAPENV_MySQLHost"))
	fmt.Printf("CAPENV_MySQLPort=%s\n", os.Getenv("CAPENV_MySQLPort"))
	fmt.Printf("CAPENV_MySQLDatabase=%s\n", os.Getenv("CAPENV_MySQLDatabase"))
	fmt.Printf("CAPENV_MySQLUser=%s\n", os.Getenv("CAPENV_MySQLUser"))
	fmt.Printf("CAPENV_MySQLPassword=%s\n", os.Getenv("CAPENV_MySQLPassword"))
}

// DB database connection
var DB *sqlx.DB

// TxTimout 事务超时的时间
var TxTimout = 60 * time.Second

// InitDB initialize database
func InitDB(config *MySQLClientConfig, maxConnection int) error {
	if DB != nil {
		return fmt.Errorf("db is already initialized")
	}
	db, err := sqlx.Open(config.DriverName(), config.DataSource())
	if err != nil {
		return err
	}
	db.SetMaxOpenConns(maxConnection)
	db.SetMaxIdleConns(maxConnection / 10)
	db.SetConnMaxLifetime(10 * time.Minute)
	DB = db
	return nil
}

// Rather than creating on init, this is created when necessary so that
// importers have time to customize the NameMapper.
var mpr *reflectx.Mapper

// mprMu protects mpr.
var mprMu sync.Mutex

// NameMapper is used to map column names to struct field names.  By default,
// it uses strings.ToLower to lowercase struct field names.  It can be set
// to whatever you want, but it is encouraged to be set before sqlx is used
// as name-to-field mappings are cached after first use on a type.
var NameMapper = strings.ToLower
var origMapper = reflect.ValueOf(NameMapper)

// mapper returns a valid mapper using the configured NameMapper func.
func mapper() *reflectx.Mapper {
	mprMu.Lock()
	defer mprMu.Unlock()

	if mpr == nil {
		mpr = reflectx.NewMapperFunc("json", sqlx.NameMapper)
	} else if origMapper != reflect.ValueOf(sqlx.NameMapper) {
		// if NameMapper has changed, create a new mapper
		mpr = reflectx.NewMapperFunc("json", sqlx.NameMapper)
		origMapper = reflect.ValueOf(sqlx.NameMapper)
	}
	return mpr
}

// InitDBV2 initialize database, use json tag
// type IndexRow struct {
// 	BeginTime time.Time  `json:"begin_time,omitempty"`
// 	EndTime   *time.Time `json:"end_time,omitempty"`
// 	Name      string     `json:"name,omitempty"`
// }
func InitDBV2(config *MySQLClientConfig, maxConnection int) error {
	if DB != nil {
		return fmt.Errorf("db is already initialized")
	}
	db, err := sqlx.Open(config.DriverName(), config.DataSource())
	if err != nil {
		return err
	}
	db.Mapper = mapper()
	db.SetMaxOpenConns(maxConnection)
	db.SetMaxIdleConns(maxConnection / 10)
	db.SetConnMaxLifetime(10 * time.Minute)
	DB = db
	return nil
}

func DeinitDB() error {
	if DB == nil {
		return fmt.Errorf("db is not initialized")
	}
	DB.Close()
	return nil
}

// GetSQLContext 创建事务
func GetSQLContext() (context.Context, error) {
	tx, err := DB.Beginx()
	if err != nil {
		return nil, err
	}

	ctx := NewSQLDatabaseContext(context.Background(), DB)
	ctx = NewSQLTxContext(ctx, tx)

	return ctx, nil
}

func GetSQLContextSaveToContext(ctx context.Context) (context.Context, error) {
	tx, err := DB.Beginx()
	if err != nil {
		return nil, err
	}
	ctx = NewSQLDatabaseContext(ctx, DB)
	ctx = NewSQLTxContext(ctx, tx)
	return ctx, nil
}

// GetSQLContextWithTimout 创建事务，超时将自动回滚
// 如果不设置超时时间，将默认使用全局超时参数TxTimout
// 全局超时时间设置TxTimeout变量，实时生效
func GetSQLContextWithTimout(timeout ...time.Duration) (context.Context, error) {
	to := TxTimout
	if len(timeout) > 0 {
		to = timeout[0]
	}
	ctx, err := GetSQLContext()
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
			CleanSQLContext(ctx, err)
		}
	}()

	ctx = context.WithValue(ctx, sqlCancelFuncCtxKey, cancel)
	return ctx, nil
}

func GetSQLContextWithTimoutSaveToContext(ctx context.Context, timeout ...time.Duration) (context.Context, error) {
	to := TxTimout
	if len(timeout) > 0 {
		to = timeout[0]
	}
	var err error
	ctx, err = GetSQLContextSaveToContext(ctx)
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
			CleanSQLContext(ctx, err)
		}
	}()
	ctx = context.WithValue(ctx, sqlCancelFuncCtxKey, cancel)
	return ctx, nil
}

// CleanSQLContext 完成事务，err != nil时将回滚
// 返回提交/回滚的结果
func CleanSQLContext(ctx context.Context, err error) error {
	cancel := ctx.Value(sqlCancelFuncCtxKey)
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
	tx, ok := GetSQLTxFromContext(ctx)
	if !ok {
		return fmt.Errorf("No tx in context")
	}

	if err != nil {
		return tx.Rollback()
	}

	err = tx.Commit()
	if err != nil {
		return tx.Rollback()
	}

	return nil
}

// CleanSQLContextByERR 完成事务，err != nil时将回滚
// 返回一个拼接的错误
func CleanSQLContextByERR(ctx context.Context, err error) error {
	ctxErr := CleanSQLContext(ctx, err)
	if ctxErr != nil {
		if err != nil {
			return fmt.Errorf(err.Error()+";"+CommitDataFailure, ctxErr.Error())
		}
		return fmt.Errorf(CommitDataFailure, ctxErr.Error())
	}
	return err
}
