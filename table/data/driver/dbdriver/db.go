package dbdriver

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/XShareGrid/cap/database/mysql"
	"github.com/XShareGrid/cap/msg/errors"
	"github.com/XShareGrid/cap/table/data/driver"
	"github.com/XShareGrid/cap/table/operator/builtin"
	cap "github.com/XShareGrid/cap/table/proto/go"
	"github.com/XShareGrid/cap/table/registry"
	"github.com/jmoiron/sqlx"
)

// DBDataDriver ...
type DBDataDriver struct {
	dbTableName string
	queryLimit  int
}

// ParseConditions parse condition to sql where segments eg: XXX = ?
// dbTag 数据库的tag，可按优先级顺序提供
func ParseConditions(desc *registry.TableColumnDescriptor, cs []*driver.Condition, dbTag ...string) (whereSegments []string, queryArgs []interface{}, err error) {
	var colID string
	for _, t := range dbTag {
		colID = desc.Tag.Get(t)
		if colID != "" {
			break
		}
	}

	supportFilterMap := make(map[string]interface{})
	for _, f := range desc.SupportedFilters {
		supportFilterMap[f.ID()] = nil
	}
	for _, c := range cs {
		if c.ColumnID != desc.ID {
			continue
		}
		if _, ok := supportFilterMap[c.OperatorID]; !ok {
			return nil, nil, errors.Wrap(ErrOperatorNotSupportedForColumn).FillDebugArgs(c.OperatorID, desc.ID)
		}
		switch c.OperatorID {
		case builtin.EQ.ID():
			whereSegments = append(whereSegments, fmt.Sprintf("(%s) = ?", colID))
			if len(c.Values) != 1 {
				err = errors.Wrap(ErrInvalidValueForCondition).FillDebugArgs(c.ColumnID + ":" + c.OperatorID)
				return
			}
			queryArgs = append(queryArgs, c.Values[0])
		case builtin.GT.ID():
			whereSegments = append(whereSegments, fmt.Sprintf("(%s) > ?", colID))
			if len(c.Values) != 1 {
				err = errors.Wrap(ErrInvalidValueForCondition).FillDebugArgs(c.ColumnID + ":" + c.OperatorID)
				return
			}
			queryArgs = append(queryArgs, c.Values[0])
		case builtin.LT.ID():
			whereSegments = append(whereSegments, fmt.Sprintf("(%s) < ?", colID))
			if len(c.Values) != 1 {
				err = errors.Wrap(ErrInvalidValueForCondition).FillDebugArgs(c.ColumnID + ":" + c.OperatorID)
				return
			}
			queryArgs = append(queryArgs, c.Values[0])
		case builtin.GE.ID():
			whereSegments = append(whereSegments, fmt.Sprintf("(%s) >= ?", colID))
			if len(c.Values) != 1 {
				err = errors.Wrap(ErrInvalidValueForCondition).FillDebugArgs(c.ColumnID + ":" + c.OperatorID)
				return
			}
			queryArgs = append(queryArgs, c.Values[0])
		case builtin.LE.ID():
			whereSegments = append(whereSegments, fmt.Sprintf("(%s) <= ?", colID))
			if len(c.Values) != 1 {
				err = errors.Wrap(ErrInvalidValueForCondition).FillDebugArgs(c.ColumnID + ":" + c.OperatorID)
				return
			}
			queryArgs = append(queryArgs, c.Values[0])
		case builtin.NE.ID():
			whereSegments = append(whereSegments, fmt.Sprintf("(%s) != ?", colID))
			if len(c.Values) != 1 {
				err = errors.Wrap(ErrInvalidValueForCondition).FillDebugArgs(c.ColumnID + ":" + c.OperatorID)
				return
			}
			queryArgs = append(queryArgs, c.Values[0])
		case builtin.CTN.ID():
			if len(c.Values) != 1 {
				err = errors.Wrap(ErrInvalidValueForCondition).FillDebugArgs(c.ColumnID + ":" + c.OperatorID)
				return
			}
			whereSegments = append(whereSegments, fmt.Sprintf("(%s) LIKE '%%%s%%'", colID, c.Values[0]))
		case builtin.LCTN.ID():
			if len(c.Values) != 1 {
				err = errors.Wrap(ErrInvalidValueForCondition).FillDebugArgs(c.ColumnID + ":" + c.OperatorID)
				return
			}
			whereSegments = append(whereSegments, fmt.Sprintf("(%s) LIKE '%s%%'", colID, c.Values[0]))

		case builtin.RCTN.ID():
			if len(c.Values) != 1 {
				err = errors.Wrap(ErrInvalidValueForCondition).FillDebugArgs(c.ColumnID + ":" + c.OperatorID)
				return
			}
			whereSegments = append(whereSegments, fmt.Sprintf("(%s) LIKE '%%%s'", colID, c.Values[0]))
		case builtin.NCTN.ID():
			if len(c.Values) != 1 {
				err = errors.Wrap(ErrInvalidValueForCondition).FillDebugArgs(c.ColumnID + ":" + c.OperatorID)
				return
			}
			whereSegments = append(whereSegments, fmt.Sprintf("(%s) NOT LIKE '%%%s%%'", colID, c.Values[0]))
		case builtin.IN.ID():
			if len(c.Values) == 0 {
				err = errors.Wrap(ErrInvalidValueForCondition).FillDebugArgs(c.ColumnID + ":" + c.OperatorID)
				return
			}
			var args []interface{}
			for _, v := range c.Values {
				args = append(args, v)
			}
			whereSegments = append(whereSegments, fmt.Sprintf("(%s) IN (?)", colID))
			queryArgs = append(queryArgs, args)
		case builtin.NIN.ID():
			if len(c.Values) == 0 {
				err = errors.Wrap(ErrInvalidValueForCondition).FillDebugArgs(c.ColumnID + ":" + c.OperatorID)
				return
			}
			var args []interface{}
			for _, v := range c.Values {
				args = append(args, v)
			}
			whereSegments = append(whereSegments, fmt.Sprintf("(%s) NOT IN (?)", colID))
			queryArgs = append(queryArgs, args)
		case builtin.ISN.ID():
			if len(c.Values) != 0 {
				err = errors.Wrap(ErrInvalidValueForCondition).FillDebugArgs(c.ColumnID + ":" + c.OperatorID)
				return
			}
			whereSegments = append(whereSegments, fmt.Sprintf("(%s) IS NULL", colID))
		case builtin.ISNN.ID():
			if len(c.Values) != 0 {
				err = errors.Wrap(ErrInvalidValueForCondition).FillDebugArgs(c.ColumnID + ":" + c.OperatorID)
				return
			}
			whereSegments = append(whereSegments, fmt.Sprintf("(%s) IS NOT NULL", colID))
		case builtin.TODAY.ID():
			if len(c.Values) != 0 {
				err = errors.Wrap(ErrInvalidValueForCondition).FillDebugArgs(c.ColumnID + ":" + c.OperatorID)
				return
			}
			whereSegments = append(whereSegments, fmt.Sprintf("(%s) >= ?", colID))
			whereSegments = append(whereSegments, fmt.Sprintf("(%s) < ?", colID))
			now := time.Now()
			queryArgs = append(queryArgs,
				time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local))
			queryArgs = append(queryArgs,
				time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 0, 0, time.Local))
		case builtin.TWEEK.ID():
			if len(c.Values) != 0 {
				err = errors.Wrap(ErrInvalidValueForCondition).FillDebugArgs(c.ColumnID + ":" + c.OperatorID)
				return
			}
			whereSegments = append(whereSegments, fmt.Sprintf("(%s) >= ?", colID))
			whereSegments = append(whereSegments, fmt.Sprintf("(%s) < ?", colID))
			now := time.Now()
			weekAgo := now.Add(-7 * 24 * time.Hour)
			queryArgs = append(queryArgs,
				time.Date(weekAgo.Year(), weekAgo.Month(), weekAgo.Day(), 0, 0, 0, 0, time.Local))
			queryArgs = append(queryArgs,
				time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 0, 0, time.Local))
		case builtin.L1MONTH.ID():
			if len(c.Values) != 0 {
				err = errors.Wrap(ErrInvalidValueForCondition).FillDebugArgs(c.ColumnID + ":" + c.OperatorID)
				return
			}
			whereSegments = append(whereSegments, fmt.Sprintf("(%s) >= ?", colID))
			whereSegments = append(whereSegments, fmt.Sprintf("(%s) < ?", colID))
			now := time.Now()
			weekAgo := now.Add(-30 * 24 * time.Hour)
			queryArgs = append(queryArgs,
				time.Date(weekAgo.Year(), weekAgo.Month(), weekAgo.Day(), 0, 0, 0, 0, time.Local))
			queryArgs = append(queryArgs,
				time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 0, 0, time.Local))
		case builtin.L3MONTH.ID():
			if len(c.Values) != 0 {
				err = errors.Wrap(ErrInvalidValueForCondition).FillDebugArgs(c.ColumnID + ":" + c.OperatorID)
				return
			}
			whereSegments = append(whereSegments, fmt.Sprintf("(%s) >= ?", colID))
			whereSegments = append(whereSegments, fmt.Sprintf("(%s) < ?", colID))
			now := time.Now()
			weekAgo := now.Add(-90 * 24 * time.Hour)
			queryArgs = append(queryArgs,
				time.Date(weekAgo.Year(), weekAgo.Month(), weekAgo.Day(), 0, 0, 0, 0, time.Local))
			queryArgs = append(queryArgs,
				time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 0, 0, time.Local))
		case builtin.YEST.ID():
			if len(c.Values) != 0 {
				err = errors.Wrap(ErrInvalidValueForCondition).FillDebugArgs(c.ColumnID + ":" + c.OperatorID)
				return
			}
			whereSegments = append(whereSegments, fmt.Sprintf("(%s) >= ?", colID))
			whereSegments = append(whereSegments, fmt.Sprintf("(%s) < ?", colID))
			yest := time.Now().Add(-24 * time.Hour)
			queryArgs = append(queryArgs,
				time.Date(yest.Year(), yest.Month(), yest.Day(), 0, 0, 0, 0, time.Local))
			queryArgs = append(queryArgs,
				time.Date(yest.Year(), yest.Month(), yest.Day(), 23, 59, 0, 0, time.Local))
		default:
			err = errors.Wrap(ErrUnknownOperator).FillDebugArgs(c.OperatorID)
			return
		}
	}
	return
}

