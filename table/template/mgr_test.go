package template

import (
	"context"
	"testing"

	db "github.com/XShareGrid/cap/database/mysql"
	"github.com/XShareGrid/cap/test"
)

var testDB *db.DB

func init() {
	var err error
	testDB, err = db.NewTestDBFromEnvVar()
	if err != nil {
		panic(err)
	}
}

func TestManager_FindTemplate(t *testing.T) {
	// filters := cap.FilterBody{}
	// filters.Conditions = append(filters.Conditions, &cap.Condition{
	// 	ColumnId:   "string",
	// 	OperatorId: "1",
	// 	Values: []*cap.FilterValue{
	// 		{
	// 			LiteralValues: &cap.Value{
	// 				V: &cap.Value_VOption{VOption: &cap.OptionValue{Id: 1}},
	// 			},
	// 		},
	// 	},
	// })
	// filters.Conditions = append(filters.Conditions, &cap.Condition{
	// 	ColumnId:   "string",
	// 	OperatorId: "1",
	// 	Values: []*cap.FilterValue{
	// 		{
	// 			LiteralValues: &cap.Value{
	// 				V: &cap.Value_VDouble{VDouble: 111.1},
	// 			},
	// 		},
	// 	},
	// })
	// filters.Conditions = append(filters.Conditions, &cap.Condition{
	// 	ColumnId:   "string",
	// 	OperatorId: "1",
	// 	Values: []*cap.FilterValue{
	// 		{
	// 			LiteralValues: &cap.Value{
	// 				V: &cap.Value_VString{VString: "1111"},
	// 			},
	// 		},
	// 	},
	// })
	// filters = cap.FilterBody{}
	// fb, _ := proto.Marshal(&filters)
	// fmt.Println(string(fb))

	// err := proto.Unmarshal(fb, &filters)
	// if err != nil {
	// 	panic(err)
	// }

	ss, err := testDB.NewSession()
	if err != nil {
		panic(err)
	}
	tpl, err := GlobalManager().FindTemplate(context.Background(), ss, "5cdc8435-d31c-11eb-b8e5-005056af603f")
	if err != nil {
		panic(err)
	}
	test.DisplayObject(tpl)
}
