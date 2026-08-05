package spmb

import "errors"

var (
	ErrKkFileRequired    = errors.New("kk file is required")
	ErrKkTooLarge        = errors.New("kk file too large")
	ErrChildNameRequired = errors.New("child_name is required")
	ErrChildNameTooLong  = errors.New("child_name must be at most 100 characters")
	ErrQuestionRequired  = errors.New("question is required")
	ErrQuestionTooLong   = errors.New("question must be at most 300 characters")
	ErrOutputInvalid     = errors.New("AI output could not be interpreted")
)
