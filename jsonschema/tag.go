package jsonschema

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Tag 标记
type Tag struct {
	Ctx        context.Context    // 必填
	Object     *Type              // 必填
	Type       JSONSchemaDataType // 字段类型
	Value      string             // Tag未转换前
	FieldName  string             // 字段名
	FieldValue string             // 字段值
}

// SetParam 设置参数
func (t *Type) SetParam(ctx context.Context, key string, value string) error {
	// 设置公共属性
	err := t.CommonKeywords(key, value)
	if err != nil {
		return err
	}
	// 设置OneOf属性
	err = t.OneOfKeywords(key, value)
	if err != nil {
		return err
	}
	// 设置自定义属性
	err = t.CustomKeywords(key, value)
	if err != nil {
		return err
	}
	// 设置定义
	if strings.HasPrefix(key, "definitions_") {
		err = t.DefinitionsKeywords(ctx, strings.TrimPrefix(key, "definitions_"), value)
		if err != nil {
			return err
		}
	}
	switch t.Type {
	case JSONSchemaObject:

	case JSONSchemaString:
		// 根据String类型设置属性
		err := t.StringKeywords(key, value)
		if err != nil {
			return err
		}
	case JSONSchemaNumber:
		// 根据Number类型设置属性
		err := t.NumberKeywords(key, value)
		if err != nil {
			return err
		}
	case JSONSchemaInteger:
		// 根据Integer类型设置属性
		err := t.NumberKeywords(key, value)
		if err != nil {
			return err
		}
	case JSONSchemaBoolean:
	case JSONSchemaArray:
		// 根据Array类型设置属性
		err := t.ArrayKeywords(key, value)
		if err != nil {
			return err
		}
	}
	return nil
}

// ParseTag 解析Tag
func (t *Type) ParseTag(ctx context.Context, fieldType JSONSchemaDataType, fieldName string, fieldValue string, tagStr string) error {
	ty := &Type{}
	ty.Type = fieldType
	if tagStr == "" {
		// 验证Object中是否存在当前字段的属性
		if _, ok := t.Properties.Get(fieldName); ok {
			return fmt.Errorf("当前属性名[%s]重复，请检查后再试", fieldName)
		}
		t.Properties.Set(fieldName, ty)
		return nil
	}
	// 当Tag不为空的时候，根据逗号切割,获取Tag数组
	tagList := strings.Split(strings.Trim(tagStr, "\""), ",")
	oneOfList := []string{}
	definitions := []string{}
	for _, tag := range tagList {
		var k, v string
		if tag == "required" {
			t.Required = append(t.Required, fieldName)
			continue
		} else if strings.HasPrefix(tag, "oneof_") {
			oneOfList = append(oneOfList, tag)
			continue
		} else if strings.HasPrefix(tag, "definitions_") {
			definitions = append(definitions, tag)
			continue
		} else if strings.HasPrefix(tag, "ref_func_option") {
			newTag := strings.TrimPrefix(tag, "ref_func_option=#/")
			directoryList := strings.Split(newTag, "/")
			if len(directoryList) != 2 {
				return fmt.Errorf("目前只支持二级目录")
			}
			if _, ok := optionCallbackMap[directoryList[1]]; !ok {
				return fmt.Errorf("option callback: not find the function[%s]", directoryList[1])
			}
			fn := optionCallbackMap[directoryList[1]]
			definitionsType := &Type{}
			optionList, err := fn(ctx)
			if err != nil {
				return err
			}
			for _, option := range optionList {
				defOneOf := &Type{}
				defOneOf.Title = option.Key
				defOneOf.Enum = []string{option.Value}
				definitionsType.OneOf = append(definitionsType.OneOf, defOneOf)
			}
			t.Definitions[ToSnakeCase(directoryList[1])] = definitionsType
			kv := strings.Split(tag, "=")
			k = "ref"
			if len(kv) > 1 {
				v = kv[1]
			}
		} else if strings.HasPrefix(tag, "default_auto") {
			if fieldValue != nullDefaultValue {
				k = "default"
				v = fieldValue
			} else {
				continue
			}
		} else {
			kv := strings.Split(tag, "=")
			if len(kv) > 1 {
				k = kv[0]
				v = kv[1]
			}
		}
		err := ty.SetParam(ctx, k, v)
		if err != nil {
			return err
		}
	}
	// 设置OneOf属性
	if len(oneOfList) > 0 {
		oneOfType := &Type{}
		for _, oneOf := range oneOfList {
			kv := strings.Split(oneOf, "=")
			if len(kv) > 1 {
				err := oneOfType.SetParam(ctx, kv[0], kv[1])
				if err != nil {
					return err
				}
			}

		}
		t.OneOf = append(t.OneOf, oneOfType)
	}
	// 设置Definitions属性
	if len(definitions) > 0 {
		for _, definition := range definitions {
			definitionsType := &Type{}
			kv := strings.Split(definition, "=")
			if len(kv) > 1 {
				err := definitionsType.SetParam(ctx, kv[0], kv[1])
				if err != nil {
					return err
				}
			}
			// 去除[definitions_]前缀
			newDefinition := strings.TrimPrefix(definition, "definitions_")
			// 找到第一个下划线[_]的索引
			index := strings.Index(newDefinition, "_")
			// 根据索引获取字段名
			fieldName := newDefinition[0:index]
			t.Definitions[ToSnakeCase(fieldName)] = definitionsType
		}
	}
	// 验证Object中是否存在当前字段的属性
	if _, ok := t.Properties.Get(fieldName); ok {
		return fmt.Errorf("当前属性名[%s]重复，请检查后再试", fieldName)
	}
	t.Properties.Set(fieldName, ty)
	return nil
}

