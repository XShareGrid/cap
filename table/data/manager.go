package data

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"sync"

	"github.com/XShareGrid/cap/database/mysql"
	"github.com/XShareGrid/cap/i18n"
	"github.com/XShareGrid/cap/msg/errors"
	"github.com/XShareGrid/cap/msg/errors/handle"
	"github.com/XShareGrid/cap/table/action"
	"github.com/XShareGrid/cap/table/data/driver"
	"github.com/XShareGrid/cap/table/data/utils"
	"github.com/XShareGrid/cap/table/operator/builtin"
	cap "github.com/XShareGrid/cap/table/proto/go"
	"github.com/XShareGrid/cap/table/registry"
	"github.com/XShareGrid/cap/table/template"
	"golang.org/x/text/language"
	"google.golang.org/grpc/codes"
)

// Manager data driver registry
type Manager struct {
	m sync.Map
}

var globalManager *Manager

func init() {
	globalManager = &Manager{}
}

// GlobalManager global data manager
func GlobalManager() *Manager {
	if globalManager == nil {
		globalManager = &Manager{}
	}
	return globalManager
}

// RegisterDriver registers driver
func (d *Manager) RegisterDriver(tableID string, driver driver.Driver) error {
	if _, ok := d.m.Load(tableID); ok {
		return errors.Wrap(ErrDuplicateDriverForTable).FillDebugArgs(tableID).Log()
	}
	d.m.Store(tableID, driver)
	return nil
}

// UnRegister registers driver
func (d *Manager) UnRegister(tableID string) error {
	if _, ok := d.m.Load(tableID); !ok {
		return errors.Wrap(ErrDriverNotFoundForTable).FillDebugArgs(tableID).Log()
	}
	d.m.Delete(tableID)
	return nil
}

// FindRow ...
func (d *Manager) FindRow(grpcCtx context.Context, ss *mysql.Session, tpl *cap.Template, id string) (*cap.GetTableRowByIDRsp, error) {
	rsp := &cap.GetTableRowByIDRsp{}
	tmd, err := registry.GlobalTableRegistry().TableMetaReg.Find(tpl.TableId)
	if err != nil {
		return rsp, errors.Wrap(err).Log()
	}
	tpl.Body.Filter = &cap.FilterBody{
		Conditions: []*cap.Condition{
			{
				ColumnId:   tmd.KeyColumnID(),
				OperatorId: builtin.EQ.ID(),
				Values: []*cap.FilterValue{
					{
						LiteralValues: &cap.Value{V: &cap.Value_VString{VString: id}},
					},
				},
			},
		},
	}
	rowsRsp, err := d.FindRows(grpcCtx, ss, tpl, &cap.PageParam{PageSize: 0}, nil)
	if err != nil {
		return rsp, errors.Wrap(err).Log()
	}
	if len(rowsRsp.Rows) == 0 {
		return nil, errors.Wrap(ErrNotResultForID).FillDebugArgs(id)
	}
	rsp.Row = rowsRsp.Rows[0]
	return rsp, nil
}

var langCode = language.Make("zh-CN")

// ParseTpl parse query template
func ParseTpl(ctx context.Context, ss *mysql.Session, tableID string, tplQuery *cap.TemplateQuery) (*cap.Template, error) {
	tpl := &cap.Template{TableId: tableID}
	tmd, err := registry.GlobalTableRegistry().TableMetaReg.Find(tableID)
	if err != nil {
		return tpl, handle.Handle(context.Background(), err).Log().GRPCErr(codes.Unknown, langCode)
	}
	if tplBody, ok := tplQuery.Tpl.(*cap.TemplateQuery_TmpTpl); ok {
		tpl.Body = tplBody.TmpTpl
	} else if tplID, ok := tplQuery.Tpl.(*cap.TemplateQuery_TplId); ok {
		if tplID.TplId == registry.EmptyTemplateID {
			return tmd.EmptyTpl(ctx), nil
		} else if tplID.TplId == registry.DefaultTemplateID {
			return tmd.DefaultTpl(ctx), nil
		} else {
			var err error
			tpl, err = template.GlobalManager().FindTemplate(ctx, ss, tplID.TplId)
			if err != nil {
				return tpl, handle.Handle(context.Background(), err).Log().GRPCErr(codes.Unknown, langCode)
			}
			tpl.TableId = tableID
		}
	}
	return tpl, nil
}

