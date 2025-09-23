package wasm

import (
	"encoding/json"
	"strconv"

	"github.com/cosmos/cosmos-sdk/codec"

	coretypes "github.com/sedaprotocol/seda-chain/x/core/types"
)

const (
	QueryPathDataRequest            = "/sedachain.core.v1.Query/DataRequest"
	QueryPathDataRequestsByStatus   = "/sedachain.core.v1.Query/DataRequestsByStatus"
	QueryPathDataRequestStatuses    = "/sedachain.core.v1.Query/DataRequestStatuses"
	QueryPathStaker                 = "/sedachain.core.v1.Query/Staker"
	QueryPathStakingConfig          = "/sedachain.core.v1.Query/StakingConfig"
	QueryPathDataRequestConfig      = "/sedachain.core.v1.Query/DataRequestConfig"
	QueryPathExecutors              = "/sedachain.core.v1.Query/Executors"
	QueryPathGetExecutorEligibility = "/sedachain.core.v1.Query/GetExecutorEligibility"
)

type CoreContractQuery struct {
	GetDataRequest          *GetDataRequest          `json:"get_data_request"`
	GetDataRequestsByStatus *GetDataRequestsByStatus `json:"get_data_requests_by_status"`
	GetDataRequestsStatuses *GetDataRequestsStatuses `json:"get_data_requests_statuses"`
	GetStaker               *GetStaker               `json:"get_staker"`
	GetStakerAndSeq         *GetStakerAndSeq         `json:"get_staker_and_seq"`
	GetStakingConfig        *GetStakingConfig        `json:"get_staking_config"`
	GetDataRequestConfig    *GetDataRequestConfig    `json:"get_dr_config"`
	GetExecutors            *GetExecutors            `json:"get_executors"`
	GetExecutorEligibility  *GetExecutorEligibility  `json:"get_executor_eligibility"`
}

type GetDataRequest struct {
	ID string `json:"dr_id"`
}

func (g GetDataRequest) ToModuleQuery() ([]byte, string, error) {
	query := &coretypes.QueryDataRequestRequest{
		DrId: g.ID,
	}
	queryProto, err := query.Marshal()
	if err != nil {
		return nil, "", err
	}
	return queryProto, QueryPathDataRequest, nil
}

func (g GetDataRequest) FromModuleQuery(cdc codec.Codec, result []byte) ([]byte, error) {
	var res coretypes.QueryDataRequestResponse
	err := cdc.Unmarshal(result, &res)
	if err != nil {
		return nil, err
	}

	responseBytes, err := json.Marshal(res.DataRequest)
	if err != nil {
		return nil, err
	}
	return responseBytes, nil
}

type GetDataRequestsByStatus struct {
	Status        string   `json:"status"`
	Limit         int      `json:"limit"`
	LastSeenIndex []string `json:"last_seen_index"`
}

type GetDataRequestsByStatusResponse struct {
	DataRequests  []coretypes.DataRequest `json:"data_requests"`
	IsPaused      bool                    `json:"is_paused"`
	LastSeenIndex []string                `json:"last_seen_index"`
	Total         uint32                  `json:"total"`
}

func (g GetDataRequestsByStatus) ToModuleQuery() ([]byte, string, error) {
	var status coretypes.DataRequestStatus
	switch g.Status {
	case "committing":
		status = coretypes.DATA_REQUEST_STATUS_COMMITTING
	case "revealing":
		status = coretypes.DATA_REQUEST_STATUS_REVEALING
	case "tallying":
		status = coretypes.DATA_REQUEST_STATUS_TALLYING
	default:
		return nil, "", coretypes.ErrInvalidDataRequestStatus.Wrapf("status: %s", g.Status)
	}

	query := &coretypes.QueryDataRequestsByStatusRequest{
		Status: status,
		//nolint:gosec
		Limit:         uint64(g.Limit),
		LastSeenIndex: g.LastSeenIndex,
	}
	queryProto, err := query.Marshal()
	if err != nil {
		return nil, "", err
	}
	return queryProto, QueryPathDataRequestsByStatus, nil
}

func (g GetDataRequestsByStatus) FromModuleQuery(cdc codec.Codec, result []byte) ([]byte, error) {
	var res coretypes.QueryDataRequestsByStatusResponse
	err := cdc.Unmarshal(result, &res)
	if err != nil {
		return nil, err
	}

	response := GetDataRequestsByStatusResponse{
		DataRequests:  res.DataRequests,
		IsPaused:      res.IsPaused,
		LastSeenIndex: res.LastSeenIndex,
		//nolint:gosec // G115: Temporary support for old version.
		Total: uint32(res.Total),
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}
	return responseBytes, nil
}

