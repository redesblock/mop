package api

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"

	"github.com/redesblock/hop/core/collection/entry"
	"github.com/redesblock/hop/core/file/pipeline"
	"github.com/redesblock/hop/core/jsonhttp"
	"github.com/redesblock/hop/core/logging"
	"github.com/redesblock/hop/core/manifest"
	"github.com/redesblock/hop/core/sctx"
	"github.com/redesblock/hop/core/storage"
	"github.com/redesblock/hop/core/swarm"
)

const (
	contentTypeHeader = "Content-Type"
	contentTypeTar    = "application/x-tar"
)

// dirUploadHandler uploads a directory supplied as a tar in an HTTP request
func (s *server) dirUploadHandler(w http.ResponseWriter, r *http.Request) {
	err := validateRequest(r)
	if err != nil {
		s.Logger.Errorf("dir upload, validate request")
		s.Logger.Debugf("dir upload, validate request err: %v", err)
		jsonhttp.BadRequest(w, "could not validate request")
		return
	}

	tag, created, err := s.getOrCreateTag(r.Header.Get(SwarmTagUidHeader))
	if err != nil {
		s.Logger.Debugf("dir upload: get or create tag: %v", err)
		s.Logger.Error("dir upload: get or create tag")
		jsonhttp.InternalServerError(w, "cannot get or create tag")
		return
	}

	// Add the tag to the context
	ctx := sctx.SetTag(r.Context(), tag)

	reference, err := storeDir(ctx, r.Body, s.Storer, requestModePut(r), s.Logger, requestEncrypt(r))
	if err != nil {
		s.Logger.Debugf("dir upload: store dir err: %v", err)
		s.Logger.Errorf("dir upload: store dir")
		jsonhttp.InternalServerError(w, "could not store dir")
		return
	}
	if created {
		_, err = tag.DoneSplit(reference)
		if err != nil {
			s.Logger.Debugf("dir upload: done split: %v", err)
			s.Logger.Error("dir upload: done split failed")
			jsonhttp.InternalServerError(w, nil)
			return
		}
	}
	w.Header().Set(SwarmTagUidHeader, fmt.Sprint(tag.Uid))
	jsonhttp.OK(w, fileUploadResponse{
		Reference: reference,
	})
}

// validateRequest validates an HTTP request for a directory to be uploaded
func validateRequest(r *http.Request) error {
	if r.Body == http.NoBody {
		return errors.New("request has no body")
	}
	contentType := r.Header.Get(contentTypeHeader)
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return err
	}
	if mediaType != contentTypeTar {
		return errors.New("content-type not set to tar")
	}
	return nil
}

