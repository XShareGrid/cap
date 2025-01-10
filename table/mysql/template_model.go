package mysql

import (
	"time"
)

// TableTemplate ...
type TableTemplate struct {
	Id          string    `db:"id"`
	Name        string    `db:"name"`
	TableId     string    `db:"table_id"`
	FAccess     int64     `db:"f_access"`
	FCreateUser string    `db:"f_create_user"`
	FCreateTime time.Time `db:"f_create_time"`
	FModTime    time.Time `db:"f_mod_time"`
	Body        []byte    `db:"body"`
}

func (TableTemplate) TableName() string {
	return "table_template"
}

type TableTemplateShare struct {
	TemplateId string `db:"template_id"`
	UserId     string `db:"user_id"`
}

func (TableTemplateShare) TableName() string {
	return "table_template_share"
}

// TableTpl ...
type TableTpl struct {
	TableTemplate
	ShareList []TableTemplateShare
}
