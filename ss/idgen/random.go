package idgen

import (
	"crypto/rand"
	r "math/rand"
	"time"
)

// randomIDGenerator generate id with random characters
type randomIDGenerator struct {
	len       int
	alphabets []byte
}

// NewRandomIDGenerator create a randomIDGenerator
// len is the default length of id to be generated
// alphabets is the characters of []bytes used to generate the id
// if alphabets is not specified, defalt alphabets with be used
func NewRandomIDGenerator(len int, alphabets ...byte) *randomIDGenerator {
	if alphabets == nil {
		return &randomIDGenerator{len,
			[]byte("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")}
	}
	return &randomIDGenerator{len, alphabets}
}

// Generate generate a id
// param is the length of id to be generated
// if param is not specified, defalt length will be used
// return is the id generated
func (g *randomIDGenerator) Generate(param ...interface{}) (string, error) {
	var n int
	if len(param) == 1 {
		n = param[0].(int)
	} else {
		n = g.len
	}
	var bytes = make([]byte, n)
	var randBy bool
	if num, err := rand.Read(bytes); num != n || err != nil {
		r.Seed(time.Now().UnixNano())
		randBy = true
	}
	for i, b := range bytes {
		if randBy {
			bytes[i] = g.alphabets[r.Intn(len(g.alphabets))]
		} else {
			bytes[i] = g.alphabets[b%byte(len(g.alphabets))]
		}
	}
	return string(bytes), nil
}
