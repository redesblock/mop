package api_test

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"path"
	"strconv"
	"testing"

	"github.com/redesblock/mop/core/api"
	"github.com/redesblock/mop/core/api/jsonhttp"
	"github.com/redesblock/mop/core/api/jsonhttp/jsonhttptest"
	"github.com/redesblock/mop/core/cluster"
	"github.com/redesblock/mop/core/file/loadsave"
	mockpost "github.com/redesblock/mop/core/incentives/voucher/mock"
	"github.com/redesblock/mop/core/log"
	"github.com/redesblock/mop/core/manifest"
	statestore "github.com/redesblock/mop/core/storer/statestore/mock"
	"github.com/redesblock/mop/core/storer/storage/mock"
	"github.com/redesblock/mop/core/tags"
)

func TestDirs(t *testing.T) {
	var (
		dirUploadResource   = "/mop"
		mopDownloadResource = func(addr, path string) string { return "/mop/" + addr + "/" + path }
		ctx                 = context.Background()
		storer              = mock.NewStorer()
		mockStatestore      = statestore.NewStateStore()
		logger              = log.Noop
		client, _, _, _     = newTestServer(t, testServerOptions{
			Storer:          storer,
			Tags:            tags.NewTags(mockStatestore, logger),
			Logger:          logger,
			PreventRedirect: true,
			Post:            mockpost.New(mockpost.WithAcceptAll()),
		})
	)

	t.Run("empty request body", func(t *testing.T) {
		jsonhttptest.Request(t, client, http.MethodPost, dirUploadResource,
			http.StatusBadRequest,
			jsonhttptest.WithRequestHeader(api.ClusterVoucherBatchIdHeader, batchOkStr),
			jsonhttptest.WithRequestBody(bytes.NewReader(nil)),
			jsonhttptest.WithRequestHeader(api.ClusterCollectionHeader, "True"),
			jsonhttptest.WithExpectedJSONResponse(jsonhttp.StatusResponse{
				Message: api.InvalidRequest.Error(),
				Code:    http.StatusBadRequest,
			}),
			jsonhttptest.WithRequestHeader(api.ContentTypeHeader, api.ContentTypeTar),
		)
	})

	t.Run("non tar file", func(t *testing.T) {
		file := bytes.NewReader([]byte("some data"))

		jsonhttptest.Request(t, client, http.MethodPost, dirUploadResource,
			http.StatusInternalServerError,
			jsonhttptest.WithRequestHeader(api.ClusterDeferredUploadHeader, "true"),
			jsonhttptest.WithRequestHeader(api.ClusterVoucherBatchIdHeader, batchOkStr),
			jsonhttptest.WithRequestBody(file),
			jsonhttptest.WithRequestHeader(api.ClusterCollectionHeader, "True"),
			jsonhttptest.WithExpectedJSONResponse(jsonhttp.StatusResponse{
				Message: api.DirectoryStoreError.Error(),
				Code:    http.StatusInternalServerError,
			}),
			jsonhttptest.WithRequestHeader(api.ContentTypeHeader, api.ContentTypeTar),
		)
	})

	t.Run("wrong content type", func(t *testing.T) {
		tarReader := tarFiles(t, []f{{
			data: []byte("some data"),
			name: "binary-file",
		}})

		// submit valid tar, but with wrong content-type
		jsonhttptest.Request(t, client, http.MethodPost, dirUploadResource,
			http.StatusBadRequest,
			jsonhttptest.WithRequestHeader(api.ClusterVoucherBatchIdHeader, batchOkStr),
			jsonhttptest.WithRequestBody(tarReader),
			jsonhttptest.WithRequestHeader(api.ClusterCollectionHeader, "True"),
			jsonhttptest.WithExpectedJSONResponse(jsonhttp.StatusResponse{
				Message: api.InvalidContentType.Error(),
				Code:    http.StatusBadRequest,
			}),
			jsonhttptest.WithRequestHeader(api.ContentTypeHeader, "other"),
		)
	})

	// valid tars
	for _, tc := range []struct {
		name                string
		expectedReference   cluster.Address
		encrypt             bool
		wantIndexFilename   string
		wantErrorFilename   string
		indexFilenameOption jsonhttptest.Option
		errorFilenameOption jsonhttptest.Option
		doMultipart         bool
		files               []f // files in dir for test case
	}{
		{
			name:              "non-nested files without extension",
			expectedReference: cluster.MustParseHexAddress("f3312af64715d26b5e1a3dc90f012d2c9cc74a167899dab1d07cdee8c107f939"),
			files: []f{
				{
					data: []byte("first file data"),
					name: "file1",
					dir:  "",
					header: http.Header{
						api.ContentTypeHeader: {""},
					},
				},
				{
					data: []byte("second file data"),
					name: "file2",
					dir:  "",
					header: http.Header{
						api.ContentTypeHeader: {""},
					},
				},
			},
		},
		{
			name:              "nested files with extension",
			doMultipart:       true,
			expectedReference: cluster.MustParseHexAddress("4c9c76d63856102e54092c38a7cd227d769752d768b7adc8c3542e3dd9fcf295"),
			files: []f{
				{
					data: []byte("robots text"),
					name: "robots.txt",
					dir:  "",
					header: http.Header{
						api.ContentTypeHeader: {"text/plain; charset=utf-8"},
					},
				},
				{
					data: []byte("image 1"),
					name: "1.png",
					dir:  "img",
					header: http.Header{
						api.ContentTypeHeader: {"image/png"},
					},
				},
				{
					data: []byte("image 2"),
					name: "2.png",
					dir:  "img",
					header: http.Header{
						api.ContentTypeHeader: {"image/png"},
					},
				},
			},
		},
		{
			name:              "no index filename",
			expectedReference: cluster.MustParseHexAddress("734b93933ed5d0a26ccfebf52a2d250c4f432a02c330fb2d49ce17e6ad46484f"),
			doMultipart:       true,
			files: []f{
				{
					data: []byte("<h1>Cluster"),
					name: "index.html",
					dir:  "",
					header: http.Header{
						api.ContentTypeHeader: {"text/html; charset=utf-8"},
					},
				},
			},
		},
		{
			name:                "explicit index filename",
			expectedReference:   cluster.MustParseHexAddress("e461445df59d0f2c1ed346efe9b826bba35dfca52e82f2857371265fb9e0ae8d"),
			wantIndexFilename:   "index.html",
			indexFilenameOption: jsonhttptest.WithRequestHeader(api.ClusterIndexDocumentHeader, "index.html"),
			doMultipart:         true,
			files: []f{
				{
					data: []byte("<h1>Cluster"),
					name: "index.html",
					dir:  "",
					header: http.Header{
						api.ContentTypeHeader: {"text/html; charset=utf-8"},
					},
				},
			},
		},
		{
			name:                "nested index filename",
			expectedReference:   cluster.MustParseHexAddress("8a5485d20caf5eaf32d6691b72671dc93b30b74701f302bba755dbbb4bac9f0c"),
			wantIndexFilename:   "index.html",
			indexFilenameOption: jsonhttptest.WithRequestHeader(api.ClusterIndexDocumentHeader, "index.html"),
			files: []f{
				{
					data: []byte("<h1>Cluster"),
					name: "index.html",
					dir:  "dir",
					header: http.Header{
						api.ContentTypeHeader: {"text/html; charset=utf-8"},
					},
				},
			},
		},
		{
			name:                "explicit index and error filename",
			expectedReference:   cluster.MustParseHexAddress("86ddb8ed182dd5729766d2988187333e35bedd62637a2b9bf901b003e1cf9f16"),
			wantIndexFilename:   "index.html",
			wantErrorFilename:   "error.html",
			indexFilenameOption: jsonhttptest.WithRequestHeader(api.ClusterIndexDocumentHeader, "index.html"),
			errorFilenameOption: jsonhttptest.WithRequestHeader(api.ClusterErrorDocumentHeader, "error.html"),
			doMultipart:         true,
			files: []f{
				{
					data: []byte("<h1>Cluster"),
					name: "index.html",
					dir:  "",
					header: http.Header{
						api.ContentTypeHeader: {"text/html; charset=utf-8"},
					},
				},
				{
					data: []byte("<h2>404"),
					name: "error.html",
					dir:  "",
					header: http.Header{
						api.ContentTypeHeader: {"text/html; charset=utf-8"},
					},
				},
			},
		},
		{
			name:              "invalid archive paths",
			expectedReference: cluster.MustParseHexAddress("f43396f8def982917869e0e09796896666cb54f0febee830c168850935d69e29"),
			files: []f{
				{
					data:     []byte("<h1>Cluster"),
					name:     "index.html",
					dir:      "",
					filePath: "./index.html",
				},
				{
					data:     []byte("body {}"),
					name:     "app.css",
					dir:      "",
					filePath: "./app.css",
				},
				{
					data: []byte(`User-agent: *
		Disallow: /`),
					name:     "robots.txt",
					dir:      "",
					filePath: "./robots.txt",
				},
			},
		},
		{
			name:    "encrypted",
			encrypt: true,
			files: []f{
				{
					data:     []byte("<h1>Cluster"),
					name:     "index.html",
					dir:      "",
					filePath: "./index.html",
				},
			},
		},
	} {
		verify := func(t *testing.T, resp api.MopUploadResponse) {
			t.Helper()
			// NOTE: reference will be different each time when encryption is enabled
			if !tc.encrypt {
				if !resp.Reference.Equal(tc.expectedReference) {
					t.Fatalf("expected root reference to match %s, got %s", tc.expectedReference, resp.Reference)
				}
			}

			// verify manifest content
			verifyManifest, err := manifest.NewDefaultManifestReference(
				resp.Reference,
				loadsave.NewReadonly(storer),
			)
			if err != nil {
				t.Fatal(err)
			}

			validateFile := func(t *testing.T, file f, filePath string) {
				t.Helper()

				jsonhttptest.Request(t, client, http.MethodGet,
					mopDownloadResource(resp.Reference.String(), filePath),
					http.StatusOK,
					jsonhttptest.WithExpectedResponse(file.data),
					jsonhttptest.WithRequestHeader(api.ContentTypeHeader, file.header.Get(api.ContentTypeHeader)),
				)
			}

			validateIsPermanentRedirect := func(t *testing.T, fromPath, toPath string) {
				t.Helper()

				expectedResponse := fmt.Sprintf("<a href=\"%s\">Permanent Redirect</a>.\n\n",
					mopDownloadResource(resp.Reference.String(), toPath))

				jsonhttptest.Request(t, client, http.MethodGet,
					mopDownloadResource(resp.Reference.String(), fromPath),
					http.StatusPermanentRedirect,
					jsonhttptest.WithExpectedResponse([]byte(expectedResponse)),
				)
			}

			validateAltPath := func(t *testing.T, fromPath, toPath string) {
				t.Helper()

				var respBytes []byte

				jsonhttptest.Request(t, client, http.MethodGet,
					mopDownloadResource(resp.Reference.String(), toPath), http.StatusOK,
					jsonhttptest.WithPutResponseBody(&respBytes),
				)

				jsonhttptest.Request(t, client, http.MethodGet,
					mopDownloadResource(resp.Reference.String(), fromPath), http.StatusOK,
					jsonhttptest.WithExpectedResponse(respBytes),
				)
			}

			// check if each file can be located and read
			for _, file := range tc.files {
				validateFile(t, file, path.Join(file.dir, file.name))
			}

			// check index filename
			if tc.wantIndexFilename != "" {
				entry, err := verifyManifest.Lookup(ctx, manifest.RootPath)
				if err != nil {
					t.Fatal(err)
				}

				manifestRootMetadata := entry.Metadata()
				indexDocumentSuffixPath, ok := manifestRootMetadata[manifest.WebsiteIndexDocumentSuffixKey]
				if !ok {
					t.Fatalf("expected index filename '%s', did not find any", tc.wantIndexFilename)
				}

				// check index suffix for each dir
				for _, file := range tc.files {
					if file.dir != "" {
						validateIsPermanentRedirect(t, file.dir, file.dir+"/")
						validateAltPath(t, file.dir+"/", path.Join(file.dir, indexDocumentSuffixPath))
					}
				}
			}

			// check error filename
			if tc.wantErrorFilename != "" {
				entry, err := verifyManifest.Lookup(ctx, manifest.RootPath)
				if err != nil {
					t.Fatal(err)
				}

				manifestRootMetadata := entry.Metadata()
				errorDocumentPath, ok := manifestRootMetadata[manifest.WebsiteErrorDocumentPathKey]
				if !ok {
					t.Fatalf("expected error filename '%s', did not find any", tc.wantErrorFilename)
				}

				// check error document
				validateAltPath(t, "_non_existent_file_path_", errorDocumentPath)
			}

		}
		t.Run(tc.name, func(t *testing.T) {
			t.Run("tar_upload", func(t *testing.T) {
				// tar all the test case files
				tarReader := tarFiles(t, tc.files)

				var resp api.MopUploadResponse

				options := []jsonhttptest.Option{
					jsonhttptest.WithRequestHeader(api.ClusterDeferredUploadHeader, "true"),
					jsonhttptest.WithRequestHeader(api.ClusterVoucherBatchIdHeader, batchOkStr),
					jsonhttptest.WithRequestBody(tarReader),
					jsonhttptest.WithRequestHeader(api.ClusterCollectionHeader, "True"),
					jsonhttptest.WithRequestHeader(api.ContentTypeHeader, api.ContentTypeTar),
					jsonhttptest.WithUnmarshalJSONResponse(&resp),
				}
				if tc.indexFilenameOption != nil {
					options = append(options, tc.indexFilenameOption)
				}
				if tc.errorFilenameOption != nil {
					options = append(options, tc.errorFilenameOption)
				}
				if tc.encrypt {
					options = append(options, jsonhttptest.WithRequestHeader(api.ClusterEncryptHeader, "true"))
				}

				// verify directory tar upload response
				jsonhttptest.Request(t, client, http.MethodPost, dirUploadResource, http.StatusCreated, options...)

				if resp.Reference.String() == "" {
					t.Fatalf("expected file reference, did not got any")
				}

				verify(t, resp)
			})
			if tc.doMultipart {
				t.Run("multipart_upload", func(t *testing.T) {
					// tar all the test case files
					mwReader, mwBoundary := multipartFiles(t, tc.files)

					var resp api.MopUploadResponse

					options := []jsonhttptest.Option{
						jsonhttptest.WithRequestHeader(api.ClusterDeferredUploadHeader, "true"),
						jsonhttptest.WithRequestHeader(api.ClusterVoucherBatchIdHeader, batchOkStr),
						jsonhttptest.WithRequestBody(mwReader),
						jsonhttptest.WithRequestHeader(api.ClusterCollectionHeader, "True"),
						jsonhttptest.WithRequestHeader(api.ContentTypeHeader, fmt.Sprintf("multipart/form-data; boundary=%q", mwBoundary)),
						jsonhttptest.WithUnmarshalJSONResponse(&resp),
					}
					if tc.indexFilenameOption != nil {
						options = append(options, tc.indexFilenameOption)
					}
					if tc.errorFilenameOption != nil {
						options = append(options, tc.errorFilenameOption)
					}
					if tc.encrypt {
						options = append(options, jsonhttptest.WithRequestHeader(api.ClusterEncryptHeader, "true"))
					}

					// verify directory tar upload response
					jsonhttptest.Request(t, client, http.MethodPost, dirUploadResource, http.StatusCreated, options...)

					if resp.Reference.String() == "" {
						t.Fatalf("expected file reference, did not got any")
					}

					verify(t, resp)
				})
			}
		})
	}
}

