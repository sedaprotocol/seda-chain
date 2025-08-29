package types

import (
	"fmt"
)

const (
	MaxUserIDLength = 66
)

func (s *SophonUser) ValidateBasic() error {
	if len(s.UserId) == 0 {
		return fmt.Errorf("user id is empty")
	}

	if len(s.UserId) > MaxUserIDLength {
		return fmt.Errorf("user id is too long; got: %d, max < %d", len(s.UserId), MaxUserIDLength)
	}

	if s.Credits.IsNil() {
		return fmt.Errorf("credits is nil")
	}

	if s.Credits.IsNegative() {
		return fmt.Errorf("credits is negative")
	}

	return nil
}
