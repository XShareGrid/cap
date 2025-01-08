package utils

import "github.com/XShareGrid/cap/msg/errors"

// ErrFailedMapValue ...
var ErrFailedMapValue = errors.New("failed to map value (%v) from dt(%s) to vt(%s)")

// ErrFailedParseConditionValue ...
var ErrFailedParseConditionValue = errors.New("failed to parse condition value for(%s): %v")