// CommonKeywords 公共属性
func (t *Type) CommonKeywords(key string, value string) error {
	switch key {
	case "title":
		t.Title = value
	case "ref":
		t.Ref = value
	// case "definitions":
	// 	tag.Definitions, err = GetDefinitions(value)
	// 	if err != nil {
	// 		return err
	// 	}
	case "description":
		t.Description = value
	case "format":
		switch value {
		case "date", "date-time", "email", "hostname", "ipv4", "ipv6", "uri":
			t.Format = value
		default:
			return fmt.Errorf("格式化参数填写错误：%s", value)
		}
	case "readOnly":
		is, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		t.ReadOnly = is
	}
	return nil
}

// CustomKeywords 自定义属性
func (t *Type) CustomKeywords(key string, value string) error {
	rt := reflect.TypeOf(t.CustomFields)
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		// 自动映射字符串字段
		if f.Type.Kind() != reflect.String {
			continue
		}
		jsonTag := f.Tag.Get("json")
		jsonTags := strings.Split(jsonTag, ",")
		for _, jt := range jsonTags {
			if jt == key {
				vf := reflect.ValueOf(&t.CustomFields).Elem().FieldByName(f.Name)
				vf.SetString(value)
			}
		}
	}
	return nil
}

// DefinitionsKeywords 定义
func (t *Type) DefinitionsKeywords(ctx context.Context, key string, value string) error {
	index := strings.Index(key, "_")
	fieldKey := key[index+1:]
	if strings.HasPrefix(fieldKey, "oneof_function") {
		if fn, ok := optionCallbackMap[value]; ok {
			optionList, err := fn(ctx)
			if err != nil {
				return err
			}
			for _, option := range optionList {
				defOneOf := &Type{}
				defOneOf.Title = option.Key
				defOneOf.Enum = []string{option.Value}
				t.OneOf = append(t.OneOf, defOneOf)
			}
			return nil
		}
		return fmt.Errorf("option callback: not find the function[%s]", value)
	}
	if strings.HasPrefix(fieldKey, "oneof_") {
		defOneOf := &Type{}
		defOneOf.OneOfKeywords(fieldKey, value)
		t.OneOf = append(t.OneOf, defOneOf)
	}
	return nil
}

// OneOfKeywords OneOf属性
func (t *Type) OneOfKeywords(key string, value string) error {
	switch key {
	case "oneof_required":
		if value != "" {
			t.Required = strings.Split(value, ":")
		}
	case "oneof_title":
		t.Title = value
	case "oneof_enum":
		if value != "" {
			t.Enum = strings.Split(value, ":")
		}
	}
	return nil
}

// StringKeywords 字符串属性
func (t *Type) StringKeywords(key string, value string) error {
	switch key {
	case "minLength":
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		t.MinLength = i
	case "maxLength":
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		t.MaxLength = i
	case "pattern":
		t.Pattern = value
	// case "format":
	// 	switch value {
	// 	case "date-time", "email", "hostname", "ipv4", "ipv6", "uri":
	// 		t.Format = value
	// 	default:
	// 		return fmt.Errorf("格式化参数填写错误：%s", value)
	// 	}
	case "default":
		t.Default = value
	case "example":
		// t.Examples = append(t.Examples, value)
	case "enum":
		if value != "" {
			t.Enum = strings.Split(value, ":")
		}
	}
	return nil
}

// NumberKeywords 数字属性
func (t *Type) NumberKeywords(key string, value string) error {
	switch key {
	case "multipleOf":
		i, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		t.MultipleOf = i
	case "minimum":
		i, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		t.Minimum = i
	case "maximum":
		i, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		t.Maximum = i
	case "exclusiveMaximum":
		b, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		t.ExclusiveMaximum = b
	case "exclusiveMinimum":
		b, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		t.ExclusiveMinimum = b
	case "default":
		if value == "" {
			value = "0"
		}
		if strings.Contains(value, ".") {
			f, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return err
			}
			t.Default = f
		} else {
			i, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return err
			}
			t.Default = i
		}
	case "example":
		// if i, err := strconv.Atoi(value); err == nil {
		// 	t.Examples = append(t.Examples, i)
		// }
	case "enum":
		if value != "" {
			strEnumList := strings.Split(value, ":")
			newEnumList := []float64{}
			for _, strEnum := range strEnumList {
				newEnum, err := strconv.ParseFloat(strEnum, 64)
				if err != nil {
					return err
				}
				newEnumList = append(newEnumList, newEnum)
			}
			t.Enum = newEnumList
		}
	}
	return nil
}

// ArrayKeywords 数组属性
func (t *Type) ArrayKeywords(key string, value string) error {
	switch key {
	case "minItems":
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		t.MinItems = i
	case "maxItems":
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		t.MaxItems = i
	case "uniqueItems":
		t.UniqueItems = true
	case "default":
		// defaultValues = append(defaultValues, val)
	}
	return nil
}
