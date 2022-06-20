package api

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"

	"github.com/ethersphere/manifest/mantaray"
	"github.com/redesblock/hop/core/collection/entry"
	"github.com/redesblock/hop/core/feeds"
	"github.com/redesblock/hop/core/file"
	"github.com/redesblock/hop/core/file/joiner"
	"github.com/redesblock/hop/core/file/loadsave"
	"github.com/redesblock/hop/core/jsonhttp"
	"github.com/redesblock/hop/core/manifest"
	"github.com/redesblock/hop/core/sctx"
	"github.com/redesblock/hop/core/storage"
	"github.com/redesblock/hop/core/swarm"
	"github.com/redesblock/hop/core/tracing"
)

func (s *server) hopDownloadHandler(w http.ResponseWriter, r *http.Request) {
	logger := tracing.NewLoggerWithTraceID(r.Context(), s.logger)
	ls := loadsave.New(s.storer, storage.ModePutRequest, false)
	feedDereferenced := false

	targets := r.URL.Query().Get("targets")
	if targets != "" {
		r = r.WithContext(sctx.SetTargets(r.Context(), targets))
	}
	ctx := r.Context()

	nameOrHex := mux.Vars(r)["address"]
	pathVar := mux.Vars(r)["path"]
	if strings.HasSuffix(pathVar, "/") {
		pathVar = strings.TrimRight(pathVar, "/")
		// NOTE: leave one slash if there was some
		pathVar += "/"
	}

	address, err := s.resolveNameOrAddress(nameOrHex)
	if err != nil {
		logger.Debugf("hop download: parse address %s: %v", nameOrHex, err)
		logger.Error("hop download: parse address")
		jsonhttp.NotFound(w, nil)
		return
	}

FETCH:
	// read manifest entry
	j, _, err := joiner.New(ctx, s.storer, address)
	if err != nil {
		logger.Debugf("hop download: joiner manifest entry %s: %v", address, err)
		logger.Errorf("hop download: joiner %s", address)
		jsonhttp.NotFound(w, nil)
		return
	}

	buf := bytes.NewBuffer(nil)
	_, err = file.JoinReadAll(ctx, j, buf)
	if err != nil {
		logger.Debugf("hop download: read entry %s: %v", address, err)
		logger.Errorf("hop download: read entry %s", address)
		jsonhttp.NotFound(w, nil)
		return
	}

	// there's a possible ambiguity here, right now the data which was
	// read can be an entry.Entry or a mantaray feed manifest. Try to
	// unmarshal as mantaray first and possibly resolve the feed, otherwise
	// go on normally.
	if !feedDereferenced {
		if l, err := s.manifestFeed(ctx, ls, buf.Bytes()); err == nil {
			//we have a feed manifest here
			ch, cur, _, err := l.At(ctx, time.Now().Unix(), 0)
			if err != nil {
				logger.Debugf("hop download: feed lookup: %v", err)
				logger.Error("hop download: feed lookup")
				jsonhttp.NotFound(w, "feed not found")
				return
			}
			if ch == nil {
				logger.Debugf("hop download: feed lookup: no updates")
				logger.Error("hop download: feed lookup")
				jsonhttp.NotFound(w, "no update found")
				return
			}
			ref, _, err := parseFeedUpdate(ch)
			if err != nil {
				logger.Debugf("hop download: parse feed update: %v", err)
				logger.Error("hop download: parse feed update")
				jsonhttp.InternalServerError(w, "parse feed update")
				return
			}
			address = ref
			feedDereferenced = true
			curBytes, err := cur.MarshalBinary()
			if err != nil {
				s.logger.Debugf("hop download: marshal feed index: %v", err)
				s.logger.Error("hop download: marshal index")
				jsonhttp.InternalServerError(w, "marshal index")
				return
			}

			w.Header().Set(SwarmFeedIndexHeader, hex.EncodeToString(curBytes))
			// this header might be overriding others. handle with care. in the future
			// we should implement an append functionality for this specific header,
			// since different parts of handlers might be overriding others' values
			// resulting in inconsistent headers in the response.
			w.Header().Set("Access-Control-Expose-Headers", SwarmFeedIndexHeader)
			goto FETCH
		}
	}
	e := &entry.Entry{}
	err = e.UnmarshalBinary(buf.Bytes())
	if err != nil {
		logger.Debugf("hop download: unmarshal entry %s: %v", address, err)
		logger.Errorf("hop download: unmarshal entry %s", address)
		jsonhttp.NotFound(w, nil)
		return
	}

	// read metadata
	j, _, err = joiner.New(ctx, s.storer, e.Metadata())
	if err != nil {
		logger.Debugf("hop download: joiner metadata %s: %v", address, err)
		logger.Errorf("hop download: joiner %s", address)
		jsonhttp.NotFound(w, nil)
		return
	}

	// read metadata
	buf = bytes.NewBuffer(nil)
	_, err = file.JoinReadAll(ctx, j, buf)
	if err != nil {
		logger.Debugf("hop download: read metadata %s: %v", address, err)
		logger.Errorf("hop download: read metadata %s", address)
		jsonhttp.NotFound(w, nil)
		return
	}
	manifestMetadata := &entry.Metadata{}
	err = json.Unmarshal(buf.Bytes(), manifestMetadata)
	if err != nil {
		logger.Debugf("hop download: unmarshal metadata %s: %v", address, err)
		logger.Errorf("hop download: unmarshal metadata %s", address)
		jsonhttp.NotFound(w, nil)
		return
	}

	// we are expecting manifest Mime type here
	m, err := manifest.NewManifestReference(
		manifestMetadata.MimeType,
		e.Reference(),
		ls,
	)
	if err != nil {
		logger.Debugf("hop download: not manifest %s: %v", address, err)
		logger.Error("hop download: not manifest")
		jsonhttp.NotFound(w, nil)
		return
	}

	if pathVar == "" {
		logger.Tracef("hop download: handle empty path %s", address)

		if indexDocumentSuffixKey, ok := manifestMetadataLoad(ctx, m, manifestRootPath, manifestWebsiteIndexDocumentSuffixKey); ok {
			pathWithIndex := path.Join(pathVar, indexDocumentSuffixKey)
			indexDocumentManifestEntry, err := m.Lookup(ctx, pathWithIndex)
			if err == nil {
				// index document exists
				logger.Debugf("hop download: serving path: %s", pathWithIndex)

				s.serveManifestEntry(w, r, address, indexDocumentManifestEntry.Reference(), !feedDereferenced)
				return
			}
		}
	}

	me, err := m.Lookup(ctx, pathVar)
	if err != nil {
		logger.Debugf("hop download: invalid path %s/%s: %v", address, pathVar, err)
		logger.Error("hop download: invalid path")

		if errors.Is(err, manifest.ErrNotFound) {

			if !strings.HasPrefix(pathVar, "/") {
				// check for directory
				dirPath := pathVar + "/"
				exists, err := m.HasPrefix(ctx, dirPath)
				if err == nil && exists {
					// redirect to directory
					u := r.URL
					u.Path += "/"
					redirectURL := u.String()

					logger.Debugf("hop download: redirecting to %s: %v", redirectURL, err)

					http.Redirect(w, r, redirectURL, http.StatusPermanentRedirect)
					return
				}
			}

			// check index suffix path
			if indexDocumentSuffixKey, ok := manifestMetadataLoad(ctx, m, manifestRootPath, manifestWebsiteIndexDocumentSuffixKey); ok {
				if !strings.HasSuffix(pathVar, indexDocumentSuffixKey) {
					// check if path is directory with index
					pathWithIndex := path.Join(pathVar, indexDocumentSuffixKey)
					indexDocumentManifestEntry, err := m.Lookup(ctx, pathWithIndex)
					if err == nil {
						// index document exists
						logger.Debugf("hop download: serving path: %s", pathWithIndex)

						s.serveManifestEntry(w, r, address, indexDocumentManifestEntry.Reference(), !feedDereferenced)
						return
					}
				}
			}

			// check if error document is to be shown
			if errorDocumentPath, ok := manifestMetadataLoad(ctx, m, manifestRootPath, manifestWebsiteErrorDocumentPathKey); ok {
				if pathVar != errorDocumentPath {
					errorDocumentManifestEntry, err := m.Lookup(ctx, errorDocumentPath)
					if err == nil {
						// error document exists
						logger.Debugf("hop download: serving path: %s", errorDocumentPath)

						s.serveManifestEntry(w, r, address, errorDocumentManifestEntry.Reference(), !feedDereferenced)
						return
					}
				}
			}

			jsonhttp.NotFound(w, "path address not found")
		} else {
			jsonhttp.NotFound(w, nil)
		}
		return
	}

	// serve requested path
	s.serveManifestEntry(w, r, address, me.Reference(), !feedDereferenced)
}

