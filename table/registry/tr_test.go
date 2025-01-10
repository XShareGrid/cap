package registry

import (
	"fmt"
	"os"
	"testing"

	"gopkg.in/ini.v1"
)

func TestExtractKeys(t *testing.T) {
	// 注册枚举
	// RegisterOptionFromProtoEnum(dict.BillStatus(0))
	//RegisterOptionFromProtoEnum(oms.MaterialStatus(0))
	// 提取翻译Key
	keys := ExtractKeys()
	// 把keys转换为ini文件打印出来
	iniFile := ini.Empty()
	sec, _ := iniFile.NewSection("Options")
	for _, key := range keys {
		sec.NewKey(key.ID, key.Src)
	}
	// 写入文件
	fmt.Println("添加以下片段到zh-CN.ini，并自行翻译en-US.ini")
	iniFile.WriteTo(os.Stdout)
	// iniFile.SaveTo("文件名.ini")
	//val, err := GlobalTableRegistry().OptionReg.LookupWithI18n(i18n.EN, "oms.Grade", 1)
	////val, err := GlobalTableRegistry().OptionReg.LookupWithI18n(i18n.EN, "oms.MaterialStatus", 5)
	//if err != nil {
	//	t.Fatal(err)
	//}
	//fmt.Printf("val:%+v\n", val)
	//fmt.Printf("val.Name:%s\n", val.Name)
}
