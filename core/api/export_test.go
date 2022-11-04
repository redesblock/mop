package api

import (
	"github.com/redesblock/mop/core/cluster"
	"github.com/redesblock/mop/core/log"
)

type (
	BytesPostResponse     = bytesPostResponse
	ChunkAddressResponse  = chunkAddressResponse
	SocPostResponse       = socPostResponse
	FeedReferenceResponse = feedReferenceResponse
	MopUploadResponse     = mopUploadResponse
	DebugTagResponse      = debugTagResponse
	TagRequest            = tagRequest
	ListTagsResponse      = listTagsResponse
	IsRetrievableResponse = isRetrievableResponse
	SecurityTokenResponse = securityTokenRsp
	SecurityTokenRequest  = securityTokenReq
)

var (
	InvalidContentType  = errInvalidContentType
	InvalidRequest      = errInvalidRequest
	DirectoryStoreError = errDirectoryStore
	EmptyDir            = errEmptyDir
)

var (
	ContentTypeTar    = contentTypeTar
	ContentTypeHeader = contentTypeHeader
)

var (
	ErrNoResolver           = errNoResolver
	ErrInvalidNameOrAddress = errInvalidNameOrAddress
)

var (
	FeedMetadataEntryOwner = feedMetadataEntryOwner
	FeedMetadataEntryTopic = feedMetadataEntryTopic
	FeedMetadataEntryType  = feedMetadataEntryType

	SuccessWsMsg = successWsMsg
)

var (
	FileSizeBucketsKBytes = fileSizeBucketsKBytes
	ToFileSizeBucket      = toFileSizeBucket
)

func (s *Service) ResolveNameOrAddress(str string) (cluster.Address, error) {
	return s.resolveNameOrAddress(str)
}

func CalculateNumberOfChunks(contentLength int64, isEncrypted bool) int64 {
	return calculateNumberOfChunks(contentLength, isEncrypted)
}

type (
	StatusResponse                    = statusResponse
	NodeResponse                      = nodeResponse
	PingpongResponse                  = pingpongResponse
	PeerConnectResponse               = peerConnectResponse
	PeersResponse                     = peersResponse
	AddressesResponse                 = addressesResponse
	WelcomeMessageRequest             = welcomeMessageRequest
	WelcomeMessageResponse            = welcomeMessageResponse
	BalancesResponse                  = balancesResponse
	PeerDataResponse                  = peerDataResponse
	PeerData                          = peerData
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
	VoucherCreateResponse             = voucherCreateResponse
	VoucherStampResponse              = voucherStampResponse
	VoucherStampsResponse             = voucherStampsResponse
	VoucherBatchResponse              = voucherBatchResponse
	VoucherStampBucketsResponse       = voucherStampBucketsResponse
	BucketData                        = bucketData
	WalletResponse                    = walletResponse
)

var (
	ErrCantBalance           = errCantBalance
	ErrCantBalances          = errCantBalances
	HttpErrGetAccountingInfo = httpErrGetAccountingInfo
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

type (
	LogRegistryIterateFn   func(fn func(string, string, log.Level, uint) bool)
	LogSetVerbosityByExpFn func(e string, v log.Level) error
)

var (
	LogRegistryIterate   = logRegistryIterate
	LogSetVerbosityByExp = logSetVerbosityByExp
)

func ReplaceLogRegistryIterateFn(fn LogRegistryIterateFn)   { logRegistryIterate = fn }
func ReplaceLogSetVerbosityByExp(fn LogSetVerbosityByExpFn) { logSetVerbosityByExp = fn }
