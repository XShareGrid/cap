package mysql

import "github.com/XShareGrid/cap/msg/errors"

// ErrRowsAffectedZero rows affected 0
var ErrRowsAffectedZero = errors.New("rows affected 0")

var ErrDuplicateEntry = errors.New("duplicate entry")

var ErrDataTooLong = errors.New("[%s]data too long")
