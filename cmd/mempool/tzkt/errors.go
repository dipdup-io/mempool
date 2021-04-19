package tzkt

import "github.com/pkg/errors"

// errors
var (
	ErrEmptyKindList           = errors.New("Empty kind list")
	ErrUnknownOperationKind    = errors.New("Unknown operation kind")
	ErrOperationDoesNotContain = errors.New("Operation doesn't contain field")
	ErrInvalidFieldType        = errors.New("Field has invalid type")
	ErrInvalidBodyType         = errors.New("Invalid message body type")
	ErrInvalidOperationType    = errors.New("Invalid operation body type")
	ErrInvalidBlockType        = errors.New("Invalid block body type")
	ErrUnknownMessageType      = errors.New("Unknown TzKT message type")
)
