package errs

import "fmt"

var (
	ErrNotFound       error = fmt.Errorf("not found")
	ErrUnjoinable           = fmt.Errorf("match unjoinable")
	ErrServerInternal       = fmt.Errorf("server internal error")
)
