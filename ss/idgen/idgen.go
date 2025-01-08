//copyright

/*
Package idgen defines the interface for generating a id string and
implemented some generators
*/
package idgen

// IDGenerator interface for generate id
type IDGenerator interface {
	Generate(param ...interface{}) (string, error)
}
