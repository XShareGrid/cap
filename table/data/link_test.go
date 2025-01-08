package data

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/XShareGrid/cap/database/mysql"
	"github.com/XShareGrid/cap/table/data/driver"
	"github.com/XShareGrid/cap/table/data/driver/dbdriver"
	"github.com/XShareGrid/cap/table/data/utils"
	"github.com/XShareGrid/cap/table/operator/builtin"
	cap "github.com/XShareGrid/cap/table/proto/go"
	"github.com/XShareGrid/cap/table/registry"
)

type MaterialData struct {
	Code  string `t_key:"true" t_name:"物料编码" t_fm:"EQ"`
	TName string `t_name:"物料名称"`
	Model string `t_name:"型号"`
	Spec  string `t_name:"规格"`
}

// Name ...
func (MaterialData) Name() string {
	return "物料列表"
}

// Desc ...
func (MaterialData) Desc() string {
	return "物料列表"
}

type MaterialDriver []*MaterialData

func (md *MaterialDriver) FindRows(ctx context.Context, ss *mysql.Session, tmd registry.TableMetaData, conditions []*driver.Condition,
	OutputColumns []string, aggCols []*driver.AggregateColumn, pageParam *cap.PageParam,
	orderParam *cap.OrderParam,
) (*driver.RowsResult, error) {
	var pageInfo *cap.PageInfo
	var results []interface{}
	var err error
	for _, md := range *md {
		matched := true
		// 匹配条件
		for _, c := range conditions {
			if c.ColumnID == "Code" {
				if c.OperatorID == builtin.EQ.ID() {
					v := c.Values[0]
					if v != md.Code {
						matched = matched && false
					}
				} else if c.OperatorID == builtin.IN.ID() {
					isInValues := false
					for i := range c.Values {
						v := c.Values[i]
						if v == md.Code {
							isInValues = true
							break
						}
					}
					if !isInValues {
						matched = matched && false
					}
				}
			}
		}
		if matched {
			results = append(results, md)
		}
	}
	// do memory paging
	results, pageInfo = utils.DoMemoryPaging(results, pageParam)

	return &driver.RowsResult{
		Rows:     results,
		PageInfo: pageInfo,
	}, err
}

// cap -t2g -tts -dsn "root:125801@tcp(localhost:3306)/cap?charset=utf8" -p data -of output.go -tn label
type Label struct {
	SN            string `t_key:"true" db:"SN" t_fm:"SSTR|SNULLABLE" t_name:"SN" t_vt:"STRING"`
	MaterialCode  string `db:"MaterialCode" t_fm:"SSTR|SNULLABLE" t_name:"物料编码" t_vt:"STRING"`
	MaterialModel string `t_link:"data.MaterialData(MaterialCode, Code, Model)" t_name:"型号"`
}

// Name ...
func (Label) Name() string {
	return "标签表"
}

// Desc ...
func (Label) Desc() string {
	return "标签表"
}
func registerTables() {
	// 注册报表
	tmdMaterial, err := registry.LoadTMDFromStruct(&MaterialData{})
	if err != nil {
		log.Fatal(err.Error())
	}

	err = registry.GlobalTableRegistry().TableMetaReg.Register(tmdMaterial)
	if err != nil {
		panic(err)
	}

	ddd := &MaterialDriver{
		{Code: "00001", TName: "T0001", Model: "物料0001", Spec: "1"},
		{Code: "00002", TName: "T0002", Model: "物料0002", Spec: "1"},
		{Code: "00003", TName: "T0003", Model: "物料0003", Spec: "1"},
		{Code: "00004", TName: "T0004", Model: "物料0004", Spec: "1"},
		{Code: "00005", TName: "T0005", Model: "物料0005", Spec: "1"},
		{Code: "00006", TName: "T0006", Model: "物料0006", Spec: "1"},
		{Code: "00007", TName: "T0007", Model: "物料0007", Spec: "1"},
		{Code: "00008", TName: "T0008", Model: "物料0008", Spec: "1"},
		{Code: "00009", TName: "T0009", Model: "物料0009", Spec: "1"},
		{Code: "00010", TName: "T0010", Model: "物料0010", Spec: "1"},
		{Code: "00011", TName: "T0011", Model: "物料0011", Spec: "1"},
		{Code: "00012", TName: "T0012", Model: "物料0012", Spec: "1"},
		{Code: "00013", TName: "T0013", Model: "物料0013", Spec: "1"},
		{Code: "00014", TName: "T0014", Model: "物料0014", Spec: "1"},
		{Code: "00015", TName: "T0015", Model: "物料0015", Spec: "1"},
		{Code: "00016", TName: "T0016", Model: "物料0016", Spec: "1"},
		{Code: "00017", TName: "T0017", Model: "物料0017", Spec: "1"},
		{Code: "00018", TName: "T0018", Model: "物料0018", Spec: "1"},
		{Code: "00019", TName: "T0019", Model: "物料0019", Spec: "1"},
		{Code: "00020", TName: "T0020", Model: "物料0020", Spec: "1"},
	}

	if err = GlobalManager().RegisterDriver(tmdMaterial.ID(), ddd); err != nil {
		panic(err)
	}

	// 注册报表
	tmdLabel, err := registry.LoadTMDFromStruct(&Label{})
	if err != nil {
		log.Fatal(err.Error())
	}

	err = registry.GlobalTableRegistry().TableMetaReg.Register(tmdLabel)
	if err != nil {
		panic(err)
	}

	dddLabel := dbdriver.NewDBDriver("label")

	if err = GlobalManager().RegisterDriver(tmdLabel.ID(), dddLabel); err != nil {
		panic(err)
	}
	if err = registry.GlobalTableRegistry().TableMetaReg.ValidateLinks(); err != nil {
		panic(err)
	}
}

func TestFindLinkRows(t *testing.T) {
	registerTables()
	// 开始查询
	ss, err := testDB.NewSession()
	if err != nil {
		panic(err)
	}
	tpl := utils.NewTmpTpl("data.MaterialData", []*driver.Condition{
		driver.NewCondition("Code", builtin.EQ.ID(), "00001")},
		[]string{"Code", "TName"},
	)

	rsp, err := GlobalManager().FindRows(context.Background(), ss, tpl, &cap.PageParam{}, &cap.OrderParam{})
	if err != nil {
		panic(err)
	}
	for _, c := range rsp.Rows {
		fmt.Println(c)
	}

	fmt.Println("=======================================================")
	tpl = utils.NewTmpTpl("data.Label", []*driver.Condition{},
		[]string{"MaterialCode", "MaterialModel"},
	)
	rsp, err = GlobalManager().FindRows(context.Background(), ss, tpl, &cap.PageParam{}, &cap.OrderParam{})
	if err != nil {
		panic(err)
	}
	for _, c := range rsp.Rows {
		fmt.Println(c)
	}

}
