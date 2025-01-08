package i18n

import (
	"context"
	"fmt"

	"github.com/XShareGrid/cap/logger"
	"github.com/XShareGrid/cap/ss/db"
	"github.com/jmoiron/sqlx"
)

// TranslationRow 翻译信息行
type TranslationRow struct {
	ID         int    `db:"id"`
	BusinessID string `db:"business_id"`
	UniqueID   string `db:"unique_id"`
	EnUS       string `db:"en_us"`
	ZhCN       string `db:"zh_cn"`
	IsDelete   bool   `db:"is_delete"`
}

const (
	translationsTableName = "translations"
)

// TranslationMapper 翻译Mapper
type TranslationMapper struct {
	db.BaseSQLMapper
}

// NewTranslationMapper 创建翻译Mapper
func NewTranslationMapper(tx *sqlx.Tx) *TranslationMapper {
	var m TranslationMapper
	m.SetTableName(translationsTableName)
	m.SetTx(tx)
	return &m
}

// NewTranslationMapperFromCtx 从Context创建翻译Mapper
func NewBusinessMapperFromCtx(ctx context.Context) (*TranslationMapper, error) {
	tx, ok := db.GetSQLTxFromContext(ctx)
	if !ok {
		return nil, db.ErrNoTx
	}
	m := NewTranslationMapper(tx)
	return m, nil
}

// InsertTranslationRow 插入翻译数据
func (m *TranslationMapper) InsertTranslationRow(row *TranslationRow) (int64, error) {
	sql := `INSERT INTO translations (business_id, unique_id, en_us, zh_cn, is_delete) VALUES (?, ?, ?, ?, ?)`

	// 打印 SQL 语句及其参数
	query := fmt.Sprintf("Executing SQL: %s with params: %s, %s, %s, %s, %d",
		sql, row.BusinessID, row.UniqueID, row.EnUS, row.ZhCN, 0)
	logger.CAP.Info(query)

	// 执行 SQL 语句
	res, err := m.GetTx().Exec(sql, row.BusinessID, row.UniqueID, row.EnUS, row.ZhCN, 0)
	if err != nil {
		logger.CAP.Error("failed to execute SQL: %s", err.Error())
		return 0, err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	if affected == 0 {
		logger.CAP.Error("no rows affected")
		return 0, fmt.Errorf("db error")
	}

	m.GetTx().Commit()

	// 获取新记录的 ID
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

// GetTranslationByID 根据ID查询翻译数据
func (m *TranslationMapper) GetTranslationByID(id int) (*TranslationRow, error) {
	sql := `SELECT * FROM translations WHERE id = ? AND is_delete = 0`
	row := &TranslationRow{}
	err := m.GetTx().Get(row, sql, id)
	if err != nil {
		return nil, err
	}
	return row, nil
}

// GetTranslationByBusinessIDAndUniqueID 根据business_id和unique_id查询翻译数据
func (m *TranslationMapper) GetTranslationByBusinessIDAndUniqueID(businessID, uniqueID string) (*TranslationRow, error) {
	sql := `SELECT * FROM translations WHERE business_id = ? AND unique_id = ? AND is_delete = 0`
	row := &TranslationRow{}
	err := m.GetTx().Get(row, sql, businessID, uniqueID)
	if err != nil {
		return nil, err
	}
	return row, nil
}

// UpdateTranslationRow 更新翻译数据
func (m *TranslationMapper) UpdateTranslationRow(row *TranslationRow) error {
	sql := `UPDATE translations SET en_us=?, zh_cn=? WHERE id=?`
	res, err := m.GetTx().Exec(sql, row.EnUS, row.ZhCN, row.ID)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	//m.GetTx().Commit()

	if affected == 0 {
		return nil
		//return fmt.Errorf("no rows were updated")
	}

	return nil
}

// GetTranslationsByBusinessID 根据business_id查询所有未删除的翻译数据
func (m *TranslationMapper) GetTranslationsByBusinessID(businessID string) ([]*TranslationRow, error) {
	sql := `SELECT * FROM translations WHERE business_id = ? AND is_delete = 0`
	var rows []*TranslationRow
	err := m.GetTx().Select(&rows, sql, businessID)
	if err != nil {
		return nil, err
	}
	return rows, nil
}
