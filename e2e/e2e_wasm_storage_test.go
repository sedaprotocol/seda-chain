package e2e

import (
	"context"
	"fmt"
	"time"

	"github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

func (s *IntegrationTestSuite) testWasmStorageStoreDataRequestWasm() {
	s.Run("store_a_data_request_wasm", func() {
		senderAddress, err := s.chain.validators[0].keyInfo.GetAddress()
		s.Require().NoError(err)
		sender := senderAddress.String()

		s.execWasmStorageStoreDataRequest(s.chain, 0, sender, standardFees.String(), false)
	})
}

func (s *IntegrationTestSuite) execWasmStorageStoreDataRequest(
	c *chain,
	valIdx int,
	from,
	fees string,
	expectErr bool,
	opt ...flagOption,
) {
	// TODO remove the hardcode opt after refactor, all methods should accept custom flags
	opt = append(opt, withKeyValue(flagFees, fees))
	opt = append(opt, withKeyValue(flagFrom, from))
	opt = append(opt, withKeyValue(flagWasmType, "data-request"))

	opts := applyOptions(c.id, opt)

	wasmFile := "burner.wasm" // TO-DO: constant

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	command := []string{
		binary,
		txCommand,
		types.ModuleName,
		"store-data-request-wasm",
		wasmFile,
		"-y",
	}
	for flag, value := range opts {
		command = append(command, fmt.Sprintf("--%s=%v", flag, value))
	}

	s.T().Logf("storing data request Wasm %s on chain %s", wasmFile, c.id)
	s.T().Logf("command %s", command)

	s.executeTx(ctx, c, command, valIdx, s.expectErrExecValidation(c, valIdx, expectErr))
}
