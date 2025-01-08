package logger

import "encoding/json"

// LogConfig struct
type logConfig struct {
	FileName   string `json:"filename"`
	Maxlines   int    `json:"maxlines"`
	MaxFiles   int    `json:"maxfiles"`
	MaxSize    int    `json:"maxsize"`
	Daily      bool   `json:"daily"`
	MaxDays    int64  `json:"maxdays"`
	Rotate     bool   `json:"rotate"`
	Perm       string `json:"perm"`
	RotatePerm string `json:"rotateperm"`
}

// FileConfig beelogger file configuration
func FileConfig(logPath string) string {
	jsonConfig := logConfig{}
	jsonConfig.Rotate = true
	jsonConfig.FileName = logPath
	jsonConfig.Maxlines = 1000000
	jsonConfig.MaxDays = 90
	// 100M
	jsonConfig.MaxSize = 100 * 1024 * 1024
	jsonConfig.Daily = true
	jsonConfig.Perm = "0660"
	jsonConfig.RotatePerm = "0440"
	jsonConfig.MaxFiles = 999
	configString, _ := json.Marshal(jsonConfig)
	return string(configString)
}
