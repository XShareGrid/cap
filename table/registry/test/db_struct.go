package rt

import (
	"encoding/json"
	"time"

	cap "github.com/XShareGrid/cap/table/proto/go"
)

type TableTemplate struct {
	Id          string             `db:"id" t_key:"true" t_fm:"SSTR" t_name:"ID" t_vt:"STRING"`            // ID|唯一id
	TName       string             `db:"name" t_fm:"SSTR" t_name:"模板名" t_vt:"STRING"`                      // 模板名|table_id+模板名需唯一
	TableId     string             `db:"table_id" t_fm:"SSTR" t_name:"表ID" t_vt:"STRING"`                  // 表ID|表ID
	FAccess     cap.FileAccessType `db:"f_access" t_fm:"SINT" t_name:"访问权限" t_vt:"OPTION"`                 // 访问权限|0-PRIVATE, 1-PUBLIC, 2-SHARED
	FCreateUser int64              `db:"f_create_user" t_agg:"SUM" t_fm:"SINT" t_name:"创建用户ID" t_vt:"INT"` // 创建用户ID|
	FCreateTime time.Time          `db:"f_create_time" t_fm:"STIME|SNULLABLE" t_name:"创建时间" t_vt:"TIME"`   // 创建时间|
	FModTime    time.Time          `db:"f_mod_time" t_fm:"STIME|SNULLABLE" t_name:"更新事件" t_vt:"TIME"`      // 更新事件|
	Body        json.RawMessage    `db:"body" t_internal:"true" t_name:"模板内容"`                             // 模板内容|
}

func (TableTemplate) TableName() string {
	return "table_template"
}

// Name ...
func (TableTemplate) Name() string {
	return "报表模板"
}

// Desc ...
func (TableTemplate) Desc() string {
	return "报表的模板列表"
}
