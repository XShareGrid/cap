package jsonschema

import (
	"context"
	"fmt"
)

// Option 选项
type Option struct {
	Key   string
	Value string
}

// OptionCallback 回调函数
type OptionCallback func(ctx context.Context) ([]Option, error)

// callbackMap 回调函数列表
var optionCallbackMap map[string]OptionCallback

// RegisterOptionFunction 注册Option方法
func RegisterOptionFunction(funcName string, callback OptionCallback) error {
	if callback == nil {
		return fmt.Errorf("option callback: Register callback is nil")
	}
	if optionCallbackMap == nil {
		optionCallbackMap = make(map[string]OptionCallback)
	}
	if _, ok := optionCallbackMap[funcName]; ok {
		return fmt.Errorf("option callback: Register called twice for callback" + funcName)
	} else {
		optionCallbackMap[funcName] = callback
	}
	return nil
}
