package spmb

const ChildNameMaxLen = 100
const KkFileMaxSize = 5 << 20

type AskRequest struct {
	Question  string `json:"question"`
	SessionID string `json:"sessionId,omitempty"`
}

type AskResponse struct {
	Output string `json:"output"`
}
