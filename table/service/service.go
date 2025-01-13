package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	proto "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/XShareGrid/cap/database/mysql"
	"github.com/XShareGrid/cap/i18n"
	"github.com/XShareGrid/cap/msg/errors"
	"github.com/XShareGrid/cap/msg/errors/handle"
	"github.com/XShareGrid/cap/table/action"
	"github.com/XShareGrid/cap/table/data"
	cap "github.com/XShareGrid/cap/table/proto/go"
	"github.com/XShareGrid/cap/table/registry"
	"github.com/XShareGrid/cap/table/template"
	"github.com/ahmetb/go-linq"

	excelize "github.com/xuri/excelize/v2"
)

// UserInfoProvider 账户信息提供者
type UserInfoProvider interface {
	GetCurrentUser(ctx context.Context) (*cap.UserInfo, error)
	GetUserInfoByID(id string) (*cap.UserInfo, error)
}

// TableWService ...
type TableWService struct {
	dbWrite          *mysql.DB
	dbRead           *mysql.DB
	userInfoProvider UserInfoProvider
}

// 读写分离
var RWSeparate = false

func (tws *TableWService) DBRead() *mysql.DB {
	if RWSeparate {
		return tws.dbRead
	}
	return tws.dbWrite
}

func (tws *TableWService) DBWrite() *mysql.DB {
	return tws.dbWrite
}

func rspOK(data proto.Message) *cap.CommonRsp {
	any, _ := anypb.New(data)
	return &cap.CommonRsp{
		Result:  0,
		Message: "ok",
		Data:    any,
	}
}

func rspErr(ctx context.Context, err error) (*cap.CommonRsp, error) {
	langCode := i18n.GetLanguageTypeByMeta(ctx)
	gErr := handle.Handle(ctx, err).Log().GRPCErr(codes.Unknown, langCode)
	errMsg := gErr.Error()
	if s, ok := status.FromError(gErr); ok {
		errMsg = fmt.Sprintf("%s: %s", s.Code().String(), s.Message())
	}
	return &cap.CommonRsp{
		Result:  1,
		Message: errMsg,
		Data:    nil,
	}, nil
}

/*************************************** 表 *******************************************/

// GetTableInfo 获取表信息
func (t *TableWService) GetTableInfo(ctx context.Context, req *cap.GetTableInfoReq) (*cap.CommonRsp, error) {
	rsp := &cap.GetTableInfoRsp{}
	userLanguage := i18n.GetUserLanguageByMeta(ctx)
	tmd, err := registry.GlobalTableRegistry().TableMetaReg.Find(req.TableId)
	if err != nil {
		return rspErr(ctx, err)
	}

	rsp.Id = tmd.ID()
	rsp.Name = i18n.GetTranslation(userLanguage, registry.TrKeyTableName(tmd.ID()))
	rsp.Desc = i18n.GetTranslation(userLanguage, registry.TrKeyTableDesc(tmd.ID()))
	rsp.ExportFilePrefix = rsp.Name + "_" + time.Now().Format("20060102")
	return rspOK(rsp), nil
}

// // 临时表格判断，query: tempTable=1
// // 临时表格不支持模板操作
func isTempTable(ctx context.Context) bool {
	// TODO. 20250108
	//ref, _ := ptk.GetRefererFromCtx(ctx)
	// if u, err := url.Parse(ref); err == nil {
	// 	if v, ok := u.Query()["tempTable"]; ok && len(v) > 0 && v[0] == "1" {
	// 		return true
	// 	}
	// }
	return false
}

/*************************************** 模板 *******************************************/