func (s *server) serveManifestEntry(w http.ResponseWriter, r *http.Request, address, manifestEntryAddress swarm.Address, etag bool) {
	var (
		logger = tracing.NewLoggerWithTraceID(r.Context(), s.logger)
		ctx    = r.Context()
		buf    = bytes.NewBuffer(nil)
	)

	// read file entry
	j, _, err := joiner.New(ctx, s.storer, manifestEntryAddress)
	if err != nil {
		logger.Debugf("hop download: joiner read file entry %s: %v", address, err)
		logger.Errorf("hop download: joiner read file entry %s", address)
		jsonhttp.NotFound(w, nil)
		return
	}

	_, err = file.JoinReadAll(ctx, j, buf)
	if err != nil {
		logger.Debugf("hop download: read file entry %s: %v", address, err)
		logger.Errorf("hop download: read file entry %s", address)
		jsonhttp.NotFound(w, nil)
		return
	}
	fe := &entry.Entry{}
	err = fe.UnmarshalBinary(buf.Bytes())
	if err != nil {
		logger.Debugf("hop download: unmarshal file entry %s: %v", address, err)
		logger.Errorf("hop download: unmarshal file entry %s", address)
		jsonhttp.NotFound(w, nil)
		return
	}

	// read file metadata
	j, _, err = joiner.New(ctx, s.storer, fe.Metadata())
	if err != nil {
		logger.Debugf("hop download: joiner read file entry %s: %v", address, err)
		logger.Errorf("hop download: joiner read file entry %s", address)
		jsonhttp.NotFound(w, nil)
		return
	}

	buf = bytes.NewBuffer(nil)
	_, err = file.JoinReadAll(ctx, j, buf)
	if err != nil {
		logger.Debugf("hop download: read file metadata %s: %v", address, err)
		logger.Errorf("hop download: read file metadata %s", address)
		jsonhttp.NotFound(w, nil)
		return
	}
	fileMetadata := &entry.Metadata{}
	err = json.Unmarshal(buf.Bytes(), fileMetadata)
	if err != nil {
		logger.Debugf("hop download: unmarshal metadata %s: %v", address, err)
		logger.Errorf("hop download: unmarshal metadata %s", address)
		jsonhttp.NotFound(w, nil)
		return
	}

	additionalHeaders := http.Header{
		"Content-Disposition": {fmt.Sprintf("inline; filename=\"%s\"", fileMetadata.Filename)},
		"Content-Type":        {fileMetadata.MimeType},
	}

	fileEntryAddress := fe.Reference()

	s.downloadHandler(w, r, fileEntryAddress, additionalHeaders, etag)
}

