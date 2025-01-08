package registry

import (
	"context"
	"fmt"
	"io"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/XShareGrid/cap/i18n"
	"github.com/XShareGrid/cap/msg/errors"
	"github.com/XShareGrid/cap/table/action"
	cap "github.com/XShareGrid/cap/table/proto/go"
)

// ColumnLink ...
type ColumnLink struct {
	// 关联表列ID
	RemoteTableID string
	// 本地列ID
	// 如果是当前列，代表外键自联，代表当前列Value = RemoteTableID.RemoteValueColID
	// 此时 Value 无需
	// 如果是非当前列，代表外键查询，代表当前列Value = VLOOKUP(RemoteTableID, RemoteSearchColID, RemoteValueColID)
	LocalColID string
	// 关联表搜索列
	RemoteSearchColID string
	// 关联表搜索值
	RemoteValueColID string
}

// String
func (c ColumnLink) String() string {
	return fmt.Sprintf("%s(%s,%s,%s)",
		c.RemoteTableID, c.LocalColID, c.RemoteSearchColID, c.RemoteValueColID)
}

// TableColumnDescriptor 表列描述符
type TableColumnDescriptor struct {
	// 列id，在表中的唯一标记
	ID string
	// 列名，用户可读
	Name string
	// DataType 数据类型
	DataType reflect.Type
	// ValueType 在表格中展示的类型
	ValueType cap.ValueType
	// ValueFormat ...
	ValueFormat string
	// 支持的筛选类型，长度为0代表不支持筛选
	SupportedFilters []Operator
	// 支持的聚合方法，长度为0代表不支持聚合
	SupportedAggregateMethod []cap.AggregateMethod
	// 关联
	Link *ColumnLink
	// TODO...
	Internal bool
	// Other tags
	Tag reflect.StructTag
	// 是否支持排序
	Orderable bool
	// 是否Key列
	IsKeyColumn bool
	// 数组分隔符
	ArrSplit string
	// Href
	Href string
	// Required
	Required bool
	// HrefStyle
	HrefStyle cap.HrefStyle
	// ColWidth 列宽，以1为单位，按比例分布，不填默认为1
	ColWidth float64
}

// ToTemplateColumn converts to *cap.TemplateColumn
func (tcd *TableColumnDescriptor) ToTemplateColumn(userLanguage, tableId string) *cap.TemplateColumn {
	return &cap.TemplateColumn{
		ColumnId:     tcd.ID,
		Visible:      true,
		ColumnDetail: tcd.BuildTableColumn(userLanguage, tableId),
	}
}

// ToTableColumn converts to *cap.ToTableColumn
func (tcd *TableColumnDescriptor) ToTableColumn(userLanguage string) *cap.TableColumn {
	col := &cap.TableColumn{
		Id:                       tcd.ID,
		Name:                     tcd.Name,
		ValueType:                tcd.ValueType,
		SupportedAggregateMethod: tcd.SupportedAggregateMethod,
		Internal:                 tcd.Internal,
		OrderAble:                tcd.Orderable,
		Required:                 tcd.Required,
		DisplayWidthRate:         tcd.ColWidth,
	}
	for _, o := range tcd.SupportedFilters {
		col.SupportedFilters = append(col.SupportedFilters,
			&cap.ColumnFilterMethod{
				Operator:  &cap.ConditionOperator{Id: o.ID(), Name: o.Name()},
				ValueType: o.FilterValueType(),
			})
	}
	if tcd.ValueType == cap.ValueType_VT_OPTION {
		col.OptionTypeId = tcd.DataType.String()
		options, _ := GlobalTableRegistry().OptionReg.GetOptionsWithI18n(userLanguage, col.OptionTypeId)
		col.Options = options
	}

	return col
}

