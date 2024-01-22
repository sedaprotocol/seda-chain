package e2e

var runWasmStorageTest = false

func (s *IntegrationTestSuite) TestWasmStorage() {
	if !runWasmStorageTest {
		s.T().Skip()
	}
	s.testWasmStorageStoreDataRequestWasm()
	s.testWasmStorageStoreOverlayWasm()         // involves gov process
	s.testInstantiateAndRegisterProxyContract() // involves gov process
}