// GetTableTemplates 获取表模板信息
func (t *TableWService) GetTableTemplates(ctx context.Context, req *cap.GetTableTemplatesReq) (*cap.CommonRsp, error) {
	rsp := &cap.GetTableTemplatesRsp{}
	ss, err := t.DBRead().NewSessionWithCtx(ctx)
	if err != nil {
		return rspErr(ctx, err)
	}
	defer func() {
		ss.Close(err)
	}()
	tmd, err := registry.GlobalTableRegistry().TableMetaReg.Find(req.TableId)
	if err != nil {
		return rspErr(ctx, err)
	}
	tplList := []*cap.Template{tmd.DefaultTpl(ctx)}
	// TODO. 20250108
	if isTempTable(ctx) {
		rsp.Templates = tplList
		return rspOK(rsp), nil
	}
	user, err := t.userInfoProvider.GetCurrentUser(ctx)
	if err != nil {
		return rspErr(ctx, err)
	}
	tplListOwn, err := template.GlobalManager().FindTemplatesByTableAndCreateUser(ctx, ss, req.TableId, user.Id)
	if err != nil {
		return rspErr(ctx, err)
	}
	tplListShared, err := template.GlobalManager().FindTemplatesByTableAndShareUser(ctx, ss, req.TableId, user.Id)
	if err != nil {
		return rspErr(ctx, err)
	}
	tplListPublic, err := template.GlobalManager().FindPublicTemplatesByTable(ctx, ss, req.TableId, user.Id)
	if err != nil {
		return rspErr(ctx, err)
	}

	tplList = append(tplList, tplListOwn...)
	tplList = append(tplList, tplListShared...)
	tplList = append(tplList, tplListPublic...)
	linq.From(tplList).
		OrderByT(
			func(tpl *cap.Template) string {
				return tpl.FileInfo.CreateTime
			},
		).DistinctByT(func(tpl *cap.Template) string {
		return tpl.Id
	}).ToSlice(&tplList)
	rsp.Templates = tplList
	return rspOK(rsp), nil
}

// CreateTableTemplate 创建模板
func (t *TableWService) CreateTableTemplate(ctx context.Context, req *cap.CreateTableTemplateReq) (*cap.CommonRsp, error) {
	langCode := i18n.GetLanguageTypeByMeta(ctx)
	userLanguage := i18n.GetUserLanguageByMeta(ctx)
	rsp := &cap.CreateTableTemplateRsp{}
	if isTempTable(ctx) {
		return rspErr(ctx, ErrTempTableNotSupportTemplateOp)
	}
	tmd, err := registry.GlobalTableRegistry().TableMetaReg.Find(req.Template.TableId)
	if err != nil {
		return rspErr(ctx, err)
	}
	err = tmd.ValidateTpl(req.Template, userLanguage)
	if err != nil {
		return rspErr(ctx, err)
	}
	ss, err := t.DBWrite().NewSessionWithCtx(ctx)
	if err != nil {
		return rspErr(ctx, err)
	}
	defer func() {
		ss.Close(err)
	}()
	user, err := t.userInfoProvider.GetCurrentUser(ctx)
	if err != nil {
		return rspErr(ctx, err)
	}

	id, err := template.GlobalManager().CreateTemplate(ctx, ss, req.Template.TableId,
		req.Template.Name, req.Template.Body, req.Template.FileInfo.Access,
		req.GetTemplate().FileInfo.ShareList, user.Id)
	if err != nil {
		err = errors.New("Duplicate plan name creation")
		if langCode.String() == "zh-CN" {
			err = errors.New("方案名称创建重复")
		}
		return rspErr(ctx, err)
	}

	created, err := template.GlobalManager().FindTemplate(ctx, ss, id)
	if err == nil {
		rsp.Template = created
	}

	return rspOK(rsp), nil
}

// DeleteTableTemplate 删除模板
func (t *TableWService) DeleteTableTemplate(ctx context.Context, req *cap.DeleteTableTemplateReq) (*cap.CommonRsp, error) {
	rsp := &cap.DeleteTableTemplateRsp{}
	ss, err := t.DBWrite().NewSessionWithCtx(ctx)
	if err != nil {
		return rspErr(ctx, err)
	}
	defer func() {
		ss.Close(err)
	}()
	if isTempTable(ctx) {
		return rspErr(ctx, ErrTempTableNotSupportTemplateOp)
	}
	user, err := t.userInfoProvider.GetCurrentUser(ctx)
	if err != nil {
		return rspErr(ctx, err)
	}
	err = template.GlobalManager().DeleteTemplate(ctx, ss, req.TemplateId, user.Id)
	if err != nil {
		return rspErr(ctx, err)
	}
	return rspOK(rsp), nil
}

