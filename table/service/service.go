package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"golang.org/x/text/language"
	"google.golang.org/grpc/codes"

	excelize "github.com/xuri/excelize/v2"

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
)

// UserInfoProvider 账户信息提供者
type UserInfoProvider interface {
	GetCurrentUser(ctx context.Context) (*cap.UserInfo, error)
	GetUserInfoByID(id int32) (*cap.UserInfo, error)
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

var langCode = language.Make("zh-CN")

/*************************************** 表 *******************************************/

// GetTableInfo 获取表信息
func (t *TableWService) GetTableInfo(ctx context.Context, req *cap.GetTableInfoReq) (*cap.GetTableInfoRsp, error) {
	langCode = i18n.GetLanguageTypeByMeta(ctx)
	rsp := &cap.GetTableInfoRsp{}
	userLanguage := i18n.GetUserLanguageByMeta(ctx)
	tmd, err := registry.GlobalTableRegistry().TableMetaReg.Find(req.TableId)
	if err != nil {
		return rsp, handle.Handle(ctx, err).Log().GRPCErr(codes.Unknown, langCode)
	}

	rsp.Id = tmd.ID()
	rsp.Name = i18n.GetTranslation(userLanguage, registry.TrKeyTableName(tmd.ID()))
	rsp.Desc = i18n.GetTranslation(userLanguage, registry.TrKeyTableDesc(tmd.ID()))
	rsp.ExportFilePrefix = rsp.Name + "_" + time.Now().Format("20060102")
	return rsp, nil
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
func (t *TableWService) GetTableTemplates(ctx context.Context, req *cap.GetTableTemplatesReq) (*cap.GetTableTemplatesRsp, error) {
	langCode = i18n.GetLanguageTypeByMeta(ctx)
	rsp := &cap.GetTableTemplatesRsp{}
	ss, err := t.DBRead().NewSessionWithCtx(ctx)
	if err != nil {
		return rsp, handle.Handle(ctx, err).Log().GRPCErr(codes.Unknown, langCode)
	}
	defer func() {
		ss.Close(err)
	}()
	tmd, err := registry.GlobalTableRegistry().TableMetaReg.Find(req.TableId)
	if err != nil {
		return rsp, handle.Handle(ctx, err).Log().GRPCErr(codes.Unknown, langCode)
	}
	tplList := []*cap.Template{tmd.DefaultTpl(ctx)}
	// TODO. 20250108
	if isTempTable(ctx) {
		rsp.Templates = tplList
		return rsp, nil
	}
	user, err := t.userInfoProvider.GetCurrentUser(ctx)
	if err != nil {
		return rsp, handle.Handle(ctx, err).Log().GRPCErr(codes.Unknown, langCode)
	}
	tplListOwn, err := template.GlobalManager().FindTemplatesByTableAndCreateUser(ctx, ss, req.TableId, int(user.Id))
	if err != nil {
		return rsp, handle.Handle(ctx, err).Log().GRPCErr(codes.Unknown, langCode)
	}
	tplListShared, err := template.GlobalManager().FindTemplatesByTableAndShareUser(ctx, ss, req.TableId, int(user.Id))
	if err != nil {
		return rsp, handle.Handle(ctx, err).Log().GRPCErr(codes.Unknown, langCode)
	}
	tplListPublic, err := template.GlobalManager().FindPublicTemplatesByTable(ctx, ss, req.TableId, int(user.Id))
	if err != nil {
		return rsp, handle.Handle(ctx, err).Log().GRPCErr(codes.Unknown, langCode)
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
	return rsp, nil
}

// CreateTableTemplate 创建模板
func (t *TableWService) CreateTableTemplate(ctx context.Context, req *cap.CreateTableTemplateReq) (*cap.CreateTableTemplateRsp, error) {
	langCode = i18n.GetLanguageTypeByMeta(ctx)
	userLanguage := i18n.GetUserLanguageByMeta(ctx)
	rsp := &cap.CreateTableTemplateRsp{}
	if isTempTable(ctx) {
		return rsp, handle.Handle(ctx, ErrTempTableNotSupportTemplateOp).GRPCErr(codes.Unknown, langCode)
	}
	tmd, err := registry.GlobalTableRegistry().TableMetaReg.Find(req.Template.TableId)
	if err != nil {
		return rsp, handle.Handle(ctx, err).Log().GRPCErr(codes.Unknown, langCode)
	}
	err = tmd.ValidateTpl(req.Template, userLanguage)
	if err != nil {
		return rsp, handle.Handle(ctx, err).Log().GRPCErr(codes.Unknown, langCode)
	}
	ss, err := t.DBWrite().NewSessionWithCtx(ctx)
	if err != nil {
		return rsp, handle.Handle(ctx, err).Log().GRPCErr(codes.Unknown, langCode)
	}
	defer func() {
		ss.Close(err)
	}()
	user, err := t.userInfoProvider.GetCurrentUser(ctx)
	if err != nil {
		return rsp, handle.Handle(ctx, err).Log().GRPCErr(codes.Unknown, langCode)
	}

	id, err := template.GlobalManager().CreateTemplate(ctx, ss, req.Template.TableId,
		req.Template.Name, req.Template.Body, req.Template.FileInfo.Access,
		req.GetTemplate().FileInfo.ShareList, int(user.Id))
	if err != nil {
		err = errors.New("Duplicate plan name creation")
		if langCode.String() == "zh-CN" {
			err = errors.New("方案名称创建重复")
		}
		return rsp, err
		//return rsp, handle.Handle(ctx, err).Log().GRPCErr(codes.Unknown, langCode)
	}

	created, err := template.GlobalManager().FindTemplate(ctx, ss, id)
	if err == nil {
		rsp.Template = created
	}

	return rsp, nil
}

// DeleteTableTemplate 删除模板
func (t *TableWService) DeleteTableTemplate(ctx context.Context, req *cap.DeleteTableTemplateReq) (*cap.DeleteTableTemplateRsp, error) {
	langCode = i18n.GetLanguageTypeByMeta(ctx)
	rsp := &cap.DeleteTableTemplateRsp{}
	ss, err := t.DBWrite().NewSessionWithCtx(ctx)
	if err != nil {
		return rsp, handle.Handle(ctx, err).Log().GRPCErr(codes.Unknown, langCode)
	}
	defer func() {
		ss.Close(err)
	}()
	if isTempTable(ctx) {
		return rsp, handle.Handle(ctx, ErrTempTableNotSupportTemplateOp).GRPCErr(codes.Unknown, langCode)
	}
	user, err := t.userInfoProvider.GetCurrentUser(ctx)
	if err != nil {
		return rsp, handle.Handle(ctx, err).Log().GRPCErr(codes.Unknown, langCode)
	}
	err = template.GlobalManager().DeleteTemplate(ctx, ss, req.TemplateId, int(user.Id))
	if err != nil {
		return rsp, handle.Handle(ctx, err).Log().GRPCErr(codes.Unknown, langCode)
	}
	return rsp, nil
}

// UpdateTableTemplate 更新模板
func (t *TableWService) UpdateTableTemplate(ctx context.Context, req *cap.CreateTableTemplateReq) (*cap.CreateTableTemplateRsp, error) {
	langCode = i18n.GetLanguageTypeByMeta(ctx)
	userLanguage := i18n.GetUserLanguageByMeta(ctx)
	rsp := &cap.CreateTableTemplateRsp{}
	if isTempTable(ctx) {
		return rsp, handle.Handle(ctx, ErrTempTableNotSupportTemplateOp).GRPCErr(codes.Unknown, langCode)
	}
	tmd, err := registry.GlobalTableRegistry().TableMetaReg.Find(req.Template.TableId)
	if err != nil {
		return rsp, handle.Handle(ctx, err).Log().GRPCErr(codes.Unknown, langCode)
	}
	err = tmd.ValidateTpl(req.Template, userLanguage)
	if err != nil {
		return rsp, handle.Handle(ctx, err).Log().GRPCErr(codes.Unknown, langCode)
	}
	ss, err := t.DBWrite().NewSessionWithCtx(ctx)
	if err != nil {
		return rsp, handle.Handle(ctx, err).Log().GRPCErr(codes.Unknown, langCode)
	}
	defer func() {
		ss.Close(err)
	}()
	user, err := t.userInfoProvider.GetCurrentUser(ctx)
	if err != nil {
		return rsp, handle.Handle(ctx, err).Log().GRPCErr(codes.Unknown, langCode)
	}
	rsp.Template, err = template.GlobalManager().UpdateTemplate(ctx, ss, req.Template, int(user.Id))
	if err != nil {
		return rsp, handle.Handle(ctx, err).Log().GRPCErr(codes.Unknown, langCode)
	}
	return rsp, nil
}

/*************************************** 数据 *******************************************/

// GetTableColumns 获取表列
func (t *TableWService) GetTableColumns(ctx context.Context, req *cap.GetTableColumnsReq) (rsp *cap.GetTableColumnsRsp, err error) {
	langCode = i18n.GetLanguageTypeByMeta(ctx)
	rsp = &cap.GetTableColumnsRsp{}
	userLanguage := i18n.GetUserLanguageByMeta(ctx)
	tmd, err := registry.GlobalTableRegistry().TableMetaReg.Find(req.TableId)
	if err != nil {
		return rsp, handle.Handle(ctx, err).Log().GRPCErr(codes.Unknown, langCode)
	}
	cols := tmd.ColumnsWithoutInternal()
	rsp.Columns = make([]*cap.TableColumn, len(cols.List()))
	for i, c := range cols.List() {
		rsp.Columns[i] = c.BuildTableColumn(userLanguage, req.TableId)
	}
	return rsp, nil
}

// GetTableRows 获取表行
func (t *TableWService) GetTableRows(ctx context.Context, req *cap.GetTableRowsReq) (*cap.GetTableRowsRsp, error) {
	langCode = i18n.GetLanguageTypeByMeta(ctx)
	rsp := &cap.GetTableRowsRsp{}
	ss, err := t.DBRead().NewSessionWithCtx(ctx)
	if err != nil {
		return rsp, handle.Handle(ctx, err).Log().GRPCErr(codes.Unknown, langCode)
	}
	defer func() {
		ss.Close(err)
	}()
	tpl, err := data.ParseTpl(ctx, ss, req.TableId, req.Tpl)
	if err != nil {
		return rsp, handle.Handle(ctx, err).Log().GRPCErr(codes.Unknown, langCode)
	}
	rsp, err = data.GlobalManager().FindRows(ctx, ss, tpl, req.Page, req.Order)
	if err != nil {
		return rsp, handle.Handle(ctx, err).Log().GRPCErr(codes.Unknown, langCode)
	}
	return rsp, nil
}

// GetTableRowsLite 简单版获取行列表接口
func (t *TableWService) GetTableRowsLite(ctx context.Context, req *cap.GetTableRowsLiteReq) (*cap.GetTableRowsLiteRsp, error) {
	langCode = i18n.GetLanguageTypeByMeta(ctx)
	rsp := &cap.GetTableRowsLiteRsp{}
	ss, err := t.DBRead().NewSessionWithCtx(ctx)
	if err != nil {
		return rsp, handle.Handle(ctx, err).Log().GRPCErr(codes.Unknown, langCode)
	}
	defer func() {
		ss.Close(err)
	}()
	rsp, err = data.GlobalManager().FindRowsLite(ctx, ss, req.TableId, req.Query)
	if err != nil {
		return rsp, handle.Handle(ctx, err).Log().GRPCErr(codes.Unknown, langCode)
	}
	return rsp, nil
}

// GetTableRowByID 精确获取一行
func (t *TableWService) GetTableRowByID(ctx context.Context, req *cap.GetTableRowByIDReq) (*cap.GetTableRowByIDRsp, error) {
	langCode = i18n.GetLanguageTypeByMeta(ctx)
	rsp := &cap.GetTableRowByIDRsp{}
	ss, err := t.DBRead().NewSessionWithCtx(ctx)
	if err != nil {
		return rsp, handle.Handle(ctx, err).Log().GRPCErr(codes.Unknown, langCode)
	}
	defer func() {
		ss.Close(err)
	}()
	tpl, err := data.ParseTpl(ctx, ss, req.TableId, req.Tpl)
	if err != nil {
		return rsp, handle.Handle(ctx, err).Log().GRPCErr(codes.Unknown, langCode)
	}
	rsp, err = data.GlobalManager().FindRow(ctx, ss, tpl, req.RowId)
	if err != nil {
		return rsp, handle.Handle(ctx, err).Log().GRPCErr(codes.Unknown, langCode)
	}
	return rsp, nil
}

// DoExportTable 导出表
func (t *TableWService) DoExportTable(ctx context.Context, req *cap.GetTableRowsReq) (*cap.ExportTableRsp, error) {
	langCode = i18n.GetLanguageTypeByMeta(ctx)
	rsp := &cap.ExportTableRsp{}
	userLanguage := i18n.GetUserLanguageByMeta(ctx)
	tmd, err := registry.GlobalTableRegistry().TableMetaReg.Find(req.TableId)
	if err != nil {
		return rsp, handle.Handle(ctx, err).Log().GRPCErr(codes.Unknown, langCode)
	}
	ss, err := t.DBRead().NewSessionWithCtx(ctx)
	if err != nil {
		return rsp, handle.Handle(ctx, err).Log().GRPCErr(codes.Unknown, langCode)
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
		return rsp, handle.Handle(ctx, err).Log().GRPCErr(codes.Unknown, langCode)
	}
	// 写表头
	// 解析模板
	tpl, err := data.ParseTpl(ctx, ss, req.TableId, req.Tpl)
	if err != nil {
		return rsp, handle.Handle(ctx, err).Log().GRPCErr(codes.Unknown, langCode)
	}
	err = tmd.ValidateTpl(tpl, userLanguage)
	if err != nil {
		return rsp, handle.Handle(ctx, err).Log().GRPCErr(codes.Unknown, langCode)
	}
	headers := make([]interface{}, len(tpl.Body.Output.VisibleColumns))
	for i, col := range tpl.Body.Output.VisibleColumns {
		_, err := tmd.Columns().Find(col.ColumnId)
		if err != nil {
			return rsp, handle.Handle(ctx, err).Log().GRPCErr(codes.Unknown, langCode)
		}
		headers[i] = i18n.GetTranslation(userLanguage, registry.TrKeyTableField(tmd.ID(), col.ColumnId), tmd.Name())
	}

	axis, _ := excelize.CoordinatesToCellName(1, 1)
	sw.SetRow(axis, headers)

	rowsRsp, err := data.GlobalManager().FindRows(ctx, ss, tpl, req.Page, req.Order)
	if err != nil {
		return rsp, handle.Handle(ctx, err).Log().GRPCErr(codes.Unknown, langCode)
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
		return rsp, handle.Handle(ctx, err).Log().GRPCErr(codes.Unknown, langCode)
	}
	rsp.ExcelData = buf.Bytes()
	return rsp, nil
}

// GetTableColumnOptions 获取列选项列表（仅ValueType = VT_OPTION时可获取）
func (t *TableWService) GetTableColumnOptions(ctx context.Context, req *cap.GetTableColumnOptionsReq) (*cap.GetTableColumnOptionsRsp, error) {
	langCode = i18n.GetLanguageTypeByMeta(ctx)
	rsp := &cap.GetTableColumnOptionsRsp{}
	tmd, err := registry.GlobalTableRegistry().TableMetaReg.Find(req.TableId)
	if err != nil {
		return rsp, handle.Handle(ctx, err).Log().GRPCErr(codes.Unknown, langCode)
	}
	desc, err := tmd.Columns().Find(req.ColumnId)
	if err != nil {
		return rsp, handle.Handle(ctx, err).Log().GRPCErr(codes.Unknown, langCode)
	}
	opts, err := registry.GlobalTableRegistry().OptionReg.GetOptionsWithI18nCtx(ctx, desc.DataType.String())
	if err != nil {
		return rsp, handle.Handle(ctx, err).Log().GRPCErr(codes.Unknown, langCode)
	}
	rsp.OptionTypeID = desc.DataType.String()
	rsp.Options = make([]*cap.OptionValue, len(opts))
	for i, opt := range opts {
		rsp.Options[i] = &cap.OptionValue{
			Id:   opt.Id,
			Name: opt.Name,
		}
	}
	return rsp, nil
}

// GetOptions 根据Option ID获取
func (t *TableWService) GetOptions(ctx context.Context, req *cap.GetOptionsReq) (*cap.GetTableColumnOptionsRsp, error) {
	//langCode = i18n.GetLanguageTypeByMeta(ctx)
	userLanguage := i18n.GetUserLanguageByMeta(ctx)
	rsp := &cap.GetTableColumnOptionsRsp{}
	opts, err := registry.GlobalTableRegistry().OptionReg.GetOptionsWithI18nCtx(ctx, req.OptionTypeID)
	if err != nil {
		return rsp, handle.Handle(ctx, err).Log().GRPCErr(codes.Unknown, langCode)
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
	return rsp, nil
}

/*************************************** 操作 *******************************************/

// DoRowFormAction ...
func (t *TableWService) DoRowFormAction(ctx context.Context, req *cap.DoRowFormActionReq) (rsp *cap.DoRowFormActionRsp, err error) {
	langCode = i18n.GetLanguageTypeByMeta(ctx)
	rsp = &cap.DoRowFormActionRsp{}
	ss, err := t.DBWrite().NewSessionWithCtx(ctx)
	if err != nil {
		return rsp, handle.Handle(ctx, err).Log().GRPCErr(codes.Unknown, langCode)
	}
	defer func() {
		ss.Close(err)
	}()
	var tmd registry.TableMetaData
	tmd, err = registry.GlobalTableRegistry().TableMetaReg.Find(req.TableId)
	if err != nil {
		return rsp, handle.Handle(ctx, err).Log().GRPCErr(codes.Unknown, langCode)
	}
	var rowAction action.RowAction
	rowAction, err = tmd.GetRowActions(ctx, nil).Find(req.ActionId)
	if err != nil {
		return rsp, handle.Handle(ctx, err).Log().GRPCErr(codes.Unknown, langCode)
	}
	if rowAction.Type() != cap.RowActionType_RAT_JSON_FORM {
		return rsp, handle.Handle(ctx, ErrNotFormAction).GRPCErr(codes.Unknown, langCode)
	}
	err = rowAction.(*action.FormRowAction).Execute(ctx, ss, req.FormJson)
	if err != nil {
		errors.Wrap(err).PrintStackTrace()
		return rsp, handle.Handle(ctx, err).Log().GRPCErr(codes.Unknown, langCode)
	}
	return rsp, nil
}

// NewTableWService creates table service
func NewTableWService(dbWrite *mysql.DB, dbRead *mysql.DB, userInfoProvider UserInfoProvider) *TableWService {
	template.AIP = func(accountID int32) (userName string, displayName string, err error) {
		info, err := userInfoProvider.GetUserInfoByID(accountID)
		if err != nil {
			return "", "", err
		}
		return info.UserName, info.DisplayName, nil
	}
	return &TableWService{dbWrite: dbWrite, dbRead: dbRead, userInfoProvider: userInfoProvider}
}
