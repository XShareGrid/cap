package rt

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	_ "github.com/XShareGrid/cap/table/operator/builtin"
	cap "github.com/XShareGrid/cap/table/proto/go"
	"github.com/XShareGrid/cap/table/registry"
	"github.com/XShareGrid/cap/test"
)

// TableExample 示例表格
type TableExample struct {
	Value1 string    `t_name:"测试1" t_key:"true" t_fm:"EQ|NE|IN" `
	Value2 int       `t_name:"测试2" t_fm:"EQ|NE|GE"`
	Value3 bool      `t_name:"测试3"`
	Value4 time.Time `t_vt:"DATE" t_link:"test_table(Value1, TValue1, TValue2)"`
	Value5 int       `t_vt:"INT" t_vt_fmt:"%02d" t_fm:"SINT" t_agg:"SUM|AVG"`
	Value6 int       `t_vt:"INT" t_link:"test_table(Value1)"`
}

// Name ...
func (TableExample) Name() string {
	return "示例表格"
}

// Desc ...
func (TableExample) Desc() string {
	return "这是一个示例表格"
}

func TestLoadTMDFromStruct(t *testing.T) {
	tmd, err := registry.LoadTMDFromStruct(&TableExample{})
	if err != nil {
		log.Fatal(err.Error())
	}
	tmd.Print(os.Stdout)
	defaultTpl := tmd.DefaultTpl(context.Background())
	test.DisplayObject(defaultTpl)
	err = registry.GlobalTableRegistry().TableMetaReg.Register(tmd)
	if err != nil {
		panic(err)
	}
	registry.GlobalTableRegistry().TableMetaReg.List()

}

func TestLoadDBStruct(t *testing.T) {
	// 注册选项
	registry.RegisterOptionFromProtoEnum(cap.FileAccessType(0))
	// 加载数据结构
	tmd, err := registry.LoadTMDFromStruct(&TableTemplate{},
		func(t *cap.Template) *cap.Template {
			t.Body.Filter = &cap.FilterBody{}
			return t
		})

	if err != nil {
		log.Fatal(err.Error())
	}
	tmd.Print(os.Stdout)
	defaultTpl := tmd.DefaultTpl(context.Background())
	test.DisplayObject(defaultTpl)
	err = registry.GlobalTableRegistry().TableMetaReg.Register(tmd)
	if err != nil {
		panic(err)
	}
	registry.GlobalTableRegistry().TableMetaReg.List()
}
