package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"strings"
	"testing"

	"github.com/redesblock/hop/core/api"
	"github.com/redesblock/hop/core/collection/entry"
	"github.com/redesblock/hop/core/file"
	"github.com/redesblock/hop/core/file/splitter"
	"github.com/redesblock/hop/core/jsonhttp"
	"github.com/redesblock/hop/core/jsonhttp/jsonhttptest"
	"github.com/redesblock/hop/core/logging"
	"github.com/redesblock/hop/core/manifest/jsonmanifest"
	smock "github.com/redesblock/hop/core/storage/mock"
	"github.com/redesblock/hop/core/swarm"
	"github.com/redesblock/hop/core/tags"
)

func TestHop(t *testing.T) {
	var (
		hopDownloadResource = func(addr, path string) string { return "/hop/" + addr + "/" + path }
		storer              = smock.NewStorer()
		sp                  = splitter.NewSimpleSplitter(storer)
		client              = newTestServer(t, testServerOptions{
			Storer: storer,
			Tags:   tags.NewTags(),
			Logger: logging.New(ioutil.Discard, 5),
		})
	)

	t.Run("download-file-by-path", func(t *testing.T) {
		fileName := "sample.html"
		filePath := "test/" + fileName
		missingFilePath := "test/missing"
		sampleHtml := `<!DOCTYPE html>
		<html>
		<body>
	
		<h1>My First Heading</h1>
	
		<p>My first paragraph.</p>
	
		</body>
		</html>`

		var err error
		var fileContentReference swarm.Address
		var fileReference swarm.Address
		var manifestFileReference swarm.Address

		// save file

		fileContentReference, err = file.SplitWriteAll(context.Background(), sp, strings.NewReader(sampleHtml), int64(len(sampleHtml)), false)
		if err != nil {
			t.Fatal(err)
		}

		fileMetadata := entry.NewMetadata(fileName)
		fileMetadataBytes, err := json.Marshal(fileMetadata)
		if err != nil {
			t.Fatal(err)
		}

		fileMetadataReference, err := file.SplitWriteAll(context.Background(), sp, bytes.NewReader(fileMetadataBytes), int64(len(fileMetadataBytes)), false)
		if err != nil {
			t.Fatal(err)
		}

		fe := entry.New(fileContentReference, fileMetadataReference)
		fileEntryBytes, err := fe.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}
		fileReference, err = file.SplitWriteAll(context.Background(), sp, bytes.NewReader(fileEntryBytes), int64(len(fileEntryBytes)), false)
		if err != nil {
			t.Fatal(err)
		}

		// save manifest

		jsonManifest := jsonmanifest.NewManifest()

		e := jsonmanifest.NewEntry(fileReference, fileName, http.Header{"Content-Type": {"text/html", "charset=utf-8"}})
		jsonManifest.Add(filePath, e)

		manifestFileBytes, err := jsonManifest.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		fr, err := file.SplitWriteAll(context.Background(), sp, bytes.NewReader(manifestFileBytes), int64(len(manifestFileBytes)), false)
		if err != nil {
			t.Fatal(err)
		}

		m := entry.NewMetadata(fileName)
		m.MimeType = api.ManifestContentType
		metadataBytes, err := json.Marshal(m)
		if err != nil {
			t.Fatal(err)
		}

		mr, err := file.SplitWriteAll(context.Background(), sp, bytes.NewReader(metadataBytes), int64(len(metadataBytes)), false)
		if err != nil {
			t.Fatal(err)
		}

		// now join both references (mr,fr) to create an entry and store it.
		newEntry := entry.New(fr, mr)
		manifestFileEntryBytes, err := newEntry.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		manifestFileReference, err = file.SplitWriteAll(context.Background(), sp, bytes.NewReader(manifestFileEntryBytes), int64(len(manifestFileEntryBytes)), false)
		if err != nil {
			t.Fatal(err)
		}

		// read file from manifest path

		rcvdHeader := jsonhttptest.ResponseDirectCheckBinaryResponse(t, client, http.MethodGet, hopDownloadResource(manifestFileReference.String(), filePath), nil, http.StatusOK, []byte(sampleHtml), nil)
		cd := rcvdHeader.Get("Content-Disposition")
		_, params, err := mime.ParseMediaType(cd)
		if err != nil {
			t.Fatal(err)
		}
		if params["filename"] != fileName {
			t.Fatal("Invalid file name detected")
		}
		if rcvdHeader.Get("ETag") != fmt.Sprintf("%q", fileContentReference) {
			t.Fatal("Invalid ETags header received")
		}
		if rcvdHeader.Get("Content-Type") != "text/html; charset=utf-8" {
			t.Fatal("Invalid content type detected")
		}

		// check on invalid path

		jsonhttptest.ResponseDirectSendHeadersAndReceiveHeaders(t, client, http.MethodGet, hopDownloadResource(manifestFileReference.String(), missingFilePath), nil, http.StatusBadRequest, jsonhttp.StatusResponse{
			Message: "invalid path address",
			Code:    http.StatusBadRequest,
		}, nil)

	})

}
