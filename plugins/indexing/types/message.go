package types

type Message struct {
	Block *BlockContext `json:"-"`
	Type  string        `json:"type"`
	Data  interface{}   `json:"data"`
}

func NewMessage(messageType string, data interface{}, block *BlockContext) *Message {
	return &Message{
		Block: block,
		Type:  messageType,
		Data:  data,
	}
}
