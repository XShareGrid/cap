package errors

import (
	"fmt"
	"log"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/XShareGrid/cap/msg"
	"github.com/XShareGrid/cap/msg/i18n"
	"golang.org/x/text/language"

	goerror "errors"
)

// UnknownErrorID error id for unknown errors
var UnknownErrorID i18n.TrID = "XXX_UNKNOWN_ERROR_ID"

// UserError struct to define error
type UserError struct {
	error
	// UID is only available for pre-define errors
	UID       uint
	debugArgs []interface{}
	next      error
	msg.Msg
	pc       []uintptr
	isPreDef bool
}

// GoError get go error
func (e *UserError) GoError() error {
	return e.error
}

// Error output debug string
func (e *UserError) Error() string {
	if e.error == nil {
		return ""
	}
	if len(e.debugArgs) == 0 {
		return e.error.Error()
	}
	return fmt.Sprintf(e.error.Error(), e.debugArgs...)
}

// FillDebugArgs fills debug arguments
func (e *UserError) FillDebugArgs(args ...interface{}) *UserError {
	if e.isPreDef {
		panic("pre-define error is not operatable, make a copy with Clone/Wrap interface")
	}
	e.debugArgs = args
	return e
}

// FillIDAndArgs fills id and args
func (e *UserError) FillIDAndArgs(id i18n.TrID, args ...interface{}) *UserError {
	if e.isPreDef {
		panic("pre-define error is not operatable, make a copy with Clone/Wrap interface")
	}
	e.ID = id
	e.Args = args
	return e
}

// DegbugArg returns the debug arg with the index, return null string if index out of range
func (e *UserError) DegbugArg(index int) interface{} {
	if index >= 0 && index < len(e.debugArgs) {
		return e.debugArgs[index]
	}
	return ""
}

// DegbugArgs returns all debug arguments
func (e *UserError) DegbugArgs() []interface{} {
	return e.debugArgs
}

// Log output debug string
func (e *UserError) Log() *UserError {
	localLogger(e.Error())
	return e
}

// Append appends error
// panics if cycle append happen.. MUST be solved before release...
// returns the last error append
func (e *UserError) append(errs ...error) *UserError {
	circleChecklist := make(map[error]interface{})
	errCount := len(errs)
	if errCount == 0 {
		return e
	}
	head := Wrap(errs[errCount-1])
	cursor := head
	moveToLast := func(userErr *UserError) *UserError {
		if _, ok := circleChecklist[userErr]; ok {
			panic(fmt.Errorf("circle point with the error list: %s", userErr))
		}
		if _, ok := circleChecklist[userErr.error]; ok {
			panic(fmt.Errorf("circle point with the error list: %s", userErr))
		}
		circleChecklist[userErr] = nil
		circleChecklist[userErr.error] = nil
		for userErr.next != nil {
			userErr = Wrap(userErr.next)
			if _, ok := circleChecklist[userErr]; ok {
				panic(fmt.Errorf("circle point with the error list: %s", userErr))
			}
			if _, ok := circleChecklist[userErr.error]; ok {
				panic(fmt.Errorf("circle point with the error list: %s", userErr))
			}
			circleChecklist[userErr] = nil
			circleChecklist[userErr.error] = nil
		}
		return userErr
	}

	cursor = moveToLast(cursor)

	if errCount > 1 {
		for i := errCount - 2; i > 0; i-- {
			cursor.next = errs[i]
			cursor = moveToLast(cursor)
		}
	}
	// check the last list
	moveToLast(e)
	cursor.next = e

	return head
}

// Triggers a new error
// returns the error triggered
func (e *UserError) Triggers(err ...error) *UserError {
	if e.isPreDef {
		panic("pre-define error is not operatable, make a copy with Clone/Wrap interface")
	}
	return e.append(err...)
}

// ErrorDump error list
type ErrorDump []error

// Log logs errors by order
func (ed ErrorDump) Log() {
	count := len(ed)
	for i, e := range ed {
		localLogger("[%d] %v", i, e)
		if i < count-1 {
			localLogger("triggered by:")
		}
	}
}

func (ed ErrorDump) Len() int {
	return len(ed)
}

func (ed ErrorDump) Less(i, j int) bool {
	return i < j
}

// Swap swaps the elements with indexes i and j.
func (ed ErrorDump) Swap(i, j int) {
	tmp := ed[j]
	ed[j] = ed[i]
	ed[i] = tmp
}

// DumpErrors dump all errors, includes current error
func (e *UserError) DumpErrors() ErrorDump {
	d := ErrorDump{e}
	for e.next != nil {
		d = append(d, e.next)
		if nextE, ok := e.next.(*UserError); ok {
			e = nextE
		} else {
			break
		}
	}
	return d
}

// MatchAll matches the error with the error dump
func (e *UserError) MatchAll(err error) bool {
	ue := e
	for ue != nil {
		if ue.Match(err) {
			return true
		}
		if ue.next != nil {
			ue = ue.next.(*UserError)
		} else {
			ue = nil
		}
	}
	return false
}

// Match matches the error with the error
func (e *UserError) Match(err error) bool {
	if ue, ok := err.(*UserError); ok {
		if ue.UID == e.UID {
			return true
		}
		return e == ue || e.error == ue || e == ue.error || e.error == ue.error
	}
	return e == err || e.error == err
}

// TrError Translated Error
func (e *UserError) TrError(lang language.Tag) (string, error) {
	if e.ID == UnknownErrorID {
		return e.Error(), nil
	}
	return e.Msg.GetMessage(lang)
}

