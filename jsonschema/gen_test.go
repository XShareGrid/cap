package jsonschema

import (
	"context"
	"fmt"
	"testing"

	"gitlab.pintechs.com/eh/energy-backend/cap/test"
)

// Position 工位
type Position struct {
	PositionNum          int     `json_schema:"required,title=分切工位号,default_auto"`
	Length               float32 `json_schema:"required,title=长度(Length),default_auto,ui:filterable=true"`
	Width                float32 `json_schema:"required,title=宽度(Width),default_auto"`
	DefCode              string  `json_schema:"required,title=不良信息(Defect Info),default_auto"`
	Grade                string  `json_schema:"required,title=等级(Grade),default_auto"`
	PrintNum             int     `json_schema:"required,title=打印数量,default_auto"`
	CustomerCode         string  `json_schema:"required,title=客户编码(Customer Code),default_auto"`
	NotPrintCustomerCode bool    `json_schema:"required,title=是否打印客户编码,default_auto"`
}

// FinSap 成品
type FinSap struct {
	BatchDate             string     `json_schema:"required,title=批次生产日期,default_auto,format=date-time,ui:filterable=true"`
	DeviceCode            string     `json_schema:"required,title=半成品设备(Semi-finished Equipment),ui:filterable=true,ref_func_option=#/definitions/device_list,default_auto"`
	OrgSapBranch          int64      `json_schema:"required,title=母卷卷号,default_auto"`
	BaseFinCutTimes       int64      `json_schema:"required,title=基膜分切次数,default_auto"`
	BaseFinCutPositionNum int64      `json_schema:"required,title=基膜分切工位数,default_auto"`
	CutTimes              int        `json_schema:"required,title=分切次数(Slitting Times),default_auto,readOnly=true"`
	Position              []Position `json_schema:"required,title=分切工位信息,default_auto"`
}

func GetDeviceList(ctx context.Context) ([]Option, error) {
	option := &Option{}
	option.Key = "A"
	option.Value = "产线A"
	optionList := []Option{}
	optionList = append(optionList, *option)
	option.Key = "B"
	option.Value = "产线B"
	optionList = append(optionList, *option)
	return optionList, nil
}

// GenJSONSchemaObject 生成JSON_Schema
func TestGenJSONSchemaObject(t *testing.T) {
	// 注册回调函数
	err := RegisterOptionFunction("device_list", GetDeviceList)
	if err != nil {
		panic(err)
	}
	// 声明数据结构，并填写默认值
	j := FinSap{}
	object, err := GenJSONSchemaObject(nil, j, nil)
	if err != nil {
		panic(err)
	}
	test.DisplayObject(object)
}

func TestString(t *testing.T) {
	fmt.Println(ToSnakeCase("userName"))
}
