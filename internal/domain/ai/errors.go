package ai

import "errors"

var (
	ErrChatMessageRequired = errors.New("chat message is required")
	ErrChatMessageTooLong  = errors.New("chat message exceeds 300 characters")
	ErrAnswersRequired     = errors.New("all 8 answers are required")
	ErrAnswerTooLong       = errors.New("each answer must be at most 500 characters")
	ErrNexxaOutputInvalid  = errors.New("AI output could not be interpreted")
	ErrN8NUnavailable      = errors.New("upstream service unavailable")
	ErrN8NTimeout          = errors.New("upstream timed out")
)
