package e2e

var runWasmStorageTest = true

func (s *IntegrationTestSuite) TestWasmStorage() {
	if !runWasmStorageTest {
		s.T().Skip()
	}
	s.testWasmStorageStoreOracleProgram()
	s.testInstantiateCoreContract() // involves gov process
}
