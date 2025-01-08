package logger

import (
	sslog "github.com/XShareGrid/cap/ss/logger"
	"github.com/astaxie/beego/logs"
)

// CAP beego logger
var CAP *logs.BeeLogger

// GRPC grpc logger
var GRPC *logs.BeeLogger

func init() {
	CAP = logs.NewLogger(1000)
	CAP.SetLogger("console", "")
	GRPC = logs.NewLogger(1000)
	GRPC.SetLogger("console", "")
}

// LogConfig struct
type LogConfig struct {
	FileName string `json:"filename"`
	Maxlines int
	Maxsize  int
	Daily    bool
}

// InitLogger initialization logger
func InitLogger(logPath string, grpcLogPath string, level string) {

	CAP.SetLogger("file", sslog.FileConfig(logPath))

	GRPC.SetLogger("file", sslog.FileConfig(grpcLogPath))
	if level == "Debug" {
		CAP.SetLevel(logs.LevelDebug)
		GRPC.SetLevel(logs.LevelDebug)
	} else if level == "Info" {
		CAP.SetLevel(logs.LevelInfo)
		GRPC.SetLevel(logs.LevelInfo)
	}

	CAP.EnableFuncCallDepth(true)
	GRPC.EnableFuncCallDepth(true)
	CAP.SetLogFuncCallDepth(2)
	GRPC.SetLogFuncCallDepth(2)
}
