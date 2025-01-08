package builtin

import (
	cap "github.com/XShareGrid/cap/table/proto/go"
	"github.com/XShareGrid/cap/table/registry"
)

// Operator operator
type Operator struct {
	id              string
	name            string
	filterValueType cap.FilterValueType
}

// registerBuiltinOperator creates an operator and register it
func registerBuiltinOperator(id, name string, filterValueType cap.FilterValueType) *Operator {
	id = "builtin." + id
	return RegisterOperator(id, name, filterValueType)
}

// RegisterOperator creates an operator and register it
func RegisterOperator(id, name string, filterValueType cap.FilterValueType) *Operator {
	o := &Operator{id: id, name: name, filterValueType: filterValueType}
	err := registry.GlobalTableRegistry().OperatorReg.Register(o)

	if err != nil {
		panic(err)
	}
	return o
}

// Name ...
func (o *Operator) Name() string {
	return o.name
}

func (o *Operator) SetName(name string) string {
	temp := o.name
	o.name = name
	return temp
}

// FilterValueType ...
func (o *Operator) FilterValueType() cap.FilterValueType { return o.filterValueType }

// ID ...
func (o *Operator) ID() string {
	return o.id
}

// EQ 等于
var EQ = registerBuiltinOperator("EQ", "等于", cap.FilterValueType_FVT_SINGLE)

// GT 大于
var GT = registerBuiltinOperator("GT", "大于", cap.FilterValueType_FVT_SINGLE)

// LT 小于
var LT = registerBuiltinOperator("LT", "小于", cap.FilterValueType_FVT_SINGLE)

// GE 大于等于
var GE = registerBuiltinOperator("GE", "大于等于", cap.FilterValueType_FVT_SINGLE)

// LE 小于等于
var LE = registerBuiltinOperator("LE", "小于等于", cap.FilterValueType_FVT_SINGLE)

// NE 不等于
var NE = registerBuiltinOperator("NE", "不等于", cap.FilterValueType_FVT_SINGLE)

// CTN 包含
var CTN = registerBuiltinOperator("CTN", "包含", cap.FilterValueType_FVT_SINGLE)

// LCTN 左包含
var LCTN = registerBuiltinOperator("LCTN", "左包含", cap.FilterValueType_FVT_SINGLE)

// RCTN 右包含
var RCTN = registerBuiltinOperator("RCTN", "右包含", cap.FilterValueType_FVT_SINGLE)

// NCTN 不包含
var NCTN = registerBuiltinOperator("NCTN", "不包含", cap.FilterValueType_FVT_SINGLE)

// IN IN
var IN = registerBuiltinOperator("IN", "IN", cap.FilterValueType_FVT_MULTIPLE)

// NIN NOT IN
var NIN = registerBuiltinOperator("NIN", "NOT IN", cap.FilterValueType_FVT_MULTIPLE)

// ISN 为空
var ISN = registerBuiltinOperator("ISN", "为空", cap.FilterValueType_FVT_NULL)

// ISNN 不为空
var ISNN = registerBuiltinOperator("ISNN", "不为空", cap.FilterValueType_FVT_NULL)

// TODAY 在今天
var TODAY = registerBuiltinOperator("TODAY", "在今天", cap.FilterValueType_FVT_NULL)

// TWEEK 最近一周
var TWEEK = registerBuiltinOperator("TWEEK", "最近一周", cap.FilterValueType_FVT_NULL)

// L1MONTH 最近1个月
var L1MONTH = registerBuiltinOperator("L1MONTH", "最近一个月", cap.FilterValueType_FVT_NULL)

// L3MONTH 最近3个月
var L3MONTH = registerBuiltinOperator("L3MONTH", "最近三个月", cap.FilterValueType_FVT_NULL)

// YEST 在昨天
var YEST = registerBuiltinOperator("YEST", "在昨天", cap.FilterValueType_FVT_NULL)
