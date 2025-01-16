package dbdriver

import "github.com/XShareGrid/cap/msg/errors"

// ErrInvalidValueForCondition ...
var ErrInvalidValueForCondition = errors.New("invalid value for condition (%s)")

// ErrUnknownOperator ...
var ErrUnknownOperator = errors.New("unknown operator (%s)")

// ErrOperatorNotSupportedForColumn ...
var ErrOperatorNotSupportedForColumn = errors.New("operator(%s) is not supported for column(%s)")

// ErrResultExceedMaxLimit ...
var ErrResultExceedMaxLimit = errors.New("result set count(%d) exceed max limit(%d)")
