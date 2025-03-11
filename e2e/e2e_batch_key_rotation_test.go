package e2e

import (
	"bytes"
	"context"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	pubkeytypes "github.com/sedaprotocol/seda-chain/x/pubkey/types"
)

func (s *IntegrationTestSuite) testBatchKeyRotation() {
	s.Run("batch key rotation", func() {
		val1Addr, err := s.chain.validators[0].keyInfo.GetAddress()
		s.Require().NoError(err)
		val1 := val1Addr.String()

		// Get currently registered pubkeyBefore of validator 1
		var val1OperatorAddr sdk.ValAddress
		val1OperatorAddr = val1Addr.Bytes()
		pubkeyBefore, err := queryPubkey(s.endpoint, val1OperatorAddr.String())
		s.Require().NoError(err)

		// Get the latest batch before rotation
		batchBefore, err := queryLatestBatch(s.endpoint)
		s.Require().NoError(err)

		s.T().Logf("Batch number before rotation: %v", batchBefore.Batch.BatchNumber)

		var ethAddrBefore []byte
		for _, entry := range batchBefore.ValidatorEntries {
			if entry.ValidatorAddress.Equals(val1Addr) {
				ethAddrBefore = entry.EthAddress
				break
			}
		}
		s.Require().NotNil(ethAddrBefore)

		// Rotate key of validator 1
		s.execBatchKeyRotation(s.chain, 0, val1, standardFees.String(), false)
		s.Require().Eventually(
			func() bool {
				batch, err := queryBatch(s.endpoint, batchBefore.Batch.BatchNumber+1)
				if err != nil {
					s.T().Logf("Error querying batch: %v", err)
					return false
				}
				for _, entry := range batch.ValidatorEntries {
					if entry.ValidatorAddress.Equals(val1Addr) {
						ethAddrChanged := !bytes.Equal(entry.EthAddress, ethAddrBefore)
						s.T().Logf("Validator 1 eth address changed: %v", ethAddrChanged)
						return ethAddrChanged
					}
				}
				// Didn't find validator 1 in the batch
				s.T().Logf("Validator 1 not found in batch %v", batch.Batch.BatchNumber+1)
				return false
			},
			30*time.Second,
			5*time.Second,
		)

		// Verify pubkey of validator 1 is rotated in registry
		pubkeyAfter, err := queryPubkey(s.endpoint, val1OperatorAddr.String())
		s.Require().NoError(err)
		s.Require().NotEqual(pubkeyBefore.ValidatorPubKeys.IndexedPubKeys[0].PubKey, pubkeyAfter.ValidatorPubKeys.IndexedPubKeys[0].PubKey)

		// Trigger new batch creation and verify that it is signed
		s.execUnbond(s.chain, 0, sdk.ValAddress(val1Addr), val1, standardFees.String(), false)
		s.Require().Eventually(
			func() bool {
				batch, err := queryBatch(s.endpoint, batchBefore.Batch.BatchNumber+2)
				if err != nil {
					s.T().Logf("Error querying batch: %v", err)
					return false
				}
				s.T().Logf("Batch signatures: %v, expected: %v", len(batch.BatchSignatures), len(s.chain.validators))
				return len(batch.BatchSignatures) == len(s.chain.validators)
			},
			30*time.Second,
			5*time.Second,
		)
	})
}

func (s *IntegrationTestSuite) execBatchKeyRotation(
	c *chain,
	valIdx int,
	from string,
	fees string,
	expectErr bool,
	opt ...flagOption,
) {
	opt = append(opt, withKeyValue(flagFees, fees))
	opt = append(opt, withKeyValue(flagFrom, from))
	opts := applyTxOptions(c.id, opt)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	command := []string{
		binary,
		txCommand,
		pubkeytypes.ModuleName,
		"add-seda-keys",
		"-y",
		"--key-file-force",
		"--key-file-no-encryption",
	}
	for flag, value := range opts {
		command = append(command, fmt.Sprintf("--%s=%v", flag, value))
	}

	s.T().Logf("add-seda-keys tx to rotate batch signing key on chain %s", c.id)

	s.executeTx(ctx, c, command, valIdx, s.expectErrExecValidation(c, valIdx, expectErr))
}
