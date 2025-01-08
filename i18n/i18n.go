package i18n

import (
	"context"
	"path/filepath"

	"github.com/XShareGrid/cap/logger"
	"github.com/XShareGrid/cap/ss/db"
	beegoCtx "github.com/astaxie/beego/context"
	"github.com/beego/i18n"
)

type UserLanguage string

const (
	DefaultLanguage              = "zh"
	defaultUserLanguageEnUS      = "en-US"
	defaultUserLanguageZhCN      = "zh-CN"
	ZH                           = "zh"
	EN                           = "en"
	GetSQLContextFailed          = "get_sql_context_failed"
	CreateMapperFailed           = "create_mapper_failed"
	GetExistingTranslationFailed = "get_existing_translation_failed"
	InsertTranslationFailed      = "insert_translation_failed"
	AccountGroupDisplayName      = "eap_account_group_display_name"
)

// TranslationItem 代表一个翻译项
type TranslationItem struct {
	BusinessID   string
	ID           string
	Translations map[string]string
}

type LanguageTranslationMap map[string]string

// 加载自定义语言包
func init() {
	// 获取 en-US 语言包文件的绝对路径
	absPath, err := filepath.Abs("./conf/en-US.ini")
	if err != nil {
		logger.CAP.Info("无法获取%v绝对路径：%s", defaultUserLanguageEnUS, err)
	}
	err = i18n.SetMessage(EN, absPath)
	if err != nil {
		logger.CAP.Info("无法加载%v语言包：%s", defaultUserLanguageEnUS, err)
	}

	// 获取 zh-CN 语言包文件的绝对路径
	absPath, err = filepath.Abs("./conf/zh-CN.ini")
	if err != nil {
		logger.CAP.Info("无法加载%v绝对路径：%s", defaultUserLanguageZhCN, err)
	}
	err = i18n.SetMessage(ZH, absPath)
	if err != nil {
		logger.CAP.Info("无法加载%v语言包：%s", defaultUserLanguageZhCN, err)
	}

	// 获取 hu-HU 语言包文件的绝对路径
	absPath, err = filepath.Abs("./conf/hu-HU.ini")
	if err != nil {
		logger.CAP.Info("无法加载hu绝对路径：%s", err)
	}
	err = i18n.SetMessage("hu", absPath)
	if err != nil {
		logger.CAP.Info("无法加载hu语言包：%s", err)
	}
}

// GetTranslation 静态字段获取翻译 如果翻译不存在则返回原始字符串
func GetTranslation(language, id string, ifNotFond ...string) string {
	translation := i18n.Tr(language, id)
	if translation == "" || translation == id {
		if len(ifNotFond) > 0 {
			return ifNotFond[0]
		}
		return id
	}
	return translation
}

type I18nString map[string]string

func NewI18nString(enUS, zhCN, huHU string) I18nString {
	return map[string]string{
		defaultUserLanguageEnUS: enUS,
		defaultUserLanguageZhCN: zhCN,
	}
}

func (s I18nString) EnUS() string {
	if t, ok := s[defaultUserLanguageEnUS]; ok {
		return t
	}
	return s[defaultUserLanguageZhCN]
}

func (s I18nString) ZhCN() string {
	if t, ok := s[defaultUserLanguageZhCN]; ok {
		return t
	}
	return s[defaultUserLanguageEnUS]
}

func (s I18nString) SetEnUS(l string) {
	s[defaultUserLanguageEnUS] = l
}

func (s I18nString) SetZhCN(l string) {
	s[defaultUserLanguageZhCN] = l
}

// GetBusinessTranslation 业务字段获取翻译
func GetBusinessTranslation(businessID, uniqueID string) I18nString {
	ctx, err := db.GetSQLContext()
	if err != nil {
		return map[string]string{}
	}
	defer db.CleanSQLContext(ctx, err)

	mapper, err := NewBusinessMapperFromCtx(ctx)
	if err != nil {
		return map[string]string{}
	}

	translations := make(map[string]string)
	row, err := mapper.GetTranslationByBusinessIDAndUniqueID(businessID, uniqueID)
	if err != nil {
		translations[defaultUserLanguageZhCN] = err.Error()
	}

	if row != nil {
		translations[defaultUserLanguageEnUS] = row.EnUS
		translations[defaultUserLanguageZhCN] = row.ZhCN
	} else {
		translations[defaultUserLanguageEnUS] = ""
		translations[defaultUserLanguageZhCN] = ""
	}

	return I18nString(translations)
}

