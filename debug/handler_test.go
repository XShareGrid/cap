package debug

import (
	"fmt"
	"testing"

	"github.com/mbndr/figlet4go"
)

func TestRegister(t *testing.T) {
	ascii := figlet4go.NewAsciiRender()
	opt := figlet4go.NewRenderOptions()
	opt.FontName = "larry3d"
	renderStr, _ := ascii.RenderOpts("EAP Welcome", opt)
	fmt.Println(renderStr)
}