// UpdateTableTemplate 更新模板
func (t *TableWService) UpdateTableTemplate(ctx context.Context, req *cap.CreateTableTemplateReq) (*cap.CommonRsp, error) {
	userLanguage := i18n.GetUserLanguageByMeta(ctx)
	rsp := &cap.CreateTableTemplateRsp{}
	if isTempTable(ctx) {
		return rspErr(ctx, ErrTempTableNotSupportTemplateOp)
	}
	tmd, err := registry.GlobalTableRegistry().TableMetaReg.Find(req.Template.TableId)
	if err != nil {
		return rspErr(ctx, err)
	}
	err = tmd.ValidateTpl(req.Template, userLanguage)
	if err != nil {
		return rspErr(ctx, err)
	}
	ss, err := t.DBWrite().NewSessionWithCtx(ctx)
	if err != nil {
		return rspErr(ctx, err)
	}
	defer func() {
		ss.Close(err)
	}()
	user, err := t.userInfoProvider.GetCurrentUser(ctx)
	if err != nil {
		return rspErr(ctx, err)
	}
	rsp.Template, err = template.GlobalManager().UpdateTemplate(ctx, ss, req.Template, user.Id)
	if err != nil {
		return rspErr(ctx, err)
	}
	return rspOK(rsp), nil
}

/*************************************** 数据 *******************************************/

// GetTableColumns 获取表列
func (t *TableWService) GetTableColumns(ctx context.Context, req *cap.GetTableColumnsReq) (*cap.CommonRsp, error) {
	rsp := &cap.GetTableColumnsRsp{}
	userLanguage := i18n.GetUserLanguageByMeta(ctx)
	tmd, err := registry.GlobalTableRegistry().TableMetaReg.Find(req.TableId)
	if err != nil {
		return rspErr(ctx, err)
	}
	cols := tmd.ColumnsWithoutInternal()
	rsp.Columns = make([]*cap.TableColumn, len(cols.List()))
	for i, c := range cols.List() {
		rsp.Columns[i] = c.BuildTableColumn(userLanguage, req.TableId)
	}
	return rspOK(rsp), nil
}

// GetTableRows 获取表行
func (t *TableWService) GetTableRows(ctx context.Context, req *cap.GetTableRowsReq) (*cap.CommonRsp, error) {
	rsp := &cap.GetTableRowsRsp{}
	ss, err := t.DBRead().NewSessionWithCtx(ctx)
	if err != nil {
		return rspErr(ctx, err)
	}
	defer func() {
		ss.Close(err)
	}()
	tpl, err := data.ParseTpl(ctx, ss, req.TableId, req.Tpl)
	if err != nil {
		return rspErr(ctx, err)
	}
	rsp, err = data.GlobalManager().FindRows(ctx, ss, tpl, req.Page, req.Order)
	if err != nil {
		return rspErr(ctx, err)
	}
	return rspOK(rsp), nil
}

// GetTableRowsLite 简单版获取行列表接口
func (t *TableWService) GetTableRowsLite(ctx context.Context, req *cap.GetTableRowsLiteReq) (*cap.CommonRsp, error) {
	rsp := &cap.GetTableRowsLiteRsp{}
	ss, err := t.DBRead().NewSessionWithCtx(ctx)
	if err != nil {
		return rspErr(ctx, err)
	}
	defer func() {
		ss.Close(err)
	}()
	rsp, err = data.GlobalManager().FindRowsLite(ctx, ss, req.TableId, req.Query, req.Page, req.PageSize)
	if err != nil {
		return rspErr(ctx, err)
	}
	return rspOK(rsp), nil
}