func (tcd *TableColumnDescriptor) BuildTableColumn(userLanguage, tableID string) *cap.TableColumn {
	trKey := TrKeyTableField(tableID, tcd.ID)
	name := i18n.GetTranslation(userLanguage, trKey, tcd.Name)
	col := &cap.TableColumn{
		Id:                       tcd.ID,
		Name:                     name,
		ValueType:                tcd.ValueType,
		SupportedAggregateMethod: tcd.SupportedAggregateMethod,
		Internal:                 tcd.Internal,
		OrderAble:                tcd.Orderable,
		Required:                 tcd.Required,
		DisplayWidthRate:         tcd.ColWidth,
	}
	for _, o := range tcd.SupportedFilters {
		opTrKey := TrKeyOperator(o.ID())
		opName := i18n.GetTranslation(userLanguage, opTrKey)
		col.SupportedFilters = append(col.SupportedFilters,
			&cap.ColumnFilterMethod{
				Operator:  &cap.ConditionOperator{Id: o.ID(), Name: opName},
				ValueType: o.FilterValueType(),
			})
	}
	if tcd.ValueType == cap.ValueType_VT_OPTION {
		col.OptionTypeId = tcd.DataType.String()
		options, _ := GlobalTableRegistry().OptionReg.GetOptionsWithI18n(userLanguage, col.OptionTypeId)
		col.Options = options
	}

	return col
}

// TableColumnDescriptorList list of columns
type TableColumnDescriptorList struct {
	list    []*TableColumnDescriptor
	idIndex map[string]int
}

// NewTableColumnDescriptorList creates TableColumnDescriptorList
func NewTableColumnDescriptorList(list []*TableColumnDescriptor) (*TableColumnDescriptorList, error) {
	tcdl := &TableColumnDescriptorList{
		list: list,
	}
	return tcdl, tcdl.generateIndex()
}

// 生成索引，方便查询
func (tcdl *TableColumnDescriptorList) generateIndex() error {
	newIdx := make(map[string]int)
	for i, col := range tcdl.list {
		if _, ok := newIdx[col.ID]; ok {
			return errors.Wrap(ErrDuplicateColumnID).FillDebugArgs(col.ID)
		}
		newIdx[col.ID] = i
	}
	tcdl.idIndex = newIdx
	return nil
}

// List ...
func (tcdl *TableColumnDescriptorList) List() []*TableColumnDescriptor {
	return tcdl.list
}

// Find find column by id
// return nil if id is not exist
func (tcdl *TableColumnDescriptorList) Find(id string) (*TableColumnDescriptor, error) {
	if i, ok := tcdl.idIndex[id]; ok {
		return tcdl.list[i], nil
	}
	return nil, errors.Wrap(ErrInvalidColumnID).FillDebugArgs(id)
}

// RowActionList list of row actions
type RowActionList struct {
	list []action.RowAction
}

// NewActionList action list
func NewActionList(list []*TableColumnDescriptor) (*TableColumnDescriptorList, error) {
	tcdl := &TableColumnDescriptorList{
		list: list,
	}
	return tcdl, tcdl.generateIndex()
}

// List ...
func (ral *RowActionList) List() []action.RowAction {
	return ral.list
}

// Find find column by id
// return nil if id is not exist
func (ral *RowActionList) Find(id string) (action.RowAction, error) {
	for _, a := range ral.list {
		if a.ID() == id {
			return a, nil
		}
	}
	return nil, errors.Wrap(ErrInvalidRowActionID).FillDebugArgs(id)
}

// RowActionFilter ...
type RowActionFilter func(grpcCtx context.Context, rowData interface{}, actionID string) (support bool)

