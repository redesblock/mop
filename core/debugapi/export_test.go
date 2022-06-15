package debugapi

type (
	StatusResponse            = statusResponse
	PingpongResponse          = pingpongResponse
	PeerConnectResponse       = peerConnectResponse
	PeersResponse             = peersResponse
	AddressesResponse         = addressesResponse
	WelcomeMessageRequest     = welcomeMessageRequest
	WelcomeMessageResponse    = welcomeMessageResponse
	BalancesResponse          = balancesResponse
	BalanceResponse           = balanceResponse
	SettlementResponse        = settlementResponse
	SettlementsResponse       = settlementsResponse
	ChequebookBalanceResponse = chequebookBalanceResponse
	ChequebookAddressResponse = chequebookAddressResponse
)

var (
	ErrCantBalance         = errCantBalance
	ErrCantBalances        = errCantBalances
	ErrNoBalance           = errNoBalance
	ErrCantSettlementsPeer = errCantSettlementsPeer
	ErrCantSettlements     = errCantSettlements
	ErrChequebookBalance   = errChequebookBalance
	ErrInvaliAddress       = errInvaliAddress
)
