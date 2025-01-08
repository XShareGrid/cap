package utils

import (
	"fmt"
	"reflect"
	"testing"

	cap "github.com/XShareGrid/cap/table/proto/go"
	"github.com/XShareGrid/cap/table/registry"
)

func TestParseConditionValue(t *testing.T) {
	v, err := ParseConditionValue(&registry.TableColumnDescriptor{
		DataType:    reflect.TypeOf(int(0)),
		ValueType:   cap.ValueType_VT_STRING,
		ValueFormat: "%03d",
	}, &cap.FilterValue{
		LiteralValues: &cap.Value{V: &cap.Value_VString{VString: "001"}},
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(v)
}
