package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"cosmossdk.io/math"
	evidencetypes "cosmossdk.io/x/evidence/types"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"

	tmconfig "github.com/cometbft/cometbft/config"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/server"
	srvconfig "github.com/cosmos/cosmos-sdk/server/config"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
)

const (
	binary                 = "seda-chaind"
	txCommand              = "tx"
	queryCommand           = "query"
	keysCommand            = "keys"
	containerChainHomePath = "/seda-chain/.seda-chain"
	containerHomePath      = "/seda-chain/"
	containerWasmDirPath   = "/seda-chain/testwasms"
	localWasmDirPath       = "../testutil/testwasms" // relative path is ok
	sedaDenom              = "seda"
	asedaDenom             = "aseda"
	stakeDenom             = asedaDenom
	initBalanceStr         = "100000000000000000seda"
	selfDelegationStr      = "10000000000000000000000000000000000aseda"
	minGasPrice            = "0.00001"

	// the test globalfee in genesis is the same as minGasPrice
	// global fee lower/higher than min_gas_price
	initialGlobalFeeAmt                   = "0.00001"
	lowGlobalFeesAmt                      = "0.000001"
	highGlobalFeeAmt                      = "0.0001"
	maxTotalBypassMinFeeMsgGasUsage       = "1"
	gas                                   = 200000
	govProposalBlockBuffer                = 35
	relayerAccountIndexHermes0            = 0
	relayerAccountIndexHermes1            = 1
	numberOfEvidences                     = 10
	slashingShares                  int64 = 10000
)

var (
	standardFees    = sdk.NewCoin(asedaDenom, math.NewInt(330000))
	proposalCounter = 0
)

type IntegrationTestSuite struct {
	suite.Suite
	tmpDirs      []string
	chain        *chain
	dkrPool      *dockertest.Pool
	dkrNet       *dockertest.Network
	valResources map[string][]*dockertest.Resource
	endpoint     string
	grpcEndpoint string //nolint:unused // unused
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up e2e integration test suite...")

	var err error
	s.chain, err = newChain()
	s.Require().NoError(err)

	s.dkrPool, err = dockertest.NewPool("")
	s.Require().NoError(err)

	s.dkrNet, err = s.dkrPool.CreateNetwork(fmt.Sprintf("%s-testnet", s.chain.id))
	s.Require().NoError(err)

	s.valResources = make(map[string][]*dockertest.Resource)

	s.T().Logf("starting e2e infrastructure; chain-id: %s; datadir: %s", s.chain.id, s.chain.dataDir)
	s.initNodes(s.chain)
	s.initGenesis(s.chain)
	s.initValidatorConfigs(s.chain)
	s.runValidators(s.chain, 0)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	if str := os.Getenv("E2E_SKIP_CLEANUP"); len(str) > 0 {
		skipCleanup, err := strconv.ParseBool(str)
		s.Require().NoError(err)

		if skipCleanup {
			return
		}
	}

	s.T().Log("tearing down e2e integration test suite...")

	for _, vr := range s.valResources {
		for _, r := range vr {
			s.Require().NoError(s.dkrPool.Purge(r))
		}
	}

	s.Require().NoError(s.dkrPool.RemoveNetwork(s.dkrNet))

	os.RemoveAll(s.chain.dataDir)

	for _, td := range s.tmpDirs {
		os.RemoveAll(td)
	}
}

func (s *IntegrationTestSuite) initNodes(c *chain) {
	err := c.createAndInitValidators(2)
	s.Require().NoError(err)

	err = c.addAccountFromMnemonic(5) // add 5 accounts to val0 local directory
	s.Require().NoError(err)

	// Initialize a genesis file for the first validator
	val0ConfigDir := c.validators[0].configDir()
	var addrAll []sdk.AccAddress
	for _, val := range c.validators {
		address, err := val.keyInfo.GetAddress()
		s.Require().NoError(err)
		addrAll = append(addrAll, address)
	}

	for _, addr := range c.genesisAccounts {
		acctAddr, err := addr.keyInfo.GetAddress()
		s.Require().NoError(err)
		addrAll = append(addrAll, acctAddr)
	}

	err = modifyGenesis(val0ConfigDir, "", initBalanceStr, addrAll, initialGlobalFeeAmt+asedaDenom, asedaDenom)
	s.Require().NoError(err)

	// copy the genesis file to the remaining validators
	for _, val := range c.validators[1:] {
		_, err := copyFile(
			filepath.Join(val0ConfigDir, "config", "genesis.json"),
			filepath.Join(val.configDir(), "config", "genesis.json"),
		)
		s.Require().NoError(err)
	}
}

