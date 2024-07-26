package e2e

var runWasmStorageTest = false

func (s *IntegrationTestSuite) TestWasmStorage() {
	if !runWasmStorageTest {
		s.T().Skip()
	}
	s.testWasmStorageStoreDataRequestWasm()
	s.testWasmStorageStoreExecutorWasm() // involves gov process
	s.testInstantiateCoreContract()      // involves gov process
}
