package e2e

var (
	runWasmStorageTest = true
)

func (s *IntegrationTestSuite) TestWasmStorage() {
	if !runWasmStorageTest {
		s.T().Skip()
	}
	s.testWasmStorageStoreDataRequestWasm()
	s.testWasmStorageStoreOverlayWasm()         // involves gov process
	s.testInstantiateAndRegisterProxyContract() // involves gov process
}
