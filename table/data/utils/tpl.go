package utils

import (
	"reflect"

	"github.com/XShareGrid/cap/table/data/driver"
	cap "github.com/XShareGrid/cap/table/proto/go"
)

// 套娃函数
func newLiteralFilterValue(v *cap.Value) *cap.FilterValue {
	return &cap.FilterValue{LiteralValues: v}
}

// Option 用于condition
type Option int32

// Date 用于condition
type Date string

// Time 用于condition
type Time string

// NewTmpTpl creates temporary template
func NewTmpTpl(tableID string, tplConditions []*driver.Condition, outputColumns []string) *cap.Template {
	tpl := &cap.Template{
		TableId: tableID,
	}
	tpl.Body = &cap.TemplateBody{
		Filter: &cap.FilterBody{},
		Output: &cap.OutputBody{},
	}
	for _, c := range tplConditions {
		con := &cap.Condition{
			ColumnId:   c.ColumnID,
			OperatorId: c.OperatorID,
		}
		for _, v := range c.Values {
			if reflect.TypeOf(v).String() == "utils.Option" {
				con.Values = append(con.Values,
					newLiteralFilterValue(&cap.Value{
						V: &cap.Value_VOption{VOption: &cap.OptionValue{Id: int32(reflect.ValueOf(v).Int())}}}))
				continue
			} else if reflect.TypeOf(v).String() == "utils.Date" {
				con.Values = append(con.Values,
					newLiteralFilterValue(&cap.Value{
						V: &cap.Value_VDate{VDate: string(v.(Date))}}))
				continue
			} else if reflect.TypeOf(v).String() == "utils.Time" {
				con.Values = append(con.Values,
					newLiteralFilterValue(&cap.Value{
						V: &cap.Value_VTime{VTime: string(v.(Time))}}))
				continue
			}
			switch reflect.TypeOf(v).Kind() {
			case reflect.Bool:
				con.Values = append(con.Values,
					newLiteralFilterValue(&cap.Value{V: &cap.Value_VBool{VBool: v.(bool)}}))
			case reflect.Int,
				reflect.Int8,
				reflect.Int16,
				reflect.Int32,
				reflect.Int64:
				con.Values = append(con.Values,
					newLiteralFilterValue(&cap.Value{V: &cap.Value_VInt{VInt: int32(reflect.ValueOf(v).Int())}}))
			case reflect.Uint,
				reflect.Uint8,
				reflect.Uint16,
				reflect.Uint32,
				reflect.Uint64:
				con.Values = append(con.Values,
					newLiteralFilterValue(&cap.Value{V: &cap.Value_VInt{VInt: int32(reflect.ValueOf(v).Uint())}}))
			case reflect.Float32:
			case reflect.Float64:
				con.Values = append(con.Values,
					newLiteralFilterValue(&cap.Value{V: &cap.Value_VDouble{VDouble: reflect.ValueOf(v).Float()}}))
			case reflect.String:
				con.Values = append(con.Values,
					newLiteralFilterValue(&cap.Value{V: &cap.Value_VString{VString: v.(string)}}))
			}
		}
		tpl.Body.Filter.Conditions = append(tpl.Body.Filter.Conditions, con)
	}

	for _, o := range outputColumns {
		tpl.Body.Output.VisibleColumns = append(tpl.Body.Output.VisibleColumns, &cap.TemplateColumn{ColumnId: o, Visible: true})
	}
	return tpl
}
