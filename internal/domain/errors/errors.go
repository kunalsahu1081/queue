package qerrors

import (
	"fmt"
)

type Code string

const (
	CodeNotFound        Code = "NOT_FOUND"
	CodeUpdateMergeFailed Code = "UPDATE_MERGE_FAILED"
	CodeInternal        Code = "INTERNAL"
)

type AppError struct {
	Code    Code
	Message string
	Err     error 
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}



func QueueNotFound(msg string, err error) error {
	return &AppError{Code: CodeNotFound, Message: msg, Err: err}
}

func UpdateMergeFailed(msg string, err error) error {
	return &AppError{Code: CodeUpdateMergeFailed, Message: msg, Err: err}
}
