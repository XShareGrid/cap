package registry

import (
	"log"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/XShareGrid/cap/msg/errors"
	cap "github.com/XShareGrid/cap/table/proto/go"
)

// TMDStruct table metadata struct interface
type TMDStruct interface {
	Name() string
	Desc() string
}

// tags
// t_key      |  是否KEY字段(主键): true/false, 默认false, id列，唯一主键，有且仅有一个，t_vt必须为string
//       	  |  t_key字段必须支持EQ操作符
// t_internal |  可选值: true 内部字段，前端隐藏显示
// t_name     |  列名
// t_fm       |  过滤方法，|分隔，默认 空
// t_agg      |  聚合方式，可选(SUM - 求和， AVG - 平均值)，|分隔
//            |  默认 空
//
// t_vt       |  列类型，ValueType，可选: INT, STRING, BOOL, DOUBLE, TIME, DATE, OPTION
//            |  默认自动根据数据类型判定，time.Time类型必填，
//            |  显式指定t_vt=0，系统将自动按照t_vt_fmt进行Sprintf转换为string，时间使用time.Format转换为string
//            |  其他不支持类型转换
//
// t_vt_fmt   |  格式化参数
// t_link     |  语法1： 表示外部数据关联，格式：TableID(locValueColumn, searchColumnID, valueColumnID)
//            |       TableID: 链接外部表ID
//            |       locValueColumn: 当前表列ID
//            |       searchColumnID: 查找外部表ID
//            |       valueColumnID: 外部表值ID
///           |       系统将自动赋值该字段，SELECT TableID.searchColumnID FROM this LEFT JOIN TableID ON this.locValueColumn = TableID.searchColumnID
//            |  语法2: 表示简单外键，格式：TableID(valueColumnID)
//            |       TableID: 链接外部表ID
//            |       valueColumnID: 外部表值ID
//            |       仅表示关系
//            |  注意：1. 关联的列，ValueType类型必须一致，2. 只支持单次链接，不支持循环链接
// t_order    |  可选项：true/false  不填默认值false

// TableExample 示例表格
type TableExample struct {
	Value1 string    `t_name:"测试1" t_fm:"EQ|NE|IN" `
	Value2 int       `t_name:"测试2" t_fm:"EQ|NE|GE"`
	Value3 bool      `t_name:"测试3"`
	Value4 time.Time `t_vt:"date" t_link:"test_table(Value1, TValue1, TValue2)"`
	Value5 int       `t_vt:"int" t_vt_fmt:"%02d" t_fm:"SINT" t_agg:"sum"`
	Value6 int       `t_internal:"true"`
}

// LoadTMDFromStruct ...
// defaultTplWrapper 默认模板包装，用户可用于自定义默认模板
func LoadTMDFromStruct(v TMDStruct, defaultTplWrapper ...func(*cap.Template) *cap.Template) (TableMetaData, error) {
	vt := reflect.TypeOf(v).Elem()
	tmd, err := NewTableMetaData(vt.String(), v.Name(), v.Desc(), []*TableColumnDescriptor{}, defaultTplWrapper...)
	if err != nil {
		return nil, errors.Wrap(err).Log()
	}
	hasIDColumn := false
	for i := 0; i < vt.NumField(); i++ {
		if vt.Field(i).Tag.Get("t_name") == "-" {
			continue
		}
		col, err := parseColumn(vt.Field(i))
		if err != nil {
			return nil, errors.Wrap(err).Log()
		}
		if col.IsKeyColumn {
			if !hasIDColumn {
				hasIDColumn = true
				tmd.keyColID = col.ID
			} else {
				return nil, errors.Wrap(ErrMultipleIDColumnNotAllowed).FillDebugArgs(tmd.ID())
			}
		}
		err = tmd.AddColumns(col)
		if err != nil {
			return nil, errors.Wrap(err).Log()
		}
	}
	if !hasIDColumn {
		return nil, errors.Wrap(ErrNoIDColumnSpecified).FillDebugArgs(tmd.ID())
	}
	tmd.rowDataType = reflect.TypeOf(v).Elem()
	return tmd, nil
}

func assertProtoEnum(v int32, m map[int32]string) {
	if _, ok := m[v]; !ok {
		log.Fatalf("value %d not in enum %v", v, m)
	}
}

// TODO. 自定义顺序
func sortOperators(ops []Operator) (sorted []Operator) {
	opsMap := make(map[string]Operator)
	for _, op := range ops {
		opsMap[op.ID()] = op
	}
	for _, v := range opsMap {
		sorted = append(sorted, v)
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].ID() < sorted[j].ID()
	})
	return
}

