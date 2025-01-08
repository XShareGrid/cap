package mysql

import "github.com/XShareGrid/cap/msg/errors"

// ErrDeleteMustContainFilters ...
var ErrDeleteMustContainFilters = errors.New("delete operation MUST contain filters")

// ErrNoSessionInCtx no session in context
var ErrNoSessionInCtx = errors.New("no session in context")

// ErrSessionTimeout session timeout
var ErrSessionTimeout = errors.New("session timeout")

// ErrSessionTimeout session timeout
var ErrSessionCanceled = errors.New("session canceled")
