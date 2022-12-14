package soc

var (
	ErrInvalidAddress = errInvalidAddress
	Hash              = hash
	RecoverAddress    = recoverAddress
)

// Signature returns the SOC signature.
func (s *SOC) Signature() []byte {
	return s.signature
}

// OwnerAddress returns the BNB Smart Chain of the SOC owner.
func (s *SOC) OwnerAddress() []byte {
	return s.owner
}

// ID returns the SOC id.
func (s *SOC) ID() []byte {
	return s.id
}