var vtMapping = map[string]cap.ValueType{
	"INT":    cap.ValueType_VT_INT,
	"STRING": cap.ValueType_VT_STRING,
	"BOOL":   cap.ValueType_VT_BOOLEAN,
	"DOUBLE": cap.ValueType_VT_DOUBLE,
	"OPTION": cap.ValueType_VT_OPTION,
	"TIME":   cap.ValueType_VT_TIME,
	"DATE":   cap.ValueType_VT_DATE,
}

var aggMapping = map[string]cap.AggregateMethod{
	"SUM": cap.AggregateMethod_AM_SUM,
	"AVG": cap.AggregateMethod_AM_AVG,
}

func parseColumn(f reflect.StructField) (*TableColumnDescriptor, error) {
	tcd := TableColumnDescriptor{
		ID: f.Name,
	}
	// Parse: Name --------------------------
	tcd.Name = f.Tag.Get("t_name")
	if tcd.Name == "" {
		tcd.Name = f.Name
	}
	var ft int = -1

	// Parse: ValueType --------------------------
	tVT := f.Tag.Get("t_vt")
	if tVT != "" {
		v, ok := vtMapping[tVT]
		if !ok {
			return nil, errors.Wrap(ErrUnsupportedFieldType).FillDebugArgs(tVT)
		}
		ft = int(v)
	}
	tcd.DataType = f.Type
	if f.Type.Kind() == reflect.Ptr {
		tcd.DataType = f.Type.Elem()
	}
	ftk := tcd.DataType.Kind()
	// 基础数据类型
	if (ftk > reflect.Invalid && ftk <= reflect.Float64) ||
		(ftk == reflect.String) {
		switch ftk {
		case reflect.Bool:
			tcd.ValueType = cap.ValueType_VT_BOOLEAN
		case reflect.Int,
			reflect.Int8,
			reflect.Int16,
			reflect.Int32,
			reflect.Int64,
			reflect.Uint,
			reflect.Uint8,
			reflect.Uint16,
			reflect.Uint32,
			reflect.Uint64:
			if ft == int(cap.ValueType_VT_OPTION) {
				tcd.ValueType = cap.ValueType_VT_OPTION
				_, err := GlobalTableRegistry().OptionReg.GetOptions(tcd.DataType.String())
				if err != nil {
					return nil, errors.Wrap(err).Log()
				}
			} else {
				tcd.ValueType = cap.ValueType_VT_INT
			}
		case reflect.Float32,
			reflect.Float64:
			tcd.ValueType = cap.ValueType_VT_DOUBLE
		case reflect.String:
			tcd.ValueType = cap.ValueType_VT_STRING
		}
	} else {
		if f.Type.String() == "time.Time" || f.Type.String() == "sql.NullTime" {
			switch ft {
			case int(cap.ValueType_VT_DATE):
				tcd.ValueType = cap.ValueType_VT_DATE
			case int(cap.ValueType_VT_TIME):
				tcd.ValueType = cap.ValueType_VT_TIME
			default:
				return nil, errors.Wrap(ErrInvalidValueTypeForTime).FillDebugArgs(tcd.ID)
			}
		} else if f.Type.String() == "sql.NullBool" {
			tcd.ValueType = cap.ValueType_VT_BOOLEAN
		} else if f.Type.String() == "sql.NullFloat64" || f.Type.String() == "sql.NullFloat32" {
			tcd.ValueType = cap.ValueType_VT_DOUBLE
		} else {
			// field type not supported
		}
	}
	// string
	if ft == int(cap.ValueType_VT_STRING) {
		tcd.ValueType = cap.ValueType_VT_STRING
		tcd.ValueFormat = f.Tag.Get("t_vt_fmt")
		if tcd.ValueFormat == "" {
			switch tcd.DataType.Kind() {
			case reflect.Int,
				reflect.Int8,
				reflect.Int16,
				reflect.Int32,
				reflect.Int64,
				reflect.Uint,
				reflect.Uint8,
				reflect.Uint16,
				reflect.Uint32,
				reflect.Uint64:
				tcd.ValueFormat = "%d"
			case reflect.Float32,
				reflect.Float64:
				tcd.ValueFormat = "%g"
			}
		}
	}
	assertProtoEnum(int32(tcd.ValueType), cap.ValueType_name)
	// Parse: ID
	isKey := f.Tag.Get("t_key")
	if isKey == "true" && tcd.ValueType == cap.ValueType_VT_STRING {
		tcd.IsKeyColumn = true
	}
	// Parse: Aggregate --------------------------
	tAggs := f.Tag.Get("t_agg")
	if tAggs != "" {
		tAggList := strings.Split(tAggs, "|")
		for _, tAgg := range tAggList {
			agg, ok := aggMapping[tAgg]
			if !ok {
				return nil, errors.Wrap(ErrParseFieldType).FillDebugArgs(tAgg)
			}
			tcd.SupportedAggregateMethod =
				append(tcd.SupportedAggregateMethod, cap.AggregateMethod(agg))
			assertProtoEnum(int32(agg), cap.AggregateMethod_name)
		}
	}

	// Parse: Filter Method --------------------------
	tFM := f.Tag.Get("t_fm")
	if tFM != "" {
		opList := strings.Split(tFM, "|")
		for _, opID := range opList {
			if !strings.Contains(opID, ".") {
				opID = "builtin." + opID
			}
			node, err := GlobalTableRegistry().OperatorReg.Find(opID)
			if err != nil {
				return nil, errors.Wrap(err).Log()
			}
			switch node.(type) {
			case Operator:
				tcd.SupportedFilters = append(tcd.SupportedFilters, node.(Operator))
			case OperatorSet:
				tcd.SupportedFilters = append(tcd.SupportedFilters, node.(OperatorSet).Operators()...)
			default:
				panic("node type not supported")
			}
		}
		sortOperators(tcd.SupportedFilters)
	}

	colWidth := f.Tag.Get("t_colwidth")
	if colWidth != "" {
		colWidthFloat, _ := strconv.ParseFloat(colWidth, 64)
		if colWidthFloat == 0 {
			panic("invalid t_colwidth: " + colWidth)
		}
		tcd.ColWidth = colWidthFloat
	} else {
		tcd.ColWidth = 1
	}
	supportEQ := false
	// 检查是否支持EQ操作
	if tcd.IsKeyColumn {
		for _, o := range tcd.SupportedFilters {
			if o.ID() == "builtin.EQ" {
				supportEQ = true
				break
			}
		}
		if !supportEQ {
			return nil, errors.Wrap(ErrKeyColumnMustSupportEQ).FillDebugArgs(tcd.ID)
		}
	}

	// Parse: ArrSplit
	tcd.ArrSplit = strings.ReplaceAll(f.Tag.Get("t_asplit"), " ", "")
	// Parse: Href
	tcd.Href = strings.ReplaceAll(f.Tag.Get("t_href"), " ", "")
	if tcd.Href != "" {
		hs := strings.ReplaceAll(f.Tag.Get("t_href_style"), " ", "")
		switch hs {
		case "newtab":
			tcd.HrefStyle = cap.HrefStyle_HSNewTab
		case "dialog":
			tcd.HrefStyle = cap.HrefStyle_HSDialog
		case "edit":
			tcd.HrefStyle = cap.HrefStyle_HSEdit
		case "custom":
			tcd.HrefStyle = cap.HrefStyle_HSCustom
		default:
			tcd.HrefStyle = cap.HrefStyle_HSNewTab
		}
	}
	// Parse: Links
	tLink := strings.ReplaceAll(f.Tag.Get("t_link"), " ", "")
	if tLink != "" {
		m, err := regexp.MatchString(`([\.\w]+)\((\w+)\,(\w+)\,(\w+)\)`, tLink)
		if !m || err != nil {
			m, err := regexp.MatchString(`([\.\w]+)\((\w+)\)`, tLink)
			if !m || err != nil {
				return nil, errors.Wrap(ErrInvalidLinkFormat).FillDebugArgs(tLink).Log()
			}
			// 外键关系
			link := &ColumnLink{}
			reg := regexp.MustCompile(`([\.\w]+)`)
			segments := reg.FindAllString(tLink, -1)
			if len(segments) != 2 {
				return nil, errors.Wrap(ErrInvalidLinkFormat).FillDebugArgs(tLink).Log()
			}
			link.RemoteTableID = segments[0]
			// 外键 local = 自己
			link.LocalColID = tcd.ID
			link.RemoteSearchColID = segments[1]
			link.RemoteValueColID = segments[1]
			tcd.Link = link
		} else {
			// 外部数据链接
			link := &ColumnLink{}
			reg := regexp.MustCompile(`([\.\w]+)`)
			segments := reg.FindAllString(tLink, -1)
			if len(segments) != 4 {
				return nil, errors.Wrap(ErrInvalidLinkFormat).FillDebugArgs(tLink).Log()
			}
			link.RemoteTableID = segments[0]
			link.LocalColID = segments[1]
			link.RemoteSearchColID = segments[2]
			link.RemoteValueColID = segments[3]
			tcd.Link = link
		}

	}
	// Parse: t_internal
	iInternal := strings.ReplaceAll(f.Tag.Get("t_internal"), " ", "")
	if iInternal == "true" {
		tcd.Internal = true
	}
	// Parse: t_required
	tRequired := strings.ReplaceAll(f.Tag.Get("t_required"), " ", "")
	if tRequired == "true" {
		tcd.Required = true
	}
	tOrder := strings.ReplaceAll(f.Tag.Get("t_order"), " ", "")
	if tOrder == "true" {
		tcd.Orderable = true
	}
	tcd.Tag = f.Tag
	return &tcd, nil

}
