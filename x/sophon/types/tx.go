package types

import (
	"fmt"
)

func (m *MsgRegisterSophon) ValidateBasic() error {
	return fmt.Errorf("not implemented")
}

func (m *MsgEditSophon) ValidateBasic() error {
	return fmt.Errorf("not implemented")
}

func (m *MsgTransferOwnership) ValidateBasic() error {
	return fmt.Errorf("not implemented")
}

func (m *MsgAcceptOwnership) ValidateBasic() error {
	return fmt.Errorf("not implemented")
}

func (m *MsgCancelOwnershipTransfer) ValidateBasic() error {
	return fmt.Errorf("not implemented")
}

func (m *MsgAddUser) ValidateBasic() error {
	return fmt.Errorf("not implemented")
}

func (m *MsgTopUpUser) ValidateBasic() error {
	return fmt.Errorf("not implemented")
}

func (m *MsgExpireCredits) ValidateBasic() error {
	return fmt.Errorf("not implemented")
}

func (m *MsgSettleCredits) ValidateBasic() error {
	return fmt.Errorf("not implemented")
}

func (m *MsgSubmitReports) ValidateBasic() error {
	return fmt.Errorf("not implemented")
}

func (m *MsgUpdateParams) ValidateBasic() error {
	return m.Params.ValidateBasic()
}
