package utils

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/XShareGrid/cap/msg/errors"
	cap "github.com/XShareGrid/cap/table/proto/go"
	"github.com/XShareGrid/cap/table/registry"
	"github.com/shopspring/decimal"
)

func deepCopy(dst, src interface{}) error {
	b, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, dst)
}

// NewTableCell creates cell from value and descriptions
func NewTableCell(userLanguage string, v interface{}, desc *registry.TableColumnDescriptor) (*cap.Cell, error) {
	cell := &cap.Cell{ColumnId: desc.ID}
	var err error
	if v != nil {
		cell.Value, err = MapValue(userLanguage, v, desc)
	} else {
		cell.Value = &cap.Value{}
	}
	if desc.Link != nil {
		cell.Link = &cap.CellLink{
			RemoteTableId:     desc.Link.RemoteTableID,
			ColId:             desc.Link.LocalColID,
			RemoteSearchColId: desc.Link.RemoteSearchColID,
			RemoteValueColId:  desc.Link.RemoteValueColID,
		}
	}
	if desc.ArrSplit != "" && desc.ValueType == cap.ValueType_VT_STRING {
		arr := strings.Split(cell.Value.V.(*cap.Value_VString).VString, desc.ArrSplit)
		d := &registry.TableColumnDescriptor{}
		deepCopy(d, desc)
		d.DataType = reflect.TypeOf("")
		for i := range arr {
			c, _ := MapValue(userLanguage, arr[i], d)
			cell.Values = append(cell.Values, c)
		}
	} else {
		cell.Values = append(cell.Values, cell.Value)
	}
	cell.HrefStyle = desc.HrefStyle
	return cell, err
}

