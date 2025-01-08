package mysql

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
)

// DB database object
type DB struct {
	*sqlx.DB
	InstanceID string
	mm         mapperMgr
	Debug      bool
}

// TxTimout 事务超时的时间
var TxTimout = 10 * time.Minute

// NewSession ...
func (db *DB) NewSession(timeout ...time.Duration) (*Session, error) {
	to := TxTimout
	if len(timeout) > 0 {
		to = timeout[0]
	}
	tx, err := db.Beginx()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), to)
	ss := &Session{Tx: *tx, cancelFunc: cancel, db: db, hasTx: true}
	go func() {
		<-ctx.Done()
		err := ctx.Err()
		// 超时处理
		if err == context.DeadlineExceeded {
			log.Printf("session timeout after %d seconds\n", int(to.Seconds()))
			ss.Close(ErrSessionTimeout)
		}
		if err == context.Canceled {
			log.Printf("session canceled")
			ss.Close(ErrSessionCanceled)
		}
	}()
	return ss, nil
}

// AsSession DB As Session, no tx
func (db *DB) AsSession(ctx context.Context, timeout ...time.Duration) (*Session, error) {
	to := TxTimout
	if len(timeout) > 0 {
		to = timeout[0]
	}

	ctx, cancel := context.WithTimeout(ctx, to)
	ss := &Session{cancelFunc: cancel, db: db}
	go func() {
		<-ctx.Done()
		err := ctx.Err()
		// 超时处理
		if err == context.DeadlineExceeded {
			log.Printf("session timeout after %d seconds\n", int(to.Seconds()))
			ss.Close(ErrSessionTimeout)
		}
	}()
	return ss, nil
}

// NewSession ...
func (db *DB) NewSessionWithCtx(ctx context.Context, timeout ...time.Duration) (*Session, error) {
	to := TxTimout
	if len(timeout) > 0 {
		to = timeout[0]
	}
	tx, err := db.Beginx()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, to)
	ss := &Session{Tx: *tx, cancelFunc: cancel, db: db, hasTx: true}
	go func() {
		<-ctx.Done()
		err := ctx.Err()
		// 超时处理
		if err == context.DeadlineExceeded {
			log.Printf("session timeout after %d seconds\n", int(to.Seconds()))
			ss.Close(ErrSessionTimeout)
		}
		if err == context.Canceled {
			log.Printf("session canceled")
			ss.Close(ErrSessionCanceled)
		}
	}()
	return ss, nil
}

// RegisterMapperInit registers mapper initializer
func (db *DB) RegisterMapperInit(m ...MapperInit) {
	db.mm.register(m...)
}

// InitMappers initialize mappers
func (db *DB) InitMappers(ss *Session) error {
	return db.mm.initMappers(ss)
}

// ConnConfig ...
type ConnConfig struct {
	Driver   string
	Host     string
	Port     string
	Database string
	User     string
	Password string

	// Connection configurations
	MaxOpenConns    int
	MaxIdelConns    int
	ConnMaxLifeTime time.Duration
	ConnMaxIdelTime time.Duration
}

// DSN ...
func (cfg *ConnConfig) DSN() string {
	if cfg.Port == "" {
		cfg.Port = "3306"
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=true&allowNativePasswords=true", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database)
}

// NewDatabase creates new database connection pool
func NewDatabase(cfg *ConnConfig) (*DB, error) {
	db, err := sqlx.Open(cfg.Driver, cfg.DSN())
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdelConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifeTime)
	// TODO 需要升级Golang版本
	// db.SetConnMaxIdleTime(cfg.ConnMaxIdelTime)
	return &DB{DB: db}, nil
}

// NewTestDBFromEnvVar create database from environment variables
func NewTestDBFromEnvVar() (*DB, error) {
	dbCfg := &ConnConfig{
		Driver:          "mysql",
		Host:            os.Getenv("TEST_MYSQL_ADDR"),
		Port:            os.Getenv("TEST_MYSQL_PORT"),
		Database:        os.Getenv("TEST_MYSQL_DB"),
		User:            os.Getenv("TEST_MYSQL_USER"),
		Password:        os.Getenv("TEST_MYSQL_PWD"),
		MaxOpenConns:    100,
		MaxIdelConns:    10,
		ConnMaxLifeTime: 10 * time.Minute,
		ConnMaxIdelTime: 10 * time.Minute,
	}
	return NewDatabase(dbCfg)
}
