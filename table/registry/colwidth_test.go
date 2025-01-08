package registry

import "testing"

func TestReloadColWidthFromConfig(t *testing.T) {
	ReloadColWidthFromConfig()
	GlobalTableRegistry().TableMetaReg.List()
}
