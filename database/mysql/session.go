package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"runtime/debug"

	"github.com/XShareGrid/cap/msg/errors"
	//"github.com/XShareGrid/ss/db"
	"github.com/jmoiron/sqlx"
)

type sessionKey struct {
	name string
}

var sk = sessionKey{name: "default"}

// Session mysql tx
type Session struct {
	sqlx.Tx
	err        error
	cancelFunc context.CancelFunc
	db         *DB
	hasTx      bool
}

// Exec ...
func (s *Session) Exec(query string, args ...interface{}) (sql.Result, error) {
	s.debugLog(query, args)
	if s.err == context.Canceled {
		return nil, context.Canceled
	}
	if s.hasTx {
		return s.Tx.Exec(query, args...)
	}
	return s.db.Exec(query, args...)
}

// Select ...
func (s *Session) Select(dest interface{}, query string, args ...interface{}) error {
	s.debugLog(query, args)
	if s.err == context.Canceled {
		return context.Canceled
	}
	if s.hasTx {
		return s.Tx.Select(dest, query, args...)
	}
	return s.db.Select(dest, query, args...)
}

// Get ...
func (s *Session) Get(dest interface{}, query string, args ...interface{}) error {
	s.debugLog(query, args)
	if s.err == context.Canceled {
		return context.Canceled
	}
	if s.hasTx {
		return s.Tx.Get(dest, query, args...)
	}
	return s.db.Get(dest, query, args...)
}

// Queryx ...
func (s *Session) Queryx(query string, args ...interface{}) (*sqlx.Rows, error) {
	s.debugLog(query, args)
	if s.err == context.Canceled {
		return nil, context.Canceled
	}
	if s.hasTx {
		return s.Tx.Queryx(query, args...)
	}
	return s.db.Queryx(query, args...)
}

// DebugLogOn debug log switch
var DebugLogOn = false

func (s *Session) debugLog(query string, args ...interface{}) {
	if DebugLogOn {
		log.Println("[MYSQL]", query, args)
	}
}

// GetSessionFromCtx ...
func GetSessionFromCtx(ctx context.Context) (*Session, error) {
	v := ctx.Value(sk)
	if v == nil {
		return nil, errors.Wrap(ErrNoSessionInCtx)
	}
	if ss, ok := v.(*Session); ok {
		return ss, nil
	}
	return nil, ErrNoSessionInCtx
}

// SaveToContext save session to context
func (s *Session) SaveToContext(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, sk, s)
}

// Close commit when seccuss
// rollback when fail
func (s *Session) Close(err error) error {
	if s.err != nil {
		return s.err
	}
	if s.cancelFunc != nil {
		s.cancelFunc()
	}
	if r := recover(); r != nil {
		log.Printf("stacktrace from panic[%s]: \n"+string(debug.Stack()), r)
		err = fmt.Errorf("recovered from panic [%v]", r)
	}
	s.err = err
	if err != nil {
		if s.hasTx {
			rollbackErr := s.Tx.Rollback()
			if rollbackErr != nil {
				fmt.Println("===>rollbackError", rollbackErr)
				return errors.Wrap(err).Triggers(rollbackErr)
			}
			return rollbackErr
		} else {
			s.err = context.Canceled
		}
	}
	if s.hasTx {
		return s.Tx.Commit()
	}
	return nil
}
