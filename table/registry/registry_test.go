package registry

import (
	"fmt"
	"testing"
)

type TestOperation struct {
	test int
}

func (TestOperation) ID() string {
	return "123"
}

func (t TestOperation) Name() string {
	return fmt.Sprintf("%d", t.test)
}

func Test_matchValue(t *testing.T) {
	fmt.Println(matchValue(&TestOperation{test: 1}, &TestOperation{}, "Name"))
}
