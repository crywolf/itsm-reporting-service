package memory

import "errors"

// ErrNotFound represents the error when object is not found in the repository
var ErrNotFound = errors.New("record was not found")