// FindRows ..
func (d *Manager) FindRows(grpcCtx context.Context, ss *mysql.Session, tpl *cap.Template, pageParam *cap.PageParam,
	orderParam *cap.OrderParam,
) (*cap.GetTableRowsRsp, error) {
	newCols := []*cap.TemplateColumn{}
	// 处理tpl，过滤invisible的列
	for _, col := range tpl.Body.Output.VisibleColumns {
		if col.Visible {
			newCols = append(newCols, col)
		}
	}
	tpl.Body.Output.VisibleColumns = newCols
	return d.findRows(grpcCtx, ss, tpl, pageParam, orderParam, false)
}

// FindRowsLite ..
func (d *Manager) FindRowsLite(grpcCtx context.Context, ss *mysql.Session, tableID string, query string) (rsp *cap.GetTableRowsLiteRsp, err error) {
	rsp = &cap.GetTableRowsLiteRsp{}
	tmd, err := registry.GlobalTableRegistry().TableMetaReg.Find(tableID)
	if err != nil {
		return rsp, errors.Wrap(err).Log()
	}
	qs, err := url.ParseQuery(query)
	if err != nil {
		return rsp, errors.Wrap(err).Log()
	}
	conditions := []*driver.Condition{}
	for k, v := range qs {
		col, err := tmd.Columns().Find(k)
		if err != nil {
			return rsp, errors.Wrap(err).Log()
		}
		operator := builtin.EQ.ID()
		if len(v) == 0 {
			continue
		} else if len(v) > 1 {
			operator = builtin.IN.ID()
		} else if len(v) == 1 {
			re := regexp.MustCompile(`^MAX\((.*?)\)$`)
			match := re.FindStringSubmatch(v[0])
			if len(match) > 1 {
				v[0] = match[1]
				operator = builtin.LT.ID()
			} else {
				re := regexp.MustCompile(`^MIN\((.*?)\)$`)
				match := re.FindStringSubmatch(v[0])
				if len(match) > 1 {
					v[0] = match[1]
					operator = builtin.GT.ID()
				}
			}
		}
		valueList := make([]interface{}, len(v))
		for i := range v {
			var vTrans interface{} = v[i]
			switch col.ValueType {
			case cap.ValueType_VT_INT:
				vTrans, err = strconv.ParseInt(v[i], 10, 64)
				if err != nil {
					return rsp, errors.Wrap(err).Log()
				}
			case cap.ValueType_VT_OPTION:
				i32, err := strconv.ParseInt(v[i], 10, 64)
				if err != nil {
					return rsp, errors.Wrap(err).Log()
				}
				vTrans = utils.Option(int32(i32))
			case cap.ValueType_VT_DOUBLE:
				vTrans, err = strconv.ParseFloat(v[i], 64)
				if err != nil {
					return rsp, errors.Wrap(err).Log()
				}
			case cap.ValueType_VT_BOOLEAN:
				vTrans, err = strconv.ParseBool(v[i])
				if err != nil {
					return rsp, errors.Wrap(err).Log()
				}
			case cap.ValueType_VT_DATE:
				vTrans = utils.Date(v[i])
			case cap.ValueType_VT_TIME:
				vTrans = utils.Time(v[i])
			}
			valueList[i] = vTrans
		}
		conditions = append(conditions, driver.NewCondition(k, operator, valueList...))
	}
	outputCols := make([]string, len(tmd.Columns().List()))
	for i, col := range tmd.Columns().List() {
		outputCols[i] = col.ID
	}
	tpl := utils.NewTmpTpl(tableID, conditions, outputCols)
	r, err := d.findRows(grpcCtx, ss, tpl, &cap.PageParam{}, &cap.OrderParam{}, false)
	if err != nil {
		return rsp, errors.Wrap(err).Log()
	}
	rsp.Rows = make([]string, len(r.Rows))
	for i := range r.Rows {
		m := make(map[string]interface{})
		for _, c := range r.Rows[i].Cells {
			switch c.Value.V.(type) {
			case *cap.Value_VString:
				m[c.ColumnId] = c.Value.GetVString()
			case *cap.Value_VInt:
				m[c.ColumnId] = c.Value.GetVInt()
			case *cap.Value_VDouble:
				m[c.ColumnId] = c.Value.GetVDouble()
			case *cap.Value_VDate:
				m[c.ColumnId] = c.Value.GetVDate()
			case *cap.Value_VTime:
				m[c.ColumnId] = c.Value.GetVTime()
			case *cap.Value_VBool:
				m[c.ColumnId] = c.Value.GetVBool()
			case *cap.Value_VOption:
				m[c.ColumnId] = c.Value.GetVOption().Id
			default:
				m[c.ColumnId] = nil
			}
			js, _ := json.Marshal(m)
			rsp.Rows[i] = string(js)
		}
	}
	return rsp, nil

}

