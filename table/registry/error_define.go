package registry

import "github.com/XShareGrid/cap/msg/errors"

// ErrDuplicateNodeID duplicate function id
var ErrDuplicateNodeID = errors.New("duplicate node id (%s)")

// ErrDuplicateNodeFiled duplicate function name
var ErrDuplicateNodeFiled = errors.New("duplicate node filed(%s)")

// ErrNodeNotExist id 404
var ErrNodeNotExist = errors.New("node is not exist id = (%s)")

// ErrDuplicateColumnID column id
var ErrDuplicateColumnID = errors.New("duplicate column id = (%s)")

// ErrDuplicateColumnName column name
var ErrDuplicateColumnName = errors.New("duplicate column name = (%s)")

// ErrParseFieldType failed to parse field type
var ErrParseFieldType = errors.New("failed to parse field type: (%s)")

// ErrUnsupportedFieldType Unsupported Field Type
var ErrUnsupportedFieldType = errors.New("unsupported value type: (%s)")

// ErrInvalidValueTypeForTime Invalid t_vt for time
var ErrInvalidValueTypeForTime = errors.New("invalid t_vt for time.Time: (%s)")

// ErrInvalidOperatorIDFormat Invalid operator id format
var ErrInvalidOperatorIDFormat = errors.New("invalid id (%s) format for operator, should be 'namespace.id'")

// ErrBuiltinNamespaceNowAllowed namespace builtin is not allowed
var ErrBuiltinNamespaceNowAllowed = errors.New("namespace 'builtin' is not allowed")

// ErrInvalidLinkFormat invalid link format (%s)
var ErrInvalidLinkFormat = errors.New("invalid link format (%s)")

// ErrInvalidColumnID invalid column id
var ErrInvalidColumnID = errors.New("invalid column id (%s)")

// ErrMultipleIDColumnNotAllowed multiple id column is not allowed
var ErrMultipleIDColumnNotAllowed = errors.New("multiple id column is not allowed for (%s)")

// ErrNoIDColumnSpecified no t_key is specified
var ErrNoIDColumnSpecified = errors.New("no valid t_key is specified for (%s)")

// ErrOptionTypeNotFound option type not found
var ErrOptionTypeNotFound = errors.New("option(%s) not found in registry")

// ErrOptionNotFound option not found
var ErrOptionNotFound = errors.New("option(%d) not found for (%s)")

// ErrKeyColumnMustSupportEQ key column must support builtin.EQ operator
var ErrKeyColumnMustSupportEQ = errors.New("key column(%s) must support builtin.EQ operator")

// ErrOperatorNotSupported operator(%s) for (%s.%s) is not supported
var ErrOperatorNotSupported = errors.New("operator(%s) for (%s.%s) is not supported")

// ErrInvalidConditionValue invalid condition value count, want %s, got %d
var ErrInvalidConditionValue = errors.New("invalid condition value count, want (%s), got (%d)")

// ErrInvalidConditionValueType invalid condition value for col(%s), want (%s), got (%s)
var ErrInvalidConditionValueType = errors.New("invalid condition value for col(%s), want (%s), got (%s)")

// ErrAggregateNotSupported operator(%s) for (%s.%s) is not supported
var ErrAggregateNotSupported = errors.New("aggregation for (%s.%s) is not supported")

// ErrTplNullOutput output columns for template is null
var ErrTplNullOutput = errors.New("output columns for template is null")

// ErrInvalidLink invalid link configuration
var ErrInvalidLink = errors.New("invalid link configuration: %s.%s -> %s")

// ErrInvalidRowActionID invalid row action id
var ErrInvalidRowActionID = errors.New("invalid row action id (%s)")

// ErrIllegalArguments illegal arguments(%s)
var ErrIllegalArguments = errors.New("illegal arguments(%s)")
