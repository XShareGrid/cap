package tables

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"

	"github.com/XShareGrid/cap/database/mysql"
	"github.com/XShareGrid/cap/jsonschema"
	"github.com/XShareGrid/cap/table/action"
	"github.com/XShareGrid/cap/table/data"
	"github.com/XShareGrid/cap/table/data/driver"
	"github.com/XShareGrid/cap/table/data/driver/dbdriver"
	"github.com/XShareGrid/cap/table/demo/tables/options"
	"github.com/XShareGrid/cap/table/operator/builtin"
	cap "github.com/XShareGrid/cap/table/proto/go"
	"github.com/XShareGrid/cap/table/registry"
)

type Testtable struct {
	F1 int64              `db:"f1" t_ignore:"true" t_vt:"STRING" t_fm:"SSTR" t_key:"true" t_name:"字段1"`                                   // 字段1|
	F2 sql.NullString     `db:"f2" t_ignore:"true"  t_fm:"SSTR" t_name:"字段2"`                                                             // 字段2|
	F3 sql.NullTime       `db:"f3" t_ignore:"true" t_vt:"TIME" t_name:"字段3" t_fm:"STIME"`                                                 // 字段3|
	F4 options.TestOption `db:"f4" t_ignore:"true" t_vt:"OPTION" t_name:"字段4" t_fm:"SOPT"`                                                // 字段4|
	F5 sql.NullString     `db:"f5" t_ignore:"true" t_href:"https://www.baidu.com/s?wd=%s" t_href_style:"newtab" t_name:"字段5" t_fm:"SSTR"` // 字段5|
	F6 sql.NullString     `db:"f6" t_ignore:"true" t_name:"字段6" t_fm:"SSTR"`                                                              // 字段6|
	F7 sql.NullString     `db:"f7" t_ignore:"true" t_name:"字段7" t_fm:"SSTR"`                                                              // 字段7|
	F8 sql.NullString     `db:"f8" t_ignore:"true" t_name:"字段8" t_fm:"SSTR"`                                                              // 字段8|
	F9 int64              `db:"f9" t_ignore:"true" t_name:"字段9" t_fm:"SINT"`                                                              // 字段9|
}

func (Testtable) TableName() string {
	return "testtable"
}

func (Testtable) Name() string {
	return "测试数据表"
}

// Desc ...
func (Testtable) Desc() string {
	return "测试数据表"
}

type TesttableDriver struct {
	ddd *dbdriver.DBDataDriver
}

// FindRows ...
func (p *TesttableDriver) FindRows(ctx context.Context, ss *mysql.Session, tmd registry.TableMetaData, conditions []*driver.Condition,
	outputColumns []string, aggCols []*driver.AggregateColumn, pageParam *cap.PageParam,
	orderParam *cap.OrderParam) (*driver.RowsResult, error) {
	// TODO 这里可以做条件优化处理
	return p.ddd.FindRows(ctx, ss, tmd, conditions, outputColumns, aggCols,
		pageParam, orderParam)
}

type deleteRowOperation struct {
	F1 string `json_schema:"required,title=字段1(Field 1),default_auto,readOnly=true"`
}

// Schema ...
func (e deleteRowOperation) Schema(rowData interface{}) ([]byte, error) {
	if rowData != nil {
		l, ok := rowData.(*Testtable)
		if !ok {
			panic("invalid row data type")
		}
		e.F1 = strconv.Itoa(int(l.F1))
	}
	js, err := jsonschema.GenJSONSchemaObject(context.Background(), e, nil)
	if err != nil {
		return nil, err
	}
	return json.Marshal(js)
}

func initTesttable() {
	// 注册选项
	registry.RegisterOptionFromProtoEnum(options.TestOption(0))

	// 加载元数据
	tmd, err := registry.LoadTMDFromStruct(&Testtable{}, func(t *cap.Template) *cap.Template {
		t.Body.Filter.Conditions = append(t.Body.Filter.Conditions,
			&cap.Condition{
				ColumnId:   "F2",
				OperatorId: builtin.CTN.ID(),
				Values:     []*cap.FilterValue{{LiteralValues: &cap.Value{V: &cap.Value_VString{VString: ""}}}},
			})
		return t
	})

	tmd.AddRowActions(
		action.NewFormRowAction(
			"delete", "删除",
			action.NewRowFormSQLExecutor("DELETE FROM testtable WHERE F1 = '{{.F1}}';"),
			&deleteRowOperation{},
		))
	if err != nil {
		panic(err)
	}
	registry.GlobalTableRegistry().TableMetaReg.Register(tmd)
	err = data.GlobalManager().RegisterDriver(tmd.ID(), &TesttableDriver{ddd: dbdriver.NewDBDriver("testtable")})
	if err != nil {
		panic(err)
	}
}