func linkAddQuery(link string, k, v string) string {
	u, err := url.Parse(link)
	if err != nil {
		return link
	}
	query := u.Query()
	query.Add(k, v)
	u.RawQuery = query.Encode()
	return u.String()
}

// 链接条件
type linkCondition struct {
	condition *cap.Condition
	desc      *registry.TableColumnDescriptor
}

// remoteTableID - remoteSearchCol - conditions
type linkConditionGroup map[string]map[string][]linkCondition

// 将链接条件转换为本地条件
func (d *Manager) transLinkConditions(grpcCtx context.Context, ss *mysql.Session, linkConditions linkConditionGroup) ([]*driver.Condition, error) {
	var localConditions []*driver.Condition
	// 1. 聚合链接remote表格的条件
	for remoteTableID, remoteSearchConditions := range linkConditions {
		// make tpl
		tpl := &cap.Template{TableId: remoteTableID}
		tplBody := &cap.TemplateBody{Filter: &cap.FilterBody{}, Output: &cap.OutputBody{}}
		tpl.Body = tplBody
		for _, remoteConditions := range remoteSearchConditions {
			// 这里都是同一列的过滤
			var outputColID string
			var localColID string
			var localColDesc *registry.TableColumnDescriptor
			for _, rc := range remoteConditions {
				tplBody.Filter.Conditions = append(tplBody.Filter.Conditions,
					&cap.Condition{
						ColumnId:   rc.desc.Link.RemoteValueColID,
						OperatorId: rc.condition.OperatorId,
						Values:     rc.condition.Values,
					})
				outputColID = rc.desc.Link.RemoteSearchColID
				localColID = rc.desc.Link.LocalColID
				localColDesc = rc.desc
			}

			tplBody.Output.VisibleColumns = append(tplBody.Output.VisibleColumns,
				&cap.TemplateColumn{ColumnId: outputColID})

			rsp, err := d.findRows(grpcCtx, ss, tpl, &cap.PageParam{}, &cap.OrderParam{}, true)
			if err != nil {
				return nil, errors.Wrap(err).Log()
			}
			driverCondition := &driver.Condition{
				ColumnID:   localColID,
				OperatorID: "builtin.IN",
			}
			for _, row := range rsp.Rows {
				for _, cell := range row.Cells {
					if cell.ColumnId == outputColID {
						dv, err := utils.ParseConditionValue(localColDesc, &cap.FilterValue{LiteralValues: cell.Value})
						if err != nil {
							return localConditions, errors.Wrap(err).Log()
						}
						driverCondition.Values = append(driverCondition.Values, dv)
					}
				}
			}
			// 有一列检索结果为0，就返回空结果，无需再查
			if len(rsp.Rows) == 0 {
				return nil, nil
			}
			localConditions = append(localConditions, driverCondition)
		}
	}
	// 2. 查询链接表
	// 3. 转换为local条件
	return localConditions, nil
}

