package e2e

var runWasmStorageTest = false

func (s *IntegrationTestSuite) TestWasmStorage() {
	if !runWasmStorageTest {
		s.T().Skip()
	}
	s.testWasmStorageStoreOracleProgram()
	s.testWasmStorageStoreExecutorWasm() // involves gov process
	s.testInstantiateCoreContract()      // involves gov process
}
