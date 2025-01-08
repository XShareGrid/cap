package service

import (
	"context"
	"fmt"
	"testing"

	db "github.com/XShareGrid/cap/database/mysql"
	"github.com/XShareGrid/cap/table/data"
	"github.com/XShareGrid/cap/table/demo/tables"
	cap "github.com/XShareGrid/cap/table/proto/go"

  	"github.com/XShareGrid/cap/test"

	// "gitlab.pintechs.com/eh/energy-backend/mts2/pkg/tables"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var testDB *db.DB

func init() {
	var err error
	testDB, err = db.NewTestDBFromEnvVar()
	if err != nil {
		panic(err)
	}
}

func TestGetTableInfo(t *testing.T) {
	conn, err := grpc.Dial("192.168.34.11:8080", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	cli := cap.NewTableWServiceClient(conn)
	req := cap.GetTableInfoReq{TableId: "tables.Material"}
	ctx := context.Background()
	// 添加token
	ctx = metadata.AppendToOutgoingContext(ctx, "Subsystem", "mes")
	rsp, err := cli.GetTableInfo(ctx, &req)
	if err != nil {
		panic(err)
	}
	fmt.Println(rsp)
}

func TestGetTableColumns(t *testing.T) {
	conn, err := grpc.Dial("oms-sh.poly-energy.com:8585", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	cli := cap.NewTableWServiceClient(conn)
	req := cap.GetTableColumnsReq{TableId: "tables.ErpPrdMo"}
	ctx := context.Background()
	// 添加token
	ctx = metadata.AppendToOutgoingContext(ctx, "Subsystem", "mes")
	rsp, err := cli.GetTableColumns(ctx, &req)
	if err != nil {
		panic(err)
	}
	fmt.Println(rsp)
}

func TestGetTableTemplates(t *testing.T) {
	conn, err := grpc.Dial("192.168.34.11:8080", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	cli := cap.NewTableWServiceClient(conn)
	req := cap.GetTableTemplatesReq{TableId: "tables.Material"}
	ctx := context.Background()
	// 添加token
	ctx = metadata.AppendToOutgoingContext(ctx, "Subsystem", "mes")
	rsp, err := cli.GetTableTemplates(ctx, &req)
	if err != nil {
		panic(err)
	}
	fmt.Println(rsp)
}

func fixedString(s string, length int) string {
	sl := len(s)
	if sl > length {
		return s[0:sl]
	}
	r := make([]byte, length)
	for i := 0; i < length; i++ {
		r[i] = byte(' ')
	}

	startIdx := (length - sl) / 2
	copy(r[startIdx:], []byte(s))

	return string(r)
}

func TestFixedString(t *testing.T) {
	fmt.Println(fixedString("1", 20))
}

func TestGetTableRows(t *testing.T) {
	conn, err := grpc.Dial("192.168.34.15:8888", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	cli := cap.NewTableWServiceClient(conn)

	// tpl := &cap.TemplateQuery_TmpTpl{TmpTpl: utils.NewTmpTpl("",
	// 	[]*driver.Condition{}, []string{"ID",
	// 		"FID",
	// 		"OrderNumber",
	// 		"MoType",
	// 		"EntryID",
	// 		"WorkshopNumber",
	// 		"WorkshopName",
	// 		"Date",
	// 		"FOraAssistant",
	// 		"FOraAssistant1",
	// 		"FOraAssistant2",
	// 		"Qty",
	// 		"MaterialCode",
	// 		"MaterialModel",
	// 		"MaterialName",
	// 		"MaterialSpec",
	// 		"MaterialSection",
	// 		"Length",
	// 		"Width"}).Body}
	req := cap.GetTableRowsReq{
		TableId: "tables.ErpPrdMo",
		Page:    &cap.PageParam{Page: 0, PageSize: 100},
		Tpl:     &cap.TemplateQuery{Tpl: &cap.TemplateQuery_TplId{TplId: "ad48c418-96bb-11eb-a8de-005056afd813"}},
	}
	ctx := context.Background()
	// 添加token
	ctx = metadata.AppendToOutgoingContext(ctx, "Subsystem", "mes")
	rsp, err := cli.GetTableRows(ctx, &req)
	if err != nil {
		panic(err)
	}
	for _, r := range rsp.Rows {
		fmt.Printf("|")
		for _, c := range r.Cells {
			if s, ok := c.Value.V.(*cap.Value_VString); ok {
				fmt.Printf("%s", fixedString(s.VString, 20))
			} else if o, ok := c.Value.V.(*cap.Value_VOption); ok {
				fmt.Printf("%s", fixedString(o.VOption.Name, 20))
			} else if o, ok := c.Value.V.(*cap.Value_VDate); ok {
				fmt.Printf("%s", fixedString(o.VDate, 20))
			}

			fmt.Printf("|")
		}
		fmt.Printf("\n")
	}
	fmt.Printf("第 %d / %d 页，每页 %d 项\n",
		rsp.PageInfo.GetCurrentPage(), rsp.PageInfo.GetTotalPages(), rsp.PageInfo.GetPageSize())
}

func Test_parseTpl(t *testing.T) {
	tables.Init()
	ss, err := testDB.NewSession()
	if err != nil {
		panic(err)
	}
	tpl, err := data.ParseTpl(context.Background(), ss, "tables.FlowOrder", &cap.TemplateQuery{
		Tpl: &cap.TemplateQuery_TplId{TplId: "ad48c418-96bb-11eb-a8de-005056afd813"}},
	)
	if err != nil {
		panic(err)
	}
	test.DisplayObject(tpl)
}
