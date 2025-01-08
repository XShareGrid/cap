package idgen

import "github.com/satori/go.uuid"

// uuidGeneratorV4 generator of UUID RFC4722 version4
type uuidGeneratorV4 struct {
}

// NewUUIDGeneratorV4 constructor of uuidGeneratorV4
func NewUUIDGeneratorV4() *uuidGeneratorV4 {
	return &uuidGeneratorV4{}
}

// Generate generate an uuid
// return is the uuid generated
func (g *uuidGeneratorV4) Generate(param ...interface{}) (string, error) {
	u := uuid.NewV4()
	return u.String(), nil
}
