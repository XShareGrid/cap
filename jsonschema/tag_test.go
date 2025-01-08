package jsonschema

import (
	"testing"
)

func TestGetDefinitions(t *testing.T) {
	// def, err := GetDefinitions("t_name[type-string;required-A:B;default-夏文杰;enum-A:B:C:D:E]")
	// if err != nil {
	// 	panic(err)
	// }
	// db.DisplayObject(def)
}

func TestCommonKeywords(t *testing.T) {
	ty := &Type{}
	err := ty.CommonKeywords("format", "date")
	if err != nil {
		panic(err)
	}
}
