package types

import (
	"encoding/base64"
	"encoding/json"

	batchingtypes "github.com/sedaprotocol/seda-chain/x/batching/types"
)

type Request struct {
	DrBinaryID        string                `json:"dr_binary_id"`
	DrInputs          string                `json:"dr_inputs"`
	GasLimit          string                `json:"gas_limit"`
	GasPrice          string                `json:"gas_price"`
	Height            uint64                `json:"height"`
	ID                string                `json:"id"`
	Memo              string                `json:"memo"`
	PaybackAddress    string                `json:"payback_address"`
	ReplicationFactor int64                 `json:"replication_factor"`
	ConsensusFilter   string                `json:"consensus_filter"`
	Reveals           map[string]RevealBody `json:"reveals"`
	SedaPayload       string                `json:"seda_payload"`
	TallyBinaryID     string                `json:"tally_binary_id"`
	TallyInputs       string                `json:"tally_inputs"`
	Version           string                `json:"version"`
}

type RevealBody struct {
	Salt         []byte   `json:"salt"`
	ExitCode     byte     `json:"exit_code"`
	GasUsed      string   `json:"gas_used"`
	Reveal       string   `json:"reveal"` // base64-encoded string
	ProxyPubKeys []string `json:"proxy_public_keys"`
}

func (u *RevealBody) MarshalJSON() ([]byte, error) {
	revealBytes, err := base64.StdEncoding.DecodeString(u.Reveal)
	if err != nil {
		return nil, err
	}

	intSlice := make([]int, len(revealBytes))
	for i, b := range revealBytes {
		intSlice[i] = int(b)
	}

	saltIntSlice := make([]int, len(u.Salt))
	for i, b := range u.Salt {
		saltIntSlice[i] = int(b)
	}

	type Alias RevealBody
	return json.Marshal(&struct {
		Reveal []int `json:"reveal"`
		Salt   []int `json:"salt"`
		*Alias
	}{
		Reveal: intSlice,
		Salt:   saltIntSlice,
		Alias:  (*Alias)(u),
	})
}

type Sudo struct {
	ID       string                   `json:"dr_id"`
	Result   batchingtypes.DataResult `json:"result"`
	ExitCode byte                     `json:"exit_code"`
}
