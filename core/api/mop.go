package api

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"

	"github.com/redesblock/mop/core/api/jsonhttp"
	"github.com/redesblock/mop/core/cluster"
	"github.com/redesblock/mop/core/feeds"
	"github.com/redesblock/mop/core/file/joiner"
	"github.com/redesblock/mop/core/file/loadsave"
	"github.com/redesblock/mop/core/incentives/voucher"
	"github.com/redesblock/mop/core/langos"
	"github.com/redesblock/mop/core/manifest"
	"github.com/redesblock/mop/core/mctx"
	"github.com/redesblock/mop/core/storer/storage"
	"github.com/redesblock/mop/core/tags"
	"github.com/redesblock/mop/core/tracer"
)

func (s *Service) mopUploadHandler(w http.ResponseWriter, r *http.Request) {
	logger := tracer.NewLoggerWithTraceID(r.Context(), s.logger)

	contentType := r.Header.Get(contentTypeHeader)
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		logger.Debug("mop upload: parse content type header string failed", "string", contentType, "error", err)
		logger.Error(nil, "mop upload: parse content type header string failed", "string", contentType)
		jsonhttp.BadRequest(w, errInvalidContentType)
		return
	}

	putter, wait, err := s.newStamperPutter(r)
	if err != nil {
		logger.Debug("mop upload: putter failed", "error", err)
		logger.Error(nil, "mop upload: putter failed")
		switch {
		case errors.Is(err, voucher.ErrNotFound):
			jsonhttp.BadRequest(w, "batch not found")
		case errors.Is(err, voucher.ErrNotUsable):
			jsonhttp.BadRequest(w, "batch not usable yet")
		default:
			jsonhttp.BadRequest(w, nil)
		}
		return
	}

	isDir := r.Header.Get(ClusterCollectionHeader)
	if strings.ToLower(isDir) == "true" || mediaType == multiPartFormData {
		s.dirUploadHandler(w, r, putter, wait)
		return
	}
	s.fileUploadHandler(w, r, putter, wait)
}

// fileUploadResponse is returned when an HTTP request to upload a file is successful
type mopUploadResponse struct {
	Reference cluster.Address `json:"reference"`
}

