package idgen

import (
	"errors"
)

// uuidGeneratorV5Recurrent
type uuidGeneratorV5Recurrent struct {
}

// NewUUIDGeneratorV5Recurrent constructor of uuidGeneratorV5
func NewUUIDGeneratorV5Recurrent() *uuidGeneratorV5Recurrent {
	return &uuidGeneratorV5Recurrent{}
}

// Generate generate an uuid
// return is the uuid generated
func (g *uuidGeneratorV5Recurrent) Generate(params ...interface{}) (string, error) {
	if len(params) == 0 {
		return "", errors.New("invalid parameters")
	}
	gen := NewUUIDGeneratorV5()
	baseKey := params[0].(string)
	var err error
	for _, key := range params[1:] {
		baseKey, err = gen.Generate(baseKey, key)
		if err != nil {
			return "", err
		}
	}
	return baseKey, nil
}
