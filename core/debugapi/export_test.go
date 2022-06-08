package debugapi

type (
	StatusResponse           = statusResponse
	PingpongResponse         = pingpongResponse
	PeerConnectResponse      = peerConnectResponse
	PeersResponse            = peersResponse
	AddressesResponse        = addressesResponse
	PinnedChunk              = pinnedChunk
	ListPinnedChunksResponse = listPinnedChunksResponse
	WelcomeMessageRequest    = welcomeMessageRequest
	WelcomeMessageResponse   = welcomeMessageResponse
	BalancesResponse         = balancesResponse
	BalanceResponse          = balanceResponse
)

var (
	ErrCantBalance   = errCantBalance
	ErrCantBalances  = errCantBalances
	ErrInvaliAddress = errInvaliAddress
)