func TestEmtpyDir(t *testing.T) {
	var (
		dirUploadResource = "/mop"
		storer            = mock.NewStorer()
		mockStatestore    = statestore.NewStateStore()
		logger            = log.Noop
		client, _, _, _   = newTestServer(t, testServerOptions{
			Storer:          storer,
			Tags:            tags.NewTags(mockStatestore, logger),
			Logger:          logger,
			PreventRedirect: true,
			Post:            mockpost.New(mockpost.WithAcceptAll()),
		})
	)

	tarReader := tarEmptyDir(t)

	jsonhttptest.Request(t, client, http.MethodPost, dirUploadResource,
		http.StatusBadRequest,
		jsonhttptest.WithRequestHeader(api.ClusterVoucherBatchIdHeader, batchOkStr),
		jsonhttptest.WithRequestBody(tarReader),
		jsonhttptest.WithRequestHeader(api.ClusterCollectionHeader, "true"),
		jsonhttptest.WithRequestHeader(api.ContentTypeHeader, api.ContentTypeTar),
		jsonhttptest.WithExpectedJSONResponse(jsonhttp.StatusResponse{
			Message: api.EmptyDir.Error(),
			Code:    http.StatusBadRequest,
		}),
	)
}

// tarFiles receives an array of test case files and creates a new tar with those files as a collection
// it returns a bytes.Buffer which can be used to read the created tar
func tarFiles(t *testing.T, files []f) *bytes.Buffer {
	t.Helper()

	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	for _, file := range files {
		filePath := path.Join(file.dir, file.name)
		if file.filePath != "" {
			filePath = file.filePath
		}

		// create tar header and write it
		hdr := &tar.Header{
			Name: filePath,
			Mode: 0600,
			Size: int64(len(file.data)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}

		// write the file data to the tar
		if _, err := tw.Write(file.data); err != nil {
			t.Fatal(err)
		}
	}

	// finally close the tar writer
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}

	return &buf
}