// manifestMetadataLoad returns the value for a key stored in the metadata of
// manifest path, or empty string if no value is present.
// The ok result indicates whether value was found in the metadata.
func manifestMetadataLoad(ctx context.Context, manifest manifest.Interface, path, metadataKey string) (string, bool) {
	me, err := manifest.Lookup(ctx, path)
	if err != nil {
		return "", false
	}

	manifestRootMetadata := me.Metadata()
	if val, ok := manifestRootMetadata[metadataKey]; ok {
		return val, ok
	}

	return "", false
}

func (s *server) manifestFeed(ctx context.Context, ls file.LoadSaver, candidate []byte) (feeds.Lookup, error) {
	node := new(mantaray.Node)
	err := node.UnmarshalBinary(candidate)
	if err != nil {
		return nil, fmt.Errorf("node unmarshal: %w", err)
	}

	e, err := node.LookupNode(context.Background(), []byte("/"), ls)
	if err != nil {
		return nil, fmt.Errorf("node lookup: %w", err)
	}
	var (
		owner, topic []byte
		t            = new(feeds.Type)
	)
	meta := e.Metadata()
	if e := meta[feedMetadataEntryOwner]; e != "" {
		owner, err = hex.DecodeString(e)
		if err != nil {
			return nil, err
		}
	}
	if e := meta[feedMetadataEntryTopic]; e != "" {
		topic, err = hex.DecodeString(e)
		if err != nil {
			return nil, err
		}
	}
	if e := meta[feedMetadataEntryType]; e != "" {
		err := t.FromString(e)
		if err != nil {
			return nil, err
		}
	}
	f := feeds.New(topic, common.BytesToAddress(owner))
	return s.feedFactory.NewLookup(*t, f)
}
