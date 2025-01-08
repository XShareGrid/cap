package registry

import (
	"testing"

	cap "github.com/XShareGrid/cap/table/proto/go"
	// mts2 "github.com/XShareGrid/cap/table/proto/go/mts2/go"
)

func TestLoadOptionFromProtoEnum(t *testing.T) {
	LoadOptionFromProtoEnum(cap.ValueType(0))
}
