package tables

import (
	"context"
	"time"

	"github.com/XShareGrid/cap/database/mysql"
	"github.com/XShareGrid/cap/table/data"
	"github.com/XShareGrid/cap/table/data/driver"
	"github.com/XShareGrid/cap/table/data/utils"
	"github.com/XShareGrid/cap/table/demo/tables/options"
	"github.com/XShareGrid/cap/table/operator/builtin"
	cap "github.com/XShareGrid/cap/table/proto/go"
	"github.com/XShareGrid/cap/table/registry"
)

type MemTable struct {
	F1  int64              `db:"f1" t_ignore:"true" t_fm:"SSTR" t_vt:"STRING" t_key:"true" t_name:"字段1"` // 字段1|
	F2  string             `db:"f2" t_ignore:"true" t_name:"字段2" t_fm:"SSTR" `                           // 字段2|
	F3  time.Time          `db:"f3" t_vt:"DATE" t_ignore:"true" t_name:"字段3" t_fm:"SDATE"`               // 字段3|
	F4  options.TestOption `db:"f4" t_ignore:"true" t_vt:"OPTION" t_fm:"SOPT" t_name:"字段4"`              // 字段4|
	F5  string             `db:"f5" t_ignore:"true" t_name:"字段5" t_fm:"SSTR"`                            // 字段5|
	F6  string             `db:"f6" t_ignore:"true" t_name:"字段6" t_fm:"SSTR"`                            // 字段6|
	F7  string             `db:"f7" t_ignore:"true" t_name:"字段7" t_fm:"SSTR"`                            // 字段7|
	F8  string             `db:"f8" t_ignore:"true" t_name:"字段8" t_fm:"SSTR"`                            // 字段8|
	F9  int64              `db:"f9" t_ignore:"true" t_name:"字段9" t_fm:"SINT"`                            // 字段9|
	F10 string             `db:"f8" t_ignore:"true" t_name:"字段10" t_fm:"SSTR"`                           // 字段8|
	F11 int64              `db:"f9" t_ignore:"true" t_name:"字段11" t_fm:"SSTR"`                           // 字段9|
}

func (MemTable) TableName() string {
	return "memtable"
}

func (MemTable) Name() string {
	return "测试内存表"
}

// Desc ...
func (MemTable) Desc() string {
	return "测试数据表"
}

type MemTableDriver struct {
}

// FindRows ...
func (p *MemTableDriver) FindRows(ctx context.Context, ss *mysql.Session, tmd registry.TableMetaData, conditions []*driver.Condition,
	outputColumns []string, aggCols []*driver.AggregateColumn, pageParam *cap.PageParam,
	orderParam *cap.OrderParam) (*driver.RowsResult, error) {
	now := time.Now()
	rows := []interface{}{}
	for i := 0; i < 10000; i++ {
		rows = append(rows,
			&MemTable{int64(i), "2", now.Add(time.Duration(i) * time.Minute), options.TestOption_O2, "421", "43", "33", "232", int64(i * 2), "ASDASD", int64(3)})
	}

	rows, pageInfo := utils.DoMemoryPaging(rows, pageParam)

	return &driver.RowsResult{
		Rows:     rows,
		PageInfo: pageInfo,
	}, nil
}

func initMemTable() {
	// 注册选项
	registry.RegisterOptionFromProtoEnum(options.TestOption(0))

	// 加载元数据
	tmd, err := registry.LoadTMDFromStruct(&MemTable{}, func(t *cap.Template) *cap.Template {
		t.Body.Filter.Conditions = append(t.Body.Filter.Conditions,
			&cap.Condition{
				ColumnId:   "F2",
				OperatorId: builtin.CTN.ID(),
				Values:     []*cap.FilterValue{{LiteralValues: &cap.Value{V: &cap.Value_VString{VString: ""}}}},
			})
		return t
	})
	if err != nil {
		panic(err)
	}
	registry.GlobalTableRegistry().TableMetaReg.Register(tmd)
	err = data.GlobalManager().RegisterDriver(tmd.ID(), &MemTableDriver{})
	if err != nil {
		panic(err)
	}
}
