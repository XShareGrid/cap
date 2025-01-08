package errors

import (
	"fmt"
	"testing"
)

func TestNew(t *testing.T) {
	err := New("test debug string")
	err.(*UserError).PrintStackTrace()
}

func Test_callers(t *testing.T) {
	fmt.Println(callers())
}

func TestUserError_Trigger(t *testing.T) {
	err0 := New("err0").(*UserError)
	err1 := New("err1").(*UserError)
	err2 := fmt.Errorf("err2")
	// err3 := err0.Triggers(err1, err2, err1)
	err3 := err0.Triggers(err1).Triggers(err2)
	Wrap(err3).DumpErrors().Log()
}

var errTest = New("test")
var errTest1 = New("test")

func TestPrintError(t *testing.T) {
	fmt.Println(Wrap(errTest).UID)
	fmt.Println(Wrap(errTest1).UID)
	Wrap(errTest).PrintStackTrace()
}