// GetBusinessTranslations 获取多个业务字段的翻译
func GetBusinessTranslations(businessID string) map[string]I18nString {
	ctx, err := db.GetSQLContext()
	if err != nil {
		return map[string]I18nString{}
	}
	defer db.CleanSQLContext(ctx, err)

	mapper, err := NewBusinessMapperFromCtx(ctx)
	if err != nil {
		return map[string]I18nString{}
	}

	rows, err := mapper.GetTranslationsByBusinessID(businessID)
	if err != nil {
		return map[string]I18nString{}
	}

	allTranslations := make(map[string]I18nString)
	for _, row := range rows {
		if _, exists := allTranslations[row.UniqueID]; !exists {
			allTranslations[row.UniqueID] = make(map[string]string)
		}
		allTranslations[row.UniqueID][defaultUserLanguageEnUS] = row.EnUS
		allTranslations[row.UniqueID][defaultUserLanguageZhCN] = row.ZhCN
	}

	return allTranslations
}

// SaveBusinessTranslation 保存或更新业务字段的翻译
//
//	translations := map[string]string{
//		"zh-CN": "",
//		"en-US": "",
//	}
//
// businessID 业务ID
// uniqueID   业务字段的唯一标识
func SaveBusinessTranslation(businessID, uniqueID string, translations I18nString) error {
	ctx, err := db.GetSQLContext()
	if err != nil {
		return err
	}
	defer func() {
		db.CleanSQLContext(ctx, err)
	}()

	mapper, err := NewBusinessMapperFromCtx(ctx)
	if err != nil {
		return err
	}

	existingRow, err := mapper.GetTranslationByBusinessIDAndUniqueID(businessID, uniqueID)
	if err != nil && !isNoRowsError(err) {
		return err
	}

	if existingRow != nil {
		existingRow.EnUS = translations[defaultUserLanguageEnUS]
		existingRow.ZhCN = translations[defaultUserLanguageZhCN]
		err = mapper.UpdateTranslationRow(existingRow)
		if err != nil {
			return err
		}
	} else {
		row := &TranslationRow{
			BusinessID: businessID,
			UniqueID:   uniqueID,
			EnUS:       translations[defaultUserLanguageEnUS],
			ZhCN:       translations[defaultUserLanguageZhCN],
			IsDelete:   false,
		}
		_, err := mapper.InsertTranslationRow(row)
		if err != nil {
			return err
		}
	}

	return nil
}

// isNoRowsError checks if the error message indicates that there were no rows in the result set.
func isNoRowsError(err error) bool {
	return err.Error() == "sql: no rows in result set"
}

func GetTranslationWithFormat(language, id string, formatArgs ...interface{}) string {
	translation := i18n.Tr(language, id, formatArgs...)
	return translation
}

// GetUserLanguage 从请求的header头中获取用户语言设置，如果不存在，则默认返回 EN
func GetUserLanguage(ctx *beegoCtx.Context) string {
	if ctx == nil || ctx.Input == nil {
		return DefaultLanguage
	}

	language := ctx.Input.Header("Accept-Language")
	if language == "" {
		return DefaultLanguage
	}
	switch language {
	case defaultUserLanguageZhCN:
		language = ZH
	case defaultUserLanguageEnUS:
		language = EN
	}

	if len(language) > 2 {
		language = language[:2]
	}

	if language != ZH && language != EN {
		language = DefaultLanguage
	}
	return language
}

func GetUserLanguageFromCtx(ctx context.Context) string {
	language, ok := ctx.Value("Accept-Language").(string)
	if !ok || language == "" {
		return DefaultLanguage
	}

	switch language {
	case defaultUserLanguageZhCN:
		language = ZH
	case defaultUserLanguageEnUS:
		language = EN
	}

	if len(language) > 2 {
		language = language[:2]
	}

	if language != ZH && language != EN {
		language = DefaultLanguage
	}

	return language
}