func (s *IntegrationTestSuite) initGenesis(c *chain) {
	var (
		serverCtx = server.NewDefaultContext()
		config    = serverCtx.Config
		validator = c.validators[0]
	)

	config.SetRoot(validator.configDir())
	config.Moniker = validator.moniker

	genFilePath := config.GenesisFile()
	appGenState, appGenesis, err := genutiltypes.GenesisStateFromGenFile(genFilePath)
	s.Require().NoError(err)

	var evidenceGenState evidencetypes.GenesisState
	s.Require().NoError(cdc.UnmarshalJSON(appGenState[evidencetypes.ModuleName], &evidenceGenState))

	evidenceGenState.Evidence = make([]*codectypes.Any, numberOfEvidences)
	for i := range evidenceGenState.Evidence {
		pk := ed25519.GenPrivKey()
		evidence := &evidencetypes.Equivocation{
			Height:           1,
			Power:            100,
			Time:             time.Now().UTC(),
			ConsensusAddress: sdk.ConsAddress(pk.PubKey().Address().Bytes()).String(),
		}
		evidenceGenState.Evidence[i], err = codectypes.NewAnyWithValue(evidence)
		s.Require().NoError(err)
	}

	appGenState[evidencetypes.ModuleName], err = cdc.MarshalJSON(&evidenceGenState)
	s.Require().NoError(err)

	var genUtilGenState genutiltypes.GenesisState
	s.Require().NoError(cdc.UnmarshalJSON(appGenState[genutiltypes.ModuleName], &genUtilGenState))

	// generate genesis txs
	genTxs := make([]json.RawMessage, len(c.validators))
	for i, val := range c.validators {
		selfDelegationCoin, err := sdk.ParseCoinNormalized(selfDelegationStr)
		s.Require().NoError(err)

		createValmsg, err := val.buildCreateValidatorMsg(selfDelegationCoin)
		s.Require().NoError(err)

		signedTx, err := val.signMsg(createValmsg)
		s.Require().NoError(err)

		txRaw, err := cdc.MarshalJSON(signedTx)
		s.Require().NoError(err)

		genTxs[i] = txRaw
	}

	genUtilGenState.GenTxs = genTxs

	appGenState[genutiltypes.ModuleName], err = cdc.MarshalJSON(&genUtilGenState)
	s.Require().NoError(err)

	appGenesis.AppState, err = json.MarshalIndent(appGenState, "", "  ")
	s.Require().NoError(err)

	err = genutil.ExportGenesisFile(appGenesis, genFilePath)
	s.Require().NoError(err)

	err = appGenesis.ValidateAndComplete()
	s.Require().NoError(err)

	bz, err := json.MarshalIndent(appGenesis, "", "  ")
	s.Require().NoError(err)

	// write the updated genesis file to each validator.
	for _, val := range c.validators {
		err = writeFile(filepath.Join(val.configDir(), "config", "genesis.json"), bz)
		s.Require().NoError(err)
	}
}

// initValidatorConfigs initializes the validator configs for the given chain.
func (s *IntegrationTestSuite) initValidatorConfigs(c *chain) {
	for i, val := range c.validators {
		tmCfgPath := filepath.Join(val.configDir(), "config", "config.toml")

		vpr := viper.New()
		vpr.SetConfigFile(tmCfgPath)
		s.Require().NoError(vpr.ReadInConfig())

		valConfig := tmconfig.DefaultConfig()

		s.Require().NoError(vpr.Unmarshal(valConfig))

		valConfig.P2P.ListenAddress = "tcp://0.0.0.0:26656"
		valConfig.P2P.AddrBookStrict = false
		valConfig.P2P.ExternalAddress = fmt.Sprintf("%s:%d", val.instanceName(), 26656)
		valConfig.RPC.ListenAddress = "tcp://0.0.0.0:26657"
		valConfig.StateSync.Enable = false
		valConfig.LogLevel = "info"

		var peers []string

		for j := 0; j < len(c.validators); j++ {
			if i == j {
				continue
			}

			peer := c.validators[j]
			peerID := fmt.Sprintf("%s@%s%d:26656", peer.nodeKey.ID(), peer.moniker, j)
			peers = append(peers, peerID)
		}

		valConfig.P2P.PersistentPeers = strings.Join(peers, ",")

		tmconfig.WriteConfigFile(tmCfgPath, valConfig)

		// set application configuration
		appCfgPath := filepath.Join(val.configDir(), "config", "app.toml")

		appConfig := srvconfig.DefaultConfig()
		appConfig.API.Enable = true
		appConfig.MinGasPrices = fmt.Sprintf("%s%s", minGasPrice, asedaDenom)
		appConfig.API.Address = "tcp://0.0.0.0:1317"

		srvconfig.SetConfigTemplate(srvconfig.DefaultConfigTemplate)
		srvconfig.WriteConfigFile(appCfgPath, appConfig)
	}
}

