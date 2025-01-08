package idgen

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
)

// md5Generator generator of UUID RFC4722 version4
type md5Generator struct {
}

// NewMD5Generator constructor of md5Generator
func NewMD5Generator() *md5Generator {
	return &md5Generator{}
}

// Generate generate an uuid
// return is the uuid generated
func (g *md5Generator) Generate(param ...interface{}) (string, error) {
	if len(param) != 1 {
		return "", errors.New("invalid param")
	}
	src, ok := param[0].(string)
	if !ok {
		return "", errors.New("invalid param")
	}
	hasher := md5.New()
	hasher.Write([]byte(src))
	return hex.EncodeToString(hasher.Sum(nil)), nil
}