// TableMetaData table metadata
type TableMetaData interface {
	ID() string
	Name() string
	Desc() string
	ExportFilePrefix() string
	// Columns 获取所有的列
	Columns() *TableColumnDescriptorList
	// ColumnsWithoutInternal 获取所有的列(排除内部列)
	ColumnsWithoutInternal() *TableColumnDescriptorList
	// Print 打印文档
	Print(io.Writer)
	// EmptyTpl 空模板
	EmptyTpl(ctx context.Context) *cap.Template
	// DefaultTpl
	DefaultTpl(ctx context.Context) *cap.Template
	// 获取Key列ID
	KeyColumnID() string
	// 校验模板合法性
	ValidateTpl(tpl *cap.Template, userLanguage string) error
	// 行数据结构（非指针）
	RowDataType() reflect.Type

	// RowActions ....
	// 添加行操作
	AddRowActions(action ...action.RowAction) error
	// 获取行操作列表
	GetRowActions(grpcCtx context.Context, rowData interface{}) *RowActionList
	// 获取所有操作
	GetAllActions() []action.RowAction
	// SetRowActionsFilter 设置行操作过滤器，根据 行数据+操作ID 返回 是否支持该操作
	// 用于做权限过滤等操作
	SetRowActionsFilter(RowActionFilter)
}

// TableMetaDataImpl table metadata implementation
type TableMetaDataImpl struct {
	id                string
	name              string
	desc              string
	columns           *TableColumnDescriptorList
	dataLock          sync.Mutex
	defaultTplWrapper func(*cap.Template) *cap.Template
	keyColID          string
	rowDataType       reflect.Type
	rowActions        *RowActionList
	rowActionFilter   RowActionFilter
}

// ID ...
func (tmd *TableMetaDataImpl) ID() string {
	return tmd.id
}

// SetRowDataType ...
func (tmd *TableMetaDataImpl) SetRowDataType(t reflect.Type) {
	tmd.rowDataType = t
}

// Name ...
func (tmd *TableMetaDataImpl) Name() string {
	return tmd.name
}

// Desc ...
func (tmd *TableMetaDataImpl) Desc() string {
	return tmd.desc
}

// ExportFilePrefix ...
func (tmd *TableMetaDataImpl) ExportFilePrefix() string {
	return tmd.name + "_"
}

// Columns ...
func (tmd *TableMetaDataImpl) Columns() *TableColumnDescriptorList {
	return tmd.columns
}

// ColumnsWithoutInternal 去除内部字段的列...
func (tmd *TableMetaDataImpl) ColumnsWithoutInternal() *TableColumnDescriptorList {
	rsp := &TableColumnDescriptorList{}
	for _, c := range tmd.columns.list {
		if !c.Internal {
			rsp.list = append(rsp.list, c)
		}
	}
	return rsp
}

// AddColumns ...
func (tmd *TableMetaDataImpl) AddColumns(columns ...*TableColumnDescriptor) error {
	for _, c := range columns {
		err := tmd.addColumns(c)
		if err != nil {
			return err
		}
	}
	return nil
}

// RowDataType 行数据结构（非指针）
func (tmd *TableMetaDataImpl) RowDataType() reflect.Type {
	return tmd.rowDataType
}

func (tmd *TableMetaDataImpl) addColumns(column *TableColumnDescriptor) error {
	tmd.dataLock.Lock()
	defer tmd.dataLock.Unlock()
	for _, c := range tmd.columns.list {
		if column.ID == c.ID {
			return errors.Wrap(ErrDuplicateColumnID).FillDebugArgs(column.Name).Log()
		}
		if column.Name == c.Name {
			return errors.Wrap(ErrDuplicateColumnName).FillDebugArgs(column.Name).Log()
		}
	}
	tmd.columns.list = append(tmd.columns.list, column)
	err := tmd.columns.generateIndex()
	if err != nil {
		return errors.Wrap(err).Log()
	}

	return nil
}

// KeyColumnID 获取Key列ID
func (tmd *TableMetaDataImpl) KeyColumnID() string {
	return tmd.keyColID
}

