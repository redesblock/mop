package urlsigner

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"net/url"
	"strconv"
	"time"
)

// SignerProvider is used to sign and verify urls.
type SignerProvider struct {
	secretKey string
	sigField  string
	expField  string
	nowFn     func() time.Time
	algorithm func() hash.Hash
}

// Sign will sign an URL object returning it updated to include the signature
func (p *SignerProvider) Sign(u url.URL) url.URL {
	signature := Sign(p.algorithm, p.secretKey, u.String())

	q := u.Query()
	q.Set(p.sigField, signature)
	u.RawQuery = q.Encode()

	return u
}

// SignTemporary will sign an URL object for a limited period of time retuning
// it updated with two new query strings: signature, expiration
func (p *SignerProvider) SignTemporary(u url.URL, expireAt time.Time) url.URL {
	q := u.Query()
	q.Set(p.expField, strconv.FormatInt(expireAt.Unix(), 10))
	u.RawQuery = q.Encode()

	return p.Sign(u)
}

// Verify will check an URL object against its signature
// This signature should be provided by the url itself in a query string
func (p *SignerProvider) Verify(u url.URL) bool {
	q := u.Query()
	signature := q.Get(p.sigField)
	q.Del(p.sigField)
	u.RawQuery = q.Encode()

	computedSignature := Sign(p.algorithm, p.secretKey, u.String())

	return Verify(signature, computedSignature)
}

// VerifyTemporary will check an URL object against its signature and check if it's expired
// This signature should be provided by the url itself in a query string
func (p *SignerProvider) VerifyTemporary(u url.URL) bool {
	q := u.Query()
	signature := q.Get(p.sigField)
	exp, _ := strconv.ParseInt(q.Get(p.expField), 10, 64)
	q.Del(p.sigField)
	u.RawQuery = q.Encode()

	computedSignature := Sign(p.algorithm, p.secretKey, u.String())

	if !Verify(signature, computedSignature) {
		return false
	}

	return time.Unix(exp, 0).After(p.nowFn())
}

// SignURL acts like Sign method but accepts the url as string instead of url.URL
func (p *SignerProvider) SignURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	newURL := p.Sign(*u)

	return newURL.String()
}

// SignTemporaryURL acts like SignTemporary method but accepts the url as string instead of url.URL
func (p *SignerProvider) SignTemporaryURL(rawURL string, expireAt time.Time) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	newURL := p.SignTemporary(*u, expireAt)

	return newURL.String()
}

// VerifyURL acts like Verify method but accepts the url as string instead of url.URL
func (p *SignerProvider) VerifyURL(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	return p.Verify(*u)
}

// VerifyTemporaryURL acts like VerifyTemporary method but accepts the url as string instead of url.URL
func (p *SignerProvider) VerifyTemporaryURL(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	return p.VerifyTemporary(*u)
}

// New will create a new SignerProvider.
//
//	urlsigner.New("secret-key")
func New(secretKey string, opts ...func(*SignerProvider)) *SignerProvider {
	provider := &SignerProvider{
		secretKey: secretKey,
		expField:  "exp",
		sigField:  "sig",
		algorithm: sha256.New,
		nowFn: func() time.Time {
			return time.Now().UTC()
		},
	}

	for _, opt := range opts {
		opt(provider)
	}

	return provider
}

// Sign will create a new signature based on a key and a string payload
// It is required to choose a hash algorithm
func Sign(algorithm func() hash.Hash, key, payload string) string {
	mac := hmac.New(algorithm, []byte(key))
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

// Verify will tell if two string signatures are equal
func Verify(a, b string) bool {
	mac1, err := hex.DecodeString(a)
	if err != nil {
		return false
	}

	mac2, err := hex.DecodeString(b)
	if err != nil {
		return false
	}

	return hmac.Equal(mac1, mac2)
}

// Algorithm allows overriding of the internal hashing algorithm
func Algorithm(alg func() hash.Hash) func(*SignerProvider) {
	return func(provider *SignerProvider) { provider.algorithm = alg }
}

// ExpirationField allows overriding of the internal field name for expiration
func ExpirationField(name string) func(*SignerProvider) {
	return func(provider *SignerProvider) {
		provider.expField = name
	}
}

// SignatureField allows overriding of the internal field name for signature
func SignatureField(name string) func(*SignerProvider) {
	return func(provider *SignerProvider) {
		provider.sigField = name
	}
}
