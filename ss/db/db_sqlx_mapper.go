package db

import (
	"context"
	"errors"

	"github.com/jmoiron/sqlx"
)

// BaseSQLMapper base SQL mapper
type BaseSQLMapper struct {
	tx        *sqlx.Tx
	tableName string
	columns   []string
}

// ErrNoTx No database tx in the context
var ErrNoTx = errors.New("No database tx in the context")

// NewSQLDatabaseContext create a sql database context
func NewSQLDatabaseContext(ctx context.Context, db *sqlx.DB) context.Context {
	return context.WithValue(ctx, sqlDatabaseCtxKey, db)
}

// GetSQLDatabaseFromContext get sql database from context
func GetSQLDatabaseFromContext(ctx context.Context) (db *sqlx.DB, ok bool) {
	db, ok = ctx.Value(sqlDatabaseCtxKey).(*sqlx.DB)
	return db, ok
}

// NewSQLTxContext create sql tx context
func NewSQLTxContext(ctx context.Context, tx *sqlx.Tx) context.Context {
	return context.WithValue(ctx, sqlTxCtxKey, tx)
}

// GetSQLTxFromContext get sql tx from context
func GetSQLTxFromContext(ctx context.Context) (tx *sqlx.Tx, ok bool) {
	tx, ok = ctx.Value(sqlTxCtxKey).(*sqlx.Tx)
	return tx, ok
}

// GetTx get tx
func (m *BaseSQLMapper) GetTx() *sqlx.Tx {
	return m.tx
}

// SetTx set tx
func (m *BaseSQLMapper) SetTx(tx *sqlx.Tx) {
	m.tx = tx
}

// GetTableName get table name
func (m *BaseSQLMapper) GetTableName() string {
	return m.tableName
}

// SetTableName set table name
func (m *BaseSQLMapper) SetTableName(name string) {
	m.tableName = name
}

// GetColumns get columns
func (m *BaseSQLMapper) GetColumns() []string {
	return m.columns
}

// SetColumns set columns
func (m *BaseSQLMapper) SetColumns(columns []string) {
	m.columns = columns
}

func (m *BaseSQLMapper) ColumnsString() string {
	str := ""
	for i, col := range m.columns {
		if i != 0 {
			str += ", "
		}
		str += col
	}
	return str
}

func (m *BaseSQLMapper) ColumnsEqualString() string {
	str := ""
	for i, col := range m.columns {
		if i != 0 {
			str += ", "
		}
		str += (col + " = ?")
	}
	return str
}

func (m *BaseSQLMapper) ColumnsNum() int {
	return len(m.columns)
}

func (m *BaseSQLMapper) QuestionMarkString() string {
	str := ""
	for i, _ := range m.columns {
		if i != 0 {
			str += ", "
		}
		str += "?"
	}
	return str
}