// findRows 内部find函数
// simpleFind = true 不处理link和action
func (d *Manager) findRows(grpcCtx context.Context, ss *mysql.Session, tpl *cap.Template, pageParam *cap.PageParam,
	orderParam *cap.OrderParam, simpleFind bool) (*cap.GetTableRowsRsp, error) {
	userLanguage := i18n.GetUserLanguageByMeta(grpcCtx)
	rsp := &cap.GetTableRowsRsp{}
	tmd, err := registry.GlobalTableRegistry().TableMetaReg.Find(tpl.TableId)
	if err != nil {
		return rsp, errors.Wrap(err).Log()
	}
	err = tmd.ValidateTpl(tpl, userLanguage)
	if err != nil {
		return rsp, errors.Wrap(err).Log()
	}
	dd, ok := d.m.Load(tpl.TableId)
	if !ok {
		return nil, errors.Wrap(ErrDriverNotFoundForTable).FillDebugArgs(tpl.TableId).Log()
	}
	dataDriver := dd.(driver.Driver)
	// 解释conditions/output/agg
	// conditions
	// link
	linkConditions := make(linkConditionGroup)
	conditions := make([]*driver.Condition, 0, len(tpl.Body.Filter.Conditions))
	for _, con := range tpl.Body.Filter.Conditions {
		values := make([]interface{}, len(con.Values))
		desc, err := tmd.Columns().Find(con.ColumnId)
		if err != nil {
			return rsp, errors.Wrap(err).Log()
		}
		// 在这里转换Link列的条件
		if desc.Link != nil {
			remoteSearchConditions, ok := linkConditions[desc.Link.RemoteTableID]
			if !ok {
				remoteSearchConditions = make(map[string][]linkCondition)
			}
			// 本地ID与remote搜索ID分组
			searchColID := desc.Link.LocalColID + desc.Link.RemoteSearchColID
			remoteConditions, ok := remoteSearchConditions[searchColID]
			if !ok {
				remoteConditions = []linkCondition{}
			}
			remoteSearchConditions[searchColID] = append(remoteConditions,
				linkCondition{condition: con, desc: desc})
			linkConditions[desc.Link.RemoteTableID] = remoteSearchConditions
			continue
		}
		for j, cv := range con.Values {
			values[j], err = utils.ParseConditionValue(desc, cv)
			if err != nil {
				return rsp, errors.Wrap(err).Log()
			}
		}
		conditions = append(conditions, driver.NewCondition(con.ColumnId, con.OperatorId, values...))
	}
	// 处理链接条件
	transLinkConditions, err := d.transLinkConditions(grpcCtx, ss, linkConditions)
	if err != nil {
		return nil, errors.Wrap(err).Log()
	}
	// 返回空结果
	if len(linkConditions) > 0 && len(transLinkConditions) == 0 {
		rsp.PageInfo = &cap.PageInfo{}
		return rsp, nil
	}
	conditions = append(conditions, transLinkConditions...)

	outputs, err := parseOutputColumns(tmd, tpl)
	if err != nil {
		return nil, errors.Wrap(err).Log()
	}
	aggParams := []*driver.AggregateColumn{}
	for _, c := range tpl.Body.Output.VisibleColumns {
		if c.AggregateMethod != cap.AggregateMethod_AM_NONE {
			aggParams = append(aggParams, &driver.AggregateColumn{
				ColumnID:        c.ColumnId,
				AggregateMethod: c.AggregateMethod,
			})
		}
	}
	rowsResults, err := dataDriver.FindRows(grpcCtx, ss, tmd, conditions, outputs, aggParams, pageParam, orderParam)
	if err != nil {
		return nil, errors.Wrap(err).Log()
	}
	var linkProcessed = false
	// index - cell link map
	var linkCols = make(map[int]*cap.Cell)
	// local search map col_id - index map
	var linkSearchCols = make(map[string]int)
	rsp.Rows = make([]*cap.TableRow, len(rowsResults.Rows))
	// TODO. 20250108 nouser
	// userInfo, _ := ptk.GetWebGRPCUserInfoFromCtx(grpcCtx)
	for i, row := range rowsResults.Rows {
		if tableRow, ok := row.(*cap.TableRow); ok {
			rsp.Rows[i] = tableRow
		} else {
			tableRow, err := mapStructToTableRow(userLanguage, row, outputs, tmd)
			if err != nil {
				return nil, errors.Wrap(err).Log()
			}
			rsp.Rows[i] = tableRow
		}
		// 假定每行提供cell组成都一致，在这里对link做一次性分析
		// 假定所有搜索列都在cells中存在
		if !simpleFind && !linkProcessed {
			for i, cell := range rsp.Rows[i].Cells {
				// 找到link的搜索列
				if cell.Link != nil && cell.Link.ColId != cell.ColumnId {
					linkCols[i] = cell
					linkSearchCols[cell.Link.ColId] = -1
				}
			}
			// 填充本地搜索列索引
			for i, cell := range rsp.Rows[i].Cells {
				if _, ok := linkSearchCols[cell.ColumnId]; ok {
					linkSearchCols[cell.ColumnId] = i
				}
			}
			linkProcessed = true
		}
		if !simpleFind {
			// Actions
			rowActions := tmd.GetRowActions(grpcCtx, row).List()
			if len(rowActions) > 0 {
				rsp.Rows[i].SupportedActions = make([]*cap.RowAction, len(rowActions))
			}
			for j, ra := range rowActions {
				actionName := i18n.GetTranslation(userLanguage, registry.TrKeyTableAction(tmd.ID(), ra.ID()))
				newAction := &cap.RowAction{
					Id:         ra.ID(),
					Name:       actionName,
					ActionType: ra.Type(),
				}
				if ra.Type() == cap.RowActionType_RAT_JSON_FORM {
					schema, err := ra.(*action.FormRowAction).Schema(row)
					if err != nil {
						return rsp, errors.Wrap(err).Log()
					}
					newAction.JsonFormSchema = schema

				} else if ra.Type() == cap.RowActionType_RAT_HREF {
					href := ra.(*action.HrefRowAction)
					hrefLink := href.Href(row)
					// TODO. 20250108 FIXME
					// if userInfo != nil {
					// 	hrefLink = linkAddQuery(hrefLink, ptk.QueryUserID, fmt.Sprintf("%d", userInfo.AccountId))
					// }
					newAction.HrefAction = &cap.HrefAction{
						Href:      hrefLink,
						HrefStyle: href.HrefStyle(),
					}
				}
				rsp.Rows[i].SupportedActions[j] = newAction
			}
		}
	}
	// TODO 根据Table和remote search聚合，提高查询效率

	// 执行link
	for cellColIndex, cell := range linkCols {
		rTmd, err := registry.GlobalTableRegistry().TableMetaReg.Find(cell.Link.RemoteTableId)
		if err != nil {
			return rsp, errors.Wrap(err).Log()
		}
		searchDesc, err := rTmd.Columns().Find(cell.Link.RemoteSearchColId)
		if err != nil {
			return rsp, errors.Wrap(err).Log()
		}
		emptyVal := utils.EmptyValue(searchDesc)

		// 使用EQ和IN两种方式，如果有IN则优先使用做批量查询
		var supportEQ, supportIN bool
		for _, f := range searchDesc.SupportedFilters {
			if f.ID() == builtin.EQ.ID() {
				supportEQ = true
			} else if f.ID() == builtin.IN.ID() {
				supportIN = true
				break
			}
		}
		linkTpl := utils.NewTmpTpl(rTmd.ID(), []*driver.Condition{},
			[]string{cell.Link.RemoteSearchColId, cell.Link.RemoteValueColId})

		linkSearchColIndex := linkSearchCols[cell.Link.ColId]
		if supportIN {
			// 批量查询
			searchValues := make([]*cap.Value, len(rowsResults.Rows))
			for i, row := range rsp.Rows {
				searchValues[i] = row.Cells[linkSearchColIndex].Value
			}
			filterValues := make([]*cap.FilterValue, len(searchValues))
			for i, sv := range searchValues {
				filterValues[i] = &cap.FilterValue{LiteralValues: sv}
			}
			// search value in
			linkTpl.Body.Filter = &cap.FilterBody{
				Conditions: []*cap.Condition{
					{
						ColumnId:   cell.Link.RemoteSearchColId,
						OperatorId: builtin.IN.ID(),
						Values:     filterValues,
					},
				},
			}
			remoteRows, err := d.findRows(grpcCtx, ss, linkTpl, &cap.PageParam{PageSize: 0}, &cap.OrderParam{}, true)
			if err != nil {
				return rsp, errors.Wrap(err).Log()
			}
			// 取结果 make a map
			var searchResults = make(map[string]*cap.Value)
			for _, r := range remoteRows.Rows {
				searchResults[r.Cells[0].Value.String()] = r.Cells[1].Value
			}
			for _, row := range rsp.Rows {
				row.Cells[cellColIndex].Value, ok = searchResults[row.Cells[linkSearchColIndex].Value.String()]
				if !ok {
					log.Printf("can not find link value for [%s] in table[%s.%s]",
						row.Cells[linkSearchColIndex].Value.String(),
						cell.Link.RemoteTableId, cell.Link.RemoteSearchColId)
					row.Cells[cellColIndex].Value = emptyVal
					row.Cells[cellColIndex].Values = []*cap.Value{emptyVal}
				} else {
					// row.Cells[cellColIndex].Value = remoteRows.Rows[0].Cells[1].Value
					row.Cells[cellColIndex].Values = []*cap.Value{row.Cells[cellColIndex].Value}
				}

			}
		} else if supportEQ {
			for _, row := range rsp.Rows {
				filterValue := &cap.FilterValue{LiteralValues: row.Cells[linkSearchColIndex].Value}
				// search value eq
				linkTpl.Body.Filter = &cap.FilterBody{
					Conditions: []*cap.Condition{
						{
							ColumnId:   cell.Link.RemoteSearchColId,
							OperatorId: builtin.EQ.ID(),
							Values:     []*cap.FilterValue{filterValue},
						},
					},
				}
				remoteRows, err := d.findRows(grpcCtx, ss, linkTpl, &cap.PageParam{PageSize: 1}, &cap.OrderParam{}, true)
				if err != nil {
					return rsp, errors.Wrap(err).Log()
				}
				if len(remoteRows.Rows) > 0 {
					row.Cells[cellColIndex].Value = remoteRows.Rows[0].Cells[1].Value
					row.Cells[cellColIndex].Values = []*cap.Value{remoteRows.Rows[0].Cells[1].Value}
				} else {
					row.Cells[cellColIndex].Value = emptyVal
					row.Cells[cellColIndex].Values = []*cap.Value{emptyVal}
				}
			}
		} else {
			return rsp, errors.Wrap(ErrTableColumnNotLinkable).FillDebugArgs(
				cell.Link.RemoteTableId, cell.Link.RemoteSearchColId)
		}

	}

	rsp.PageInfo = rowsResults.PageInfo

	if len(rsp.Rows) == 0 {
		return rsp, nil
	}
	// 找出映射关系
	firstRow := rsp.Rows[0]
	// columnid - index map
	rowMap := make(map[string]int)
	for i, cell := range firstRow.Cells {
		rowMap[cell.ColumnId] = i
	}
	// template index - data row index
	idxMapping := make(map[int]int)
	for tplIdx, tplCol := range tpl.Body.GetOutput().VisibleColumns {
		if rowIdx, ok := rowMap[tplCol.ColumnId]; ok {
			idxMapping[tplIdx] = rowIdx
		}
	}
	// 数据转换
	for _, row := range rsp.Rows {
		newCellList := make([]*cap.Cell,
			len(tpl.Body.GetOutput().VisibleColumns))
		for i := range newCellList {
			dataIdx := idxMapping[i]
			newCellList[i] = row.Cells[dataIdx]
		}
		row.Cells = newCellList
	}

	// 聚合结果
	if len(rowsResults.AggResults) > 0 {
		rsp.AggregateResult = make([]*cap.AggregateResult, len(rowsResults.AggResults))
		for i, r := range rowsResults.AggResults {
			desc, err := tmd.Columns().Find(r.ColumnID)
			if err != nil {
				return nil, errors.Wrap(err).Log()
			}
			v, err := utils.MapValue(userLanguage, r.Result, desc)
			if err != nil {
				return nil, errors.Wrap(err).Log()
			}
			rsp.AggregateResult[i] = &cap.AggregateResult{
				ColumnId:   r.ColumnID,
				ColumnName: desc.Name,
				Value:      v,
			}
		}
	}
	return rsp, nil
}

