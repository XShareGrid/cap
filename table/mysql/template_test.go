package mysql

import (
	"encoding/json"
	"log"
	"testing"
	"time"

	db "github.com/XShareGrid/cap/database/mysql"
	"github.com/XShareGrid/cap/ss/idgen"
	"github.com/XShareGrid/cap/test"
	_ "github.com/astaxie/beego/session/mysql"
)

var testDB *db.DB

func init() {
	var err error
	testDB, err = db.NewTestDBFromEnvVar()
	if err != nil {
		panic(err)
	}
}

func TestTableTemplateMapper_CreateTemplate(t *testing.T) {
	ss, err := testDB.NewSession()
	if err != nil {
		t.Fatal(err)
	}
	rand := idgen.NewRandomIDGenerator(5)
	defer ss.Close(err)
	mapper := NewTableTemplateMapper(ss)
	for i := 0; i < 10; i++ {
		id, _ := rand.Generate()
		err = mapper.CreateTemplate(&TableTpl{
			TableTemplate: TableTemplate{
				Id:          id,
				Name:        id + "_n",
				TableId:     "111",
				FAccess:     0,
				FCreateUser: "1",
				FCreateTime: time.Now(),
				FModTime:    time.Now(),
				Body:        []byte(`{"a":1}`),
			},
			ShareList: []TableTemplateShare{
				{UserId: "1"},
				{UserId: "2"},
				{UserId: "3"},
			},
		})
	}
	if err != nil {
		t.Fatal(err)
	}
}

func TestTableTemplateMapper_FindTemplates(t *testing.T) {
	ss, err := testDB.NewSession()
	if err != nil {
		t.Fatal(err)
	}
	defer ss.Close(err)
	mapper := NewTableTemplateMapper(ss)
	templates, err := mapper.FindTemplates(FilterTemplateIDEquals("392109bf-96b4-11eb-ab69-005056afd813"))
	if err != nil {
		t.Fatal(err)
	}
	test.DisplayObject(templates)
}

func TestTableTemplateMapper_FindTemplate(t *testing.T) {
	ss, err := testDB.NewSession()
	if err != nil {
		t.Fatal(err)
	}
	defer ss.Close(err)
	mapper := NewTableTemplateMapper(ss)
	template, err := mapper.FindTemplate("STgf0", true)
	if err != nil {
		t.Fatal(err)
	}
	test.DisplayObject(template)
}

func TestTableTemplateMapper_DeleteTemplates(t *testing.T) {
	ss, err := testDB.NewSession()
	if err != nil {
		t.Fatal(err)
	}
	defer ss.Close(err)
	mapper := NewTableTemplateMapper(ss)
	affected, err := mapper.DeleteTemplates(FilterTemplateIDEquals("STgf0"))
	if err != nil {
		t.Fatal(err)
	}
	log.Println("affected =", affected)
}

func TestTableTemplateMapper_UpdateTableTemplate(t *testing.T) {
	ss, err := testDB.NewSession()
	if err != nil {
		t.Fatal(err)
	}
	defer ss.Close(err)
	mapper := NewTableTemplateMapper(ss)
	err = mapper.UpdateTableTemplate("p7Cj9", "测试", 2, json.RawMessage(`{"test":77777777}`),
		[]TableTemplateShare{
			{UserId: "4"},
			{UserId: "5"},
			{UserId: "6"},
		})
	if err != nil {
		t.Fatal(err)
	}
}

func TestTableTemplateMapper_FindTemplatesByShareUser(t *testing.T) {
	ss, err := testDB.NewSession()
	if err != nil {
		t.Fatal(err)
	}
	defer ss.Close(err)
	mapper := NewTableTemplateMapper(ss)
	templates, err := mapper.FindTemplatesByShareUserAndTableID("1", "111")
	if err != nil {
		t.Fatal(err)
	}
	test.DisplayObject(templates)
}