// Print ...
func (tmd *TableMetaDataImpl) Print(w io.Writer) {
	fmt.Fprintln(w, "Table\t\t:\t", tmd.id, tmd.name)
	fmt.Fprintln(w, "Desc\t\t:\t", tmd.desc)
	fmt.Fprintln(w, "Exported as\t:\t", tmd.ExportFilePrefix())
	fmt.Fprintln(w, "BindStruct\t:\t", tmd.rowDataType.String())
	fmt.Fprintln(w, "Actions\t\t:")
	if tmd.rowActions != nil {
		for _, a := range tmd.rowActions.list {
			fmt.Fprintf(w, "[%s:\t%s(%s)]\n", cap.RowActionType_name[int32(a.Type())], a.Name(), a.ID())
		}
	}
	for _, c := range tmd.columns.list {
		fmt.Fprintln(w, "\t---------------------------------------------------------------")
		fmt.Fprintln(w, "\t", "Profile\t|\t", c.ID, c.Name)
		fmt.Fprintln(w, "\t", "Type\t\t|\t", cap.ValueType_name[int32(c.ValueType)], c.ValueFormat)
		fmt.Fprintln(w, "\t", "ColWidth\t|\t", c.ColWidth)
		aggMethods := []string{}
		for _, agg := range c.SupportedAggregateMethod {
			aggMethods = append(aggMethods, cap.AggregateMethod_name[int32(agg)])
		}
		if len(aggMethods) > 0 {
			fmt.Fprintln(w, "\t", "Aggregate\t|\t", strings.Join(aggMethods, ","))
		}
		if len(c.SupportedFilters) > 0 {
			var ops []string
			for _, o := range c.SupportedFilters {
				ops = append(ops, o.Name())
			}
			fmt.Fprintln(w, "\t", "Filters\t|\t", strings.Join(ops, ", "))
		}
		if c.Link != nil {
			fmt.Fprintln(w, "\t", "Link\t\t|\t", fmt.Sprintf("%s.%s -> %s.%s:%s",
				tmd.ID(), c.Link.LocalColID, c.Link.RemoteTableID,
				c.Link.RemoteSearchColID, c.Link.RemoteValueColID))
		}
		if c.ValueType == cap.ValueType_VT_OPTION {
			fmt.Fprintln(w, "\t", "OptionID\t|\t", fmt.Sprintf(`<a href="/options/%s">%s</a>`, c.DataType.String(), c.DataType.String()))
			options, _ := GlobalTableRegistry().OptionReg.GetOptions(c.DataType.String())
			optionStr := ""
			for j, opt := range options {
				optionStr += fmt.Sprintf("%s(%d)", opt.Name, opt.Id)
				if j != len(options)-1 {
					optionStr += ", "
				}
			}
			fmt.Fprintln(w, "\t", "Options\t|\t", optionStr)
		}
	}
	fmt.Fprintln(w, "\t---------------------------------------------------------------")
}

// NewTableMetaData ...
func NewTableMetaData(id string, name string, desc string,
	columns []*TableColumnDescriptor, defaultTplWrapper ...func(*cap.Template) *cap.Template) (*TableMetaDataImpl, error) {
	tcdl, err := NewTableColumnDescriptorList(columns)
	if err != nil {
		return nil, errors.Wrap(err).Log()
	}
	tmdl := &TableMetaDataImpl{
		id:      id,
		name:    name,
		desc:    desc,
		columns: tcdl,
	}
	if len(defaultTplWrapper) > 0 {
		tmdl.defaultTplWrapper = defaultTplWrapper[0]
	}
	return tmdl, err
}

// TableMetaReg table metadata registry
type TableMetaReg struct {
	*registryContainer
}

// Register registers function
func (vr *TableMetaReg) Register(tmd TableMetaData) error {
	defer fmt.Println("Table ==>", tmd.ID(), "registered")
	return vr.registryContainer.Register(tmd)
}

// Find find function with id
func (vr *TableMetaReg) Find(id string) (TableMetaData, error) {
	v, err := vr.registryContainer.Find(id)
	if err != nil {
		return nil, err
	}
	return v.(TableMetaData), nil
}

