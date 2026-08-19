package transfer

import (
	"fmt"
	"strings"
)

// Operation is the transfer mode applied to the destination collection.
type Operation string

const (
	OperationReplace Operation = "replace"
	OperationAdd     Operation = "add"
)

// ParseOperation validates a --operation flag value.
func ParseOperation(value string) (Operation, error) {
	switch Operation(strings.ToLower(strings.TrimSpace(value))) {
	case OperationReplace:
		return OperationReplace, nil
	case OperationAdd:
		return OperationAdd, nil
	default:
		return "", fmt.Errorf("invalid --operation %q: must be %q or %q", value, OperationReplace, OperationAdd)
	}
}