// PrintStackTrace prints stack trace
func (e *UserError) PrintStackTrace() *UserError {
	localLogger("error time: %s, stack trace:", e.Timestamp)
	pc, _, _, _ := runtime.Caller(1)
	f := runtime.FuncForPC(pc)
	file, line := f.FileLine(pc)
	localLogger("at %s:%d#%s", file, line, f.Name())
	for _, p := range e.pc {
		if p == 0 {
			return e
		}
		f := runtime.FuncForPC(p)
		file, line := f.FileLine(p)
		localLogger("at %s:%d#%s", file, line, f.Name())
	}
	return e
}

// GetStackTrace gets stack trace
func (e *UserError) GetStackTrace() (st []string) {
	st = []string{}
	for _, p := range e.pc {
		if p == 0 {
			return st
		}
		f := runtime.FuncForPC(p)
		file, line := f.FileLine(p)
		st = append(st, fmt.Sprintf("at %s:%d#%s", file, line, f.Name()))
	}
	return st
}

// GetPos gets stack trace
func (e *UserError) GetPos() (file string, line int, funcName string) {
	if len(e.pc) == 0 {
		return "", 0, ""
	}
	pc := e.pc[0]

	if pc == 0 {
		return "", 0, ""
	}
	f := runtime.FuncForPC(pc)
	file, line = f.FileLine(pc)
	return file, line, f.Name()

}

// NewUserD new user error with debug string
// debug string will not be displayed to TrError
func NewUserD(debugString string, errID i18n.TrID, args ...interface{}) error {
	return &UserError{
		error: goerror.New(debugString),
		Msg: msg.Msg{
			Timestamp: getTimeStamp(),
			ID:        errID,
			Args:      args},
		pc: callers(),
	}
}

// NewUser new error with debug string
func NewUser(errID i18n.TrID, args ...interface{}) error {
	return &UserError{
		error: goerror.New(""),
		Msg: msg.Msg{
			Timestamp: getTimeStamp(),
			ID:        errID,
			Args:      args,
		},
		pc: callers(),
	}
}

// Newf new error with fmt
func Newf(format string, args ...interface{}) error {
	pc, _, _, _ := runtime.Caller(1)
	callerFunc := runtime.FuncForPC(pc)
	if callerFunc != nil && strings.HasSuffix(callerFunc.Name(), ".init") {
		return PreDef(fmt.Sprintf(format, args...))
	}
	return &UserError{
		error: fmt.Errorf(format, args...),
		Msg: msg.Msg{
			Timestamp: getTimeStamp(),
			ID:        UnknownErrorID,
		},
		pc: callers(),
	}
}

// New new error with debug string
func New(debugString string) error {
	pc, _, _, _ := runtime.Caller(1)
	callerFunc := runtime.FuncForPC(pc)
	if callerFunc != nil && strings.HasSuffix(callerFunc.Name(), ".init") {
		return PreDef(debugString)
	}
	return &UserError{
		error: goerror.New(debugString),
		Msg: msg.Msg{
			Timestamp: getTimeStamp(),
			ID:        UnknownErrorID,
		},
		pc: callers(),
	}
}

// Clone clones a user error, only single level, not deep clone
func (e *UserError) Clone() *UserError {
	return &UserError{
		error:     goerror.New(e.Error()),
		UID:       e.UID,
		debugArgs: e.DegbugArgs(),
		Msg:       e.Msg,
		pc:        e.pc,
	}
}

var uidSeed uint
var uidMutex sync.Mutex

// InvalidErrorUID invalid error uid
const InvalidErrorUID uint = 0

func genUID() uint {
	uidMutex.Lock()
	defer uidMutex.Unlock()
	uidSeed++
	return uidSeed
}

// PreDef pre-define errors
func PreDef(text string) error {
	return &UserError{
		UID:   genUID(),
		error: goerror.New(text),
		Msg: msg.Msg{
			Timestamp: getTimeStamp(),
			ID:        UnknownErrorID,
		},
		isPreDef: true,
	}
}

// Wrap wrap go error with user error
func Wrap(err error) *UserError {
	if err == nil {
		panic("err is nil")
	}
	if ue, ok := err.(*UserError); ok {
		if ue.isPreDef {
			ue = ue.Clone()
		}
		if len(ue.pc) == 0 {
			ue.pc = callers()
		}
		return ue
	}
	return &UserError{
		error: err,
		Msg: msg.Msg{
			Timestamp: getTimeStamp(),
		},
		pc: callers(),
	}
}

// MatchAll compares two errors with the whole error dump
func MatchAll(err0 error, err1 error) bool {
	if ue0, ok := err0.(*UserError); ok {
		return ue0.MatchAll(err1)
	}
	if ue1, ok := err1.(*UserError); ok {
		return ue1.MatchAll(err0)
	}
	return err0 == err1
}

// Match compares two errors
func Match(err0 error, err1 error) bool {
	if ue0, ok := err0.(*UserError); ok {
		return ue0.Match(err1)
	}
	if ue1, ok := err1.(*UserError); ok {
		return ue1.Match(err0)
	}
	return err0 == err1
}

func getTimeStamp() time.Time {
	return time.Now()
}

var localLogger ErrorLogger = func(format string, args ...interface{}) {
	log.Printf(format, args...)
}

// SetLogger set logger for error
func SetLogger(l ErrorLogger) {
	localLogger = l
}

// ErrorLogger for error
type ErrorLogger func(v string, args ...interface{})

func callers() []uintptr {
	pc := make([]uintptr, 32)
	runtime.Callers(3, pc)
	return pc
}
