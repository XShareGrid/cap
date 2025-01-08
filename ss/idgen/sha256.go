package idgen

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

// sha256Generator generator of UUID RFC4722 version4
type sha256Generator struct {
}

// NewSHA256Generator constructor of sha256Generator
func NewSHA256Generator() *sha256Generator {
	return &sha256Generator{}
}

// Generate generate an uuid
// return is the uuid generated
func (g *sha256Generator) Generate(param ...interface{}) (string, error) {
	if len(param) != 1 {
		return "", errors.New("invalid param")
	}
	src, ok := param[0].(string)
	if !ok {
		return "", errors.New("invalid param")
	}
	hasher := sha256.New()
	hasher.Write([]byte(src))
	return hex.EncodeToString(hasher.Sum(nil)), nil
}
