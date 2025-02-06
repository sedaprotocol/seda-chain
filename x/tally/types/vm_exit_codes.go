package types

const (
	TallyExitCodeNotEnoughCommits   uint32 = 200 // tally VM not executed due to not enough commits
	TallyExitCodeInvalidRequest     uint32 = 201 // tally VM not executed due to invalid request
	TallyExitCodeContractPaused     uint32 = 202 // tally VM not executed due to contract being paused
	TallyExitCodeInvalidFilterInput uint32 = 253 // tally VM not executed due to invalid filter input
	TallyExitCodeFilterError        uint32 = 254 // tally VM not executed due to filter error
	TallyExitCodeExecError          uint32 = 255 // error while executing tally VM
)