// ValidateLinks ...
func (vr *TableMetaReg) ValidateLinks() error {
	var validateErr error
	vr.reg.Range(func(_, v interface{}) bool {
		tmd := v.(TableMetaData)
		for _, c := range tmd.Columns().List() {
			// 三个key都必须存在，并且被搜索的key支持EQ/IN
			if c.Link != nil && c.Link.LocalColID != c.ID {
				localCol, err := tmd.Columns().Find(c.Link.LocalColID)
				if err != nil {
					validateErr = errors.Wrap(err).Triggers(ErrInvalidLink).FillDebugArgs(tmd.ID(), c.ID, c.Link.String()).Log()
					break
				}
				remoteTmd, err := vr.Find(c.Link.RemoteTableID)
				if err != nil {
					validateErr = errors.Wrap(err).Triggers(ErrInvalidLink).FillDebugArgs(tmd.ID(), c.ID, c.Link.String()).Log()
					break
				}
				rsc, err := remoteTmd.Columns().Find(c.Link.RemoteSearchColID)
				if err != nil {
					validateErr = errors.Wrap(err).Triggers(ErrInvalidLink).FillDebugArgs(tmd.ID(), c.ID, c.Link.String()).Log()
					break
				}
				supportLink := false
				for _, f := range rsc.SupportedFilters {
					if f.ID() == "builtin.EQ" || f.ID() == "builtin.IN" {
						supportLink = true
					}
				}
				if !supportLink {
					validateErr = errors.Wrap(ErrInvalidLink).FillDebugArgs(tmd.ID(), c.ID, c.Link.String()).Log()
					break
				}
				remoteCol, err := remoteTmd.Columns().Find(c.Link.RemoteValueColID)
				if err != nil {
					validateErr = errors.Wrap(err).Triggers(ErrInvalidLink).FillDebugArgs(tmd.ID(), c.ID, c.Link.String()).Log()
					break
				}
				// local列支持IN的前提下，复制remote列的筛选条件
				for _, sf := range localCol.SupportedFilters {
					if sf.ID() == "builtin.IN" {
						c.SupportedFilters = remoteCol.SupportedFilters
						break
					}
				}
			}
		}
		if validateErr != nil {
			return false
		}
		return true
	})
	return validateErr
}

const DefaultTemplateID = "TPL_DEFAULT"
const DefaultTemplateName = "缺省模板"
const DefaultTemplateTrKey = "default_template_name"

const EmptyTemplateID = "TPL_EMPTY"
const EmptyTemplateName = "空模板"

var defaultTemplateUser = &cap.UserInfo{
	Id:          -1,
	UserName:    "system",
	DisplayName: "系统",
}

var defaultTemplateTime = time.Date(2000, 1, 1, 0, 0, 0, 0, time.Local).Format(time.RFC3339)

