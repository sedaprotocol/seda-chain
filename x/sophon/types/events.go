package types

const (
	EventTypeSophon                  = "sophon"
	EventTypeUser                    = "user"
	EventTypeRegisterSophon          = "register_sophon"
	EventTypeUpdateSophon            = "update_sophon"
	EventTypeTransferOwnership       = "transfer_ownership"
	EventTypeAcceptOwnership         = "accept_ownership"
	EventTypeCancelOwnershipTransfer = "cancel_ownership_transfer"
	EventTypeAddUser                 = "add_user"

	AttributeSophonPubKey    = "sophon_pub_key"
	AttributeSophonID        = "sophon_id"
	AttributeOwnerAddress    = "owner_address"
	AttributeAdminAddress    = "admin_address"
	AttributeAddress         = "address"
	AttributeMemo            = "memo"
	AttributeBalance         = "balance"
	AttributeUsedCredits     = "used_credits"
	AttributeNewOwnerAddress = "new_owner_address"
	AttributeUserID          = "user_id"
	//nolint:gosec // G101: These are not sensitive values
	AttributeInitialCredits = "user_initial_credits"
	AttributeCredits        = "credits"
)
