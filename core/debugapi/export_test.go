package debugapi

type (
	StatusResponse                    = statusResponse
	PingpongResponse                  = pingpongResponse
	PeerConnectResponse               = peerConnectResponse
	PeersResponse                     = peersResponse
	AddressesResponse                 = addressesResponse
	WelcomeMessageRequest             = welcomeMessageRequest
	WelcomeMessageResponse            = welcomeMessageResponse
	BalancesResponse                  = balancesResponse
	BalanceResponse                   = balanceResponse
	SettlementResponse                = settlementResponse
	SettlementsResponse               = settlementsResponse
	ChequebookBalanceResponse         = chequebookBalanceResponse
	ChequebookAddressResponse         = chequebookAddressResponse
	ChequebookLastChequePeerResponse  = chequebookLastChequePeerResponse
	ChequebookLastChequesResponse     = chequebookLastChequesResponse
	ChequebookLastChequesPeerResponse = chequebookLastChequesPeerResponse
	ChequebookTxResponse              = chequebookTxResponse
	SwapCashoutResponse               = swapCashoutResponse
	SwapCashoutStatusResponse         = swapCashoutStatusResponse
	SwapCashoutStatusResult           = swapCashoutStatusResult
	TransactionInfo                   = transactionInfo
	TransactionPendingList            = transactionPendingList
	TransactionHashResponse           = transactionHashResponse
	TagResponse                       = tagResponse
	ReserveStateResponse              = reserveStateResponse
	ChainStateResponse                = chainStateResponse
	PostageCreateResponse             = postageCreateResponse
	PostageVouchResponse              = postageVouchResponse
	PostageVouchsResponse             = postageVouchsResponse
	PostageVouchBucketsResponse       = postageVouchBucketsResponse
	BucketData                        = bucketData
)

var (
	ErrCantBalance           = errCantBalance
	ErrCantBalances          = errCantBalances
	ErrNoBalance             = errNoBalance
	ErrCantSettlementsPeer   = errCantSettlementsPeer
	ErrCantSettlements       = errCantSettlements
	ErrChequebookBalance     = errChequebookBalance
	ErrInvalidAddress        = errInvalidAddress
	ErrUnknownTransaction    = errUnknownTransaction
	ErrCantGetTransaction    = errCantGetTransaction
	ErrCantResendTransaction = errCantResendTransaction
	ErrAlreadyImported       = errAlreadyImported
)