// func (tmd *TableMetaDataImpl) parseURLQuery(ctx context.Context, tpl *cap.Template) {
// 	// parse conditions from url query and set into default url
// 	ref, _ := ptk.GetRefererFromCtx(ctx)
// 	u, _ := url.Parse(ref)
// 	if u != nil {
// 		queries := u.Query()
// 		// enableUrlFilter to enable this filter
// 		if _, ok := queries["enableUrlFilter"]; !ok {
// 			return
// 		}
// 		for k, vs := range queries {
// 			if col, _ := tmd.Columns().Find(k); col != nil {
// 				// k=v1+v2+v3
// 				// support builtin.EQ && builtin.IN
// 				// params cannot be parsed will be ignored
// 				var parsedValues []*cap.Value
// 				{
// 					// parse values
// 					switch col.ValueType {
// 					case cap.ValueType_VT_STRING:
// 						for _, v := range vs {
// 							parsedValues = append(parsedValues, &cap.Value{V: &cap.Value_VString{VString: v}})
// 						}
// 					case cap.ValueType_VT_INT:
// 						for _, v := range vs {
// 							if vInt, err := strconv.Atoi(v); err == nil {
// 								parsedValues = append(parsedValues, &cap.Value{V: &cap.Value_VInt{VInt: int32(vInt)}})
// 							}
// 						}
// 					case cap.ValueType_VT_OPTION:
// 						for _, v := range vs {
// 							if vInt, err := strconv.Atoi(v); err == nil {
// 								parsedValues = append(parsedValues,
// 									&cap.Value{V: &cap.Value_VOption{VOption: &cap.OptionValue{Id: int32(vInt)}}})
// 							}
// 						}
// 					case cap.ValueType_VT_DOUBLE:
// 						for _, v := range vs {
// 							if vFloat, err := strconv.ParseFloat(v, 64); err == nil {
// 								parsedValues = append(parsedValues, &cap.Value{V: &cap.Value_VDouble{VDouble: vFloat}})
// 							}
// 						}
// 					case cap.ValueType_VT_BOOLEAN:
// 						for _, v := range vs {
// 							if vBool, err := strconv.ParseBool(v); err == nil {
// 								parsedValues = append(parsedValues, &cap.Value{V: &cap.Value_VBool{VBool: vBool}})
// 							}
// 						}
// 					case cap.ValueType_VT_TIME:
// 						for _, v := range vs {
// 							parsedValues = append(parsedValues, &cap.Value{V: &cap.Value_VTime{VTime: v}})
// 						}
// 					case cap.ValueType_VT_DATE:
// 						for _, v := range vs {
// 							parsedValues = append(parsedValues, &cap.Value{V: &cap.Value_VDate{VDate: v}})
// 						}
// 					default:
// 						// not supported
// 					}
// 				}
// 				if len(parsedValues) > 0 {
// 					var supportEQ, supportIN bool
// 					for _, sf := range col.SupportedFilters {
// 						if sf.ID() == "builtin.EQ" {
// 							supportEQ = true
// 						} else if sf.ID() == "builtin.IN" {
// 							supportIN = true
// 						}
// 					}
// 					if len(parsedValues) == 1 && supportEQ {
// 						tpl.Body.Filter.Conditions = append(tpl.Body.Filter.Conditions, &cap.Condition{
// 							ColumnId:   k,
// 							OperatorId: "builtin.EQ",
// 							Values:     []*cap.FilterValue{{LiteralValues: parsedValues[0]}},
// 						})
// 					} else if len(parsedValues) > 1 && supportIN {
// 						condition := &cap.Condition{
// 							ColumnId:   k,
// 							OperatorId: "builtin.IN",
// 							Values:     []*cap.FilterValue{},
// 						}
// 						for _, pv := range parsedValues {
// 							condition.Values =
// 								append(condition.Values, &cap.FilterValue{LiteralValues: pv})
// 						}
// 						tpl.Body.Filter.Conditions = append(tpl.Body.Filter.Conditions, condition)
// 					}
// 				}
// 			}
// 		}
// 	}
// }

// DefaultTpl gets default template
func (tmd *TableMetaDataImpl) DefaultTpl(ctx context.Context) *cap.Template {
	userLanguage := i18n.GetUserLanguageByMeta(ctx)
	tpl := tmd.EmptyTpl(ctx)
	tpl.Id = DefaultTemplateID
	tpl.Name = i18n.GetTranslation(userLanguage, DefaultTemplateTrKey)
	// TODO. 20250108
	// tmd.parseURLQuery(ctx, tpl)
	// url 指定后，不再做用户自定义默认条件
	if len(tpl.Body.Filter.Conditions) == 0 {
		if tmd.defaultTplWrapper != nil {
			tpl = tmd.defaultTplWrapper(tpl)
		}
	}
	return tpl
}

// EmptyTpl gets empty template
func (tmd *TableMetaDataImpl) EmptyTpl(ctx context.Context) *cap.Template {
	userLanguage := i18n.GetUserLanguageByMeta(ctx)
	tpl := &cap.Template{
		Id:      EmptyTemplateID,
		Name:    EmptyTemplateName,
		TableId: tmd.id,
		FileInfo: &cap.FileInfo{
			Access:     cap.FileAccessType_TA_PUBLIC,
			CreateUser: defaultTemplateUser,
			CreateTime: defaultTemplateTime,
		},
		Body: &cap.TemplateBody{
			Filter: &cap.FilterBody{},
			Output: &cap.OutputBody{},
		},
	}
	for _, c := range tmd.columns.list {
		if !c.Internal {
			tpl.Body.Output.VisibleColumns = append(tpl.Body.Output.VisibleColumns,
				c.ToTemplateColumn(userLanguage, tmd.ID()))
		}
	}
	return tpl
}

