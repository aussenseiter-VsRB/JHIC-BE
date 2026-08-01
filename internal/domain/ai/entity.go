package ai

const (
	ChatMessageMaxLen = 300
	NexxaAnswerCount  = 8
	NexxaAnswerMaxLen = 500
)

type ChatRequest struct {
	ChatInput string `json:"chatInput"`
	SessionID string `json:"sessionId"`
}

type ChatResponse struct {
	Output string `json:"output"`
}

type NexxaRequest struct {
	Jawaban1 string `json:"jawaban_1"`
	Jawaban2 string `json:"jawaban_2"`
	Jawaban3 string `json:"jawaban_3"`
	Jawaban4 string `json:"jawaban_4"`
	Jawaban5 string `json:"jawaban_5"`
	Jawaban6 string `json:"jawaban_6"`
	Jawaban7 string `json:"jawaban_7"`
	Jawaban8 string `json:"jawaban_8"`
}

func (n NexxaRequest) Answers() []string {
	return []string{
		n.Jawaban1, n.Jawaban2, n.Jawaban3, n.Jawaban4,
		n.Jawaban5, n.Jawaban6, n.Jawaban7, n.Jawaban8,
	}
}

type NexxaResponse struct {
	NamaJurusan string `json:"nama_jurusan"`
	Alasan      string `json:"alasan"`
}
