package registry

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/XShareGrid/cap/logger"
)

const colWidthPath = "./conf/table/"

// ReloadColWidthFromConfig 通过配置文件设置列宽
func ReloadColWidthFromConfig() {
	// 判断colWidthPath目录是否存在
	if _, err := os.Stat(colWidthPath); os.IsNotExist(err) {
		return
	}
	// 遍历路径下的json文件
	filepath.Walk(colWidthPath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(strings.ToLower(info.Name()), ".json") {
			if err := parseWidthFromJsonFile(info.Name(), path); err != nil {
				logger.CAP.Error("failed to parse width from json file[%s]: %s", info.Name(), err.Error())
			} else {
				logger.CAP.Info("Loaded width from json file[%s]", info.Name())
			}
		}
		return nil
	})
}

func parseWidthFromJsonFile(fileName, path string) error {
	fileName = strings.TrimSuffix(fileName, ".json")
	jsonBytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	// 列名-宽度
	var configWidths map[string]float64
	err = json.Unmarshal(jsonBytes, &configWidths)
	if err != nil {
		return err
	}

	tmd, err := GlobalTableRegistry().TableMetaReg.Find(fileName)
	if err != nil {
		return err
	}
	tmdImpl := tmd.(*TableMetaDataImpl)
	for k, v := range configWidths {
		for i, col := range tmdImpl.columns.list {
			if col.ID == k {
				// 前端默认单位是80px
				tmdImpl.columns.list[i].ColWidth = v / 80.0
				logger.CAP.Debug("width set for[%s] = %f", k, v)
			}
		}
	}
	return nil
}
