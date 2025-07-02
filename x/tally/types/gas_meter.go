package types

import (
	"encoding/binary"

	"cosmossdk.io/math"
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

// NewGasMeter creates a new gas meter and incurs the base gas cost.
func NewGasMeter(tallyGasLimit, execGasLimit, maxTallyGasLimit uint64, postedGasPrice math.Int, baseGasCost uint64) *GasMeter {
	gasMeter := &GasMeter{
		tallyGasLimit:     min(tallyGasLimit, maxTallyGasLimit),
		tallyGasRemaining: min(tallyGasLimit, maxTallyGasLimit),
		execGasLimit:      execGasLimit,
		execGasRemaining:  execGasLimit,
		postedGasPrice:    postedGasPrice,
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
func (g *GasMeter) ConsumeExecGasForProxy(proxyPubkey, payoutAddr string, gasUsedPerExec uint64, replicationFactor uint16) {
	amount := gasUsedPerExec * uint64(replicationFactor)

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

func GetEntropy(drID string, height int64) []byte {
	heightBytes := make([]byte, 8)
	//nolint:gosec // G115: We shouldn't get negative block heights anyway.
	binary.BigEndian.PutUint64(heightBytes, uint64(height))
	return append([]byte(drID), heightBytes...)
}