// storeDir stores all files recursively contained in the directory given as a tar
// it returns the hash for the uploaded manifest corresponding to the uploaded dir
func storeDir(ctx context.Context, reader io.ReadCloser, s storage.Storer, mode storage.ModePut, logger logging.Logger, encrypt bool) (swarm.Address, error) {

	dirManifest, err := manifest.NewDefaultManifest(encrypt, s)
	if err != nil {
		return swarm.ZeroAddress, err
	}

	// set up HTTP body reader
	tarReader := tar.NewReader(reader)
	defer reader.Close()

	filesAdded := 0

	// iterate through the files in the supplied tar
	for {
		fileHeader, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return swarm.ZeroAddress, fmt.Errorf("read tar stream: %w", err)
		}

		filePath := fileHeader.Name

		// only store regular files
		if !fileHeader.FileInfo().Mode().IsRegular() {
			logger.Warningf("skipping file upload for %s as it is not a regular file", filePath)
			continue
		}

		fileName := fileHeader.FileInfo().Name()
		contentType := mime.TypeByExtension(filepath.Ext(fileHeader.Name))

		// upload file
		fileInfo := &fileUploadInfo{
			name:        fileName,
			size:        fileHeader.FileInfo().Size(),
			contentType: contentType,
			reader:      tarReader,
		}
		fileReference, err := storeFile(ctx, fileInfo, s, mode, encrypt)
		if err != nil {
			return swarm.ZeroAddress, fmt.Errorf("store dir file: %w", err)
		}
		logger.Tracef("uploaded dir file %v with reference %v", filePath, fileReference)

		// add file entry to dir manifest
		err = dirManifest.Add(filePath, manifest.NewEntry(fileReference))
		if err != nil {
			return swarm.ZeroAddress, fmt.Errorf("add to manifest: %w", err)
		}

		filesAdded++
	}

	// check if files were uploaded through the manifest
	if filesAdded == 0 {
		return swarm.ZeroAddress, fmt.Errorf("no files in tar")
	}

	// save manifest
	manifestBytesReference, err := dirManifest.Store(ctx, mode)
	if err != nil {
		return swarm.ZeroAddress, fmt.Errorf("store manifest: %w", err)
	}

	// store the manifest metadata and get its reference
	m := entry.NewMetadata(manifestBytesReference.String())
	m.MimeType = dirManifest.Type()
	metadataBytes, err := json.Marshal(m)
	if err != nil {
		return swarm.ZeroAddress, fmt.Errorf("metadata marshal: %w", err)
	}

	pipe := pipeline.NewPipelineBuilder(ctx, s, mode, encrypt)
	mr, err := pipeline.FeedPipeline(ctx, pipe, bytes.NewReader(metadataBytes), int64(len(metadataBytes)))
	if err != nil {
		return swarm.ZeroAddress, fmt.Errorf("split metadata: %w", err)
	}

	// now join both references (fr, mr) to create an entry and store it
	e := entry.New(manifestBytesReference, mr)
	fileEntryBytes, err := e.MarshalBinary()
	if err != nil {
		return swarm.ZeroAddress, fmt.Errorf("entry marshal: %w", err)
	}

	pipe = pipeline.NewPipelineBuilder(ctx, s, mode, encrypt)
	manifestFileReference, err := pipeline.FeedPipeline(ctx, pipe, bytes.NewReader(fileEntryBytes), int64(len(fileEntryBytes)))
	if err != nil {
		return swarm.ZeroAddress, fmt.Errorf("split entry: %w", err)
	}

	return manifestFileReference, nil
}

// storeFile uploads the given file and returns its reference
// this function was extracted from `fileUploadHandler` and should eventually replace its current code
func storeFile(ctx context.Context, fileInfo *fileUploadInfo, s storage.Storer, mode storage.ModePut, encrypt bool) (swarm.Address, error) {
	// first store the file and get its reference
	pipe := pipeline.NewPipelineBuilder(ctx, s, mode, encrypt)
	fr, err := pipeline.FeedPipeline(ctx, pipe, fileInfo.reader, fileInfo.size)
	if err != nil {
		return swarm.ZeroAddress, fmt.Errorf("split file: %w", err)
	}

	// if filename is still empty, use the file hash as the filename
	if fileInfo.name == "" {
		fileInfo.name = fr.String()
	}

	// then store the metadata and get its reference
	m := entry.NewMetadata(fileInfo.name)
	m.MimeType = fileInfo.contentType
	metadataBytes, err := json.Marshal(m)
	if err != nil {
		return swarm.ZeroAddress, fmt.Errorf("metadata marshal: %w", err)
	}

	pipe = pipeline.NewPipelineBuilder(ctx, s, mode, encrypt)
	mr, err := pipeline.FeedPipeline(ctx, pipe, bytes.NewReader(metadataBytes), int64(len(metadataBytes)))
	if err != nil {
		return swarm.ZeroAddress, fmt.Errorf("split metadata: %w", err)
	}

	// now join both references (mr, fr) to create an entry and store it
	e := entry.New(fr, mr)
	fileEntryBytes, err := e.MarshalBinary()
	if err != nil {
		return swarm.ZeroAddress, fmt.Errorf("entry marshal: %w", err)
	}
	pipe = pipeline.NewPipelineBuilder(ctx, s, mode, encrypt)
	reference, err := pipeline.FeedPipeline(ctx, pipe, bytes.NewReader(fileEntryBytes), int64(len(fileEntryBytes)))
	if err != nil {
		return swarm.ZeroAddress, fmt.Errorf("split entry: %w", err)
	}

	return reference, nil
}
