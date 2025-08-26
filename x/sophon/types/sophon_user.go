package types

import (
	"fmt"
)

func (s *SophonUser) ValidateBasic() error {
	if len(s.UserId) == 0 {
		return fmt.Errorf("user id is empty")
	}

	if s.Credits.IsNil() {
		return fmt.Errorf("credits is nil")
	}

	return nil
}
