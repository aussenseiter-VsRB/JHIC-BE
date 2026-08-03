package cvreview

type CvReviewRequest struct {
	CvText    string `json:"cv_text"`
	WordCount int    `json:"word_count"`
	PageCount int    `json:"page_count"`
}

type ValidateInputRequest map[string]any

type NormalizeOutputRequest struct {
	Raw string `json:"raw"`
}