// fileUploadHandler uploads the file and its metadata supplied in the file body and
// the headers
func (s *Service) fileUploadHandler(w http.ResponseWriter, r *http.Request, storer storage.Storer, waitFn func() error) {
	logger := tracer.NewLoggerWithTraceID(r.Context(), s.logger)
	var (
		reader   io.Reader
		fileName string
	)

	// Content-Type has already been validated by this time
	contentType := r.Header.Get(contentTypeHeader)

	tag, created, err := s.getOrCreateTag(r.Header.Get(ClusterTagHeader))
	if err != nil {
		logger.Debug("mop upload file: get or create tag failed", "error", err)
		logger.Error(nil, "mop upload file: get or create tag failed")
		jsonhttp.InternalServerError(w, "mop upload file: get or create tag failed")
		return
	}

	if !created {
		// only in the case when tag is sent via header (i.e. not created by this request)
		if estimatedTotalChunks := requestCalculateNumberOfChunks(r); estimatedTotalChunks > 0 {
			err = tag.IncN(tags.TotalChunks, estimatedTotalChunks)
			if err != nil {
				s.logger.Debug("mop upload file: increment tag failed", "error", err)
				s.logger.Error(nil, "mop upload file: increment tag failed")
				jsonhttp.InternalServerError(w, "mop upload file: increment tag failed")
				return
			}
		}
	}

	// Add the tag to the context
	ctx := mctx.SetTag(r.Context(), tag)

	fileName = r.URL.Query().Get("name")
	reader = r.Body

	p := requestPipelineFn(storer, r)

	// first store the file and get its reference
	fr, err := p(ctx, reader)
	if err != nil {
		logger.Debug("mop upload file: file store failed", "file_name", fileName, "error", err)
		logger.Error(nil, "mop upload file: file store failed", "file_name", fileName)
		switch {
		case errors.Is(err, voucher.ErrBucketFull):
			jsonhttp.PaymentRequired(w, "batch is overissued")
		default:
			jsonhttp.InternalServerError(w, errFileStore)
		}
		return
	}

	// If filename is still empty, use the file hash as the filename
	if fileName == "" {
		fileName = fr.String()
	}

	encrypt := requestEncrypt(r)
	factory := requestPipelineFactory(ctx, storer, r)
	l := loadsave.New(storer, factory)

	m, err := manifest.NewDefaultManifest(l, encrypt)
	if err != nil {
		logger.Debug("mop upload file: create manifest failed", "file_name", fileName, "error", err)
		logger.Error(nil, "mop upload file: create manifest failed", "file_name", fileName)
		jsonhttp.InternalServerError(w, "mop upload file: create manifest failed")
		return
	}

	rootMetadata := map[string]string{
		manifest.WebsiteIndexDocumentSuffixKey: fileName,
	}

	err = m.Add(ctx, manifest.RootPath, manifest.NewEntry(cluster.ZeroAddress, rootMetadata))
	if err != nil {
		logger.Debug("mop upload file: adding metadata to manifest failed", "file_name", fileName, "error", err)
		logger.Error(nil, "mop upload file: adding metadata to manifest failed", "file_name", fileName)
		jsonhttp.InternalServerError(w, "mop upload file: add metadata failed")
		return
	}

	fileMtdt := map[string]string{
		manifest.EntryMetadataContentTypeKey: contentType,
		manifest.EntryMetadataFilenameKey:    fileName,
	}

	err = m.Add(ctx, fileName, manifest.NewEntry(fr, fileMtdt))
	if err != nil {
		logger.Debug("mop upload file: adding file to manifest failed", "file_name", fileName, "error", err)
		logger.Error(nil, "mop upload file: adding file to manifest failed", "file_name", fileName)
		jsonhttp.InternalServerError(w, "mop upload file: add file failed")
		return
	}

	logger.Debug("mop upload file: info", "encrypt", encrypt, "file_name", fileName, "hash", fr, "metadata", fileMtdt)

	storeSizeFn := []manifest.StoreSizeFunc{}
	if !created {
		// only in the case when tag is sent via header (i.e. not created by this request)
		// each content that is saved for manifest
		storeSizeFn = append(storeSizeFn, func(dataSize int64) error {
			if estimatedTotalChunks := calculateNumberOfChunks(dataSize, encrypt); estimatedTotalChunks > 0 {
				err = tag.IncN(tags.TotalChunks, estimatedTotalChunks)
				if err != nil {
					return fmt.Errorf("increment tag: %w", err)
				}
			}
			return nil
		})
	}

	manifestReference, err := m.Store(ctx, storeSizeFn...)
	if err != nil {
		logger.Debug("mop upload file: manifest store failed", "file_name", fileName, "error", err)
		logger.Error(nil, "mop upload file: manifest store failed", "file_name", fileName)
		switch {
		case errors.Is(err, voucher.ErrBucketFull):
			jsonhttp.PaymentRequired(w, "batch is overissued")
		default:
			jsonhttp.InternalServerError(w, "mop upload file: manifest store failed")
		}
		return
	}
	logger.Debug("mop upload file: store", "manifest_reference", manifestReference)

	if created {
		_, err = tag.DoneSplit(manifestReference)
		if err != nil {
			logger.Debug("mop upload file: done split failed", "error", err)
			logger.Error(nil, "mop upload file: done split failed")
			jsonhttp.InternalServerError(w, "mop upload file: done split failed")
			return
		}
	}

	if strings.ToLower(r.Header.Get(ClusterPinHeader)) == "true" {
		if err := s.pinning.CreatePin(ctx, manifestReference, false); err != nil {
			logger.Debug("mop upload file: pins creation failed", "manifest_reference", manifestReference, "error", err)
			logger.Error(nil, "mop upload file: pins creation failed")
			jsonhttp.InternalServerError(w, "mop upload file: create pins failed")
			return
		}
	}

	if err = waitFn(); err != nil {
		s.logger.Debug("mop upload: chainsync chunks failed", "error", err)
		s.logger.Error(nil, "mop upload: chainsync chunks failed")
		jsonhttp.InternalServerError(w, "mop upload file: chainsync chunks failed")
		return
	}

	w.Header().Set("ETag", fmt.Sprintf("%q", manifestReference.String()))
	w.Header().Set(ClusterTagHeader, fmt.Sprint(tag.Uid))
	w.Header().Set("Access-Control-Expose-Headers", ClusterTagHeader)
	jsonhttp.Created(w, mopUploadResponse{
		Reference: manifestReference,
	})
}

