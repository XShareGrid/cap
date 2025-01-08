package utils

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/XShareGrid/cap/logger"
	"github.com/XShareGrid/cap/msg/errors"
	cap "github.com/XShareGrid/cap/table/proto/go"
	"github.com/XShareGrid/cap/table/registry"
)

// ParseConditionValue cap.Value -> 实际的datatype
func ParseConditionValue(desc *registry.TableColumnDescriptor, v *cap.FilterValue) (interface{}, error) {
	switch desc.ValueType {
	// vt为string时，dt可能是int/float/bool/time/date等等
	case cap.ValueType_VT_STRING:
		// 默认为空字符串
		if v.GetLiteralValues() == nil || v.GetLiteralValues().V == nil {
			v.LiteralValues = &cap.Value{V: &cap.Value_VString{VString: ""}}
		}
		vString := strings.TrimSpace(v.GetLiteralValues().GetVString())
		switch desc.DataType.Kind() {
		case reflect.Int,
			reflect.Int8,
			reflect.Int16,
			reflect.Int32,
			reflect.Int64,
			reflect.Uint,
			reflect.Uint8,
			reflect.Uint16,
			reflect.Uint32,
			reflect.Uint64:
			var vInt int32
			f := desc.ValueFormat
			if f == "" {
				f = "%d"
			}
			_, err := fmt.Sscanf(vString, f, &vInt)
			if err == nil {
				return vInt, nil
			}
			// 解析失败直接返回-9999 当无效条件
			logger.CAP.Error("error while ParseInt(%s = %s)", desc.ID, vString)
			return -9999, nil
		case reflect.Float32,
			reflect.Float64:
			var vFloat float64
			f := desc.ValueFormat
			if f == "" {
				f = "%g"
			}
			_, err := fmt.Sscanf(vString, f, &vFloat)
			if err == nil {
				return vFloat, nil
			}
			logger.CAP.Error("error while ParseFloat(%s = %s)", desc.ID, vString)
			return nil, errors.Wrap(err).Log()
		case reflect.String:
			var vStringUnFmt string
			f := desc.ValueFormat
			if f == "" {
				return vString, nil
			}
			_, err := fmt.Sscanf(vString, f, &vStringUnFmt)
			if err == nil {
				return vStringUnFmt, err
			}
			return nil, errors.Wrap(err).Log()
		default:
			return vString, nil
		}
	case cap.ValueType_VT_INT:
		return v.GetLiteralValues().GetVInt(), nil
	case cap.ValueType_VT_DOUBLE:
		return v.GetLiteralValues().GetVDouble(), nil
	case cap.ValueType_VT_DATE:
		f := desc.ValueFormat
		if f == "" {
			f = "2006-01-02"
		}
		t, err := time.Parse(f, v.GetLiteralValues().GetVDate())
		if err == nil {
			return t, err
		}
		logger.CAP.Error("error while parse date as format(%s): "+err.Error(), desc.ValueFormat)
		return nil, errors.Wrap(err).Log()
	case cap.ValueType_VT_TIME:
		f := desc.ValueFormat
		if f == "" {
			f = time.RFC3339
		}
		t, err := time.Parse(f, v.GetLiteralValues().GetVTime())
		if err == nil {
			return t, err
		}
		logger.CAP.Error("error while parse time as format(%s): "+err.Error(), desc.ValueFormat)
		return nil, errors.Wrap(err).Log()
	case cap.ValueType_VT_BOOLEAN:
		return v.GetLiteralValues().GetVBool(), nil
	case cap.ValueType_VT_OPTION:
		if lv := v.GetLiteralValues(); lv != nil {
			opt := lv.GetVOption()
			if opt == nil {
				return 0, nil
			}
			return lv.GetVOption().Id, nil
		}
	}
	return nil, errors.Wrap(ErrFailedParseConditionValue).FillDebugArgs(desc.ID, v)
}
