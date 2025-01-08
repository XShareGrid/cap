package jsonschema

import (
	"reflect"
)

// JSONSchemaDataType 数据类型
type JSONSchemaDataType string

const (
	JSONSchemaObject  JSONSchemaDataType = "object"
	JSONSchemaArray   JSONSchemaDataType = "array"
	JSONSchemaString  JSONSchemaDataType = "string"
	JSONSchemaNumber  JSONSchemaDataType = "number"
	JSONSchemaInteger JSONSchemaDataType = "integer"
	JSONSchemaBoolean JSONSchemaDataType = "boolean"
)

// GetJSONSchemaDataType 获取数据类型
func GetJSONSchemaDataType(t reflect.Type) JSONSchemaDataType {
	var dataType JSONSchemaDataType
	switch t.Kind() {
	case reflect.String:
		dataType = JSONSchemaString
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		dataType = JSONSchemaInteger
	case reflect.Float32, reflect.Float64:
		dataType = JSONSchemaNumber
	case reflect.Struct:
		dataType = JSONSchemaObject
	case reflect.Bool:
		dataType = JSONSchemaBoolean
	case reflect.Slice:
		dataType = JSONSchemaArray
	}
	return dataType
}
