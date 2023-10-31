package e2e

// import (
// 	"context"
// 	"fmt"
// 	"time"

// 	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
// )

// func (s *IntegrationTestSuite) testBankTokenTransfer() {
// 	s.Run("send_between_accounts", func() {
// 		senderAddress, err := s.chain.validators[0].keyInfo.GetAddress()
// 		s.Require().NoError(err)
// 		sender := senderAddress.String()

// 		recipientAddress, err := s.chain.validators[1].keyInfo.GetAddress()
// 		s.Require().NoError(err)
// 		recipient := recipientAddress.String()

// 		// chainAAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chain.id][0].GetHostPort("1317/tcp"))

// 		// var (
// 		// 	beforeSenderUAtomBalance    sdk.Coin
// 		// 	beforeRecipientUAtomBalance sdk.Coin
// 		// )

// 		// s.Require().Eventually(
// 		// 	func() bool {
// 		// 		beforeSenderUAtomBalance, err = getSpecificBalance(chainAAPIEndpoint, sender, asedaDenom)
// 		// 		s.Require().NoError(err)

// 		// 		beforeRecipientUAtomBalance, err = getSpecificBalance(chainAAPIEndpoint, recipient, asedaDenom)
// 		// 		s.Require().NoError(err)

// 		// 		return beforeSenderUAtomBalance.IsValid() && beforeRecipientUAtomBalance.IsValid()
// 		// 	},
// 		// 	10*time.Second,
// 		// 	5*time.Second,
// 		// )

// 		s.execBankSend(s.chain, 0, sender, recipient, tokenAmount.String(), standardFees.String(), false)

// 		// s.Require().Eventually(
// 		// 	func() bool {
// 		// 		afterSenderUAtomBalance, err := getSpecificBalance(chainAAPIEndpoint, sender, asedaDenom)
// 		// 		s.Require().NoError(err)

// 		// 		afterRecipientUAtomBalance, err := getSpecificBalance(chainAAPIEndpoint, recipient, asedaDenom)
// 		// 		s.Require().NoError(err)

// 		// 		decremented := beforeSenderUAtomBalance.Sub(tokenAmount).Sub(standardFees).IsEqual(afterSenderUAtomBalance)
// 		// 		incremented := beforeRecipientUAtomBalance.Add(tokenAmount).IsEqual(afterRecipientUAtomBalance)

// 		// 		return decremented && incremented
// 		// 	},
// 		// 	time.Minute,
// 		// 	5*time.Second,
// 		// )
// 	})
// }

// func (s *IntegrationTestSuite) execBankSend(
// 	c *chain,
// 	valIdx int,
// 	from,
// 	to,
// 	amt,
// 	fees string,
// 	expectErr bool,
// 	opt ...flagOption,
// ) {
// 	// TODO remove the hardcode opt after refactor, all methods should accept custom flags
// 	opt = append(opt, withKeyValue(flagFees, fees))
// 	opt = append(opt, withKeyValue(flagFrom, from))
// 	opts := applyOptions(c.id, opt)

// 	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
// 	defer cancel()

// 	s.T().Logf("sending %s tokens from %s to %s on chain %s", amt, from, to, c.id)

// 	command := []string{
// 		binary,
// 		txCommand,
// 		banktypes.ModuleName,
// 		"send",
// 		from,
// 		to,
// 		amt,
// 		"-y",
// 	}
// 	for flag, value := range opts {
// 		command = append(command, fmt.Sprintf("--%s=%v", flag, value))
// 	}

// 	s.executeTx(ctx, c, command, valIdx, s.expectErrExecValidation(c, valIdx, expectErr))
// }
