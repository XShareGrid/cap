package registry

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"reflect"

	"github.com/XShareGrid/cap/proto/extension"
	cap "github.com/XShareGrid/cap/table/proto/go"
	"github.com/golang/protobuf/proto"
	protobuf "github.com/golang/protobuf/protoc-gen-go/descriptor"
)

// EnumOption 枚举选项
type EnumOption struct {
	TypeID  string
	Options []*cap.OptionValue
}

// PBEnum pb enum
type PBEnum interface {
	EnumDescriptor() ([]byte, []int)
}

// COPY from pb source code
func extractFile(gz []byte) (*protobuf.FileDescriptorProto, error) {
	r, err := gzip.NewReader(bytes.NewReader(gz))
	if err != nil {
		return nil, fmt.Errorf("failed to open gzip reader: %v", err)
	}
	defer r.Close()

	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to uncompress descriptor: %v", err)
	}

	fd := new(protobuf.FileDescriptorProto)
	if err := proto.Unmarshal(b, fd); err != nil {
		return nil, fmt.Errorf("malformed FileDescriptorProto: %v", err)
	}

	return fd, nil
}

// LoadOptionFromProtoEnum 通过pb生成的枚举来自动加载选项数据
func LoadOptionFromProtoEnum(en PBEnum) (*EnumOption, error) {
	b, _ := en.EnumDescriptor()
	fd, err := extractFile(b)
	if err != nil {
		return nil, err
	}
	opt := &EnumOption{TypeID: reflect.TypeOf(en).String()}

	enName := reflect.TypeOf(en).Name()
	for _, ed := range fd.EnumType {
		if *ed.Name == enName {
			opt.Options = []*cap.OptionValue{}
			for _, o := range ed.Value {
				var optName string

				ext, err := proto.GetExtension(o.Options, extension.E_OptionName)
				if err == nil && ext.(*string) != nil {
					optName = *(ext.(*string))
				} //else {
				// //如果上面扩展名没有找到，继续第二个扩展名
				// ext, err := proto.GetExtension(o.Options, dict.E_OperationTypeName)
				// if err == nil && ext.(*string) != nil {
				// 	optName = *(ext.(*string))
				// } else {
				// 	//如果上面扩展名没有找到，继续第三个扩展名
				// 	ext, err := proto.GetExtension(o.Options, dict.E_BillStatusName)
				// 	if err == nil && ext.(*string) != nil {
				// 		optName = *(ext.(*string))
				// 	} else {
				// 		optName = *o.Name
				// 	}
				// }
				//}
				ov := &cap.OptionValue{Id: *o.Number, Name: optName}
				if optName == "-" {
					continue
				}
				opt.Options = append(opt.Options, ov)
			}
		}
	}
	return opt, nil
}

// RegisterOptionFromProtoEnum load + register
// 将proto枚举注册为option
func RegisterOptionFromProtoEnum(en PBEnum) error {
	eo, err := LoadOptionFromProtoEnum(en)
	if err != nil {
		return err
	}
	err = GlobalTableRegistry().OptionReg.Register(eo.TypeID, eo.Options)
	if err != nil {
		return err
	}
	return nil
}
