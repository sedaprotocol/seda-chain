package types

type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func NewMessage(messageType string, data interface{}) *Message {
	return &Message{
		Type: messageType,
		Data: data,
	}
}
