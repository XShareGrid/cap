package mysql

import (
	"context"
	"fmt"
	"strings"

	db "github.com/XShareGrid/cap/database/mysql"
	"github.com/XShareGrid/cap/logger"
	"github.com/XShareGrid/cap/msg/errors"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// TableTemplateMapper table template mapper
type TableTemplateMapper struct {
	*db.BaseMapper
}

// NewTableTemplateMapperFromDBContext creates table template mapper
func NewTableTemplateMapperFromDBContext(ctx context.Context) (*TableTemplateMapper, error) {
	ss, err := db.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	return NewTableTemplateMapper(ss), err
}

// NewTableTemplateMapper creates table template mapper from session
func NewTableTemplateMapper(ss *db.Session) *TableTemplateMapper {
	return &TableTemplateMapper{BaseMapper: db.NewBaseMapper(ss, "table_template")}
}

// InitMapper ...
func (m *TableTemplateMapper) InitMapper(ss *db.Session) error {
	/*
		CREATE TABLE cap.table_template (
		  id varchar(64) NOT NULL DEFAULT '',
		  name varchar(128) NOT NULL DEFAULT '' COMMENT '模板名，table_id+模板名需唯一',
		  table_id varchar(64) NOT NULL DEFAULT '',
		  f_access tinyint NOT NULL,
		  f_create_user int NOT NULL,
		  f_create_time timestamp NULL DEFAULT NULL,
		  f_mod_time timestamp NULL DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,
		  body blob DEFAULT NULL,
		  PRIMARY KEY (id)
		)
		ENGINE = INNODB,
		CHARACTER SET utf8mb4,
		COLLATE utf8mb4_0900_ai_ci;

		ALTER TABLE cap.table_template
		ADD UNIQUE INDEX name (name);

		ALTER TABLE cap.table_template
		ADD UNIQUE INDEX UK_table_template (name, table_id);



		CREATE TABLE cap.table_template_share (
		  template_id varchar(64) NOT NULL DEFAULT '' COMMENT '模板id',
		  user_id int NOT NULL COMMENT '用户id',
		  PRIMARY KEY (template_id, user_id)
		)
		ENGINE = INNODB,
		CHARACTER SET utf8mb4,
		COLLATE utf8mb4_0900_ai_ci,
		COMMENT = '报表模板共享信息';

		ALTER TABLE cap.table_template_share
		ADD CONSTRAINT FK_table_template_share_template_id FOREIGN KEY (template_id)
		REFERENCES cap.table_template (id) ON DELETE CASCADE ON UPDATE CASCADE;
	*/

	return nil
}

// FilterTableIDEquals 按表ID过滤
func FilterTableIDEquals(tableID string) *db.RowsFilter {
	return &db.RowsFilter{Key: "table_id", Operator: db.RFEqual, Value: tableID}
}

// FilterTableAccessEquals 按共享方式过滤
func FilterTableAccessEquals(access int) *db.RowsFilter {
	return &db.RowsFilter{Key: "f_access", Operator: db.RFEqual, Value: access}
}

// FilterTemplateIDEquals 按模板ID过滤
func FilterTemplateIDEquals(tplID string) *db.RowsFilter {
	return &db.RowsFilter{Key: "id", Operator: db.RFEqual, Value: tplID}
}

// FilterCreateUserEquals 按创建用户过滤
func FilterCreateUserEquals(userID int64) *db.RowsFilter {
	return &db.RowsFilter{Key: "f_create_user", Operator: db.RFEqual, Value: userID}
}

// FilterUserIDIn 按ID列表过滤
func FilterUserIDIn(idList []string) *db.RowsFilter {
	return &db.RowsFilter{Key: "id", Operator: db.RFIn, Value: idList}
}

// FindTemplates 支持过滤器 FilterTableIDEquals， FilterCreateUserEquals
func (m *TableTemplateMapper) FindTemplates(filters ...*db.RowsFilter) ([]*TableTpl, error) {
	sqlSelect := `SELECT id, name, table_id, f_access, f_create_user, f_create_time,
	 f_mod_time, body FROM table_template `
	if len(filters) != 0 {
		sqlSelect += "WHERE "
	}
	var values []interface{}
	hasInOperator := false
	for i, f := range filters {
		sqlSelect += fmt.Sprintf("%s %s (?) ", f.Key, f.Operator)
		if i != len(filters)-1 {
			sqlSelect += "AND "
		}
		if f.Operator == db.RFIn {
			hasInOperator = true
		}
		values = append(values, f.Value)
	}
	if hasInOperator {
		var err error
		sqlSelect, values, err = sqlx.In(sqlSelect, values...)
		if err != nil {
			return nil, errors.Wrap(err).Log()
		}
	}
	var tplList []TableTemplate
	err := m.GetTx().Select(&tplList, sqlSelect, values...)
	if err != nil {
		return nil, errors.Wrap(err).Log()
	}
	var tableTplList = make([]*TableTpl, len(tplList))
	for i, tpl := range tplList {
		tableTplList[i] = &TableTpl{TableTemplate: tpl}
		// FindShareList
		tableTplList[i].ShareList, err = m.findShareList(tpl.Id)
		if err != nil {
			return nil, errors.Wrap(err).Log()
		}
	}
	return tableTplList, nil
}

// FindTemplatesByShareUserAndTableID ...
func (m *TableTemplateMapper) FindTemplatesByShareUserAndTableID(userID int64, tableID string) ([]*TableTpl, error) {
	templateIDList := []string{}
	sqlSelect := `SELECT template_id FROM table_template_share WHERE user_id = ?`
	err := m.GetTx().Select(&templateIDList, sqlSelect, userID)
	if err != nil {
		return nil, errors.Wrap(err).Log()
	}
	var tplList []string
	err = m.GetTx().Select(&tplList, sqlSelect, userID)
	if err != nil {
		return nil, errors.Wrap(err).Log()
	}
	if len(tplList) == 0 {
		return nil, nil
	}
	return m.FindTemplates(FilterTableIDEquals(tableID), FilterUserIDIn(tplList))
}

// FindTemplate find template for update
func (m *TableTemplateMapper) FindTemplate(id string, forUpdate ...bool) (*TableTpl, error) {
	tpl := &TableTpl{}
	sqlSelect := `SELECT id, name, table_id, f_access, f_create_user, f_create_time,
	f_mod_time, body FROM table_template WHERE id = ?`
	if len(forUpdate) > 0 && forUpdate[0] {
		sqlSelect += " FOR UPDATE"
	}
	err := m.GetTx().Get(&tpl.TableTemplate, sqlSelect, id)
	if err != nil {
		return nil, errors.Wrap(err).Log()
	}

	tpl.ShareList, err = m.findShareList(id)
	if err != nil {
		return nil, errors.Wrap(err).Log()
	}
	return tpl, nil
}

func (m *TableTemplateMapper) findShareList(tplID string) ([]TableTemplateShare, error) {
	var shareList []TableTemplateShare
	err := m.GetTx().Select(&shareList,
		"SELECT template_id, user_id FROM table_template_share WHERE template_id = ?", tplID)
	return shareList, err
}

// CreateTemplate ...
func (m *TableTemplateMapper) CreateTemplate(tt *TableTpl) error {
	// template
	sqlInsert :=
		`INSERT INTO table_template (
		id, name, table_id, f_access, f_create_user, f_create_time, f_mod_time, body)  
		VALUES (?,?,?,?,?,?,?,?);`
	ret, err := m.GetTx().Exec(sqlInsert, tt.Id, tt.Name, tt.TableId,
		tt.FAccess, tt.FCreateUser, tt.FCreateTime, tt.FModTime, tt.Body)
	if err != nil {
		if me, ok := err.(*mysql.MySQLError); ok {
			if me.Number == 1062 {
				return errors.Wrap(err).Triggers(ErrDuplicateEntry).Log()
			} else if me.Number == 1406 {
				return errors.Wrap(err).Triggers(ErrDataTooLong).FillDebugArgs(tt.Name)
			}
		}
		return err
	}

	affected, err := ret.RowsAffected()
	if err != nil {
		return errors.Wrap(err).Log()
	}

	if affected == 0 {
		logger.CAP.Error("add template error, rows affected = 0")
		return ErrRowsAffectedZero
	}
	if len(tt.ShareList) > 0 {
		err = m.insertShareList(tt.Id, tt.ShareList)
		if err != nil {
			return errors.Wrap(err).Log()
		}
	}
	return err
}

func (m *TableTemplateMapper) insertShareList(templateID string, shareList []TableTemplateShare) error {
	var shareListSQL []string
	for _, s := range shareList {
		s.TemplateId = templateID
		shareListSQL = append(shareListSQL, fmt.Sprintf("('%s', %d)", s.TemplateId, s.UserId))
	}
	sqlInsertShare := fmt.Sprintf(
		`INSERT INTO table_template_share (template_id, user_id) VALUES %s;`, strings.Join(shareListSQL, ","))
	ret, err := m.GetTx().Exec(sqlInsertShare)
	if err != nil {
		return err
	}
	affected, _ := ret.RowsAffected()
	if affected == 0 {
		return errors.Wrap(ErrRowsAffectedZero).Log()
	}
	return nil
}

// UpdateTableTemplate ...
func (m *TableTemplateMapper) UpdateTableTemplate(id, name string, access int,
	body []byte, shareList []TableTemplateShare) error {
	sqlUpdate := `UPDATE table_template SET name = ?, f_access = ?, body = ? WHERE id = ?;`
	_, err := m.GetTx().Exec(sqlUpdate, name, access, body, id)
	if err != nil {
		return errors.Wrap(err).Log()
	}
	sqlDelete := `DELETE FROM table_template_share WHERE template_id = ?;`
	_, err = m.GetTx().Exec(sqlDelete, id)
	if err != nil {
		return errors.Wrap(err).Log()
	}
	if len(shareList) > 0 {
		err = m.insertShareList(id, shareList)
		if err != nil {
			errors.Wrap(err).Log()
		}
	}
	return nil
}

// DeleteTemplates ...
func (m *TableTemplateMapper) DeleteTemplates(filters ...*db.RowsFilter) (affected int64, err error) {
	return m.Delete("table_template", filters...)
}