func getVT(fv *cap.FilterValue) (cap.ValueType, error) {
	lv := fv.LiteralValues
	if lv == nil {
		return 0, errors.Wrap(ErrIllegalArguments).FillDebugArgs("FilterValue.LiteralValues == NULL")
	}
	switch lv.V.(type) {
	case *cap.Value_VString:
		return cap.ValueType_VT_STRING, nil
	case *cap.Value_VInt:
		return cap.ValueType_VT_INT, nil
	case *cap.Value_VDouble:
		return cap.ValueType_VT_DOUBLE, nil
	case *cap.Value_VDate:
		return cap.ValueType_VT_DATE, nil
	case *cap.Value_VTime:
		return cap.ValueType_VT_TIME, nil
	case *cap.Value_VBool:
		return cap.ValueType_VT_BOOLEAN, nil
	case *cap.Value_VOption:
		return cap.ValueType_VT_OPTION, nil
	default:
		return cap.ValueType_VT_STRING, nil
	}

}

// ValidateTpl 校验模板合法性
func (tmd *TableMetaDataImpl) ValidateTpl(tpl *cap.Template, userLanguage string) error {
	if tpl.Body == nil {
		tpl.Body = &cap.TemplateBody{}
	}
	if tpl.Body.Filter == nil {
		tpl.Body.Filter = &cap.FilterBody{}
	}
	if tpl.Body.Output == nil {
		tpl.Body.Output = &cap.OutputBody{}
	}
	// 校验参数
	for _, f := range tpl.Body.Filter.Conditions {
		desc, err := tmd.Columns().Find(f.ColumnId)
		if err != nil {
			return errors.Wrap(err).Log()
		}
		filterSupported := false
		for _, sf := range desc.SupportedFilters {
			if f.OperatorId == sf.ID() {
				switch sf.FilterValueType() {
				case cap.FilterValueType_FVT_NULL:
					if len(f.Values) == 0 {
						filterSupported = true
					} else {
						return errors.Wrap(ErrInvalidConditionValue).FillDebugArgs("0", len(f.Values))
					}
				case cap.FilterValueType_FVT_SINGLE:
					if len(f.Values) == 1 {
						// 字符串默认空值
						if f.Values[0].LiteralValues == nil || f.Values[0].LiteralValues.V == nil {
							if desc.ValueType == cap.ValueType_VT_STRING {
								f.Values[0].LiteralValues = &cap.Value{V: &cap.Value_VString{VString: ""}}
							} else {
								return errors.Wrap(ErrInvalidConditionValueType).FillDebugArgs(
									i18n.GetTranslation(userLanguage, desc.ID)+" "+f.OperatorId, cap.ValueType_name[int32(desc.ValueType)],
									"NULL",
								)
							}
						}
						vvt, err := getVT(f.Values[0])
						if err != nil {
							return errors.Wrap(err).Log()
						}
						if desc.ValueType != vvt {
							return errors.Wrap(ErrInvalidConditionValueType).FillDebugArgs(
								i18n.GetTranslation(userLanguage, desc.ID)+" "+f.OperatorId, cap.ValueType_name[int32(desc.ValueType)],
								cap.ValueType_name[int32(vvt)],
							)
						}
						filterSupported = true
					} else {
						return errors.Wrap(ErrInvalidConditionValue).FillDebugArgs("1", len(f.Values))
					}
				case cap.FilterValueType_FVT_MULTIPLE:
					if len(f.Values) >= 1 {
						for i, v := range f.Values {
							// 字符串默认空值
							if f.Values[i].LiteralValues == nil || f.Values[i].LiteralValues.V == nil {
								if desc.ValueType == cap.ValueType_VT_STRING {
									f.Values[i].LiteralValues = &cap.Value{V: &cap.Value_VString{VString: ""}}
								} else {
									return errors.Wrap(ErrInvalidConditionValueType).FillDebugArgs(
										i18n.GetTranslation(userLanguage, desc.ID)+" "+f.OperatorId, cap.ValueType_name[int32(desc.ValueType)],
										"NULL",
									)
								}
							}
							vvt, err := getVT(v)
							if err != nil {
								return errors.Wrap(err).Log()
							}
							if desc.ValueType != vvt {
								return errors.Wrap(ErrInvalidConditionValueType).FillDebugArgs(
									i18n.GetTranslation(userLanguage, desc.ID)+" "+f.OperatorId, cap.ValueType_name[int32(desc.ValueType)],
									cap.ValueType_name[int32(vvt)],
								)
							}
						}
						filterSupported = true
					} else {
						return errors.Wrap(ErrInvalidConditionValue).FillDebugArgs("1+", len(f.Values))
					}
				}
				break
			}
		}
		if !filterSupported {
			return errors.Wrap(ErrOperatorNotSupported).FillDebugArgs(f.OperatorId, tmd.id, desc.ID).Log()
		}
	}
	// 校验输出
	if len(tpl.Body.Output.VisibleColumns) == 0 {
		return errors.Wrap(ErrTplNullOutput).Log()
	}
	cols := make(map[string]interface{})
	for _, c := range tpl.Body.Output.VisibleColumns {
		if _, ok := cols[c.ColumnId]; ok {
			return errors.Wrap(ErrDuplicateColumnID).FillDebugArgs(c.ColumnId).Log()
		}
		cols[c.ColumnId] = nil
		desc, err := tmd.Columns().Find(c.ColumnId)
		if err != nil {
			return err
		}
		if c.AggregateMethod != cap.AggregateMethod_AM_NONE {
			aggSupported := false
			for _, method := range desc.SupportedAggregateMethod {
				if method == c.AggregateMethod {
					aggSupported = true
					break
				}
			}
			if !aggSupported {
				return errors.Wrap(ErrAggregateNotSupported).FillDebugArgs(tmd.Name, desc.ID).Log()
			}
		}
	}
	return nil
}

