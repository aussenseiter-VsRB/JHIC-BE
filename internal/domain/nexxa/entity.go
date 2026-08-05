package nexxa

type ChatResponse struct {
	Output string      `json:"output"`
	Button *ChatButton `json:"button,omitempty"`
}

type ChatButton struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}
