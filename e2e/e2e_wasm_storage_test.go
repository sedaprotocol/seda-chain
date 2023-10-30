package e2e

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/hyperledger/burrow/crypto"
	"github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

func (s *IntegrationTestSuite) testWasmStorageStoreDataRequestWasm() {
	chainEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chain.id][0].GetHostPort("1317/tcp"))

	s.Run("store_a_data_request_wasm", func() {
		senderAddress, err := s.chain.validators[0].keyInfo.GetAddress()
		s.Require().NoError(err)
		sender := senderAddress.String()

		wasmFileName := "burner.wasm"
		bytecode, err := os.ReadFile("testwasms/burner.wasm")
		if err != nil {
			panic("failed to read file")
		}

		hashBytes := crypto.Keccak256(bytecode)
		if hashBytes == nil {
			panic("failed to compute hash")
		}
		hashString := hex.EncodeToString(hashBytes)

		s.execWasmStorageStoreDataRequest(s.chain, 0, wasmFileName, sender, standardFees.String(), false)

		s.Require().Eventually(
			func() bool {
				res, err := queryDataRequestWasm(chainEndpoint, hashString)
				s.Require().NoError(err)

				return bytes.Equal(hashBytes, res.Wasm.Hash)
			},
			20*time.Second,
			5*time.Second,
		)
	})
}

func (s *IntegrationTestSuite) execWasmStorageStoreDataRequest(
	c *chain,
	valIdx int,
	wasmFileName,
	from,
	fees string,
	expectErr bool,
	opt ...flagOption,
) {
	opt = append(opt, withKeyValue(flagFees, fees))
	opt = append(opt, withKeyValue(flagFrom, from))
	opt = append(opt, withKeyValue(flagWasmType, "data-request"))

	opts := applyOptions(c.id, opt)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	wasmFilePath := filepath.Join(containerWasmDirPath, wasmFileName)

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
