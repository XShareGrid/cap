package mysql

import (
	"fmt"
	"log"
	"sort"

	"github.com/XShareGrid/cap/msg/errors"
)

// MapperInit mapper initializer
type MapperInit interface {
	Name() string
	InitMapper(ss *Session) error
	// Dependencies mapper names
	Dependencies() []string
}

type mapperMgr struct {
	mapperList []MapperInit
}

// register mappers
func (mm *mapperMgr) register(m ...MapperInit) {
	mm.mapperList = append(mm.mapperList, m...)
}

// initialize mappers
func (mm *mapperMgr) initMappers(ss *Session) error {
	// sort mappers
	sort.SliceStable(mm.mapperList, func(i, j int) bool {
		for _, md := range mm.mapperList[i].Dependencies() {
			if md == mm.mapperList[j].Name() {
				return false
			}
		}
		return true
	})
	for _, m := range mm.mapperList {
		log.Printf("initializing mapper[%s]", m.Name())
		err := m.InitMapper(ss)
		if err != nil {
			log.Printf("initialize mapper[%s], failed with error: %s\n", m.Name(), err.Error())
			return err
		}
		log.Printf("initialize mapper[%s], succeeded\n", m.Name())
	}
	return nil
}

// BaseMapper base mapper
type BaseMapper struct {
	tx        *Session
	tableName string
}

// NewBaseMapper creates a BaseMapper
func NewBaseMapper(ss *Session, tableName string) *BaseMapper {
	return &BaseMapper{tx: ss, tableName: tableName}
}

// GetTx gets session
func (m *BaseMapper) GetTx() *Session {
	return m.tx
}

// Name gets mapper name
func (m *BaseMapper) Name() string {
	return m.tableName
}

// Delete ...
func (m *BaseMapper) Delete(tableName string, filters ...*RowsFilter) (affected int64, err error) {
	affected = -1
	sqlDelete := fmt.Sprintf(`DELETE FROM %s WHERE `, tableName)
	if len(filters) == 0 {
		return affected, ErrDeleteMustContainFilters
	}
	var values []interface{}
	for i, f := range filters {
		sqlDelete += fmt.Sprintf("%s %s (?) ", f.Key, f.Operator)
		if i != len(filters)-1 {
			sqlDelete += "AND "
		}
		values = append(values, f.Value)
	}
	ret, err := m.GetTx().Exec(sqlDelete, values...)
	if err != nil {
		return affected, errors.Wrap(err).Log()
	}
	return ret.RowsAffected()
}

// RowsFilterOperator rows filter operator
type RowsFilterOperator string

const (
	// RFEqual ..
	RFEqual = "="
	// RFNotEqual ...
	RFNotEqual = "!="
	// RFIn ...
	RFIn = "IN"
)

// RowsFilter rows filter
type RowsFilter struct {
	Key      string
	Operator RowsFilterOperator
	Value    interface{}
}
