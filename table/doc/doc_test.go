package doc

import (
	"testing"

	"github.com/XShareGrid/cap/table/registry"
)

func Test_docHander_serveData(t *testing.T) {
	// tables.Init()
	dh := docHander{reg: registry.GlobalTableRegistry()}
	dh.serveData("tables.OperationRecord", 0, 30, nil)
	dh.serveExtractKeys(nil)
}
