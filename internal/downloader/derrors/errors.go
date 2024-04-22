package derrors

import (
	"errors"
)

var (
	ErrSizeLimitReached = errors.New("size limit reached")
	ErrNotFound         = errors.New("not found")
)
