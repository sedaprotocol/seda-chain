package testwasms

import (
	_ "embed"

	"cosmossdk.io/math"
)

func BurnerWasmFileName() string {
	return "burner.wasm"
}

func ReflectWasmFileName() string {
	return "reflect.wasm"
}

func CoreContractWasmFileName() string {
	return "core_contract.wasm"
}

var (
	//go:embed core_contract.wasm
	coreContract []byte

	//go:embed core_contract_upgrade_ready.wasm
	coreContractUpgradeReadyWasm []byte

	//go:embed sample_tally.wasm
	sampleTallyWasm []byte

	//go:embed sample_tally_2.wasm
	sampleTallyWasm2 []byte

	//go:embed random_string_tally.wasm
	randomStringTallyWasm []byte

	//go:embed invalid_import.wasm
	invalidImportWasm []byte

	//go:embed chaos-dr.wasm
	chaosDrWasm []byte

	//go:embed data-proxy.wasm
	dataProxyWasm []byte

	//go:embed http-heavy.wasm
	httpHeavyWasm []byte

	//go:embed long-http.wasm
	longHTTPWasm []byte

	//go:embed max-dr.wasm
	maxDrWasm []byte

	//go:embed max-result.wasm
	maxResultWasm []byte

	//go:embed memory.wasm
	memoryWasm []byte

	//go:embed mock-api.wasm
	mockAPIWasm []byte

	//go:embed price-feed.wasm
	priceFeedWasm []byte

	//go:embed random-number.wasm
	randomNumberWasm []byte

	//go:embed big.wasm
	bigWasm []byte

	//go:embed oversized.wasm
	oversizedWasm []byte

	//go:embed hello-world.wasm
	helloWorldWasm []byte
)

var TestWasms = [][]byte{
	SampleTallyWasm(),
	SampleTallyWasm2(),
	RandomStringTallyWasm(),
	InvalidImportWasm(),
	ChaosDrWasm(),
	DataProxyWasm(),
	HTTPHeavyWasm(),
	LongHTTPWasm(),
	MaxDrWasm(),
	MaxResultWasm(),
	MemoryWasm(),
	MockAPIWasm(),
	PriceFeedWasm(),
	RandomNumberWasm(),
	BigWasm(),
	HelloWorldWasm(),
}

var TestWasmNames = []string{
	"sample_tally",
	"sample_tally_2",
	"random_string_tally",
	"invalid_import",
	"chaos_dr",
	"data_proxy",
	"http_heavy",
	"long_http",
	"max_dr",
	"max_result",
	"memory",
	"mock_api",
	"price_feed",
	"random_number",
	"big",
	"hello_world",
}

// v1.0.16 Core Contract with commit/reveal refund tx call removed
func CoreContractWasm() []byte {
	return coreContract
}

// v1.0.16 Core Contract with drain data request pool tx
// and without commit/reveal refund tx call
func CoreContractUpgradeReadyWasm() []byte {
	return coreContractUpgradeReadyWasm
}

// SampleTallyWasm returns a sample tally wasm binary whose Keccak256
// hash is 8ade60039246740faa80bf424fc29e79fe13b32087043e213e7bc36620111f6b.
func SampleTallyWasm() []byte {
	return sampleTallyWasm
}

// SampleTallyWasm2 returns a tally wasm binary that prints all the
// tally VM environment variables.
// Hash: 5f3b31bff28c64a143119ee6389d62e38767672daace9c36db54fa2d18e9f391
func SampleTallyWasm2() []byte {
	return sampleTallyWasm2
}

// RandomStringTallyWasm returns a tally wasm binary that concatenates the
// strings found in the 'value' attribute of the reveals. The random
// element from the name only applies to the execution phase, not the tally.
// taken from https://github.com/sedaprotocol/dr-playground 8b40cbc
// Hash: dae38b4ed8c00031a12c1cd506b8f4949b3a314720939f7b5400d1f6b9978337
func RandomStringTallyWasm() []byte {
	return randomStringTallyWasm
}

// InvalidImportsWasm returns a tally wasm binary that has invalid imports.
// Custom build with a non-existent import.
// Hash: 18d962cac6d9f931546f111da8d11b3ed54ccd77b79637b1c49a86ead16edb78
func InvalidImportWasm() []byte {
	return invalidImportWasm
}

func ChaosDrWasm() []byte {
	return chaosDrWasm
}

func DataProxyWasm() []byte {
	return dataProxyWasm
}

func HTTPHeavyWasm() []byte {
	return httpHeavyWasm
}

func LongHTTPWasm() []byte {
	return longHTTPWasm
}

func MaxDrWasm() []byte {
	return maxDrWasm
}

func MaxResultWasm() []byte {
	return maxResultWasm
}

func MemoryWasm() []byte {
	return memoryWasm
}

func MockAPIWasm() []byte {
	return mockAPIWasm
}

func PriceFeedWasm() []byte {
	return priceFeedWasm
}

func RandomNumberWasm() []byte {
	return randomNumberWasm
}

// OversizedWasm returns a wasm that is over 1MB in size.
func OversizedWasm() []byte {
	return oversizedWasm
}

// BigWasm returns a wasm that is slightly less than 1MB in size.
func BigWasm() []byte {
	return bigWasm
}

func HelloWorldWasm() []byte {
	return helloWorldWasm
}

// TallyTestItem is a bundle of tally program, reveal, and expected VM result.
type TallyTestItem struct {
	TallyProgram []byte
	Reveal       []byte
	GasUsed      uint64

	// Expected values of data result:
	ExpectedExitCode uint32
	ExpectedResult   []byte
	ExpectedGasUsed  math.Int
}

var TallyTestItems = []TallyTestItem{
	TallyTestItemSampleTally(),
	TallyTestItemRandomString(),
	TallyTestItemRandomNumber(),
	TallyTestInvalidImport(),
}

func TallyTestItemSampleTally() TallyTestItem {
	return TallyTestItem{
		TallyProgram:     sampleTallyWasm,
		Reveal:           []byte("{\"value\":\"one\"}"),
		GasUsed:          150000000000000000,
		ExpectedResult:   []byte("tally_inputs__REST__0"),
		ExpectedExitCode: 0,
		ExpectedGasUsed:  math.NewInt(100013725103196250),
	}
}

func TallyTestItemRandomString() TallyTestItem {
	return TallyTestItem{
		TallyProgram:     randomStringTallyWasm,
		Reveal:           []byte("{\"value\":\"one\"}"),
		GasUsed:          150000000000000000,
		ExpectedResult:   []byte("one"),
		ExpectedExitCode: 0,
		ExpectedGasUsed:  math.NewInt(100014761092378750),
	}
}

func TallyTestItemRandomNumber() TallyTestItem {
	return TallyTestItem{
		TallyProgram:     randomNumberWasm,
		Reveal:           []byte("{\"value\":\"one\"}"),
		GasUsed:          150000000000000000,
		ExpectedResult:   []byte("Not ok"),
		ExpectedExitCode: 1,
		ExpectedGasUsed:  math.NewInt(100017279108241250),
	}
}

func TallyTestInvalidImport() TallyTestItem {
	return TallyTestItem{
		TallyProgram:     invalidImportWasm,
		Reveal:           nil,
		GasUsed:          150000000000000000,
		ExpectedResult:   []byte("Error: Failed to create WASMER instance: Error while importing \"seda_v1\".\"this_does_not_exist\": unknown import. Expected Function(FunctionType { params: [], results: [] })"),
		ExpectedExitCode: 4,
		ExpectedGasUsed:  math.NewInt(100006000002140000),
	}
}
