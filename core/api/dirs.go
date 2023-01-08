package api

import (
	"archive/tar"
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/redesblock/mop/core/api/jsonhttp"
	"github.com/redesblock/mop/core/cluster"
	"github.com/redesblock/mop/core/file"
	"github.com/redesblock/mop/core/file/loadsave"
	"github.com/redesblock/mop/core/incentives/voucher"
	"github.com/redesblock/mop/core/log"
	"github.com/redesblock/mop/core/manifest"
	"github.com/redesblock/mop/core/mctx"
	"github.com/redesblock/mop/core/storer/storage"
	"github.com/redesblock/mop/core/tags"
	"github.com/redesblock/mop/core/tracer"
)

var errEmptyDir = errors.New("no files in root directory")

// dirUploadHandler uploads a directory supplied as a tar in an HTTP request
func (s *Service) dirUploadHandler(w http.ResponseWriter, r *http.Request, storer storage.Storer, waitFn func() error) {
	logger := tracer.NewLoggerWithTraceID(r.Context(), s.logger)
	if r.Body == http.NoBody {
		logger.Error(nil, "mop upload dir: request has no body")
		jsonhttp.BadRequest(w, errInvalidRequest)
		return
	}
	contentType := r.Header.Get(contentTypeHeader)
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		logger.Error(nil, "mop upload dir: parse media type failed")
		logger.Debug("mop upload dir: parse media type failed", "error", err)
		jsonhttp.BadRequest(w, errInvalidContentType)
		return
	}

	var dReader dirReader
	switch mediaType {
	case contentTypeTar:
		dReader = &tarReader{r: tar.NewReader(r.Body), logger: s.logger}
	case multiPartFormData:
		dReader = &multipartReader{r: multipart.NewReader(r.Body, params["boundary"])}
	default:
		logger.Error(nil, "mop upload dir: invalid content-type for directory upload")
		jsonhttp.BadRequest(w, errInvalidContentType)
		return
	}
	defer r.Body.Close()

	tag, created, err := s.getOrCreateTag(r.Header.Get(ClusterTagHeader))
	if err != nil {
		logger.Debug("mop upload dir: get or create tag failed", "error", err)
		logger.Error(nil, "mop upload dir: get or create tag failed")
		jsonhttp.InternalServerError(w, "mop upload dir: get or create tag failed")
		return
	}

	// Add the tag to the context
	ctx := mctx.SetTag(r.Context(), tag)

	reference, err := storeDir(
		ctx,
		requestEncrypt(r),
		dReader,
		s.logger,
		requestPipelineFn(storer, r),
		loadsave.New(storer, requestPipelineFactory(ctx, storer, r)),
		r.Header.Get(ClusterIndexDocumentHeader),
		r.Header.Get(ClusterErrorDocumentHeader),
		tag,
		created,
	)
	if err != nil {
		logger.Debug("mop upload dir: store dir failed", "error", err)
		logger.Error(nil, "mop upload dir: store dir failed")
		switch {
		case errors.Is(err, voucher.ErrBucketFull):
			jsonhttp.PaymentRequired(w, "batch is overissued")
		case errors.Is(err, errEmptyDir):
			jsonhttp.BadRequest(w, errEmptyDir)
		case errors.Is(err, tar.ErrHeader):
			jsonhttp.BadRequest(w, "invalid filename in tar archive")
		default:
			jsonhttp.InternalServerError(w, errDirectoryStore)
		}
		return
	}
	if created {
		_, err = tag.DoneSplit(reference)
		if err != nil {
			logger.Debug("mop upload dir: done split failed", "error", err)
			logger.Error(nil, "mop upload dir: done split failed")
			jsonhttp.InternalServerError(w, "mop upload dir: done split failed")
			return
		}
	}

	if strings.ToLower(r.Header.Get(ClusterPinHeader)) == "true" {
		if err := s.pinning.CreatePin(r.Context(), reference, false); err != nil {
			logger.Debug("mop upload dir: pins creation failed", "address", reference, "error", err)
			logger.Error(nil, "mop upload dir: pins creation failed")
			jsonhttp.InternalServerError(w, "mop upload dir: create pins failed")
			return
		}
	}

	if err = waitFn(); err != nil {
		s.logger.Debug("mop upload: chainsync chunks failed", "error", err)
		s.logger.Error(nil, "mop upload: chainsync chunks failed")
		jsonhttp.InternalServerError(w, "mop upload: chainsync chunks failed")
		return
	}

	w.Header().Set("Access-Control-Expose-Headers", ClusterTagHeader)
	w.Header().Set(ClusterTagHeader, fmt.Sprint(tag.Uid))
	jsonhttp.Created(w, mopUploadResponse{
		Reference: reference,
	})
}