// GetTableRowByID 精确获取一行
func (t *TableWService) GetTableRowByID(ctx context.Context, req *cap.GetTableRowByIDReq) (*cap.CommonRsp, error) {
	rsp := &cap.GetTableRowByIDRsp{}
	ss, err := t.DBRead().NewSessionWithCtx(ctx)
	if err != nil {
		return rspErr(ctx, err)
	}
	defer func() {
		ss.Close(err)
	}()
	tpl, err := data.ParseTpl(ctx, ss, req.TableId, req.Tpl)
	if err != nil {
		return rspErr(ctx, err)
	}
	rsp, err = data.GlobalManager().FindRow(ctx, ss, tpl, req.RowId)
	if err != nil {
		return rspErr(ctx, err)
	}
	return rspOK(rsp), nil
}

// DoExportTable 导出表
func (t *TableWService) DoExportTable(ctx context.Context, req *cap.GetTableRowsReq) (*cap.CommonRsp, error) {
	rsp := &cap.ExportTableRsp{}
	userLanguage := i18n.GetUserLanguageByMeta(ctx)
	tmd, err := registry.GlobalTableRegistry().TableMetaReg.Find(req.TableId)
	if err != nil {
		return rspErr(ctx, err)
	}
	ss, err := t.DBRead().NewSessionWithCtx(ctx)
	if err != nil {
		return rspErr(ctx, err)
	}
	defer func() {
		ss.Close(err)
	}()
	// 创建文件
	fileName := i18n.GetTranslation(userLanguage, registry.TrKeyTableName(req.GetTableId())) + "_" + time.Now().Format("20060102150405")
	rsp.FileName = fileName
	f := excelize.NewFile()
	f.NewSheet(fileName)
	f.DeleteSheet("Sheet1")
	sw, err := f.NewStreamWriter(fileName)
	if err != nil {
		return rspErr(ctx, err)
	}
	// 写表头
	// 解析模板
	tpl, err := data.ParseTpl(ctx, ss, req.TableId, req.Tpl)
	if err != nil {
		return rspErr(ctx, err)
	}
	err = tmd.ValidateTpl(tpl, userLanguage)
	if err != nil {
		return rspErr(ctx, err)
	}
	headers := make([]interface{}, len(tpl.Body.Output.VisibleColumns))
	for i, col := range tpl.Body.Output.VisibleColumns {
		_, err := tmd.Columns().Find(col.ColumnId)
		if err != nil {
			return rspErr(ctx, err)
		}
		headers[i] = i18n.GetTranslation(userLanguage, registry.TrKeyTableField(tmd.ID(), col.ColumnId), tmd.Name())
	}

	axis, _ := excelize.CoordinatesToCellName(1, 1)
	sw.SetRow(axis, headers)

	rowsRsp, err := data.GlobalManager().FindRows(ctx, ss, tpl, req.Page, req.Order)
	if err != nil {
		return rspErr(ctx, err)
	}
	for i, rowRsp := range rowsRsp.Rows {
		row := make([]interface{}, len(rowRsp.Cells))
		for j, cell := range rowRsp.Cells {
			switch cell.Value.V.(type) {
			case *cap.Value_VString:
				row[j] = cell.Value.GetVString()
			case *cap.Value_VInt:
				row[j] = cell.Value.GetVInt()
			case *cap.Value_VDouble:
				row[j] = cell.Value.GetVDouble()
			case *cap.Value_VDate:
				row[j] = cell.Value.GetVDate()
			case *cap.Value_VTime:
				t, _ := time.Parse(time.RFC3339, cell.Value.GetVTime())
				if !t.IsZero() {
					row[j] = t.Local().Format("2006-01-02 15:04:05")
				} else {
					row[j] = "N/A"
				}
			case *cap.Value_VBool:
				row[j] = cell.Value.GetVBool()
			case *cap.Value_VOption:
				row[j] = cell.Value.GetVOption().Name
			default:
				row[j] = "N/A"
			}
		}
		axis, _ := excelize.CoordinatesToCellName(1, i+2)
		sw.SetRow(axis, row)
	}
	sw.Flush()
	buf, err := f.WriteToBuffer()
	if err != nil {
		return rspErr(ctx, err)
	}
	rsp.ExcelData = buf.Bytes()
	return rspOK(rsp), nil
}

