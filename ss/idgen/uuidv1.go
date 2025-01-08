package idgen

import "github.com/satori/go.uuid"

// uuidGeneratorV1 generator of UUID with RFC4122 version1
type uuidGeneratorV1 struct {
}

// NewUUIDGeneratorV1 constructor of uuidGeneratorV1
func NewUUIDGeneratorV1() *uuidGeneratorV1 {
	return &uuidGeneratorV1{}
}

// Generate generate an uuid v4
// return is the uuid generated
func (g *uuidGeneratorV1) Generate(param ...interface{}) (string, error) {
	u := uuid.NewV1()
	return u.String(), nil
}
