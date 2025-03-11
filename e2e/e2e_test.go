package e2e

var (
	runWasmStorageTest = false
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
	s.testUnbond()           // to trigger batch creation and signing, and ensuring there are batches for the rotation test
	s.testBatchKeyRotation() // to verify that when > 1/3 of the voting power rotates the signing process continues as expected
}
