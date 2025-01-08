package registry

import (
	"log"
	"reflect"
	"sync"

	"github.com/XShareGrid/cap/msg/errors"
)

// TableRegistry table registry
type TableRegistry struct {
	OperatorReg  *OperatorReg
	TableMetaReg *TableMetaReg
	OptionReg    *OptionReg
}

var theTableRegistry *TableRegistry

func init() {
	theTableRegistry = &TableRegistry{
		OperatorReg:  NewOperatorReg(),
		TableMetaReg: NewTableMetaReg(),
		OptionReg:    NewOptionReg(),
	}
}

// GlobalTableRegistry table registry
func GlobalTableRegistry() *TableRegistry {
	return theTableRegistry
}

type registryNode interface {
	ID() string
	Name() string
}

type registryContainer struct {
	nodeType     reflect.Type
	uniqueFields []string
	reg          sync.Map
}

func matchValue(objSrc interface{}, objDst interface{}, valueFunction string) (match bool) {
	tSrc := reflect.ValueOf(objSrc)
	sf := tSrc.MethodByName(valueFunction)
	tDst := reflect.ValueOf(objDst)
	df := tDst.MethodByName(valueFunction)
	if sf.IsZero() || df.IsZero() {
		return false
	}
	return sf.Call([]reflect.Value{})[0].Interface() == df.Call([]reflect.Value{})[0].Interface()
}

// Register registers function
func (rc *registryContainer) Register(node registryNode) error {
	if reflect.TypeOf(node).Elem().Implements(rc.nodeType) {
		log.Fatalf("registry node not match (%s) != (%s)", reflect.TypeOf(node).Name(), rc.nodeType.Name())
	}
	var err error
	// check
	rc.reg.Range(func(key, value interface{}) bool {
		if node.ID() == key.(string) {
			err = errors.Wrap(ErrDuplicateNodeID).FillDebugArgs(node.ID()).Log()
			return false
		}
		for _, uf := range rc.uniqueFields {
			if matchValue(value, node, uf) {
				err = errors.Wrap(ErrDuplicateNodeFiled).FillDebugArgs(uf).Log()
				return false
			}
		}
		return true
	})

	if err != nil {
		return err
	}
	rc.reg.Store(node.ID(), node)
	return nil
}

// Find find function with id
func (rc *registryContainer) Find(id string) (registryNode, error) {
	vff, ok := rc.reg.Load(id)
	if ok {
		return vff.(registryNode), nil
	}
	return nil, errors.Wrap(ErrNodeNotExist).FillDebugArgs(id)
}

// Find find function with id
func (rc *registryContainer) FindAll() []registryNode {
	nodes := []registryNode{}
	rc.reg.Range(func(key, value interface{}) bool {
		nodes = append(nodes, value.(registryNode))
		return true
	})
	return nodes
}

// Find find function with id
func (rc *registryContainer) List() {
	rc.reg.Range(func(key, value interface{}) bool {
		log.Println("Registered:", key, "-", value.(registryNode).Name())
		return true
	})

}

// newRegistryContainer 创建注册表容器
// nodeType: 节点类型interface{}
// uniqueFields: 唯一字段(方法名)
func newRegistryContainer(nodeType reflect.Type, uniqueFields []string) *registryContainer {
	return &registryContainer{nodeType: nodeType, uniqueFields: uniqueFields}
}

func (rc *registryContainer) Store() *sync.Map {
	return &rc.reg
}