// NewDBDriver create db driver
func NewDBDriver(tableName string, queryLimit ...int) *DBDataDriver {
	defaultQueryLimit := 100000
	if len(queryLimit) > 0 {
		defaultQueryLimit = queryLimit[0]
	}
	return &DBDataDriver{dbTableName: tableName, queryLimit: defaultQueryLimit}
}

// FindRows ...
func (ddd *DBDataDriver) FindRows(ctx context.Context, ss *mysql.Session, tmd registry.TableMetaData, conditions []*driver.Condition,
	outputColumns []string, aggCols []*driver.AggregateColumn, pageParam *cap.PageParam,
	orderParam *cap.OrderParam,
) (*driver.RowsResult, error) {
	// SELECT ... FROM dbTableName WHERE filters ORDER BY orderParam
	query := "SELECT %s FROM " + ddd.dbTableName
	countQuery := "SELECT COUNT(*) FROM " + ddd.dbTableName
	aggQuery := "SELECT %s FROM " + ddd.dbTableName
	var fieldList []string
	var aggFieldList []string
	var queryArgs []interface{}
	var whereSegments []string
	var orderString string
	var err error
	var aggResults []*driver.AggregateResult
	var pageInfo *cap.PageInfo
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
		ws, qa, err := ParseConditions(desc, conditions, "dbc", "db")
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
	countQueryArgs := make([]interface{}, len(queryArgs))
	aggQueryArgs := make([]interface{}, len(queryArgs))
	copy(countQueryArgs, queryArgs)
	copy(aggQueryArgs, queryArgs)
	// 聚合列
	for _, col := range aggCols {
		desc, err := tmd.Columns().Find(col.ColumnID)
		if err != nil {
			return nil, errors.Wrap(err).Log()
		}
		if col.AggregateMethod == cap.AggregateMethod_AM_AVG {
			aggFieldList = append(aggFieldList, fmt.Sprintf("AVG(%s)", desc.Tag.Get("db")))
		} else if col.AggregateMethod == cap.AggregateMethod_AM_SUM {
			aggFieldList = append(aggFieldList, fmt.Sprintf("SUM(%s)", desc.Tag.Get("db")))
		} else {
			continue
		}
		aggregateColumns = append(aggregateColumns, desc)
	}
	aggFields := strings.Join(aggFieldList, ", ")
	aggQuery = fmt.Sprintf(aggQuery, aggFields)

	// TODO. 条件排序
	if len(whereSegments) > 0 {
		whereQuery := strings.Join(whereSegments, " AND ")
		query += " WHERE " + whereQuery
		countQuery += " WHERE " + whereQuery
		aggQuery += " WHERE " + whereQuery
	}
	if orderString != "" {
		query += " " + orderString
	}
	query, queryArgs, err = sqlx.In(query, queryArgs...)
	if err != nil {
		return nil, errors.Wrap(err).Log()
	}
	countQuery, countQueryArgs, err = sqlx.In(countQuery, countQueryArgs...)
	if err != nil {
		return nil, errors.Wrap(err).Log()
	}
	aggQuery, aggQueryArgs, err = sqlx.In(aggQuery, aggQueryArgs...)
	if err != nil {
		return nil, errors.Wrap(err).Log()
	}
	aggValues := []interface{}{}
	for _, desc := range aggregateColumns {
		aggValues = append(aggValues, reflect.New(desc.DataType).Interface())
	}
	if len(aggregateColumns) > 0 {
		aggRow := ss.QueryRow(aggQuery, aggQueryArgs...)
		if aggRow != nil {
			err = aggRow.Scan(aggValues...)
			if err != nil {
				return nil, errors.Wrap(err).Log()
			}
			for i, col := range aggregateColumns {
				aggResults = append(aggResults, &driver.AggregateResult{
					ColumnID: col.ID,
					Result:   reflect.ValueOf(aggValues[i]).Elem().Interface(),
				})
			}
		}
	}
	// COUNT
	resultCount := 0
	err = ss.Get(&resultCount, countQuery, queryArgs...)
	if err != nil {
		return nil, errors.Wrap(err).Log()
	}
	if pageParam.PageSize == 0 && resultCount > ddd.queryLimit {
		return nil, errors.Wrap(ErrResultExceedMaxLimit).FillDebugArgs(resultCount, ddd.queryLimit)
	}
	// Paging
	if pageParam.PageSize > 0 {
		query += fmt.Sprintf(" LIMIT %d, %d",
			pageParam.Page*pageParam.PageSize, pageParam.PageSize)
	}

	rows, err := ss.Queryx(query, queryArgs...)
	if err != nil {
		return nil, errors.Wrap(err).Log()
	}

	var results []interface{}
	// parse data
	for rows.Next() {
		var selectRow interface{}
		if tmd.RowDataType() == reflect.TypeOf(map[string]interface{}{}) {
			selectRow = map[string]interface{}{}
			err = rows.MapScan(selectRow.(map[string]interface{}))
			if err != nil {
				return nil, errors.Wrap(err).Log()
			}
			results = append(results, selectRow)
		} else {
			selectRow = reflect.New(tmd.RowDataType()).Interface()
			err = rows.StructScan(selectRow)
			if err != nil {
				return nil, errors.Wrap(err).Log()
			}
			results = append(results, selectRow)
		}

	}

	var pageCount = int32(1)
	if pageParam.PageSize > 0 {
		pageCount = int32(resultCount) / pageParam.PageSize
		if resultCount%int(pageParam.PageSize) > 0 {
			pageCount++
		}
	}
	pageInfo = &cap.PageInfo{
		CurrentPage:  pageParam.Page,
		PageSize:     pageParam.PageSize,
		TotalPages:   pageCount,
		TotalResults: int32(resultCount),
	}

	return &driver.RowsResult{
		Rows:       results,
		AggResults: aggResults,
		PageInfo:   pageInfo,
	}, err
}
