package e2e

var (
	runWasmStorageTest = true
	runBatchingTest    = true
)

func (s *IntegrationTestSuite) TestWasmStorage() {
	if !runWasmStorageTest {
		s.T().Skip()
	}
	s.testWasmStorageStoreOracleProgram()
	s.testInstantiateCoreContract() // involves gov process
}

func (s *IntegrationTestSuite) TestBatching() {
	if !runBatchingTest {
		s.T().Skip()
	}
	s.testUnbond() // to trigger batch creation and signing
}