// storeDir stores all files recursively contained in the directory given as a tar/multipart
// it returns the hash for the uploaded manifest corresponding to the uploaded dir
func storeDir(
	ctx context.Context,
	encrypt bool,
	reader dirReader,
	log log.Logger,
	p pipelineFunc,
	ls file.LoadSaver,
	indexFilename,
	errorFilename string,
	tag *tags.Tag,
	tagCreated bool,
) (cluster.Address, error) {
	logger := tracer.NewLoggerWithTraceID(ctx, log)
	loggerV1 := logger.V(1).Build()

	dirManifest, err := manifest.NewDefaultManifest(ls, encrypt)
	if err != nil {
		return cluster.ZeroAddress, err
	}

	if indexFilename != "" && strings.ContainsRune(indexFilename, '/') {
		return cluster.ZeroAddress, errors.New("index document suffix must not include slash character")
	}

	filesAdded := 0

	// iterate through the files in the supplied tar
	for {
		fileInfo, err := reader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return cluster.ZeroAddress, fmt.Errorf("read tar stream: %w", err)
		}

		if !tagCreated {
			// only in the case when tag is sent via header (i.e. not created by this request)
			// for each file
			if estimatedTotalChunks := calculateNumberOfChunks(fileInfo.Size, encrypt); estimatedTotalChunks > 0 {
				err = tag.IncN(tags.TotalChunks, estimatedTotalChunks)
				if err != nil {
					return cluster.ZeroAddress, fmt.Errorf("increment tag: %w", err)
				}
			}
		}

		fileReference, err := p(ctx, fileInfo.Reader)
		if err != nil {
			debug.PrintStack()
			return cluster.ZeroAddress, fmt.Errorf("store dir file: %w", err)
		}
		loggerV1.Debug("mop upload dir: file dir uploaded", "file_path", fileInfo.Path, "address", fileReference)

		fileMtdt := map[string]string{
			manifest.EntryMetadataContentTypeKey: fileInfo.ContentType,
			manifest.EntryMetadataFilenameKey:    fileInfo.Name,
		}
		// add file entry to dir manifest
		err = dirManifest.Add(ctx, fileInfo.Path, manifest.NewEntry(fileReference, fileMtdt))
		if err != nil {
			return cluster.ZeroAddress, fmt.Errorf("add to manifest: %w", err)
		}

		filesAdded++
	}

	// check if files were uploaded through the manifest
	if filesAdded == 0 {
		return cluster.ZeroAddress, errEmptyDir
	}

	// store website information
	if indexFilename != "" || errorFilename != "" {
		metadata := map[string]string{}
		if indexFilename != "" {
			metadata[manifest.WebsiteIndexDocumentSuffixKey] = indexFilename
		}
		if errorFilename != "" {
			metadata[manifest.WebsiteErrorDocumentPathKey] = errorFilename
		}
		rootManifestEntry := manifest.NewEntry(cluster.ZeroAddress, metadata)
		err = dirManifest.Add(ctx, manifest.RootPath, rootManifestEntry)
		if err != nil {
			return cluster.ZeroAddress, fmt.Errorf("add to manifest: %w", err)
		}
	}

	storeSizeFn := []manifest.StoreSizeFunc{}
	if !tagCreated {
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

	// save manifest
	manifestReference, err := dirManifest.Store(ctx, storeSizeFn...)
	if err != nil {
		return cluster.ZeroAddress, fmt.Errorf("store manifest: %w", err)
	}
	loggerV1.Debug("mop upload dir: uploaded dir finished", "address", manifestReference)

	return manifestReference, nil
}

type FileInfo struct {
	Path        string
	Name        string
	ContentType string
	Size        int64
	Reader      io.Reader
}

type dirReader interface {
	Next() (*FileInfo, error)
}

type tarReader struct {
	r      *tar.Reader
	logger log.Logger
}

func (t *tarReader) Next() (*FileInfo, error) {
	for {
		fileHeader, err := t.r.Next()
		if err != nil {
			return nil, err
		}

		fileName := fileHeader.FileInfo().Name()
		contentType := mime.TypeByExtension(filepath.Ext(fileHeader.Name))
		fileSize := fileHeader.FileInfo().Size()
		filePath := filepath.Clean(fileHeader.Name)

		if filePath == "." {
			t.logger.Warning("skipping file upload empty path")
			continue
		}
		if runtime.GOOS == "windows" {
			// always use Unix path separator
			filePath = filepath.ToSlash(filePath)
		}
		// only store regular files
		if !fileHeader.FileInfo().Mode().IsRegular() {
			t.logger.Warning("mop upload dir: skipping file upload as it is not a regular file", "file_path", filePath)
			continue
		}

		return &FileInfo{
			Path:        filePath,
			Name:        fileName,
			ContentType: contentType,
			Size:        fileSize,
			Reader:      t.r,
		}, nil
	}
}

// multipart reader returns files added as a multipart form. We will ensure all the
// part headers are passed correctly
type multipartReader struct {
	r *multipart.Reader
}

func (m *multipartReader) Next() (*FileInfo, error) {
	part, err := m.r.NextPart()
	if err != nil {
		return nil, err
	}

	filePath := part.FileName()
	if filePath == "" {
		filePath = part.FormName()
	}
	if filePath == "" {
		return nil, errors.New("filepath missing")
	}

	fileName := filepath.Base(filePath)

	contentType := part.Header.Get(contentTypeHeader)
	if contentType == "" {
		return nil, errors.New("content-type missing")
	}

	contentLength := part.Header.Get("Content-Length")
	if contentLength == "" {
		return nil, errors.New("content-length missing")
	}
	fileSize, err := strconv.ParseInt(contentLength, 10, 64)
	if err != nil {
		return nil, errors.New("invalid file size")
	}

	return &FileInfo{
		Path:        filePath,
		Name:        fileName,
		ContentType: contentType,
		Size:        fileSize,
		Reader:      part,
	}, nil
}