// GetTableColumnOptions 获取列选项列表（仅ValueType = VT_OPTION时可获取）
func (t *TableWService) GetTableColumnOptions(ctx context.Context, req *cap.GetTableColumnOptionsReq) (*cap.CommonRsp, error) {
	rsp := &cap.GetTableColumnOptionsRsp{}
	tmd, err := registry.GlobalTableRegistry().TableMetaReg.Find(req.TableId)
	if err != nil {
		return rspErr(ctx, err)
	}
	desc, err := tmd.Columns().Find(req.ColumnId)
	if err != nil {
		return rspErr(ctx, err)
	}
	opts, err := registry.GlobalTableRegistry().OptionReg.GetOptionsWithI18nCtx(ctx, desc.DataType.String())
	if err != nil {
		return rspErr(ctx, err)
	}
	rsp.OptionTypeID = desc.DataType.String()
	rsp.Options = make([]*cap.OptionValue, len(opts))
	for i, opt := range opts {
		rsp.Options[i] = &cap.OptionValue{
			Id:   opt.Id,
			Name: opt.Name,
		}
	}
	return rspOK(rsp), nil
}

// GetOptions 根据Option ID获取
func (t *TableWService) GetOptions(ctx context.Context, req *cap.GetOptionsReq) (*cap.CommonRsp, error) {
	//langCode := i18n.GetLanguageTypeByMeta(ctx)
	userLanguage := i18n.GetUserLanguageByMeta(ctx)
	rsp := &cap.GetTableColumnOptionsRsp{}
	opts, err := registry.GlobalTableRegistry().OptionReg.GetOptionsWithI18nCtx(ctx, req.OptionTypeID)
	if err != nil {
		return rspErr(ctx, err)
	}
	rsp.OptionTypeID = req.OptionTypeID
	rsp.Options = make([]*cap.OptionValue, len(opts))
	for i, opt := range opts {
		key := strings.ReplaceAll(fmt.Sprintf("table_opt_%s_%d", req.OptionTypeID, int(opt.Id)), ".", "_")
		name := i18n.GetTranslation(userLanguage, key)
		rsp.Options[i] = &cap.OptionValue{
			Id:   opt.Id,
			Name: name,
		}
	}
	return rspOK(rsp), nil
}

/*************************************** 操作 *******************************************/

// DoRowFormAction ...
func (t *TableWService) DoRowFormAction(ctx context.Context, req *cap.DoRowFormActionReq) (*cap.CommonRsp, error) {
	rsp := &cap.DoRowFormActionRsp{}
	ss, err := t.DBWrite().NewSessionWithCtx(ctx)
	if err != nil {
		return rspErr(ctx, err)
	}
	defer func() {
		ss.Close(err)
	}()
	var tmd registry.TableMetaData
	tmd, err = registry.GlobalTableRegistry().TableMetaReg.Find(req.TableId)
	if err != nil {
		return rspErr(ctx, err)
	}
	var rowAction action.RowAction
	rowAction, err = tmd.GetRowActions(ctx, nil).Find(req.ActionId)
	if err != nil {
		return rspErr(ctx, err)
	}
	if rowAction.Type() != cap.RowActionType_RAT_JSON_FORM {
		return rspErr(ctx, ErrNotFormAction)
	}
	err = rowAction.(*action.FormRowAction).Execute(ctx, ss, req.FormJson)
	if err != nil {
		errors.Wrap(err).PrintStackTrace()
		return rspErr(ctx, err)
	}
	return rspOK(rsp), nil
}

// NewTableWService creates table service
func NewTableWService(dbWrite *mysql.DB, dbRead *mysql.DB, userInfoProvider UserInfoProvider) *TableWService {
	template.AIP = func(accountID string) (userName string, displayName string, err error) {
		info, err := userInfoProvider.GetUserInfoByID(accountID)
		if err != nil {
			return "", "", err
		}
		return info.UserName, info.DisplayName, nil
	}
	return &TableWService{dbWrite: dbWrite, dbRead: dbRead, userInfoProvider: userInfoProvider}
}
