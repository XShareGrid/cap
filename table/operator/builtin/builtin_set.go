package builtin

import (
	"github.com/XShareGrid/cap/table/registry"
)

// OperatorSet operator set for int
type OperatorSet struct {
	id   string
	name string
	ops  []registry.Operator
}

// Operators get operators
func (os *OperatorSet) Operators() []registry.Operator {
	return os.ops
}

// ID get id
func (os *OperatorSet) ID() string {
	return os.id
}

// Name get name
func (os *OperatorSet) Name() string {
	return os.name
}

func registerBuiltinOperatorSet(id, name string, ops ...registry.Operator) *OperatorSet {
	id = "builtin." + id
	return RegisterOperatorSet(id, name, ops...)
}

// RegisterOperatorSet creates an operator set and register it
func RegisterOperatorSet(id, name string, ops ...registry.Operator) *OperatorSet {
	os := &OperatorSet{id: id, name: name, ops: ops}
	err := registry.GlobalTableRegistry().OperatorReg.Register(os)
	if err != nil {
		panic(err)
	}
	return os
}

// SINT set for int fields
var SINT = registerBuiltinOperatorSet("SINT", "Set of Int", EQ, GT, LT, GE, LE, NE, IN, NIN)

// SBOOL set for bool fields
var SBOOL = registerBuiltinOperatorSet("SBOOL", "Set of Boolean", EQ)

// SOPT set for option fields
var SOPT = registerBuiltinOperatorSet("SOPT", "Set of Options", EQ, NE, IN, NIN)

// SDOUBLE set for int fields
var SDOUBLE = registerBuiltinOperatorSet("SDOUBLE", "Set of Double", EQ, GT, LT, GE, LE, NE, IN, NIN)

// SSTR set for string fields
var SSTR = registerBuiltinOperatorSet("SSTR", "Set of String", LCTN, EQ, CTN, GT, LT, GE, LE, NE, RCTN, NCTN, IN, NIN)

// SNULLABLE set for nullable fields
var SNULLABLE = registerBuiltinOperatorSet("SNULLABLE", "Set of Nullable", ISN, ISNN)

// STIME set for time fields
var STIME = registerBuiltinOperatorSet("STIME", "Set of Time", EQ, GT, LT, GE, LE, NE, TODAY, YEST, TWEEK, L1MONTH, L3MONTH)

// SDATE set for date fields
var SDATE = registerBuiltinOperatorSet("SDATE", "Set of Date", EQ, GT, LT, GE, LE, NE, TODAY, YEST)