func parseOutputColumns(tmd registry.TableMetaData, tpl *cap.Template) ([]string, error) {
	visibleColumns := make(map[string]interface{})
	for _, col := range tpl.Body.Output.VisibleColumns {
		visibleColumns[col.ColumnId] = nil
		desc, err := tmd.Columns().Find(col.ColumnId)
		if err != nil {
			return nil, errors.Wrap(err).Log()
		}
		// 非自链接，添加被链接的列
		if desc.Link != nil && desc.Link.LocalColID != desc.ID {
			visibleColumns[desc.Link.LocalColID] = nil
		}
	}
	// 添加ID列，必选
	visibleColumns[tmd.KeyColumnID()] = nil
	// 添加列，添加
	for _, c := range tpl.Body.Filter.Conditions {
		visibleColumns[c.ColumnId] = nil
	}
	results := make([]string, 0, len(visibleColumns))
	for k := range visibleColumns {
		results = append(results, k)
	}
	return results, nil
}

func mapStructToTableRow(userLanguage string, v interface{}, outputColumns []string, tmd registry.TableMetaData) (*cap.TableRow, error) {
	// map
	if mv, ok := v.(map[string]interface{}); ok {
		row := &cap.TableRow{}
		row.Cells = make([]*cap.Cell, len(outputColumns))
		for i, c := range outputColumns {
			desc, err := tmd.Columns().Find(c)
			if err != nil {
				return nil, errors.Wrap(err).Log()
			}
			cell, err := utils.NewTableCell(userLanguage, mv[desc.Tag.Get("db")], desc)
			if err != nil {
				return nil, errors.Wrap(err).Log()
			}
			row.Cells[i] = cell
			if desc.IsKeyColumn {
				row.Id = cell.Value.GetVString()
			}
		}
		// 补ID列
		if row.Id == "" {
			keyColumn, err := tmd.Columns().Find(tmd.KeyColumnID())
			if err != nil {
				return nil, errors.Wrap(err).Log()
			}
			if rowID, ok := mv[keyColumn.Tag.Get("db")].(string); ok {
				row.Id = rowID
			} else {
				row.Id = fmt.Sprintf(keyColumn.ValueFormat, keyColumn.Tag.Get("db"))
			}
		}
		return row, nil
	}
	row := &cap.TableRow{}
	elem := reflect.ValueOf(v).Elem()
	row.Cells = make([]*cap.Cell, len(outputColumns))
	for i, c := range outputColumns {
		desc, err := tmd.Columns().Find(c)
		if err != nil {
			return nil, errors.Wrap(err).Log()
		}
		fv := elem.FieldByName(c)
		if !fv.IsValid() {
			return nil, errors.Wrap(ErrInvalidColumnFromStruct).FillDebugArgs(c, v)
		}
		cell, err := utils.NewTableCell(userLanguage, fv.Interface(), desc)
		if err != nil {
			return nil, errors.Wrap(err).Log()
		}
		row.Cells[i] = cell
		if desc.IsKeyColumn {
			row.Id = cell.Value.GetVString()
		}
	}
	// 补ID列
	if row.Id == "" {
		keyColumn, err := tmd.Columns().Find(tmd.KeyColumnID())
		if err != nil {
			return nil, errors.Wrap(err).Log()
		}
		fv := elem.FieldByName(keyColumn.ID)
		if !fv.IsValid() {
			return nil, errors.Wrap(ErrInvalidColumnFromStruct).FillDebugArgs(keyColumn.ID, v)
		}
		if rowID, ok := fv.Interface().(string); ok {
			row.Id = rowID
		} else {
			row.Id = fmt.Sprintf(keyColumn.ValueFormat, fv.Interface())
		}
	}
	return row, nil
}
