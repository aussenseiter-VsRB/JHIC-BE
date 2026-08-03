package cvreview

import "errors"

var (
	ErrCvTextRequired  = errors.New("cv_text is required")
	ErrCvTextTooLong   = errors.New("cv_text must be at most 50000 characters")
	ErrInvalidCounts   = errors.New("word_count and page_count must be zero or positive")
	ErrCvOutputInvalid = errors.New("AI output could not be interpreted")
)