package postgresql

import "errors"

var (
	ErrDuplicateLogin = errors.New("duplicate login")
	ErrTemporary      = errors.New("temporary error")
)
