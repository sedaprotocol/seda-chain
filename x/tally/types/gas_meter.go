package types

import (
	"cosmossdk.io/math"
)

// GasMeter stores the results of the canonical gas consumption calculations.
type GasMeter struct {
	Burn              uint64 // filter and tally gas to be burned
	Proxies           []ProxyGasUsed
	Executors         []ExecutorGasUsed
	ReducedPayout     bool
	tallyGasLimit     uint64
	tallyGasRemaining uint64 // TODO to be burned
	execGasLimit      uint64
	execGasRemaining  uint64
	gasPrice          math.Int // gas price for the request
}

type ProxyGasUsed struct {
	PublicKey []byte
	Amount    math.Int
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

	// Consume base gas cost, ignoring even if gas runs out.
	gasMeter.ConsumeTallyGas(baseGasCost)
	return gasMeter
}

func (g GasMeter) TotalGasUsed() uint64 {
	return g.TallyGasUsed() + g.ExecutionGasUsed()
}

func (g GasMeter) TallyGasUsed() uint64 {
	return g.tallyGasLimit - g.tallyGasRemaining
}

func (g GasMeter) ExecutionGasUsed() uint64 {
	return g.execGasLimit - g.execGasRemaining
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
// the tally gas runs out during the process. Consumed tally gas gets burned.
func (g *GasMeter) ConsumeTallyGas(amount uint64) bool {
	if amount > g.tallyGasRemaining {
		g.tallyGasRemaining = 0
		g.Burn += amount
		return true
	}
	g.tallyGasRemaining -= amount
	g.Burn += amount
	return false
}

// ConsumeExecGasForProxy consumes execution gas for data proxy payout and records
// the payout information. It returns true if the execution gas runs out during
// the process.
func (g *GasMeter) ConsumeExecGasForProxy(proxyPubKey []byte, amount uint64) bool {
	g.Proxies = append(g.Proxies, ProxyGasUsed{
		PublicKey: proxyPubKey,
		Amount:    math.NewIntFromUint64(amount),
	})

	if amount > g.execGasRemaining {
		g.execGasRemaining = 0
		return true
	}
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