func tarEmptyDir(t *testing.T) *bytes.Buffer {
	t.Helper()

	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	hdr := &tar.Header{
		Name: "empty/",
		Mode: 0600,
	}

	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatal(err)
	}

	// finally close the tar writer
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}

	return &buf
}

func multipartFiles(t *testing.T, files []f) (*bytes.Buffer, string) {
	t.Helper()

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	for _, file := range files {
		filePath := path.Join(file.dir, file.name)
		if file.filePath != "" {
			filePath = file.filePath
		}

		hdr := make(textproto.MIMEHeader)
		hdr.Set("Content-Disposition", fmt.Sprintf("form-data; name=%q", filePath))

		contentType := file.header.Get(api.ContentTypeHeader)
		if contentType != "" {
			hdr.Set(api.ContentTypeHeader, contentType)

		}
		if len(file.data) > 0 {
			hdr.Set("Content-Length", strconv.Itoa(len(file.data)))

		}
		part, err := mw.CreatePart(hdr)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = io.Copy(part, bytes.NewBuffer(file.data)); err != nil {
			t.Fatal(err)
		}
	}

	// finally close the tar writer
	if err := mw.Close(); err != nil {
		t.Fatal(err)
	}

	return &buf, mw.Boundary()
}

// struct for dir files for test cases
type f struct {
	data     []byte
	name     string
	dir      string
	filePath string
	header   http.Header
}
