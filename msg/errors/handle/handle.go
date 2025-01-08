package handle

import (
	"context"
	"sync"

	"github.com/XShareGrid/cap/msg/errors"
	i18n "github.com/XShareGrid/cap/msg/i18n"
)

// ErrorHandler handle error function
type ErrorHandler func(context.Context, *errors.UserError) *errors.UserError

var gReg *hdRegistry

func init() {
	gReg = &hdRegistry{
		reg: make(map[uint]ErrorHandler),
	}
}

type hdRegistry struct {
	reg map[uint]ErrorHandler
	m   sync.RWMutex
}

func (hd *hdRegistry) handle(ctx context.Context, err error, retIfNoMatch ...*errors.UserError) *errors.UserError {
	defaultErr := errors.Wrap(err).FillIDAndArgs(ERR_UNKNOWN, err.Error())
	if len(retIfNoMatch) > 0 {
		defaultErr = retIfNoMatch[0]
	}
	ue, ok := err.(*errors.UserError)
	if !ok || ue.UID == errors.InvalidErrorUID {
		return defaultErr
	}
	hd.m.RLock()
	defer hd.m.RUnlock()
	if handler, ok := hd.reg[ue.UID]; ok {
		return handler(ctx, ue)
	}
	return defaultErr
}

func (hd *hdRegistry) register(err error, handler ErrorHandler) {
	hd.m.Lock()
	defer hd.m.Unlock()
	ue, ok := err.(*errors.UserError)
	if !ok || ue.UID == errors.InvalidErrorUID {
		panic("only pre-define error is registerable")
	}
	hd.reg[ue.UID] = handler
}

// DefaultHandler 默认处理，直接填入语言ID和调试参数
func DefaultHandler(trID i18n.TrID) ErrorHandler {
	return func(ctx context.Context, err *errors.UserError) *errors.UserError {
		return err.Clone().FillIDAndArgs(trID, err.DegbugArgs()...)
	}
}

// RegisterDefaultHandler 注册一个默认转换的处理器，直接填入语言ID和调试参数
func RegisterDefaultHandler(err error, trID i18n.TrID) {
	Register(err, DefaultHandler(trID))
}

// Register register handler
func Register(err error, handler ErrorHandler) {
	gReg.register(err, handler)
}

// Handle handles errors
// ctx context
// err 需要转换的error
// retIfNoMatch 如果没有匹配的转换器，默认返回的错误
func Handle(ctx context.Context, err error, retIfNoMatch ...*errors.UserError) *errors.UserError {
	return gReg.handle(ctx, err)
}
