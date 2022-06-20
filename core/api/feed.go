package api

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	"github.com/redesblock/hop/core/feeds"
	"github.com/redesblock/hop/core/file/loadsave"
	"github.com/redesblock/hop/core/jsonhttp"
	"github.com/redesblock/hop/core/manifest"
	"github.com/redesblock/hop/core/soc"
	"github.com/redesblock/hop/core/swarm"
)

const (
	feedMetadataEntryOwner = "swarm-feed-owner"
	feedMetadataEntryTopic = "swarm-feed-topic"
	feedMetadataEntryType  = "swarm-feed-type"
)

var (
	errInvalidFeedUpdate = errors.New("invalid feed update")
)

type feedReferenceResponse struct {
	Reference swarm.Address `json:"reference"`
}

func (s *server) feedGetHandler(w http.ResponseWriter, r *http.Request) {
	owner, err := hex.DecodeString(mux.Vars(r)["owner"])
	if err != nil {
		s.Logger.Debugf("feed get: decode owner: %v", err)
		s.Logger.Error("feed get: bad owner")
		jsonhttp.BadRequest(w, "bad owner")
		return
	}

	topic, err := hex.DecodeString(mux.Vars(r)["topic"])
	if err != nil {
		s.Logger.Debugf("feed get: decode topic: %v", err)
		s.Logger.Error("feed get: bad topic")
		jsonhttp.BadRequest(w, "bad topic")
		return
	}

	var at int64
	atStr := r.URL.Query().Get("at")
	if atStr != "" {
		at, err = strconv.ParseInt(atStr, 10, 64)
		if err != nil {
			s.Logger.Debugf("feed get: decode at: %v", err)
			s.Logger.Error("feed get: bad at")
			jsonhttp.BadRequest(w, "bad at")
			return
		}
	} else {
		at = time.Now().Unix()
	}

	f := feeds.New(topic, common.BytesToAddress(owner))
	lookup, err := s.feedFactory.NewLookup(feeds.Sequence, f)
	if err != nil {
		s.Logger.Debugf("feed get: new lookup: %v", err)
		s.Logger.Error("feed get: new lookup")
		jsonhttp.InternalServerError(w, "new lookup")
		return
	}

	ch, cur, next, err := lookup.At(r.Context(), at, 0)
	if err != nil {
		s.Logger.Debugf("feed get: lookup: %v", err)
		s.Logger.Error("feed get: lookup error")
		jsonhttp.NotFound(w, "lookup failed")
		return
	}

	// KLUDGE: if a feed was never updated, the chunk will be nil
	if ch == nil {
		s.Logger.Debugf("feed get: no update found: %v", err)
		s.Logger.Error("feed get: no update found")
		jsonhttp.NotFound(w, "lookup failed")
		return
	}

	ref, _, err := parseFeedUpdate(ch)
	if err != nil {
		s.Logger.Debugf("feed get: parse update: %v", err)
		s.Logger.Error("feed get: parse update")
		jsonhttp.InternalServerError(w, "parse update")
		return
	}

	curBytes, err := cur.MarshalBinary()
	if err != nil {
		s.Logger.Debugf("feed get: marshal current index: %v", err)
		s.Logger.Error("feed get: marshal index")
		jsonhttp.InternalServerError(w, "marshal index")
		return
	}

	nextBytes, err := next.MarshalBinary()
	if err != nil {
		s.Logger.Debugf("feed get: marshal next index: %v", err)
		s.Logger.Error("feed get: marshal index")
		jsonhttp.InternalServerError(w, "marshal index")
		return
	}

	w.Header().Set(SwarmFeedIndexHeader, hex.EncodeToString(curBytes))
	w.Header().Set(SwarmFeedIndexNextHeader, hex.EncodeToString(nextBytes))
	w.Header().Set("Access-Control-Expose-Headers", fmt.Sprintf("%s, %s", SwarmFeedIndexHeader, SwarmFeedIndexNextHeader))

	jsonhttp.OK(w, feedReferenceResponse{Reference: ref})
}

func (s *server) feedPostHandler(w http.ResponseWriter, r *http.Request) {
	owner, err := hex.DecodeString(mux.Vars(r)["owner"])
	if err != nil {
		s.Logger.Debugf("feed put: decode owner: %v", err)
		s.Logger.Error("feed put: bad owner")
		jsonhttp.BadRequest(w, "bad owner")
		return
	}

	topic, err := hex.DecodeString(mux.Vars(r)["topic"])
	if err != nil {
		s.Logger.Debugf("feed put: decode topic: %v", err)
		s.Logger.Error("feed put: bad topic")
		jsonhttp.BadRequest(w, "bad topic")
		return
	}
	l := loadsave.New(s.Storer, requestModePut(r), false)
	feedManifest, err := manifest.NewDefaultManifest(l, false)
	if err != nil {
		s.Logger.Debugf("feed put: new manifest: %v", err)
		s.Logger.Error("feed put: new manifest")
		jsonhttp.InternalServerError(w, "create manifest")
		return
	}

	meta := map[string]string{
		feedMetadataEntryOwner: hex.EncodeToString(owner),
		feedMetadataEntryTopic: hex.EncodeToString(topic),
		feedMetadataEntryType:  feeds.Sequence.String(), // only sequence allowed for now
	}

	emptyAddr := make([]byte, 32)

	// a feed manifest stores the metadata at the root "/" path
	err = feedManifest.Add(r.Context(), "/", manifest.NewEntry(swarm.NewAddress(emptyAddr), meta))
	if err != nil {
		s.Logger.Debugf("feed post: add manifest entry: %v", err)
		s.Logger.Error("feed post: add manifest entry")
		jsonhttp.InternalServerError(w, nil)
		return
	}
	ref, err := feedManifest.Store(r.Context())
	if err != nil {
		s.Logger.Debugf("feed post: store manifest: %v", err)
		s.Logger.Error("feed post: store manifest")
		jsonhttp.InternalServerError(w, nil)
		return
	}
	jsonhttp.Created(w, feedReferenceResponse{Reference: ref})
}

func parseFeedUpdate(ch swarm.Chunk) (swarm.Address, int64, error) {
	sch, err := soc.FromChunk(ch)
	if err != nil {
		return swarm.ZeroAddress, 0, fmt.Errorf("soc unmarshal: %w", err)
	}

	update := sch.Chunk.Data()
	// split the timestamp and reference
	// possible values right now:
	// unencrypted ref: span+timestamp+ref => 8+8+32=48
	// encrypted ref: span+timestamp+ref+decryptKey => 8+8+64=80
	if len(update) != 48 && len(update) != 80 {
		return swarm.ZeroAddress, 0, errInvalidFeedUpdate
	}
	ts := binary.BigEndian.Uint64(update[8:16])
	ref := swarm.NewAddress(update[16:])
	return ref, int64(ts), nil
}
