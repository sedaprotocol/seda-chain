package e2e

import (
	"bytes"
	"context"
	"fmt"
	"slices"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	batchingtypes "github.com/sedaprotocol/seda-chain/x/batching/types"
)

func (s *IntegrationTestSuite) testUnbond() {
	s.Run("unbond", func() {
		val1Addr, err := s.chain.validators[0].keyInfo.GetAddress()
		s.Require().NoError(err)
		val1 := val1Addr.String()

		val2Addr, err := s.chain.validators[1].keyInfo.GetAddress()
		s.Require().NoError(err)
		val2 := val2Addr.String()

		// Ensure that only the first batch is created
		firstBatch, err := queryBatch(s.endpoint, 0)
		s.Require().NoError(err)
		_, err = queryBatch(s.endpoint, 1)
		s.Require().Error(err)

		validatorEntriesBefore := firstBatch.ValidatorEntries
		slices.SortFunc(validatorEntriesBefore, sortValidatorEntries)
		s.T().Logf("validatorEntriesBefore: %v", validatorEntriesBefore)

		s.execUnbond(s.chain, 0, sdk.ValAddress(val1Addr), val1, standardFees.String(), false)
		s.Require().Eventually(
			func() bool {
				batch, err := queryBatch(s.endpoint, firstBatch.Batch.BatchNumber+1)
				if err != nil {
					return false
				}
				return len(batch.BatchSignatures) == len(s.chain.validators)
			},
			30*time.Second,
			5*time.Second,
		)

		batchMiddle, err := queryBatch(s.endpoint, firstBatch.Batch.BatchNumber+1)
		s.Require().NoError(err)
		validatorEntriesMiddle := batchMiddle.ValidatorEntries
		slices.SortFunc(validatorEntriesMiddle, sortValidatorEntries)
		s.T().Logf("validatorEntriesMiddle: %v", validatorEntriesMiddle)

		for i, entry := range validatorEntriesBefore {
			s.Require().Equal(entry.ValidatorAddress, validatorEntriesMiddle[i].ValidatorAddress)

			// Due to the small amount being unbonded only the first validator should have had their voting power change: 50.000000% -> 49.999999%
			// The other validator should have had their voting power remain the same: 50.000000% -> 50.000000%
			if entry.ValidatorAddress.Equals(val1Addr) {
				s.Require().NotEqual(entry.VotingPowerPercent, validatorEntriesMiddle[i].VotingPowerPercent, "Voting power percent should have changed %d", entry.VotingPowerPercent)
			} else {
				s.Require().Equal(entry.VotingPowerPercent, validatorEntriesMiddle[i].VotingPowerPercent, "Voting power percent should not have changed %d", entry.VotingPowerPercent)
			}
		}

		s.execUnbond(s.chain, 1, sdk.ValAddress(val2Addr), val2, standardFees.String(), false)
		s.Require().Eventually(
			func() bool {
				batch, err := queryBatch(s.endpoint, 2)
				if err != nil {
					return false
				}

				if len(batch.BatchSignatures) != len(s.chain.validators) {
					return false
				}
				for _, sig := range batch.BatchSignatures {
					if len(sig.Secp256K1Signature) != 65 {
						return false
					}
					if sig.ValidatorAddress.Empty() {
						return false
					}
				}
				return true
			},
			90*time.Second,
			5*time.Second,
		)

		batchAfter, err := queryBatch(s.endpoint, batchMiddle.Batch.BatchNumber+1)
		s.Require().NoError(err)
		validatorEntriesAfter := batchAfter.ValidatorEntries
		slices.SortFunc(validatorEntriesAfter, sortValidatorEntries)
		s.T().Logf("validatorEntriesAfter: %v", validatorEntriesAfter)

		for i, entry := range validatorEntriesBefore {
			s.Require().Equal(entry.ValidatorAddress, validatorEntriesAfter[i].ValidatorAddress)
			s.Require().Equal(entry.VotingPowerPercent, validatorEntriesAfter[i].VotingPowerPercent, "Voting power percent should not have changed %d", entry.VotingPowerPercent)
		}
	})
}

func sortValidatorEntries(a, b batchingtypes.ValidatorTreeEntry) int {
	return bytes.Compare(a.ValidatorAddress, b.ValidatorAddress)
}

func (s *IntegrationTestSuite) execUnbond(
	c *chain,
	valIdx int,
	valAddr sdktypes.ValAddress,
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
		stakingtypes.ModuleName,
		"unbond",
		valAddr.String(),
		"1seda",
		"-y",
	}
	for flag, value := range opts {
		command = append(command, fmt.Sprintf("--%s=%v", flag, value))
	}

	s.T().Logf("unbond tx to trigger batch creation on chain %s", c.id)

	s.executeTx(ctx, c, command, valIdx, s.expectErrExecValidation(c, valIdx, expectErr))
}