// MapValue from interface to cap.Value
func MapValue(userLanguage string, v interface{}, desc *registry.TableColumnDescriptor) (*cap.Value, error) {
	cv := &cap.Value{}
	ddk := desc.DataType.Kind()
	dt := desc.DataType.String()
	if reflect.TypeOf(v).Kind() == reflect.Ptr {
		// 空值
		if reflect.ValueOf(v).IsNil() {
			v = reflect.New(reflect.TypeOf(v).Elem()).Elem().Interface()
		} else {
			v = reflect.ValueOf(v).Elem().Interface()
		}
	}
	// bytes => v转换
	switch dt {
	case "sql.NullString":
		if vb, ok := v.([]byte); ok {
			v = string(vb)
		} else {
			v = v.(sql.NullString).String
		}
		ddk = reflect.TypeOf(v).Kind()
		dt = reflect.TypeOf(v).String()
	case "sql.NullInt32":
		if vb, ok := v.([]byte); ok {
			v, _ = strconv.Atoi(string(vb))
		} else {
			v = v.(sql.NullInt32).Int32
		}
		ddk = reflect.TypeOf(v).Kind()
		dt = reflect.TypeOf(v).String()
	case "sql.NullInt64":
		if vb, ok := v.([]byte); ok {
			v, _ = strconv.Atoi(string(vb))
		} else if vb, ok := v.(int64); ok {
			v = vb
		} else {
			v = v.(sql.NullInt64).Int64
		}
		ddk = reflect.TypeOf(v).Kind()
		dt = reflect.TypeOf(v).String()
	case "sql.NullTime":
		v = v.(sql.NullTime).Time
		ddk = reflect.TypeOf(v).Kind()
		dt = reflect.TypeOf(v).String()
	case "sql.NullBool":
		if vb, ok := v.([]byte); ok {
			v, _ = strconv.Atoi(string(vb))
			v = (v != 0)
		} else {
			v = v.(sql.NullBool).Bool
		}
		ddk = reflect.TypeOf(v).Kind()
		dt = reflect.TypeOf(v).String()
	case "sql.NullFloat64":
		if vb, ok := v.([]byte); ok {
			v, _ = strconv.ParseFloat(string(vb), 64)
		} else {
			v = v.(sql.NullFloat64).Float64
		}
		ddk = reflect.TypeOf(v).Kind()
		dt = reflect.TypeOf(v).String()
	}
	if desc.Href != "" {
		cv.Href = fmt.Sprintf(desc.Href, v)
	}
	if (ddk > reflect.Invalid && ddk <= reflect.Float64) || ddk == reflect.String {
		valueFmt := ""
		switch ddk {
		case reflect.Bool:
			if vb, ok := v.([]byte); ok {
				v, _ = strconv.Atoi(string(vb))
				v = (v != 0)
			}
			cv.V = &cap.Value_VBool{VBool: v.(bool)}
			valueFmt = "%d"
		case reflect.Int,
			reflect.Int8,
			reflect.Int16,
			reflect.Int32,
			reflect.Int64:
			if vb, ok := v.([]byte); ok {
				v, _ = strconv.Atoi(string(vb))
			}
			cv.V = &cap.Value_VInt{VInt: int32(reflect.ValueOf(v).Int())}
			valueFmt = "%d"
		case reflect.Uint,
			reflect.Uint8,
			reflect.Uint16,
			reflect.Uint32,
			reflect.Uint64:
			if vb, ok := v.([]byte); ok {
				v, _ = strconv.Atoi(string(vb))
			}
			cv.V = &cap.Value_VInt{VInt: int32(reflect.ValueOf(v).Uint())}
			valueFmt = "%d"
		case reflect.Float32:
			if vb, ok := v.([]byte); ok {
				v, _ = strconv.ParseFloat(string(vb), 64)
			}
			fv, _ := decimal.NewFromFloat32(v.(float32)).Float64()
			cv.V = &cap.Value_VDouble{VDouble: fv}
			valueFmt = "%g"
		case reflect.Float64:
			if vb, ok := v.([]byte); ok {
				v, _ = strconv.ParseFloat(string(vb), 64)
			}
			cv.V = &cap.Value_VDouble{VDouble: v.(float64)}
			valueFmt = "%g"
		case reflect.String:
			if vb, ok := v.([]byte); ok {
				v = string(vb)
			}
			if vString, ok := v.(string); ok {
				cv.V = &cap.Value_VString{VString: vString}
				valueFmt = "%s"
			}
		}
		if desc.ValueFormat != "" {
			valueFmt = desc.ValueFormat
		}
		// 处理string
		if desc.ValueType == cap.ValueType_VT_STRING {
			cv.V = &cap.Value_VString{VString: fmt.Sprintf(valueFmt, v)}
		} else if desc.ValueType == cap.ValueType_VT_OPTION {
			if vInt, ok := cv.V.(*cap.Value_VInt); ok {
				opt, err := registry.GlobalTableRegistry().OptionReg.LookupWithI18n(userLanguage, dt, vInt.VInt)
				if err != nil {
					cv.V = &cap.Value_VOption{VOption: &cap.OptionValue{Id: vInt.VInt, Name: "N/A"}}
				} else {
					cv.V = &cap.Value_VOption{VOption: opt}
				}
			}
		}
		return cv, nil
	} else if dt == "time.Time" {
		valueFmt := time.RFC3339
		if desc.ValueType == cap.ValueType_VT_DATE {
			valueFmt = "2006-01-02"
		}
		if desc.ValueFormat != "" {
			valueFmt = desc.ValueFormat
		}
		if desc.ValueType == cap.ValueType_VT_DATE {
			cv.V = &cap.Value_VDate{VDate: v.(time.Time).Format(valueFmt)}
		} else if desc.ValueType == cap.ValueType_VT_TIME {
			cv.V = &cap.Value_VTime{VTime: v.(time.Time).Format(valueFmt)}
		} else {
			return nil, errors.Wrap(ErrFailedMapValue).FillDebugArgs(
				v, desc.DataType.String(), cap.ValueType_name[int32(desc.ValueType)])
		}
		return cv, nil
	}
	return nil, errors.Wrap(ErrFailedMapValue).FillDebugArgs(
		v, desc.DataType.String(), cap.ValueType_name[int32(desc.ValueType)])
}

// MapValue from interface to cap.Value
func EmptyValue(desc *registry.TableColumnDescriptor) *cap.Value {
	switch desc.ValueType {
	case cap.ValueType_VT_STRING:
		return &cap.Value{V: &cap.Value_VString{VString: ""}}
	case cap.ValueType_VT_INT:
		return &cap.Value{V: &cap.Value_VInt{VInt: 0}}
	case cap.ValueType_VT_DOUBLE:
		return &cap.Value{V: &cap.Value_VDouble{VDouble: 0.0}}
	case cap.ValueType_VT_DATE:
		return &cap.Value{V: &cap.Value_VDate{VDate: "0000-00-00"}}
	case cap.ValueType_VT_TIME:
		return &cap.Value{V: &cap.Value_VTime{VTime: "0000-00-00T00:00:00Z"}}
	case cap.ValueType_VT_BOOLEAN:
		return &cap.Value{V: &cap.Value_VBool{VBool: false}}
	case cap.ValueType_VT_OPTION:
		return &cap.Value{V: &cap.Value_VOption{VOption: &cap.OptionValue{Id: -1, Name: "N/A"}}}
	}
	return nil
}
