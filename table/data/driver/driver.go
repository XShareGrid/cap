package driver

import (
	"context"

	"github.com/XShareGrid/cap/database/mysql"
	cap "github.com/XShareGrid/cap/table/proto/go"
	"github.com/XShareGrid/cap/table/registry"
)

// NewCondition 创建模板过滤条件
func NewCondition(columnID, operatorID string, values ...interface{}) *Condition {
	return &Condition{
		ColumnID:   columnID,
		OperatorID: operatorID,
		Values:     values,
	}
}

// Condition ...
type Condition struct {
	ColumnID   string
	OperatorID string
	Values     []interface{}
}

// AggregateColumn 聚合列
type AggregateColumn struct {
	ColumnID        string
	AggregateMethod cap.AggregateMethod
}

// AggregateResult 聚合结果
type AggregateResult struct {
	ColumnID string
	Result   interface{}
}

// RowsResult 结果
type RowsResult struct {
	Rows       []interface{}      // 返回数据行列表，支持metadata对应【数据结构】指针数组
	AggResults []*AggregateResult // 聚合结果
	PageInfo   *cap.PageInfo      // 分页信息
}

// Driver        数据驱动接口
// ss:           数据库会话
// tmd:          表格元数据
// conditions:   筛选条件
// outputColumn: 期望输出列
//
//	期望输出的列必须有数据，其他列可有可无
//	可用于优化查询
//
// aggCols:      聚合列
// pageParam:    分页参数
// orderParam:   排序参数
type Driver interface {
	FindRows(ctx context.Context, ss *mysql.Session, tmd registry.TableMetaData, conditions []*Condition, outputColumns []string,
		aggCols []*AggregateColumn, pageParam *cap.PageParam, orderParam *cap.OrderParam) (result *RowsResult, err error)
}

// Deletable 可删除
type Deletable interface {
	DeleteRows(ctx context.Context, ss *mysql.Session, tmd registry.TableMetaData, rowIDs []string) (err error)
}

// RowsResultStream 结果
type RowsResultStream interface {
	// 获取下一行结果，空代表结束
	Next() (interface{}, error)
}

// Streamer 流式查询，应对大量数据导出，不支持聚合，不支持分页
type Streamer interface {
	FindRowsStream(ctx context.Context, ss *mysql.Session, tmd registry.TableMetaData, conditions []*Condition, outputColumns []string,
		orderParam *cap.OrderParam) (result *RowsResultStream, err error)
}