// runValidators runs the validators in the chain
func (s *IntegrationTestSuite) runValidators(c *chain, portOffset int) {
	s.T().Logf("starting %s validator containers...", c.id)

	s.valResources[c.id] = make([]*dockertest.Resource, len(c.validators))
	for i, val := range c.validators {
		runOpts := &dockertest.RunOptions{
			Name:      val.instanceName(),
			NetworkID: s.dkrNet.Network.ID,
			Mounts: []string{
				fmt.Sprintf("%s/:%s", val.configDir(), containerChainHomePath),
			},
			Repository: "sedaprotocol/seda-chaind-e2e",
		}

		s.Require().NoError(exec.Command("chmod", "-R", "0777", val.configDir()).Run()) //nolint:gosec // this is a test

		// expose the first validator for debugging and communication
		if val.index == 0 {
			runOpts.PortBindings = map[docker.Port][]docker.PortBinding{
				"1317/tcp":  {{HostIP: "", HostPort: fmt.Sprintf("%d", 1317+portOffset)}},
				"6060/tcp":  {{HostIP: "", HostPort: fmt.Sprintf("%d", 6060+portOffset)}},
				"6061/tcp":  {{HostIP: "", HostPort: fmt.Sprintf("%d", 6061+portOffset)}},
				"6062/tcp":  {{HostIP: "", HostPort: fmt.Sprintf("%d", 6062+portOffset)}},
				"6063/tcp":  {{HostIP: "", HostPort: fmt.Sprintf("%d", 6063+portOffset)}},
				"6064/tcp":  {{HostIP: "", HostPort: fmt.Sprintf("%d", 6064+portOffset)}},
				"6065/tcp":  {{HostIP: "", HostPort: fmt.Sprintf("%d", 6065+portOffset)}},
				"9090/tcp":  {{HostIP: "", HostPort: fmt.Sprintf("%d", 9090+portOffset)}},
				"26656/tcp": {{HostIP: "", HostPort: fmt.Sprintf("%d", 26656+portOffset)}},
				"26657/tcp": {{HostIP: "", HostPort: fmt.Sprintf("%d", 26657+portOffset)}},
			}
		}

		resource, err := s.dkrPool.RunWithOptions(runOpts, noRestartAndMountWasmDir)
		s.Require().NoError(err)

		s.valResources[c.id][i] = resource
		s.T().Logf("started %s validator container: %s", c.id, resource.Container.ID)
	}

	s.endpoint = fmt.Sprintf("http://%s", s.valResources[s.chain.id][0].GetHostPort("1317/tcp"))

	rpcClient, err := rpchttp.New("tcp://localhost:26657", "/websocket")
	s.Require().NoError(err)

	s.Require().Eventually(
		func() bool {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()

			status, err := rpcClient.Status(ctx)
			if err != nil {
				return false
			}

			// let the node produce a few blocks
			if status.SyncInfo.CatchingUp || status.SyncInfo.LatestBlockHeight < 3 {
				return false
			}

			return true
		},
		5*time.Minute,
		time.Second,
		"Node has failed to produce blocks",
	)
}

func noRestartAndMountWasmDir(hc *docker.HostConfig) {
	hc.RestartPolicy = docker.RestartPolicy{
		Name: "no",
	}

	testWasmsDir, err := filepath.Abs(localWasmDirPath)
	if err != nil {
		panic(err)
	}

	hc.Mounts = append(hc.Mounts,
		docker.HostMount{
			Type:     "bind",
			Target:   containerWasmDirPath,
			Source:   testWasmsDir,
			ReadOnly: true,
		},
	)
}
