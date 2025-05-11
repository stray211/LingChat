package types

type RawResponse []byte

// MessageHandler 定义消息处理接口
type MessageHandler func([]byte) ([]RawResponse, error)

// Message 表示预期的 JSON 结构
type Message struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

// Response 表示服务器响应结构
type Response struct {
	Type            string `json:"type" yaml:"type"`
	Emotion         string `json:"emotion" yaml:"emotion"`
	OriginalTag     string `json:"originalTag" yaml:"originalTag"`
	Message         string `json:"message" yaml:"message"`
	MotionText      string `json:"motionText" yaml:"motionText"`
	AudioFile       string `json:"audioFile" yaml:"audioFile"`
	OriginalMessage string `json:"originalMessage" yaml:"originalMessage"`
	IsMultiPart     bool   `json:"isMultiPart" yaml:"isMultiPart"`
	PartIndex       int    `json:"partIndex" yaml:"partIndex"`
	TotalParts      int    `json:"totalParts" yaml:"totalParts"`
	Error           string `json:"error,omitempty"`
}