// MarshalJSON prevents omission of the DataRequests field in the response.
func (r GetDataRequestsByStatusResponse) MarshalJSON() ([]byte, error) {
	type Alias GetDataRequestsByStatusResponse
	if r.DataRequests == nil {
		r.DataRequests = []coretypes.DataRequest{}
	}
	return json.Marshal(Alias(r))
}

type GetDataRequestsStatuses struct {
	DataRequestIDs []string `json:"dr_ids"`
}

func (g GetDataRequestsStatuses) ToModuleQuery() ([]byte, string, error) {
	query := &coretypes.QueryDataRequestStatusesRequest{
		DataRequestIds: g.DataRequestIDs,
	}
	queryProto, err := query.Marshal()
	if err != nil {
		return nil, "", err
	}
	return queryProto, QueryPathDataRequestStatuses, nil
}

func (g GetDataRequestsStatuses) FromModuleQuery(cdc codec.Codec, result []byte) ([]byte, error) {
	var res coretypes.QueryDataRequestStatusesResponse
	err := cdc.Unmarshal(result, &res)
	if err != nil {
		return nil, err
	}

	responseBytes, err := json.Marshal(res.Statuses)
	if err != nil {
		return nil, err
	}
	return responseBytes, nil
}

type GetStaker struct {
	PublicKey string `json:"public_key"`
}

type GetStakerResponse struct {
	StakerResponse
}

type StakerResponse struct {
	PublicKey               string `json:"public_key"`
	Memo                    []byte `json:"memo"`
	TokensPendingWithdrawal string `json:"tokens_pending_withdrawal"`
	TokensStaked            string `json:"tokens_staked"`
}

func (g GetStaker) ToModuleQuery() ([]byte, string, error) {
	query := &coretypes.QueryStakerRequest{
		PublicKey: g.PublicKey,
	}
	queryProto, err := query.Marshal()
	if err != nil {
		return nil, "", err
	}
	return queryProto, QueryPathStaker, nil
}

func (g GetStaker) FromModuleQuery(cdc codec.Codec, result []byte) ([]byte, error) {
	var res coretypes.QueryStakerResponse
	err := cdc.Unmarshal(result, &res)
	if err != nil {
		return nil, err
	}

	response := GetStakerResponse{
		StakerResponse{
			Memo:                    []byte(res.Staker.Memo),
			TokensPendingWithdrawal: res.Staker.PendingWithdrawal.String(),
			TokensStaked:            res.Staker.Staked.String(),
		},
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}
	return responseBytes, nil
}

type GetExecutors struct {
	Limit  uint32 `json:"limit"`
	Offset uint32 `json:"offset"`
}

type GetExecutorsResponse struct {
	Executors []StakerResponse `json:"executors"`
}

func (g GetExecutors) ToModuleQuery() ([]byte, string, error) {
	query := &coretypes.QueryExecutorsRequest{
		Limit:  g.Limit,
		Offset: g.Offset,
	}
	queryProto, err := query.Marshal()
	if err != nil {
		return nil, "", err
	}
	return queryProto, QueryPathExecutors, nil
}

func (g GetExecutors) FromModuleQuery(cdc codec.Codec, result []byte) ([]byte, error) {
	var res coretypes.QueryExecutorsResponse
	err := cdc.Unmarshal(result, &res)
	if err != nil {
		return nil, err
	}

	response := make([]StakerResponse, len(res.Executors))
	for i, executor := range res.Executors {
		response[i] = StakerResponse{
			PublicKey:               executor.PublicKey,
			Memo:                    []byte(executor.Memo),
			TokensPendingWithdrawal: executor.PendingWithdrawal.String(),
			TokensStaked:            executor.Staked.String(),
		}
	}

	responseBytes, err := json.Marshal(GetExecutorsResponse{
		Executors: response,
	})
	if err != nil {
		return nil, err
	}
	return responseBytes, nil
}

type GetStakerAndSeq struct {
	PublicKey string `json:"public_key"`
}

type GetStakerAndSeqResponse struct {
	Seq    string         `json:"seq"`
	Staker StakerResponse `json:"staker"`
}

func (g GetStakerAndSeq) ToModuleQuery() ([]byte, string, error) {
	query := &coretypes.QueryStakerRequest{
		PublicKey: g.PublicKey,
	}
	queryProto, err := query.Marshal()
	if err != nil {
		return nil, "", err
	}
	return queryProto, QueryPathStaker, nil
}

func (g GetStakerAndSeq) FromModuleQuery(cdc codec.Codec, result []byte) ([]byte, error) {
	var res coretypes.QueryStakerResponse
	err := cdc.Unmarshal(result, &res)
	if err != nil {
		return nil, err
	}

	response := GetStakerAndSeqResponse{
		Seq: strconv.FormatUint(res.Staker.SequenceNum, 10),
		Staker: StakerResponse{
			Memo:                    []byte(res.Staker.Memo),
			TokensPendingWithdrawal: res.Staker.PendingWithdrawal.String(),
			TokensStaked:            res.Staker.Staked.String(),
		},
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}
	return responseBytes, nil
}

