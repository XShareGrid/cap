package jsonschema

import (
	"github.com/iancoleman/orderedmap"
)

// CustomFields ...
type CustomFields struct {
	UIFilterable string `json:"ui:filterable,omitempty"`
	UIHidden     string `json:"ui:hidden,omitempty"`
}

// Type 数据类型
type Type struct {
	CustomFields
	MinLength        int64                  `json:"minLength,omitempty"`        // 规定值的长度必须大于等于该项
	MaxLength        int64                  `json:"maxLength,omitempty"`        // 规定值的长度必须小于等于该项
	Pattern          string                 `json:"pattern,omitempty"`          // 正则表达式，规定值必须匹配该项
	Enum             interface{}            `json:"enum,omitempty"`             // 枚举值，即值只能是enum数组中的某一项
	Default          interface{}            `json:"default,omitempty"`          // 默认值
	MultipleOf       float64                `json:"multipleOf,omitempty"`       // 规定值必须为该项的倍数
	Minimum          float64                `json:"minimum,omitempty"`          // 规定值必须大于等于该项
	ExclusiveMinimum float64                `json:"exclusiveMinimum,omitempty"` // 规定值就必须大于minimum
	Maximum          float64                `json:"maximum,omitempty"`          // 规定值必须小于等于该项
	ExclusiveMaximum float64                `json:"exclusiveMaximum,omitempty"` // 规定值就必须小于maximum
	Items            interface{}            `json:"items,omitempty"`            // 子元素可能是对象,也可能是数组
	AdditionalItems  *Type                  `json:"additionalItems,omitempty"`  // section 5.9
	MinItems         int64                  `json:"minitems,omitempty"`         // 规定子元素数量必须大于等于该项
	MaxItems         int64                  `json:"maxitems,omitempty"`         // 规定子元素数量必须小于等于该项
	UniqueItems      bool                   `json:"uniqueItems,omitempty"`      // 每个元素都不相同,唯一
	Required         []string               `json:"required,omitempty"`         // 规定object下哪些键是必须的
	Properties       *orderedmap.OrderedMap `json:"properties,omitempty"`       //
	Type             JSONSchemaDataType     `json:"type,omitempty"`             // 规定值的类型
	Title            string                 `json:"title,omitempty"`            // 标题
	Ref              string                 `json:"$ref,omitempty"`             // 用来引用其他的schema
	Description      string                 `json:"description,omitempty"`      // 描述信息
	Schema           string                 `json:"$schema,omitempty"`          // JSONSchema文件遵守的规范
	Definitions      map[string]*Type       `json:"definitions,omitempty"`      // 创建内部结构体,定义公共约束
	AllOf            []*Type                `json:"allOf,omitempty"`            // section 5.22
	AnyOf            []*Type                `json:"anyOf,omitempty"`            // section 5.23
	OneOf            []*Type                `json:"oneOf,omitempty"`            // section 5.24
	Not              *Type                  `json:"not,omitempty"`              // section 5.25
	Format           string                 `json:"format,omitempty"`           // section 7
	ReadOnly         bool                   `json:"readOnly,omitempty"`         // section 7
	Const            interface{}            `json:"const,omitempty"`            // http://json-schema.org/understanding-json-schema/reference/generic.html#constant-values
}
