package dbdriver

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/XShareGrid/cap/database/mysql"
	"github.com/XShareGrid/cap/msg/errors"
	"github.com/XShareGrid/cap/table/data/driver"
	"github.com/XShareGrid/cap/table/data/driver/dbdriver"
	cap "github.com/XShareGrid/cap/table/proto/go"
	"github.com/XShareGrid/cap/table/registry"
	"github.com/jmoiron/sqlx"
)

// DBDataDriver ...
type DBDataDriver struct {
	dbTableName string
	queryLimit  int
}

// NewDBDriver create db driver
func NewStreamDriver(tableName string, queryLimit ...int) *DBDataDriver {
	defaultQueryLimit := 100000
	if len(queryLimit) > 0 {
		defaultQueryLimit = queryLimit[0]
	}
	return &DBDataDriver{dbTableName: tableName, queryLimit: defaultQueryLimit}
}

// FindRows ...
func (ddd *DBDataDriver) FindRowsStream(ctx context.Context, ss *mysql.Session, tmd registry.TableMetaData, conditions []*driver.Condition, outputColumns []string,
	orderParam *cap.OrderParam) (driver.RowsResultStream, error) {
	// SELECT ... FROM dbTableName WHERE filters ORDER BY orderParam
	query := "SELECT %s FROM " + ddd.dbTableName
	countQuery := "SELECT COUNT(*) FROM " + ddd.dbTableName
	var fieldList []string
	var queryArgs []interface{}
	var whereSegments []string
	var orderString string
	var err error
	// select列
	selectColumns := []*registry.TableColumnDescriptor{}
	// 聚合列
	aggregateColumns := []*registry.TableColumnDescriptor{}

	type linkNullValue struct {
		colID     string
		nullValue interface{}
	}
	var linkColumns []*linkNullValue

	// 选择列
	for _, col := range outputColumns {
		desc, err := tmd.Columns().Find(col)
		if err != nil {
			return nil, errors.Wrap(err).Log()
		}
		if desc.Link != nil {
			// 外链，代表需要外部JOIN，这里不做数据初始化
			if desc.Link.LocalColID != desc.ID {
				linkColumns = append(linkColumns,
					&linkNullValue{colID: desc.ID, nullValue: reflect.New(desc.DataType).Elem().Interface()})
				continue
			}
		}
		colID := desc.Tag.Get("db")
		dbc := desc.Tag.Get("dbc")
		if dbc == "" {
			dbc = colID
		}
		fieldList = append(fieldList, fmt.Sprintf("%s AS `%s`", dbc, colID))

		selectColumns = append(selectColumns, desc)
	}

	for _, desc := range tmd.Columns().List() {
		ws, qa, err := dbdriver.ParseConditions(desc, conditions, "dbc", "db")
		if err != nil {
			return nil, errors.Wrap(err).Log()
		}
		if len(qa) > 0 {
			queryArgs = append(queryArgs, qa...)
		}
		if len(ws) > 0 {
			whereSegments = append(whereSegments, ws...)
		}
		if orderParam != nil && orderParam.ColumnId == desc.ID && desc.Orderable {
			order := "DESC"
			if orderParam.Order == cap.Order_O_ASC {
				order = "ASC"
			}
			colID := desc.Tag.Get("db")
			dbc := desc.Tag.Get("dbc")
			if dbc == "" {
				dbc = colID
			}
			orderString = fmt.Sprintf("ORDER BY %s %s", colID, order)
		}
	}

	fields := strings.Join(fieldList, ", ")
	query = fmt.Sprintf(query, fields)

	// TODO. 条件排序
	if len(whereSegments) > 0 {
		whereQuery := strings.Join(whereSegments, " AND ")
		query += " WHERE " + whereQuery
	}
	if orderString != "" {
		query += " " + orderString
	}
	query, queryArgs, err = sqlx.In(query, queryArgs...)
	if err != nil {
		return nil, errors.Wrap(err).Log()
	}
	aggValues := []interface{}{}
	for _, desc := range aggregateColumns {
		aggValues = append(aggValues, reflect.New(desc.DataType).Interface())
	}
	// COUNT
	resultCount := 0
	err = ss.Get(&resultCount, countQuery, queryArgs...)
	if err != nil {
		return nil, errors.Wrap(err).Log()
	}

	rows, err := ss.Queryx(query, queryArgs...)
	if err != nil {
		return nil, errors.Wrap(err).Log()
	}

	return &DBResult{dbRows: rows, tmd: tmd}, nil

	// return &driver.RowsResult{
	// 	Rows:       results,
	// 	AggResults: aggResults,
	// 	PageInfo:   pageInfo,
	// }, err
}

type DBResult struct {
	dbRows *sqlx.Rows
	tmd    registry.TableMetaData
}

func (r *DBResult) Next() (interface{}, error) {
	// parse data
	if r.dbRows.Next() {
		var selectRow interface{}
		if r.tmd.RowDataType() == reflect.TypeOf(map[string]interface{}{}) {
			selectRow = map[string]interface{}{}
			err := r.dbRows.MapScan(selectRow.(map[string]interface{}))
			if err != nil {
				return nil, errors.Wrap(err).Log()
			}
			return selectRow, nil
		} else {
			selectRow = reflect.New(r.tmd.RowDataType()).Interface()
			err := r.dbRows.StructScan(selectRow)
			if err != nil {
				return nil, errors.Wrap(err).Log()
			}
			return selectRow, nil
		}

	} else {
		return nil, nil
	}
}
