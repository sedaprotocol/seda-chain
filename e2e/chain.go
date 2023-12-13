package e2e

import (
	"fmt"
	"os"
	"sync"

	tmrand "github.com/cometbft/cometbft/libs/rand"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/sedaprotocol/seda-chain/app"
)

const (
	keyringPassphrase = "testpassphrase"
	keyringAppName    = "testnet"
)

var (
	encodingConfig app.EncodingConfig
	cdc            codec.Codec
	txConfig       client.TxConfig
)

func init() {
	encodingConfig = app.GetEncodingConfig()
	cdc = encodingConfig.Marshaler
	txConfig = encodingConfig.TxConfig
}

type chain struct {
	dataDir         string
	id              string
	validators      []*validator
	genesisAccounts []*account // initial accounts in genesis
}

func newChain() (*chain, error) {
	tmpDir, err := os.MkdirTemp("", "e2e-testnet-")
	if err != nil {
		return nil, err
	}

	return &chain{
		id:      "chain-" + tmrand.Str(6),
		dataDir: tmpDir,
	}, nil
}

func (c *chain) configDir() string {
	return fmt.Sprintf("%s/%s", c.dataDir, c.id)
}

func (c *chain) createAndInitValidators(count int) error {
	// Preallocate with a capacity to avoid rellocation
	c.validators = make([]*validator, 0, count)

	var wg sync.WaitGroup
	errChan := make(chan error, count)

	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			node := c.createValidator(index)

			// generate genesis files
			if err := node.init(); err != nil {
				errChan <- err
				return
			}

			// create keys
			if err := c.createKeys(node); err != nil {
				errChan <- err
				return
			}

			c.validators = append(c.validators, node)
		}(i)
	}

	wg.Wait()
	close(errChan)

	// Return the first error that occurred, if any
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *chain) createKeys(node *validator) error {
	keys := []string{"val", "node", "consensus"}
	for _, key := range keys {
		if err := node.createKey(key); err != nil {
			return err
		}
	}
	return nil
}

func (c *chain) createValidator(index int) *validator {
	return &validator{
		chain:   c,
		index:   index,
		moniker: fmt.Sprintf("%s-seda-%d", c.id, index),
	}
}
