package e2e

var (
	runBankTest        = true
	runWasmStorageTest = true
	runGovTest         = true
)

// func (s *IntegrationTestSuite) TestBank() {
// 	if !runBankTest {
// 		s.T().Skip()
// 	}
// 	s.testBankTokenTransfer()
// }

func (s *IntegrationTestSuite) TestWasmStorage() {
	if !runWasmStorageTest {
		s.T().Skip()
	}
	s.testWasmStorageStoreDataRequestWasm()
}

func (s *IntegrationTestSuite) TestGov() {
	if !runGovTest {
		s.T().Skip()
	}
	s.testWasmStorageStoreOverlayWasm()
}
