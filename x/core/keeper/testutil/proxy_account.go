package testutil

import (
	"encoding/hex"

	"cosmossdk.io/math"
	"github.com/cometbft/cometbft/crypto/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/crypto"
	dataproxytypes "github.com/sedaprotocol/seda-chain/x/data-proxy/types"
	"github.com/stretchr/testify/require"
)

type ProxyAccount struct {
	name       string
	addr       sdk.AccAddress
	privateKey secp256k1.PrivKey
	publicKey  secp256k1.PubKey
	fixture    *Fixture
}

func (pa *ProxyAccount) Name() string {
	return pa.name
}

func (pa *ProxyAccount) Address() string {
	return pa.addr.String()
}

func (pa *ProxyAccount) AccAddress() sdk.AccAddress {
	return pa.addr
}

func (pa *ProxyAccount) PublicKeyHex() string {
	return hex.EncodeToString(pa.publicKey.Bytes())
}

func (pa *ProxyAccount) Register(feeSeda int64, payoutAddress *string) {
	fee := sdk.NewCoin(BondDenom, pa.fixture.SedaToAseda(feeSeda))
	msg := &dataproxytypes.MsgRegisterDataProxy{
		AdminAddress:  pa.Address(),
		PayoutAddress: pa.Address(),
		Memo:          "",
		PubKey:        pa.PublicKeyHex(),
		Fee:           &fee,
	}

	if payoutAddress != nil {
		msg.PayoutAddress = *payoutAddress
	}

	// Generate the signature
	feeBytes := []byte(msg.Fee.String())
	adminAddressBytes := []byte(msg.AdminAddress)
	payoutAddressBytes := []byte(msg.PayoutAddress)
	memoBytes := []byte(msg.Memo)
	chainIDBytes := []byte(pa.fixture.Context().ChainID())

	payloadSize := len(feeBytes) + len(adminAddressBytes) + len(payoutAddressBytes) + len(memoBytes) + len(chainIDBytes)
	payload := make([]byte, 0, payloadSize)

	payload = append(payload, feeBytes...)
	payload = append(payload, adminAddressBytes...)
	payload = append(payload, payoutAddressBytes...)
	payload = append(payload, memoBytes...)
	payload = append(payload, chainIDBytes...)
	hash := crypto.Keccak256(payload)

	signature, err := pa.privateKey.Sign(hash)
	require.NoError(pa.fixture.tb, err)
	msg.Signature = hex.EncodeToString(signature)

	_, err = pa.fixture.DataProxyMsgServer.RegisterDataProxy(pa.fixture.Context(), msg)
	require.NoError(pa.fixture.tb, err)
}

func (pa *ProxyAccount) Balance() math.Int {
	return pa.fixture.BankKeeper.GetBalance(pa.fixture.Context(), pa.addr, BondDenom).Amount
}
