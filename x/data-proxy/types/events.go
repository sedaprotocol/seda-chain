package types

const (
	EventTypeRegisterProxy = "register_data_proxy"
	EventTypeEditProxy     = "edit_data_proxy"
	EventTypeFeeUpdate     = "fee_update"
	EventTypeTransferAdmin = "transfer_admin"

	AttributePubKey        = "pub_key"
	AttributePayoutAddress = "payout_address"
	AttributeFee           = "fee"
	AttributeMemo          = "memo"
	AttributeAdminAddress  = "admin_address"
	AttributeNewFee        = "new_fee"
	AttributeNewFeeHeight  = "new_fee_height"
)