type GetStakingConfig struct{}

type GetStakingConfigResponse struct {
	MinimumStake     string `json:"minimum_stake"`
	AllowlistEnabled bool   `json:"allowlist_enabled"`
}

func (g GetStakingConfig) ToModuleQuery() ([]byte, string, error) {
	query := &coretypes.QueryStakingConfigRequest{}
	queryProto, err := query.Marshal()
	if err != nil {
		return nil, "", err
	}
	return queryProto, QueryPathStakingConfig, nil
}

func (g GetStakingConfig) FromModuleQuery(cdc codec.Codec, result []byte) ([]byte, error) {
	var res coretypes.QueryStakingConfigResponse
	err := cdc.Unmarshal(result, &res)
	if err != nil {
		return nil, err
	}

	response := GetStakingConfigResponse{
		MinimumStake:     res.StakingConfig.MinimumStake.String(),
		AllowlistEnabled: res.StakingConfig.AllowlistEnabled,
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}
	return responseBytes, nil
}

type GetDataRequestConfig struct{}

type GetDataRequestConfigResponse struct {
	coretypes.DataRequestConfig
}

func (g GetDataRequestConfig) ToModuleQuery() ([]byte, string, error) {
	query := &coretypes.QueryDataRequestConfigRequest{}
	queryProto, err := query.Marshal()
	if err != nil {
		return nil, "", err
	}
	return queryProto, QueryPathDataRequestConfig, nil
}

func (g GetDataRequestConfig) FromModuleQuery(cdc codec.Codec, result []byte) ([]byte, error) {
	var res coretypes.QueryDataRequestConfigResponse
	err := cdc.Unmarshal(result, &res)
	if err != nil {
		return nil, err
	}

	responseBytes, err := json.Marshal(GetDataRequestConfigResponse{res.DataRequestConfig})
	if err != nil {
		return nil, err
	}
	return responseBytes, nil
}

type GetExecutorEligibility struct {
	Data string `json:"data"`
}

type GetExecutorEligibilityResponse struct {
	BlockHeight uint64                  `json:"block_height"`
	Status      LegacyEligibilityStatus `json:"status"`
}

type LegacyEligibilityStatus string

const (
	LegacyEligibilityStatusEligible            LegacyEligibilityStatus = "eligible"
	LegacyEligibilityStatusNotEligible         LegacyEligibilityStatus = "not_eligible"
	LegacyEligibilityStatusDataRequestNotFound LegacyEligibilityStatus = "data_request_not_found"
	LegacyEligibilityStatusNotStaker           LegacyEligibilityStatus = "not_staker"
	LegacyEligibilityStatusInvalidSignature    LegacyEligibilityStatus = "invalid_signature"
)

func (g GetExecutorEligibility) ToModuleQuery() ([]byte, string, error) {
	query := &coretypes.QueryGetExecutorEligibilityRequest{
		Data: g.Data,
	}
	queryProto, err := query.Marshal()
	if err != nil {
		return nil, "", err
	}
	return queryProto, QueryPathGetExecutorEligibility, nil
}

func (g GetExecutorEligibility) FromModuleQuery(cdc codec.Codec, result []byte) ([]byte, error) {
	var res coretypes.QueryGetExecutorEligibilityResponse
	err := cdc.Unmarshal(result, &res)
	if err != nil {
		return nil, err
	}

	response := GetExecutorEligibilityResponse{
		BlockHeight: res.BlockHeight,
	}

	// Convert eligibility status to legacy status.
	switch res.Status {
	case coretypes.ELIGIBILITY_STATUS_ELLIGIBLE:
		response.Status = LegacyEligibilityStatusEligible
	case coretypes.ELIGIBILITY_STATUS_NOT_ELIGIBLE:
		response.Status = LegacyEligibilityStatusNotEligible
	case coretypes.ELIGIBILITY_STATUS_DATA_REQUEST_NOT_FOUND:
		response.Status = LegacyEligibilityStatusDataRequestNotFound
	case coretypes.ELIGIBILITY_STATUS_NOT_STAKER:
		response.Status = LegacyEligibilityStatusNotStaker
	case coretypes.ELIGIBILITY_STATUS_INSUFFICIENT_STAKE:
		response.Status = LegacyEligibilityStatusNotStaker
	case coretypes.ELIGIBILITY_STATUS_NOT_ALLOWLISTED:
		response.Status = LegacyEligibilityStatusNotStaker
	case coretypes.ELIGIBILITY_STATUS_INVALID_SIGNATURE:
		response.Status = LegacyEligibilityStatusInvalidSignature
	default: // coretypes.ELIGIBILITY_STATUS_UNSPECIFIED
		response.Status = LegacyEligibilityStatusNotEligible
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}
	return responseBytes, nil
}
