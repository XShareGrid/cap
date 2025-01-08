package registry

import (
	"fmt"
	"sort"
	"strings"
)

func TrKeyOperator(operatorID string) string {
	return strings.ReplaceAll(fmt.Sprintf("table_op_%s", operatorID), ".", "_")
}

func TrKeyOption(optionTypeID string, optionID int) string {
	return strings.ReplaceAll(fmt.Sprintf("table_opt_%s_%d", optionTypeID, optionID), ".", "_")
}

func TrKeyTableName(tableID string) string {
	return strings.ReplaceAll(fmt.Sprintf("table_tn_%s", tableID), ".", "_")
}

func TrKeyTableDesc(tableID string) string {
	return strings.ReplaceAll(fmt.Sprintf("table_td_%s", tableID), ".", "_")
}

func TrKeyTableField(tableID string, fieldID string) string {
	return strings.ReplaceAll(fmt.Sprintf("table_tf_%s_%s", tableID, fieldID), ".", "_")
}

func TrKeyTableAction(tableID string, actionID string) string {
	return strings.ReplaceAll(fmt.Sprintf("table_ta_%s_%s", tableID, actionID), ".", "_")
}

type TrKey struct {
	ID  string
	Src string
}

func ExtractKeys() []TrKey {
	keys := []TrKey{}
	// Operators 操作符
	operators := GlobalTableRegistry().OperatorReg.FindAll()
	for _, op := range operators {
		keys = append(keys, TrKey{ID: TrKeyOperator(op.ID()), Src: op.Name()})
	}

	// Options 枚举
	optionTypes := GlobalTableRegistry().OptionReg.GetAllOptions()
	for _, opt := range optionTypes {
		for _, op := range opt.Options {
			keys = append(keys, TrKey{ID: TrKeyOption(opt.ID, int(op.Id)), Src: op.Name})
		}
	}

	// table fields
	// 表名/描述/字段名/操作
	tables := GlobalTableRegistry().TableMetaReg.FindAll()
	for _, table := range tables {
		tmd := table.(TableMetaData)
		keys = append(keys, TrKey{ID: TrKeyTableName(tmd.ID()), Src: tmd.Name()})
		keys = append(keys, TrKey{ID: TrKeyTableDesc(tmd.ID()), Src: tmd.Desc()})
		fields := tmd.Columns()
		for _, field := range fields.List() {
			keys = append(keys, TrKey{ID: TrKeyTableField(tmd.ID(), field.ID), Src: field.Name})
		}
		actions := tmd.GetAllActions()
		for _, action := range actions {
			keys = append(keys, TrKey{ID: TrKeyTableAction(tmd.ID(), action.ID()), Src: action.Name()})
		}
	}

	// sort
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].ID < keys[j].ID
	})

	return keys
}
