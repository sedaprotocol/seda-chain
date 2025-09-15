package types

import (
	"encoding/binary"
	"fmt"
	"strconv"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GasMeter stores the results of the canonical gas consumption calculations.
type GasMeter struct {
	proxies              []ProxyGasUsed
	executors            []ExecutorGasUsed
	ReducedPayout        bool
	tallyGasLimit        uint64
	tallyGasRemaining    uint64
	execGasLimit         uint64
	execGasRemaining     uint64
	totalProxyGasPerExec uint64
	postedGasPrice       math.Int // gas price as posted, can be higher than the GasPrice on the request
	poster               string
	escrow               math.Int
}

// NewGasMeter creates a new gas meter and incurs the base gas cost.
func NewGasMeter(dr *DataRequest, maxTallyGasLimit uint64, baseGasCost uint64) *GasMeter {
	gasMeter := &GasMeter{
		tallyGasLimit:     min(dr.TallyGasLimit, maxTallyGasLimit),
		tallyGasRemaining: min(dr.TallyGasLimit, maxTallyGasLimit),
		execGasLimit:      dr.ExecGasLimit,
		execGasRemaining:  dr.ExecGasLimit,
		postedGasPrice:    dr.PostedGasPrice,
		poster:            dr.Poster,
		escrow:            dr.Escrow,
	}

	// For normal operations we first check if the gas limit is enough to cover
	// the operation. For the base gas cost we always want to consume it as even
	// running out of gas incurs a cost to the network.
	if baseGasCost > gasMeter.tallyGasLimit {
		gasMeter.tallyGasRemaining = 0
	} else {
		gasMeter.ConsumeTallyGas(baseGasCost)
	}

	return gasMeter
}

func (g *GasMeter) GetPoster() string {
	return g.poster
}

func (g *GasMeter) GetEscrow() math.Int {
	return g.escrow
}

func (g GasMeter) TotalGasUsed() *math.Int {
	totalGasUsed := math.NewIntFromUint64(g.TallyGasUsed()).Add(math.NewIntFromUint64(g.ExecutionGasUsed()))
	return &totalGasUsed
}

func (g GasMeter) TallyGasUsed() uint64 {
	return g.tallyGasLimit - g.tallyGasRemaining
}

func (g GasMeter) ExecutionGasUsed() uint64 {
	return g.execGasLimit - g.execGasRemaining
}

func (g GasMeter) CorrectExecGasReportWithProxyGas(gasReport uint64) uint64 {
	if gasReport <= g.totalProxyGasPerExec {
		return 0
	}
	return gasReport - g.totalProxyGasPerExec
}

func (g GasMeter) RemainingTallyGas() uint64 {
	return g.tallyGasRemaining
}

func (g GasMeter) RemainingExecGas() uint64 {
	return g.execGasRemaining
}

func (g GasMeter) GasPrice() math.Int {
	return g.postedGasPrice
}

func (g *GasMeter) SetReducedPayoutMode() {
	g.ReducedPayout = true
}

// ConsumeTallyGas consumes tally gas as much as possible, returning true if
// the tally gas limit is not sufficient to cover the amount.
func (g *GasMeter) ConsumeTallyGas(amount uint64) bool {
	if amount > g.tallyGasRemaining {
		g.tallyGasRemaining = 0
		return true
	}

	g.tallyGasRemaining -= amount
	return false
}

// ConsumeExecGasForProxy consumes execution gas for data proxy payout and records
// the payout information. It returns true if the execution gas runs out during
// the process.
func (g *GasMeter) ConsumeExecGasForProxy(proxyPubkey, payoutAddr string, gasUsedPerExec, replicationFactor uint64) {
	amount := gasUsedPerExec * replicationFactor

	g.proxies = append(g.proxies, ProxyGasUsed{
		PayoutAddress: payoutAddr,
		PublicKey:     proxyPubkey,
		Amount:        math.NewIntFromUint64(amount),
	})

	if amount > g.execGasRemaining {
		g.execGasRemaining = 0
	} else {
		g.execGasRemaining -= amount
	}

	// It does not matter that totalProxyGasPerExec would not reflect the actual
	// execution gas consumption in the out of gas case because this would not
	// affect the executor gas report correction in CorrectExecGasReportWithProxyGas()
	// anyways.
	g.totalProxyGasPerExec += gasUsedPerExec
}

// ConsumeExecGasForExecutor consumes execution gas for executor payout and
// records the payout information. It returns true if the execution gas runs
// out during the process.
func (g *GasMeter) ConsumeExecGasForExecutor(executorPubKey string, amount uint64) {
	g.executors = append(g.executors, ExecutorGasUsed{
		PublicKey: executorPubKey,
		Amount:    math.NewIntFromUint64(amount),
	})

	if amount > g.execGasRemaining {
		g.execGasRemaining = 0
	} else {
		g.execGasRemaining -= amount
	}
}

// GetProxyGasUsed returns a list of gas used amounts by the data proxies.
// The list is sorted by their public keys with entropy from data request ID
// and block height.
func (g *GasMeter) GetProxyGasUsed(drID string, height int64) []ProxyGasUsed {
	return HashSort(g.proxies, GetEntropy(drID, height))
}

// GetExecutorGasUsed returns a list of gas used amounts by the executors. The
// list should already have been sorted with entropy by SanitizeReveals() except
// in the case where the number of commits is less than the replication factor.
func (g *GasMeter) GetExecutorGasUsed() []ExecutorGasUsed {
	return g.executors
}

// ReadGasMeter reads the given gas meter to construct a list of distributions.
func (g GasMeter) ReadGasMeter(ctx sdk.Context, drID string, drHeight uint64, burnRatio math.LegacyDec) []Distribution {
	dists := []Distribution{}
	attrs := []sdk.Attribute{
		sdk.NewAttribute(AttributeDataRequestID, drID),
		sdk.NewAttribute(AttributeDataRequestHeight, strconv.FormatUint(drHeight, 10)),
		sdk.NewAttribute(AttributeReducedPayout, strconv.FormatBool(g.ReducedPayout)),
	}

	// First distribution message is the combined burn.
	burn := NewBurn(math.NewIntFromUint64(g.TallyGasUsed()), g.GasPrice())
	dists = append(dists, burn)
	attrs = append(attrs, sdk.NewAttribute(AttributeTallyGas, strconv.FormatUint(g.TallyGasUsed(), 10)))

	// Append distribution messages for data proxies.
	for _, proxy := range g.GetProxyGasUsed(drID, ctx.BlockHeight()) {
		proxyDist := NewDataProxyReward(proxy.PublicKey, proxy.PayoutAddress, proxy.Amount, g.GasPrice())
		dists = append(dists, proxyDist)
		attrs = append(attrs, sdk.NewAttribute(AttributeDataProxyGas,
			fmt.Sprintf("%s,%s,%s", proxy.PublicKey, proxy.PayoutAddress, proxy.Amount.String())))
	}

	// Append distribution messages for executors, burning a portion of their
	// payouts in case of a reduced payout scenario.
	reducedPayoutBurn := math.ZeroInt()
	for _, executor := range g.GetExecutorGasUsed() {
		payoutAmt := executor.Amount
		if g.ReducedPayout {
			burnAmt := burnRatio.MulInt(executor.Amount).TruncateInt()
			payoutAmt = executor.Amount.Sub(burnAmt)
			reducedPayoutBurn = reducedPayoutBurn.Add(burnAmt)
		}

		executorDist := NewExecutorReward(executor.PublicKey, payoutAmt, g.GasPrice())
		dists = append(dists, executorDist)
		attrs = append(attrs, sdk.NewAttribute(AttributeExecutorGas,
			fmt.Sprintf("%s,%s", executor.PublicKey, payoutAmt.String())))
	}

	dists[0].Burn.Amount = dists[0].Burn.Amount.Add(reducedPayoutBurn.Mul(g.GasPrice()))
	attrs = append(attrs, sdk.NewAttribute(AttributeReducedPayoutBurn, reducedPayoutBurn.String()))

	ctx.EventManager().EmitEvent(sdk.NewEvent(EventTypeGasMeter, attrs...))

	return dists
}

var _ HashSortable = ProxyGasUsed{}

type ProxyGasUsed struct {
	PayoutAddress string
	PublicKey     string
	Amount        math.Int
}

func (p ProxyGasUsed) GetSortKey() []byte {
	return []byte(p.PublicKey)
}

var _ HashSortable = ExecutorGasUsed{}

type ExecutorGasUsed struct {
	PublicKey string
	Amount    math.Int
}

func (e ExecutorGasUsed) GetSortKey() []byte {
	return []byte(e.PublicKey)
}

func GetEntropy(drID string, height int64) []byte {
	heightBytes := make([]byte, 8)
	//nolint:gosec // G115: We shouldn't get negative block heights anyway.
	binary.BigEndian.PutUint64(heightBytes, uint64(height))
	return append([]byte(drID), heightBytes...)
}

type Distribution struct {
	Burn            *DistributionBurn            `json:"burn,omitempty"`
	ExecutorReward  *DistributionExecutorReward  `json:"executor_reward,omitempty"`
	DataProxyReward *DistributionDataProxyReward `json:"data_proxy_reward,omitempty"`
}

type DistributionBurn struct {
	Amount math.Int `json:"amount"`
}

type DistributionDataProxyReward struct {
	PayoutAddress string   `json:"payout_address"`
	Amount        math.Int `json:"amount"`
	// The public key of the data proxy as a hex string
	PublicKey string `json:"public_key"`
}

type DistributionExecutorReward struct {
	Amount math.Int `json:"amount"`
	// The public key of the executor as a hex string
	Identity string `json:"identity"`
}

func NewBurn(amount, gasPrice math.Int) Distribution {
	return Distribution{
		Burn: &DistributionBurn{Amount: amount.Mul(gasPrice)},
	}
}

func NewDataProxyReward(pubkey, payoutAddr string, amount, gasPrice math.Int) Distribution {
	return Distribution{
		DataProxyReward: &DistributionDataProxyReward{
			PayoutAddress: payoutAddr,
			Amount:        amount.Mul(gasPrice),
			PublicKey:     pubkey,
		},
	}
}

func NewExecutorReward(identity string, amount, gasPrice math.Int) Distribution {
	return Distribution{
		ExecutorReward: &DistributionExecutorReward{
			Identity: identity,
			Amount:   amount.Mul(gasPrice),
		},
	}
}
