package data

import (
	"fmt"
	"testing"

	db "github.com/XShareGrid/cap/database/mysql"
	_ "github.com/go-sql-driver/mysql"
)

var testDB *db.DB

func init() {
	var err error
	testDB, err = db.NewTestDBFromEnvVar()
	if err != nil {
		panic(err)
	}
}

// func TestManager_FindRows(t *testing.T) {
// 	// 注册选项
// 	registry.RegisterOptionFromProtoEnum(cap.FileAccessType(0))
// 	// 注册报表
// 	tmd, err := registry.LoadTMDFromStruct(&rt.TableTemplate{},
// 		func(t *cap.Template) *cap.Template {
// 			t.Body.Filter = &cap.FilterBody{}
// 			return t
// 		})
// 	if err != nil {
// 		log.Fatal(err.Error())
// 	}

// 	err = registry.GlobalTableRegistry().TableMetaReg.Register(tmd)
// 	if err != nil {
// 		panic(err)
// 	}

// 	ddd := .NewDBDriver("table_template")
// 	tpl := tmd.DefaultTpl()
// 	if err = GlobalManager().RegisterDriver(tmd.ID(), ddd); err != nil {
// 		t.Fatal(err)
// 	}

// 	// 开始查询
// 	ss, err := testDB.NewSession()
// 	if err != nil {
// 		panic(err)
// 	}

// 	tpl.Body.Filter.Conditions = append(tpl.Body.Filter.Conditions, &cap.Condition{
// 		ColumnId: "TName", OperatorId: "builtin.RCTN",
// 		Values: []*cap.FilterValue{
// 			{LiteralValues: &cap.Value{V: &cap.Value_VString{VString: "n"}}}},
// 	},
// 	// &cap.Condition{
// 	// 	ColumnId: "FCreateUser", OperatorId: "builtin.EQ",
// 	// 	Values: []*cap.FilterValue{
// 	// 		{Values: &cap.FilterValue_LiteralValues{LiteralValues: &cap.Value{V: &cap.Value_VInt{VInt: 1}}}}},
// 	// },
// 	)
// 	tpl.Body.Output.VisibleColumns[4].AggregateMethod = cap.AggregateMethod_AM_SUM
// 	results, err := GlobalManager().FindRows(context.Background(), ss, tpl, &cap.PageParam{Page: 0, PageSize: 100}, &cap.OrderParam{})
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	for _, r := range results.Rows {
// 		for _, c := range r.Cells {
// 			fmt.Printf("%5.30v| ", c.Value.V)
// 		}
// 		fmt.Printf("\n")
// 	}
// 	fmt.Println(results.AggregateResult)
// }

func Test_linkAddQuery(t *testing.T) {
	fmt.Println(linkAddQuery("/abc/def?aaa=b", "user", "111"))
	fmt.Println(linkAddQuery("http://1.2.3.4:87476/abc/def?aaa=b", "user", "111"))
	fmt.Println(linkAddQuery("https://1.2.3.4:87476/abc/def?aaa=b", "user", "111"))
}
