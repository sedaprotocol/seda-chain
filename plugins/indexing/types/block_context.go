package types

import "time"

type BlockContext struct {
	Height int64
	Time   time.Time
}

func NewBlockContext(height int64, timestamp time.Time) *BlockContext {
	return &BlockContext{
		Height: height,
		Time:   timestamp,
	}
}
