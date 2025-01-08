package registry

import (
	"reflect"
	"regexp"
	"runtime"
	"strings"

	"github.com/XShareGrid/cap/msg/errors"
	cap "github.com/XShareGrid/cap/table/proto/go"
)

// OperatorNode ...
type OperatorNode interface {
	ID() string
	Name() string
}

// Operator 操作符
type Operator interface {
	OperatorNode
	FilterValueType() cap.FilterValueType
}

// OperatorSet a set of operators
type OperatorSet interface {
	OperatorNode
	Operators() []Operator
}

// OperatorReg ValueFilterFunction registry
type OperatorReg struct {
	*registryContainer
}

func verifyIDFmt(id string) error {
	m, err := regexp.MatchString(`(\w+)\.(\w+)`, id)
	if !m || err != nil {
		return errors.Wrap(ErrInvalidOperatorIDFormat).FillDebugArgs(id).Log()
	}
	seg := strings.Split(id, ".")
	if len(seg) != 2 {
		return errors.Wrap(ErrInvalidOperatorIDFormat).FillDebugArgs(id).Log()
	}
	// namespace := seg[0]
	pc, _, _, _ := runtime.Caller(3)
	caller := runtime.FuncForPC(pc).Name()
	namespace := seg[0]
	// builtin namespace reserved
	if namespace == "builtin" {
		if caller != "github.com/XShareGrid/cap/table/operator/builtin.registerBuiltinOperator" &&
			caller != "github.com/XShareGrid/cap/table/operator/builtin.registerBuiltinOperatorSet" {
			return errors.Wrap(ErrBuiltinNamespaceNowAllowed).Log()
		}
	}
	return nil
}

// Register registers function
func (vr *OperatorReg) Register(ops ...OperatorNode) error {
	for _, op := range ops {
		if err := verifyIDFmt(op.ID()); err != nil {
			return err
		}
		err := vr.registryContainer.Register(op)
		if err != nil {
			return err
		}
	}
	return nil
}

// Find find function with id
// for operator set, returns an array
func (vr *OperatorReg) Find(id string) (OperatorNode, error) {
	v, err := vr.registryContainer.Find(id)
	if err != nil {
		return nil, err
	}
	return v.(OperatorNode), nil
}

func (vr *OperatorReg) FindAll() []OperatorNode {
	all := []OperatorNode{}
	nodes := vr.registryContainer.FindAll()
	for _, node := range nodes {
		all = append(all, node.(OperatorNode))
	}
	return all
}

// NewOperatorReg creates value filter function registry
func NewOperatorReg() *OperatorReg {
	return &OperatorReg{registryContainer: newRegistryContainer(reflect.TypeOf((*OperatorNode)(nil)).Elem(), []string{"Name"})}
}
