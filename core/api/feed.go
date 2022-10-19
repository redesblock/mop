package api

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	"github.com/redesblock/mop/core/api/jsonhttp"
	"github.com/redesblock/mop/core/chunk/soc"
	"github.com/redesblock/mop/core/cluster"
	"github.com/redesblock/mop/core/feeds"
	"github.com/redesblock/mop/core/file/loadsave"
	"github.com/redesblock/mop/core/incentives/voucher"
	"github.com/redesblock/mop/core/manifest"
)

const (
	feedMetadataEntryOwner = "cluster-feed-owner"
	feedMetadataEntryTopic = "cluster-feed-topic"
	feedMetadataEntryType  = "cluster-feed-type"
)

var errInvalidFeedUpdate = errors.New("invalid feed update")

type feedReferenceResponse struct {
	Reference cluster.Address `json:"reference"`
}

func (s *Service) feedGetHandler(w http.ResponseWriter, r *http.Request) {
	str := mux.Vars(r)["owner"]
	owner, err := hex.DecodeString(str)
	if err != nil {
		s.logger.Debug("feed get: decode owner string failed", "string", str, "error", err)
		s.logger.Error(nil, "feed get: decode owner string failed")
		jsonhttp.BadRequest(w, "bad owner")
		return
	}

	str = mux.Vars(r)["topic"]
	topic, err := hex.DecodeString(str)
	if err != nil {
		s.logger.Debug("feed get: decode topic string failed", "error", err)
		s.logger.Error(nil, "feed get: decode topic string failed")
		jsonhttp.BadRequest(w, "bad topic")
		return
	}

	var at int64
	atStr := r.URL.Query().Get("at")
	if atStr != "" {
		at, err = strconv.ParseInt(atStr, 10, 64)
		if err != nil {
			s.logger.Debug("feed get: decode at string failed", "string", atStr, "error", err)
			s.logger.Error(nil, "feed get: decode at string failed")
			jsonhttp.BadRequest(w, "bad at")
			return
		}
	} else {
		at = time.Now().Unix()
	}

	f := feeds.New(topic, common.BytesToAddress(owner))
	lookup, err := s.feedFactory.NewLookup(feeds.Sequence, f)
	if err != nil {
		s.logger.Debug("feed get: new lookup failed", "owner", owner, "error", err)
		s.logger.Error(nil, "feed get: new lookup failed")
		jsonhttp.InternalServerError(w, "new lookup failed")
		return
	}

	ch, cur, next, err := lookup.At(r.Context(), at, 0)
	if err != nil {
		s.logger.Debug("feed get: lookup at failed", "at", at, "error", err)
		s.logger.Error(nil, "feed get: lookup at failed")
		jsonhttp.NotFound(w, "lookup at failed")
		return
	}

	// KLUDGE: if a feed was never updated, the chunk will be nil
	if ch == nil {
		s.logger.Debug("feed get: no update found")
		s.logger.Error(nil, "feed get: no update found")
		jsonhttp.NotFound(w, "no update found")
		return
	}

	ref, _, err := parseFeedUpdate(ch)
	if err != nil {
		s.logger.Debug("feed get: parse feed update failed", "error", err)
		s.logger.Error(nil, "feed get: parse feed update failed")
		jsonhttp.InternalServerError(w, "parse feed update failed")
		return
	}

	curBytes, err := cur.MarshalBinary()
	if err != nil {
		s.logger.Debug("feed get: marshal current index failed", "error", err)
		s.logger.Error(nil, "feed get: marshal current index failed")
		jsonhttp.InternalServerError(w, "marshal current index failed")
		return
	}

	nextBytes, err := next.MarshalBinary()
	if err != nil {
		s.logger.Debug("feed get: marshal next index failed", "error", err)
		s.logger.Error(nil, "feed get: marshal next index failed")
		jsonhttp.InternalServerError(w, "marshal next index failed")
		return
	}

	w.Header().Set(ClusterFeedIndexHeader, hex.EncodeToString(curBytes))
	w.Header().Set(ClusterFeedIndexNextHeader, hex.EncodeToString(nextBytes))
	w.Header().Set("Access-Control-Expose-Headers", fmt.Sprintf("%s, %s", ClusterFeedIndexHeader, ClusterFeedIndexNextHeader))

	jsonhttp.OK(w, feedReferenceResponse{Reference: ref})
}

