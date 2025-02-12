package types

import (
	"cosmossdk.io/math"
)

// GasMeter stores the results of the canonical gas consumption calculations.
type GasMeter struct {
	Proxies              []ProxyGasUsed
	Executors            []ExecutorGasUsed
	ReducedPayout        bool
	tallyGasLimit        uint64
	tallyGasRemaining    uint64
	execGasLimit         uint64
	execGasRemaining     uint64
	totalProxyGasPerExec uint64
	gasPrice             math.Int // gas price for the request
}

type ProxyGasUsed struct {
	PayoutAddress string
	PublicKey     string
	Amount        math.Int
}

type ExecutorGasUsed struct {
	PublicKey string
	Amount    math.Int
}

// NewGasMeter creates a new gas meter and incurs the base gas cost.
func NewGasMeter(tallyGasLimit, execGasLimit, maxTallyGasLimit uint64, gasPrice math.Int, baseGasCost uint64) *GasMeter {
	gasMeter := &GasMeter{
		tallyGasLimit:     min(tallyGasLimit, maxTallyGasLimit),
		tallyGasRemaining: min(tallyGasLimit, maxTallyGasLimit),
		execGasLimit:      execGasLimit,
		execGasRemaining:  execGasLimit,
		gasPrice:          gasPrice,
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
	if gasReport < g.totalProxyGasPerExec {
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
	return g.gasPrice
}

func (g *GasMeter) SetReducedPayoutMode() {
	g.ReducedPayout = true
}

// ConsumeTallyGas consumes tally gas as much as possible, returning true if
// the tally gas limit is not sufficient to cover the amount.
func (g *GasMeter) ConsumeTallyGas(amount uint64) bool {
	if amount > g.tallyGasRemaining {
		return true
	}

	g.tallyGasRemaining -= amount
	return false
}

// ConsumeExecGasForProxy consumes execution gas for data proxy payout and records
// the payout information. It returns true if the execution gas runs out during
// the process.
func (g *GasMeter) ConsumeExecGasForProxy(proxyPubkey, payoutAddr string, gasUsedPerExec uint64, replicationFactor uint16) bool {
	amount := gasUsedPerExec * uint64(replicationFactor)

	g.Proxies = append(g.Proxies, ProxyGasUsed{
		PayoutAddress: payoutAddr,
		PublicKey:     proxyPubkey,
		Amount:        math.NewIntFromUint64(amount),
	})

	if amount > g.execGasRemaining {
		g.execGasRemaining = 0
		return true
	}

	g.totalProxyGasPerExec += gasUsedPerExec
	g.execGasRemaining -= amount
	return false
}

// ConsumeExecGasForExecutor consumes execution gas for executor payout and
// records the payout information. It returns true if the execution gas runs
// out during the process.
func (g *GasMeter) ConsumeExecGasForExecutor(executorPubKey string, amount uint64) bool {
	g.Executors = append(g.Executors, ExecutorGasUsed{
		PublicKey: executorPubKey,
		Amount:    math.NewIntFromUint64(amount),
	})

	if amount > g.execGasRemaining {
		g.execGasRemaining = 0
		return true
	}
	g.execGasRemaining -= amount
	return false
}
