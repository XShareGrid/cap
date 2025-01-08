package jsonschema

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/iancoleman/orderedmap"
)

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

// ToSnakeCase 转换为蛇形命名
func ToSnakeCase(str string) string {
	return str
	// snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	// snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	// return strings.ToLower(snake)
}

// GenJSONSchemaObject 生成JSONSchema对象
func GenJSONSchemaObject(ctx context.Context, v interface{}, neglectKeywordList []string) (*Type, error) {
	if reflect.TypeOf(v).Kind() == reflect.Ptr {
		return nil, fmt.Errorf("请传入非指针对象")
	}
	keywordMap := make(map[string]string)
	if len(neglectKeywordList) > 0 {
		for _, neglectKeyword := range neglectKeywordList {
			if _, ok := keywordMap[neglectKeyword]; !ok {
				keywordMap[neglectKeyword] = neglectKeyword
			}
		}
	} else {
		keywordMap = nil
	}
	object, err := genJSONSchemaObject(ctx, v, keywordMap)
	if err != nil {
		return nil, err
	}
	if _, ok := object.(*Type); !ok {
		return nil, fmt.Errorf("当前对象类型不是Type，请检查后再试")
	}
	return object.(*Type), nil
}

const nullDefaultValue = "XXX_NULL_DEFAULT_VALUE"

// GenJSONSchemaObject 生成JSONSchema对象
func genJSONSchemaObject(ctx context.Context, v interface{}, keywordMap map[string]string) (interface{}, error) {

	var err error
	rt := reflect.TypeOf(v)
	rv := reflect.ValueOf(v)
	object := &Type{}
	object.OneOf = []*Type{}
	// 创建对象属性map
	object.Properties = orderedmap.New()
	// properties := make(map[string]*Type)
	// object.Properties = properties
	definitions := make(map[string]*Type)
	object.Definitions = definitions
	if rt.Kind() == reflect.Struct {
		object.Type = JSONSchemaObject
		// 遍历对象设置属性值
		for i := 0; i < rt.NumField(); i++ {
			field := rt.Field(i)
			fieldName := ToSnakeCase(field.Name)
			if keywordMap != nil {
				if _, ok := keywordMap[fieldName]; ok {
					continue
				}
			}
			name := field.Tag.Get("json")
			if name != "" {
				fieldName = name
			}
			// 获取JSON_Schema Tag信息
			tagStr := fmt.Sprint(field.Tag.Get("json_schema"))
			if strings.HasPrefix(fmt.Sprint(field.Type), "jsonschema") {
				definition, err := genJSONSchemaObject(ctx, rv.Field(i).Interface(), keywordMap)
				if err != nil {
					return nil, err
				}
				definitions[ToSnakeCase(reflect.TypeOf(rv.Field(i).Interface()).Name())] = definition.(*Type)
				object.Definitions = definitions
				continue
			}
			defaultValue := ""
			if rv.Kind() != reflect.Slice {
				if !rv.Field(i).IsZero() {
					defaultValue = fmt.Sprintf("%v", rv.Field(i).Interface())
				} else {
					defaultValue = nullDefaultValue
				}
			}
			// 解析Tag
			err = object.ParseTag(ctx, GetJSONSchemaDataType(field.Type), fieldName, defaultValue, tagStr)
			if err != nil {
				return nil, err
			}
			if strings.HasPrefix(fmt.Sprint(field.Type), "[]jsonschema") {
				arrayPropertie, err := genJSONSchemaObject(ctx, rv.Field(i).Interface(), keywordMap)
				if err != nil {
					return nil, err
				}
				if _, ok := object.Properties.Get(fieldName); ok {
					object.Properties.Set(fieldName, arrayPropertie)
				}
				continue
			}

		}
	} else if rt.Kind() == reflect.Slice {
		object.Type = JSONSchemaObject
		for i := 0; i < rt.Elem().NumField(); i++ {
			field := rt.Elem().Field(i)
			fieldName := ToSnakeCase(field.Name)
			if keywordMap != nil {
				if _, ok := keywordMap[fieldName]; ok {
					continue
				}
			}
			name := field.Tag.Get("json")
			if name != "" {
				fieldName = name
			}
			// 获取JSON_Schema Tag信息
			tagStr := fmt.Sprint(field.Tag.Get("json_schema"))
			if strings.HasPrefix(fmt.Sprint(field.Type), "schema") {
				definition, err := genJSONSchemaObject(ctx, rv.Field(i).Interface(), keywordMap)
				if err != nil {
					return nil, err
				}
				definitions[ToSnakeCase(reflect.TypeOf(rv.Field(i).Interface()).Name())] = definition.(*Type)
				object.Definitions = definitions
				continue
			}
			if strings.HasPrefix(fmt.Sprint(field.Type), "[]jsonschema") {
				arrayPropertie, err := genJSONSchemaObject(ctx, rv.Field(i).Interface(), keywordMap)
				if err != nil {
					return nil, err
				}
				object.Properties.Set(fieldName, arrayPropertie.(*Type))
				continue
			}
			defaultValue := ""
			if rv.Kind() != reflect.Slice {
				if !rv.Field(i).IsZero() {
					defaultValue = fmt.Sprintf("%v", rv.Field(i).Interface())
				} else {
					defaultValue = nullDefaultValue
				}
			}
			// 解析Tag
			err = object.ParseTag(ctx, GetJSONSchemaDataType(field.Type), fieldName, defaultValue, tagStr)
			if err != nil {
				return nil, err
			}
		}
	}
	return object, nil
}