func (s *Service) feedPostHandler(w http.ResponseWriter, r *http.Request) {
	str := mux.Vars(r)["owner"]
	owner, err := hex.DecodeString(str)
	if err != nil {
		s.logger.Debug("feed post: decode owner string failed", "string", str, "error", err)
		s.logger.Error(nil, "feed post: decode owner string failed")
		jsonhttp.BadRequest(w, "bad owner")
		return
	}

	str = mux.Vars(r)["topic"]
	topic, err := hex.DecodeString(str)
	if err != nil {
		s.logger.Debug("feed post: decode topic string failed", "string", str, "error", err)
		s.logger.Error(nil, "feed post: decode topic string failed")
		jsonhttp.BadRequest(w, "bad topic")
		return
	}

	putter, wait, err := s.newStamperPutter(r)
	if err != nil {
		s.logger.Debug("feed post: putter failed", "error", err)
		s.logger.Error(nil, "feed post: putter failed")
		switch {
		case errors.Is(err, voucher.ErrNotFound):
			jsonhttp.BadRequest(w, "batch not found")
		case errors.Is(err, voucher.ErrNotUsable):
			jsonhttp.BadRequest(w, "batch not usable yet")
		case errors.Is(err, errInvalidVoucherBatch):
			jsonhttp.BadRequest(w, "invalid voucher batch id")
		default:
			jsonhttp.BadRequest(w, nil)
		}
		return
	}

	l := loadsave.New(putter, requestPipelineFactory(r.Context(), putter, r))
	feedManifest, err := manifest.NewDefaultManifest(l, false)
	if err != nil {
		s.logger.Debug("feed post: create manifest failed", "error", err)
		s.logger.Error(nil, "feed post: create manifest failed")
		jsonhttp.InternalServerError(w, "create manifest failed")
		return
	}

	meta := map[string]string{
		feedMetadataEntryOwner: hex.EncodeToString(owner),
		feedMetadataEntryTopic: hex.EncodeToString(topic),
		feedMetadataEntryType:  feeds.Sequence.String(), // only sequence allowed for now
	}

	emptyAddr := make([]byte, 32)

	// a feed manifest stores the metadata at the root "/" path
	err = feedManifest.Add(r.Context(), "/", manifest.NewEntry(cluster.NewAddress(emptyAddr), meta))
	if err != nil {
		s.logger.Debug("feed post: add manifest entry failed", "error", err)
		s.logger.Error(nil, "feed post: add manifest entry failed")
		jsonhttp.InternalServerError(w, "feed post: add manifest entry failed")
		return
	}
	ref, err := feedManifest.Store(r.Context())
	if err != nil {
		s.logger.Debug("feed post: store manifest failed", "error", err)
		s.logger.Error(nil, "feed post: store manifest failed")
		switch {
		case errors.Is(err, voucher.ErrBucketFull):
			jsonhttp.PaymentRequired(w, "batch is overissued")
		default:
			jsonhttp.InternalServerError(w, "feed post: store manifest failed")
		}
		return
	}

	if strings.ToLower(r.Header.Get(ClusterPinHeader)) == "true" {
		if err := s.pinning.CreatePin(r.Context(), ref, false); err != nil {
			s.logger.Debug("feed post: pins creation failed: %v", "address", ref, "error", err)
			s.logger.Error(nil, "feed post: pins creation failed")
			jsonhttp.InternalServerError(w, "feed post: creation of pins failed")
			return
		}
	}

	if err = wait(); err != nil {
		s.logger.Debug("feed post: chainsync chunks failed", "error", err)
		s.logger.Error(nil, "feed post: chainsync chunks failed")
		jsonhttp.InternalServerError(w, "feed upload: chainsync failed")
		return
	}

	jsonhttp.Created(w, feedReferenceResponse{Reference: ref})
}

func parseFeedUpdate(ch cluster.Chunk) (cluster.Address, int64, error) {
	s, err := soc.FromChunk(ch)
	if err != nil {
		return cluster.ZeroAddress, 0, fmt.Errorf("soc unmarshal: %w", err)
	}

	update := s.WrappedChunk().Data()
	// split the timestamp and reference
	// possible values right now:
	// unencrypted ref: span+timestamp+ref => 8+8+32=48
	// encrypted ref: span+timestamp+ref+decryptKey => 8+8+64=80
	if len(update) != 48 && len(update) != 80 {
		return cluster.ZeroAddress, 0, errInvalidFeedUpdate
	}
	ts := binary.BigEndian.Uint64(update[8:16])
	ref := cluster.NewAddress(update[16:])
	return ref, int64(ts), nil
}
