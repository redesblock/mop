package api

import "github.com/redesblock/hop/core/swarm"

type Server = server

type (
	BytesPostResponse  = bytesPostResponse
	FileUploadResponse = fileUploadResponse
	TagResponse        = tagResponse
	TagRequest         = tagRequest
)

var (
	ContentTypeTar = contentTypeTar
)

var (
	ErrNoResolver           = errNoResolver
	ErrInvalidNameOrAddress = errInvalidNameOrAddress
)

func (s *Server) ResolveNameOrAddress(str string) (swarm.Address, error) {
	return s.resolveNameOrAddress(str)
}
