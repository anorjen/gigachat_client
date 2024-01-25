package dto

type Request struct {
	Model             string    `json:"model"`
	Messages          []Message `json:"messages"`
	Temperature       float64   `json:"temperature"`
	TopP              float64   `json:"top_p"`
	N                 int64     `json:"n"`
	Stream            bool      `json:"stream"`
	MaxTokens         int64     `json:"max_tokens"`
	RepetitionPenalty float64   `json:"repetition_penalty"`
	UpdateInterval    float64   `json:"update_interval"`
}
