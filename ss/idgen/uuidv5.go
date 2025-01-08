package idgen

import (
	"errors"

	"github.com/satori/go.uuid"
)

// uuidGeneratorV5 generator of UUID RFC4722 version4
type uuidGeneratorV5 struct {
}

// NewUUIDGeneratorV5 constructor of uuidGeneratorV5
func NewUUIDGeneratorV5() *uuidGeneratorV5 {
	return &uuidGeneratorV5{}
}

// Generate generate an uuid
// return is the uuid generated
func (g *uuidGeneratorV5) Generate(param ...interface{}) (string, error) {
	if len(param) != 2 {
		return "", errors.New("invalid param")
	}
	nsStr, ok := param[0].(string)
	if !ok {
		return "", errors.New("invalid param")
	}
	name, ok := param[1].(string)
	if !ok {
		return "", errors.New("invalid param")
	}
	var ns uuid.UUID
	ns.UnmarshalText([]byte(nsStr))
	return uuid.NewV5(ns, name).String(), nil
}