// AddRowActions ....
// 添加行操作
func (tmd *TableMetaDataImpl) AddRowActions(action ...action.RowAction) error {
	if tmd.rowActions == nil {
		tmd.rowActions = &RowActionList{}
	}
	tmd.rowActions.list = append(tmd.rowActions.list, action...)
	return nil
}

// GetRowActions 获取行操作列表
func (tmd *TableMetaDataImpl) GetRowActions(grpcCtx context.Context, rowData interface{}) *RowActionList {
	if tmd.rowActions == nil {
		tmd.rowActions = &RowActionList{}
	}
	if tmd.rowActionFilter == nil || rowData == nil {
		return tmd.rowActions
	}
	var ret []action.RowAction
	for _, ra := range tmd.rowActions.list {
		if tmd.rowActionFilter(grpcCtx, rowData, ra.ID()) {
			ret = append(ret, ra)
		}
	}
	return &RowActionList{list: ret}
}

// GetAllActions 获取所有操作
func (tmd *TableMetaDataImpl) GetAllActions() []action.RowAction {
	if tmd.rowActions == nil {
		return []action.RowAction{}
	}
	var actions []action.RowAction
	for _, action := range tmd.rowActions.list {
		if action != nil {
			actions = append(actions, action)
		}
	}
	return actions
}

// SetRowActionsFilter 设置行操作过滤器，根据 行数据+操作ID 返回 是否支持该操作
// 用于做权限过滤等操作
func (tmd *TableMetaDataImpl) SetRowActionsFilter(filter RowActionFilter) {
	tmd.rowActionFilter = filter
}

// NewTableMetaReg creates value filter function registry
func NewTableMetaReg() *TableMetaReg {
	return &TableMetaReg{registryContainer: newRegistryContainer(reflect.TypeOf((*TableMetaData)(nil)).Elem(), []string{"Name"})}
}