func (s *Service) mopDownloadHandler(w http.ResponseWriter, r *http.Request) {
	logger := tracer.NewLoggerWithTraceID(r.Context(), s.logger)

	nameOrHex := mux.Vars(r)["address"]
	pathVar := mux.Vars(r)["path"]
	if strings.HasSuffix(pathVar, "/") {
		pathVar = strings.TrimRight(pathVar, "/")
		// NOTE: leave one slash if there was some
		pathVar += "/"
	}

	address, err := s.resolveNameOrAddress(nameOrHex)
	if err != nil {
		logger.Debug("mop download: parse address string failed", "string", nameOrHex, "error", err)
		logger.Error(nil, "mop download: parse address string failed")
		jsonhttp.NotFound(w, nil)
		return
	}
	s.serveReference(address, pathVar, w, r)
}

func (s *Service) serveReference(address cluster.Address, pathVar string, w http.ResponseWriter, r *http.Request) {
	logger := tracer.NewLoggerWithTraceID(r.Context(), s.logger)
	ls := loadsave.NewReadonly(s.storer)
	feedDereferenced := false

	ctx := r.Context()

FETCH:
	// read manifest entry
	m, err := manifest.NewDefaultManifestReference(
		address,
		ls,
	)
	if err != nil {
		logger.Debug("mop download: not manifest", "address", address, "error", err)
		logger.Error(nil, "mop download: not manifest")
		jsonhttp.NotFound(w, nil)
		return
	}

	// there's a possible ambiguity here, right now the data which was
	// read can be an entry.Entry or a mantaray feed manifest. Try to
	// unmarshal as mantaray first and possibly resolve the feed, otherwise
	// go on normally.
	if !feedDereferenced {
		if l, err := s.manifestFeed(ctx, m); err == nil {
			//we have a feed manifest here
			ch, cur, _, err := l.At(ctx, time.Now().Unix(), 0)
			if err != nil {
				logger.Debug("mop download: feed lookup failed", "error", err)
				logger.Error(nil, "mop download: feed lookup failed")
				jsonhttp.NotFound(w, "feed not found")
				return
			}
			if ch == nil {
				logger.Debug("mop download: feed lookup: no updates")
				logger.Error(nil, "mop download: feed lookup")
				jsonhttp.NotFound(w, "no update found")
				return
			}
			ref, _, err := parseFeedUpdate(ch)
			if err != nil {
				logger.Debug("mop download: parse feed update failed", "error", err)
				logger.Error(nil, "mop download: parse feed update failed")
				jsonhttp.InternalServerError(w, "parse feed update")
				return
			}
			address = ref
			feedDereferenced = true
			curBytes, err := cur.MarshalBinary()
			if err != nil {
				s.logger.Debug("mop download: marshal feed index failed", "error", err)
				s.logger.Error(nil, "mop download: marshal index failed")
				jsonhttp.InternalServerError(w, "marshal index")
				return
			}

			w.Header().Set(ClusterFeedIndexHeader, hex.EncodeToString(curBytes))
			// this header might be overriding others. handle with care. in the future
			// we should implement an append functionality for this specific header,
			// since different parts of handlers might be overriding others' values
			// resulting in inconsistent headers in the response.
			w.Header().Set("Access-Control-Expose-Headers", ClusterFeedIndexHeader)
			goto FETCH
		}
	}

	tarVar := mux.Vars(r)["tar"]
	if strings.ToLower(tarVar) == "true" {
		s.serveManifestEntryTar(w, r, address, m, !feedDereferenced)
		return
	}

	if pathVar == "" {
		logger.Debug("mop download: handle empty path", "address", address)

		if indexDocumentSuffixKey, ok := manifestMetadataLoad(ctx, m, manifest.RootPath, manifest.WebsiteIndexDocumentSuffixKey); ok {
			pathWithIndex := path.Join(pathVar, indexDocumentSuffixKey)
			indexDocumentManifestEntry, err := m.Lookup(ctx, pathWithIndex)
			if err == nil {
				// index document exists
				logger.Debug("mop download: serving path", "path", pathWithIndex)

				s.serveManifestEntry(w, r, address, indexDocumentManifestEntry, !feedDereferenced)
				return
			}
		}
	}
	me, err := m.Lookup(ctx, pathVar)
	if err != nil {
		logger.Debug("mop download: invalid path", "address", address, "path", pathVar, "error", err)

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

					logger.Debug("mop download: redirecting failed", "url", redirectURL, "error", err)

					http.Redirect(w, r, redirectURL, http.StatusPermanentRedirect)
					return
				}
			}

			// check index suffix path
			if indexDocumentSuffixKey, ok := manifestMetadataLoad(ctx, m, manifest.RootPath, manifest.WebsiteIndexDocumentSuffixKey); ok {
				if !strings.HasSuffix(pathVar, indexDocumentSuffixKey) {
					// check if path is directory with index
					pathWithIndex := path.Join(pathVar, indexDocumentSuffixKey)
					indexDocumentManifestEntry, err := m.Lookup(ctx, pathWithIndex)
					if err == nil {
						// index document exists
						logger.Debug("mop download: serving path", "path", pathWithIndex)

						s.serveManifestEntry(w, r, address, indexDocumentManifestEntry, !feedDereferenced)
						return
					}
				}
			}

			// check if error document is to be shown
			if errorDocumentPath, ok := manifestMetadataLoad(ctx, m, manifest.RootPath, manifest.WebsiteErrorDocumentPathKey); ok {
				if pathVar != errorDocumentPath {
					errorDocumentManifestEntry, err := m.Lookup(ctx, errorDocumentPath)
					if err == nil {
						// error document exists
						logger.Debug("mop download: serving path", "path", errorDocumentPath)

						s.serveManifestEntry(w, r, address, errorDocumentManifestEntry, !feedDereferenced)
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
	s.serveManifestEntry(w, r, address, me, !feedDereferenced)
}

func (s *Service) serveManifestEntryTar(
	w http.ResponseWriter,
	r *http.Request,
	address cluster.Address,
	mf manifest.Interface,
	etag bool,
) {
	//var buf bytes.Buffer
	//tw := tar.NewWriter(&buf)
	var files []string
	if err := mf.IterateAddresses(r.Context(), func(address cluster.Address, mtdt map[string]string) error {
		fname, ok := mtdt[manifest.EntryMetadataFilenameKey]
		if !ok {
			return fmt.Errorf("file name not found")
		}
		files = append(files, fname)
		//reader, l, err := joiner.New(r.Context(), s.storer, address)
		//if err != nil {
		//	return fmt.Errorf("joiner failed %v", err)
		//}
		//_ = l
		//hdr := &tar.Header{
		//	Name: fname,
		//	Mode: 0600,
		//	Size: reader.Size(),
		//}
		//
		//if err := tw.WriteHeader(hdr); err != nil {
		//	return fmt.Errorf("write header failed %v", err)
		//}
		//bts, err := ioutil.ReadAll(reader)
		//if err != nil {
		//	return fmt.Errorf("write header failed %v", err)
		//}
		//if _, err := tw.Write(bts); err != nil {
		//	return fmt.Errorf("write failed %v", err)
		//}
		return nil
	}); err != nil {
		jsonhttp.InternalServerError(w, err)
		return
	}

	//l := int64(buf.Len())
	//w.Header().Set(contentTypeHeader, contentTypeTar)
	//w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s.tar\"", address))
	//w.Header().Set("Content-Length", strconv.FormatInt(l, 10))
	//w.Header().Set("Decompressed-Content-Length", strconv.FormatInt(l, 10))
	//w.Header().Set("Access-Control-Expose-Headers", "Content-Disposition")
	if etag {
		w.Header().Set("ETag", fmt.Sprintf("%q", address))
	}
	jsonhttp.OK(w, files)
	//http.ServeContent(w, r, "", time.Now(), buf)
}

func (s *Service) serveManifestEntry(
	w http.ResponseWriter,
	r *http.Request,
	address cluster.Address,
	manifestEntry manifest.Entry,
	etag bool,
) {
	additionalHeaders := http.Header{}
	mtdt := manifestEntry.Metadata()
	if fname, ok := mtdt[manifest.EntryMetadataFilenameKey]; ok {
		fname = filepath.Base(fname) // only keep the file name
		additionalHeaders["Content-Disposition"] =
			[]string{fmt.Sprintf("inline; filename=\"%s\"", fname)}
	}
	if mimeType, ok := mtdt[manifest.EntryMetadataContentTypeKey]; ok {
		additionalHeaders[contentTypeHeader] = []string{mimeType}
	}

	s.downloadHandler(w, r, manifestEntry.Reference(), additionalHeaders, etag)
}

// downloadHandler contains common logic for dowloading Cluster file from API
func (s *Service) downloadHandler(w http.ResponseWriter, r *http.Request, reference cluster.Address, additionalHeaders http.Header, etag bool) {
	logger := tracer.NewLoggerWithTraceID(r.Context(), s.logger)

	reader, l, err := joiner.New(r.Context(), s.storer, reference)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			logger.Debug("api download: not found ", "address", reference, "error", err)
			logger.Error(nil, "api download: not found")
			jsonhttp.NotFound(w, nil)
			return
		}
		logger.Debug("api download: unexpected error", "address", reference, "error", err)
		logger.Error(nil, "api download: unexpected error")
		jsonhttp.InternalServerError(w, "api download: joiner failed")
		return
	}

	// include additional headers
	for name, values := range additionalHeaders {
		w.Header().Set(name, strings.Join(values, "; "))
	}
	if etag {
		w.Header().Set("ETag", fmt.Sprintf("%q", reference))
	}
	w.Header().Set("Content-Length", strconv.FormatInt(l, 10))
	w.Header().Set("Decompressed-Content-Length", strconv.FormatInt(l, 10))
	w.Header().Set("Access-Control-Expose-Headers", "Content-Disposition")
	http.ServeContent(w, r, "", time.Now(), langos.NewBufferedLangos(reader, lookaheadBufferSize(l)))
}

// manifestMetadataLoad returns the value for a key stored in the metadata of
// manifest path, or empty string if no value is present.
// The ok result indicates whether value was found in the metadata.
func manifestMetadataLoad(
	ctx context.Context,
	manifest manifest.Interface,
	path, metadataKey string,
) (string, bool) {
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

func (s *Service) manifestFeed(
	ctx context.Context,
	m manifest.Interface,
) (feeds.Lookup, error) {
	e, err := m.Lookup(ctx, "/")
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
	if len(owner) == 0 || len(topic) == 0 {
		return nil, fmt.Errorf("node lookup: %s", "feed metadata absent")
	}
	f := feeds.New(topic, common.BytesToAddress(owner))
	return s.feedFactory.NewLookup(*t, f)
}
