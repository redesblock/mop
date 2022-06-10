package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/redesblock/hop/core/collection/entry"
	"github.com/redesblock/hop/core/encryption"
	"github.com/redesblock/hop/core/file"
	"github.com/redesblock/hop/core/file/joiner"
	"github.com/redesblock/hop/core/jsonhttp"
	"github.com/redesblock/hop/core/manifest"
	"github.com/redesblock/hop/core/sctx"
	"github.com/redesblock/hop/core/swarm"
)

func (s *server) hopDownloadHandler(w http.ResponseWriter, r *http.Request) {
	targets := r.URL.Query().Get("targets")
	r = r.WithContext(sctx.SetTargets(r.Context(), targets))
	ctx := r.Context()

	addressHex := mux.Vars(r)["address"]
	path := mux.Vars(r)["path"]

	address, err := swarm.ParseHexAddress(addressHex)
	if err != nil {
		s.Logger.Debugf("hop download: parse address %s: %v", addressHex, err)
		s.Logger.Error("hop download: parse address")
		jsonhttp.BadRequest(w, "invalid address")
		return
	}

	toDecrypt := len(address.Bytes()) == (swarm.HashSize + encryption.KeyLength)

	// read manifest entry
	j := joiner.NewSimpleJoiner(s.Storer)
	buf := bytes.NewBuffer(nil)
	_, err = file.JoinReadAll(ctx, j, address, buf, toDecrypt)
	if err != nil {
		s.Logger.Debugf("hop download: read entry %s: %v", address, err)
		s.Logger.Errorf("hop download: read entry %s", address)
		jsonhttp.NotFound(w, nil)
		return
	}
	e := &entry.Entry{}
	err = e.UnmarshalBinary(buf.Bytes())
	if err != nil {
		s.Logger.Debugf("hop download: unmarshal entry %s: %v", address, err)
		s.Logger.Errorf("hop download: unmarshal entry %s", address)
		jsonhttp.InternalServerError(w, "error unmarshaling entry")
		return
	}

	// read metadata
	buf = bytes.NewBuffer(nil)
	_, err = file.JoinReadAll(ctx, j, e.Metadata(), buf, toDecrypt)
	if err != nil {
		s.Logger.Debugf("hop download: read metadata %s: %v", address, err)
		s.Logger.Errorf("hop download: read metadata %s", address)
		jsonhttp.NotFound(w, nil)
		return
	}
	manifestMetadata := &entry.Metadata{}
	err = json.Unmarshal(buf.Bytes(), manifestMetadata)
	if err != nil {
		s.Logger.Debugf("hop download: unmarshal metadata %s: %v", address, err)
		s.Logger.Errorf("hop download: unmarshal metadata %s", address)
		jsonhttp.InternalServerError(w, "error unmarshaling metadata")
		return
	}

	// we are expecting manifest Mime type here
	m, err := manifest.NewManifestReference(
		ctx,
		manifestMetadata.MimeType,
		e.Reference(),
		toDecrypt,
		s.Storer,
	)
	if err != nil {
		s.Logger.Debugf("hop download: not manifest %s: %v", address, err)
		s.Logger.Error("hop download: not manifest")
		jsonhttp.BadRequest(w, "not manifest")
		return
	}

	me, err := m.Lookup(path)
	if err != nil {
		s.Logger.Debugf("hop download: invalid path %s/%s: %v", address, path, err)
		s.Logger.Error("hop download: invalid path")

		if errors.Is(err, manifest.ErrNotFound) {
			jsonhttp.NotFound(w, "path address not found")
		} else {
			jsonhttp.BadRequest(w, "invalid path address")
		}
		return
	}

	manifestEntryAddress := me.Reference()

	// read file entry
	buf = bytes.NewBuffer(nil)
	_, err = file.JoinReadAll(ctx, j, manifestEntryAddress, buf, toDecrypt)
	if err != nil {
		s.Logger.Debugf("hop download: read file entry %s: %v", address, err)
		s.Logger.Errorf("hop download: read file entry %s", address)
		jsonhttp.NotFound(w, nil)
		return
	}
	fe := &entry.Entry{}
	err = fe.UnmarshalBinary(buf.Bytes())
	if err != nil {
		s.Logger.Debugf("hop download: unmarshal file entry %s: %v", address, err)
		s.Logger.Errorf("hop download: unmarshal file entry %s", address)
		jsonhttp.InternalServerError(w, "error unmarshaling file entry")
		return
	}

	// read file metadata
	buf = bytes.NewBuffer(nil)
	_, err = file.JoinReadAll(ctx, j, fe.Metadata(), buf, toDecrypt)
	if err != nil {
		s.Logger.Debugf("hop download: read file metadata %s: %v", address, err)
		s.Logger.Errorf("hop download: read file metadata %s", address)
		jsonhttp.NotFound(w, nil)
		return
	}
	fileMetadata := &entry.Metadata{}
	err = json.Unmarshal(buf.Bytes(), fileMetadata)
	if err != nil {
		s.Logger.Debugf("hop download: unmarshal metadata %s: %v", address, err)
		s.Logger.Errorf("hop download: unmarshal metadata %s", address)
		jsonhttp.InternalServerError(w, "error unmarshaling metadata")
		return
	}

	additionalHeaders := http.Header{
		"Content-Disposition": {fmt.Sprintf("inline; filename=\"%s\"", fileMetadata.Filename)},
		"Content-Type":        {fileMetadata.MimeType},
	}

	fileEntryAddress := fe.Reference()

	s.downloadHandler(w, r, fileEntryAddress, additionalHeaders)
}
