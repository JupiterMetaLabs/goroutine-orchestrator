package Errors

import "fmt"

var (
	ErrGlobalManagerNotFound = fmt.Errorf("global manager not found")
	ErrAppManagerNotFound    = fmt.Errorf("app manager not found")
	ErrLocalManagerNotFound  = fmt.Errorf("local manager not found")
	ErrLockContextCancelled  = fmt.Errorf("lock acquisition cancelled due to context cancellation")
	ErrRoutineNotFound       = fmt.Errorf("routine not found")
	ErrFunctionWgNotFound    = fmt.Errorf("function wg not found")
)

// this is for warnings
var (
	WrngLocalManagerAlreadyExists = fmt.Errorf("local manager already exists")
)