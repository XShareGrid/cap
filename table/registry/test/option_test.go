package rt

import (
	"log"
	"testing"

	"github.com/XShareGrid/cap/table/registry"
)

func TestLoadProtoEnum(t *testing.T) {
	eo, err := registry.LoadOptionFromProtoEnum(TestEnum(0))
	if err != nil {
		t.Fatal(err)
	}
	err = registry.GlobalTableRegistry().OptionReg.Register(eo.TypeID, eo.Options)
	if err != nil {
		t.Fatal(err)
	}
	opts, err := registry.GlobalTableRegistry().OptionReg.GetOptions("rt.TestEnum")
	if err != nil {
		t.Fatal(err)
	}
	for _, o := range opts {
		log.Println(o.Id, o.Name)
	}

}
