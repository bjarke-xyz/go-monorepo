package controller

import "fmt"

type ControllerError struct {
	StatusCode int
	Err        error
}

func CError(status int, err error) *ControllerError {
	return &ControllerError{
		StatusCode: status,
		Err:        err,
	}
}

func (e *ControllerError) Error() string {
	return fmt.Sprintf("Status %d: err %v", e.StatusCode, e.Err)
}
