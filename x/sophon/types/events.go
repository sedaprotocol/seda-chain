package types

const (
	EventTypeRegisterSophon          = "register_sophon"
	EventTypeUpdateSophon            = "update_sophon"
	EventTypeTransferOwnership       = "transfer_ownership"
	EventTypeAcceptOwnership         = "accept_ownership"
	EventTypeCancelOwnershipTransfer = "cancel_ownership_transfer"

	AttributeSophonPubKey    = "sophon_pub_key"
	AttributeSophonID        = "sophon_id"
	AttributeOwnerAddress    = "owner_address"
	AttributeAdminAddress    = "admin_address"
	AttributeAddress         = "address"
	AttributeMemo            = "memo"
	AttributeBalance         = "balance"
	AttributeUsedCredits     = "used_credits"
	AttributeNewOwnerAddress = "new_owner_address"
)
