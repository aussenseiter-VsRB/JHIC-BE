package chat

const ChatMessageMaxLen = 300

type ChatRequest struct {
	ChatInput string `json:"chatInput"`
	SessionID string `json:"sessionId"`
	Topic     string `json:"topic,omitempty"`
}
