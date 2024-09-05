package keeper

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"golang.org/x/crypto/sha3"

	"github.com/sedaprotocol/seda-chain/cmd/sedad/utils"
	"github.com/sedaprotocol/seda-chain/x/batching/types"
	tallytypes "github.com/sedaprotocol/seda-chain/x/tally/types"
)

func (k Keeper) EndBlock(ctx sdk.Context) (err error) {
	// Use defer to prevent returning an error, which would cause
	// the chain to halt.
	defer func() {
		// Handle a panic.
		if r := recover(); r != nil {
			k.Logger(ctx).Error("recovered from panic in batching EndBlock", "err", r)
		}
		// Handle an error.
		if err != nil {
			k.Logger(ctx).Error("error in batching EndBlock", "err", err)
		}
		err = nil
	}()

	batch, err := k.ConstructBatch(ctx)
	if err != nil {
		return err
	}

	err = k.SetBatch(ctx, batch)
	if err != nil {
		return err
	}
	err = k.IncrementCurrentBatchNum(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (k Keeper) ConstructBatch(ctx sdk.Context) (types.Batch, error) {
	curBatchNum, err := k.GetCurrentBatchNum(ctx)
	if err != nil {
		return types.Batch{}, err
	}
	dataRootHex, err := k.ConstructDataResultTree(ctx)
	if err != nil {
		return types.Batch{}, err
	}
	valRootHex, err := k.ConstructValidatorTree(ctx)
	if err != nil {
		return types.Batch{}, err
	}

	return types.Batch{
		BatchNumber:    curBatchNum,
		BlockHeight:    ctx.BlockHeight(),
		DataResultRoot: dataRootHex,
		ValidatorRoot:  valRootHex,
		BlockTime:      ctx.BlockTime(),
	}, nil
}

// ConstructDataResultTree constructs a data result tree based on the
// batching-ready data results returned from the core contract and
// returns a hex-encoded tree root.
func (k Keeper) ConstructDataResultTree(ctx sdk.Context) (string, error) {
	coreContract, err := k.wasmStorageKeeper.GetCoreContractAddr(ctx)
	if err != nil {
		return "", err
	}
	// TODO: Deal with offset and limits. (#313)
	queryRes, err := k.wasmViewKeeper.QuerySmart(ctx, coreContract, []byte(`{"get_data_results_by_status":{"status": "tallied", "offset": 0, "limit": 100}}`))
	if err != nil {
		return "", err
	}
	if string(queryRes) == "[]" {
		return "", err
	}

	var dataResults []tallytypes.DataResult
	err = json.Unmarshal(queryRes, &dataResults)
	if err != nil {
		return "", err
	}

	leaves := make([][]byte, len(dataResults))
	for _, res := range dataResults {
		resHash, err := hex.DecodeString(res.ID)
		if err != nil {
			return "", err
		}
		leaves = append(leaves, resHash)
	}

	// TODO construct the whole tree. (and merkle proofs?)
	// curRoot := merkle.HashFromByteSlices(leaves)
	curRoot := utils.HashFromByteSlices(leaves)
	prevRoot, err := k.GetPreviousDataResultRoot(ctx)
	if err != nil {
		return "", err
	}
	// root := merkle.HashFromByteSlices([][]byte{prevRoot, curRoot})
	root := utils.HashFromByteSlices([][]byte{prevRoot, curRoot})

	// TODO update data result status on contract

	return hex.EncodeToString(root), nil
}

type validatorPower struct {
	ValAddr sdk.ValAddress
	Power   int64
}

// ConstructValidatorTree constructs a validator tree based on the
// validators in the active set and their registered public keys.
func (k Keeper) ConstructValidatorTree(ctx sdk.Context) (string, error) {
	var activeSet []validatorPower
	err := k.stakingKeeper.IterateLastValidatorPowers(ctx, func(valAddr sdk.ValAddress, power int64) (stop bool) {
		activeSet = append(activeSet, validatorPower{ValAddr: valAddr, Power: power})
		return false
	})
	if err != nil {
		return "", err
	}

	var leaves [][]byte
	var votes []types.Vote
	for _, vp := range activeSet {
		pubKey, err := k.pubKeyKeeper.GetValidatorKeyAtIndex(ctx, vp.ValAddr, utils.SEDAKeysIndexSecp256k1)
		if err != nil {
			if errors.Is(err, collections.ErrNotFound) {
				continue // TODO check
			}
		}

		// Construct a leaf content and hash it.
		pkBytes := pubKey.Bytes()
		buf := make([]byte, len(pkBytes)+8)
		copy(buf[:len(pkBytes)], pkBytes)
		binary.BigEndian.PutUint64(buf[len(pkBytes):], uint64(vp.Power))

		hash := sha3.New256()
		hash.Write(buf)
		hashed := hash.Sum(nil)

		leaves = append(leaves, hashed)
		votes = append(votes, types.Vote{
			ValidatorAddr: vp.ValAddr.String(),
			VotingPower:   vp.Power,
			Signatures: []*types.Signature{{
				Scheme:      utils.SEDAKeysIndexSecp256k1,
				Signature:   "",
				PublicKey:   pubKey.String(),
				MerkleProof: "", // TODO populate this
			}},
		})
	}

	// TODO construct the whole tree and populate merkle proof fields.
	// secp256k1Root := merkle.HashFromByteSlices(leaves)
	// root := merkle.HashFromByteSlices([][]byte{{}, secp256k1Root})
	secp256k1Root := utils.HashFromByteSlices(leaves)
	root := utils.HashFromByteSlices([][]byte{{}, secp256k1Root})

	return hex.EncodeToString(root), nil
}
