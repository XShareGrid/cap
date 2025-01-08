package debug

import (
	"github.com/mbndr/figlet4go"
	telnet "github.com/reiver/go-telnet"
	"github.com/reiver/go-telnet/telsh"
)

var shellHandler *telsh.ShellHandler

func init() {
	shellHandler = telsh.NewShellHandler()
	ascii := figlet4go.NewAsciiRender()
	opt := figlet4go.NewRenderOptions()
	opt.FontName = "larry3d"
	renderStr, _ := ascii.RenderOpts("EAP Welcome", opt)
	shellHandler.WelcomeMessage = renderStr
}

func Register(name string, f telsh.HandlerFunc) error {
	nf := func(ctx telnet.Context, name string, args ...string) telsh.Handler {
		return telsh.PromoteHandlerFunc(f, args...)
	}
	return shellHandler.Register(name, telsh.ProducerFunc(nf))
}
