package e2e

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

const (
	drWasm      = "burner.wasm"
	tallyWasm   = "reflect.wasm"
	overlayWasm = "staking.wasm"
	coreWasm    = "core_contract.wasm"
)

func (s *IntegrationTestSuite) testWasmStorageStoreDataRequestWasm() {
	s.Run("store_a_data_request_wasm", func() {
		senderAddress, err := s.chain.validators[0].keyInfo.GetAddress()
		s.Require().NoError(err)
		sender := senderAddress.String()

		bytecode, err := os.ReadFile(filepath.Join(localWasmDirPath, drWasm))
		if err != nil {
			panic("failed to read data request Wasm file")
		}
		drHashBytes := crypto.Keccak256(bytecode)
		if drHashBytes == nil {
			panic("failed to compute hash")
		}
		drHashStr := hex.EncodeToString(drHashBytes)

		//
		bytecode, err = os.ReadFile(filepath.Join(localWasmDirPath, tallyWasm))
		if err != nil {
			panic("failed to read tally Wasm file")
		}
		tallyHashBytes := crypto.Keccak256(bytecode)
		if tallyHashBytes == nil {
			panic("failed to compute hash")
		}
		tallyHashStr := hex.EncodeToString(tallyHashBytes)

		s.execWasmStorageStoreDataRequest(s.chain, 0, drWasm, "data-request", sender, standardFees.String(), false)
		s.execWasmStorageStoreDataRequest(s.chain, 0, tallyWasm, "tally", sender, standardFees.String(), false)

		s.Require().Eventually(
			func() bool {
				drWasmRes, err := queryDataRequestWasm(s.endpoint, drHashStr)
				s.Require().NoError(err)
				s.Require().True(bytes.Equal(drHashBytes, drWasmRes.Wasm.Hash))

				tallyWasmRes, err := queryDataRequestWasm(s.endpoint, tallyHashStr)
				s.Require().NoError(err)
				s.Require().True(bytes.Equal(tallyHashBytes, tallyWasmRes.Wasm.Hash))

				wasms, err := queryDataRequestWasms(s.endpoint)
				s.Require().NoError(err)

				if fmt.Sprintf("%s,%s", drHashStr, types.WasmTypeDataRequest.String()) == wasms.HashTypePairs[0] {
					return fmt.Sprintf("%s,%s", tallyHashStr, types.WasmTypeDataRequest.String()) == wasms.HashTypePairs[1]
				}
				if fmt.Sprintf("%s,%s", tallyHashStr, types.WasmTypeDataRequest.String()) == wasms.HashTypePairs[0] {
					return fmt.Sprintf("%s,%s", drHashStr, types.WasmTypeDataRequest.String()) == wasms.HashTypePairs[1]
				}
				return false
			},
			30*time.Second,
			5*time.Second,
		)
	})
}

func (s *IntegrationTestSuite) execWasmStorageStoreDataRequest(
	c *chain,
	valIdx int,
	drWasm,
	wasmType,
	from,
	fees string,
	expectErr bool,
	opt ...flagOption,
) {
	opt = append(opt, withKeyValue(flagFees, fees))
	opt = append(opt, withKeyValue(flagFrom, from))
	opt = append(opt, withKeyValue(flagWasmType, wasmType))

	opts := applyTxOptions(c.id, opt)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	wasmFilePath := filepath.Join(containerWasmDirPath, drWasm)

	command := []string{
		binary,
		txCommand,
		types.ModuleName,
		"store-data-request-wasm",
		wasmFilePath,
		"-y",
	}
	for flag, value := range opts {
		command = append(command, fmt.Sprintf("--%s=%v", flag, value))
	}

	s.T().Logf("storing data request Wasm %s on chain %s", wasmFilePath, c.id)

	s.executeTx(ctx, c, command, valIdx, s.expectErrExecValidation(c, valIdx, expectErr))
}
