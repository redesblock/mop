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
	"github.com/redesblock/mop/core/file/loadsave"
	"github.com/redesblock/mop/core/flock"
	"github.com/redesblock/mop/core/jsonhttp"
	"github.com/redesblock/mop/core/jsonhttp/jsonhttptest"
	"github.com/redesblock/mop/core/logging"
	"github.com/redesblock/mop/core/manifest"
	mockpost "github.com/redesblock/mop/core/postage/mock"
	statestore "github.com/redesblock/mop/core/statestore/mock"
	"github.com/redesblock/mop/core/storage/mock"
	"github.com/redesblock/mop/core/tags"
)

func TestDirs(t *testing.T) {
	var (
		dirUploadResource   = "/mop"
		mopDownloadResource = func(addr, path string) string { return "/mop/" + addr + "/" + path }
		ctx                 = context.Background()
		storer              = mock.NewStorer()
		mockStatestore      = statestore.NewStateStore()
		logger              = logging.New(io.Discard, 0)
		client, _, _        = newTestServer(t, testServerOptions{
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
			jsonhttptest.WithRequestHeader(api.FlockPostageBatchIdHeader, batchOkStr),
			jsonhttptest.WithRequestBody(bytes.NewReader(nil)),
			jsonhttptest.WithRequestHeader(api.FlockCollectionHeader, "True"),
			jsonhttptest.WithExpectedJSONResponse(jsonhttp.StatusResponse{
				Message: api.InvalidRequest.Error(),
				Code:    http.StatusBadRequest,
			}),
			jsonhttptest.WithRequestHeader("Content-Type", api.ContentTypeTar),
		)
	})

	t.Run("non tar file", func(t *testing.T) {
		file := bytes.NewReader([]byte("some data"))

		jsonhttptest.Request(t, client, http.MethodPost, dirUploadResource,
			http.StatusInternalServerError,
			jsonhttptest.WithRequestHeader(api.FlockPostageBatchIdHeader, batchOkStr),
			jsonhttptest.WithRequestBody(file),
			jsonhttptest.WithRequestHeader(api.FlockCollectionHeader, "True"),
			jsonhttptest.WithExpectedJSONResponse(jsonhttp.StatusResponse{
				Message: api.DirectoryStoreError.Error(),
				Code:    http.StatusInternalServerError,
			}),
			jsonhttptest.WithRequestHeader("Content-Type", api.ContentTypeTar),
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
			jsonhttptest.WithRequestHeader(api.FlockPostageBatchIdHeader, batchOkStr),
			jsonhttptest.WithRequestBody(tarReader),
			jsonhttptest.WithRequestHeader(api.FlockCollectionHeader, "True"),
			jsonhttptest.WithExpectedJSONResponse(jsonhttp.StatusResponse{
				Message: api.InvalidContentType.Error(),
				Code:    http.StatusBadRequest,
			}),
			jsonhttptest.WithRequestHeader("Content-Type", "other"),
		)
	})

	// valid tars
	for _, tc := range []struct {
		name                string
		expectedReference   flock.Address
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
			expectedReference: flock.MustParseHexAddress("f3312af64715d26b5e1a3dc90f012d2c9cc74a167899dab1d07cdee8c107f939"),
			files: []f{
				{
					data: []byte("first file data"),
					name: "file1",
					dir:  "",
					header: http.Header{
						"Content-Type": {""},
					},
				},
				{
					data: []byte("second file data"),
					name: "file2",
					dir:  "",
					header: http.Header{
						"Content-Type": {""},
					},
				},
			},
		},
		{
			name:              "nested files with extension",
			expectedReference: flock.MustParseHexAddress("4c9c76d63856102e54092c38a7cd227d769752d768b7adc8c3542e3dd9fcf295"),
			files: []f{
				{
					data: []byte("robots text"),
					name: "robots.txt",
					dir:  "",
					header: http.Header{
						"Content-Type": {"text/plain; charset=utf-8"},
					},
				},
				{
					data: []byte("image 1"),
					name: "1.png",
					dir:  "img",
					header: http.Header{
						"Content-Type": {"image/png"},
					},
				},
				{
					data: []byte("image 2"),
					name: "2.png",
					dir:  "img",
					header: http.Header{
						"Content-Type": {"image/png"},
					},
				},
			},
		},
		{
			name:              "no index filename",
			expectedReference: flock.MustParseHexAddress("350dd938021b8c68d6de9e23003e57219301061b6c0bb1a3c9ea537a8b246e4c"),
			doMultipart:       true,
			files: []f{
				{
					data: []byte("<h1>Flock"),
					name: "index.html",
					dir:  "",
					header: http.Header{
						"Content-Type": {"text/html; charset=utf-8"},
					},
				},
			},
		},
		{
			name:                "explicit index filename",
			expectedReference:   flock.MustParseHexAddress("8e25122f32b69302b7134e697eff6aba4752c7ef8a45b7d3ff92fdad0c6bff1b"),
			wantIndexFilename:   "index.html",
			indexFilenameOption: jsonhttptest.WithRequestHeader(api.FlockIndexDocumentHeader, "index.html"),
			doMultipart:         true,
			files: []f{
				{
					data: []byte("<h1>Flock"),
					name: "index.html",
					dir:  "",
					header: http.Header{
						"Content-Type": {"text/html; charset=utf-8"},
					},
				},
			},
		},
		{
			name:                "nested index filename",
			expectedReference:   flock.MustParseHexAddress("3db829746e08889ed0b216fe435df490998997c3187f4168e9a3e603ec21b087"),
			wantIndexFilename:   "index.html",
			indexFilenameOption: jsonhttptest.WithRequestHeader(api.FlockIndexDocumentHeader, "index.html"),
			files: []f{
				{
					data: []byte("<h1>Flock"),
					name: "index.html",
					dir:  "dir",
					header: http.Header{
						"Content-Type": {"text/html; charset=utf-8"},
					},
				},
			},
		},
		{
			name:                "explicit index and error filename",
			expectedReference:   flock.MustParseHexAddress("12e5cf61226e15f514da16b0787ca351d075619aeea75a2d320bd34b4475834f"),
			wantIndexFilename:   "index.html",
			wantErrorFilename:   "error.html",
			indexFilenameOption: jsonhttptest.WithRequestHeader(api.FlockIndexDocumentHeader, "index.html"),
			errorFilenameOption: jsonhttptest.WithRequestHeader(api.FlockErrorDocumentHeader, "error.html"),
			doMultipart:         true,
			files: []f{
				{
					data: []byte("<h1>Flock"),
					name: "index.html",
					dir:  "",
					header: http.Header{
						"Content-Type": {"text/html; charset=utf-8"},
					},
				},
				{
					data: []byte("<h2>404"),
					name: "error.html",
					dir:  "",
					header: http.Header{
						"Content-Type": {"text/html; charset=utf-8"},
					},
				},
			},
		},
		{
			name:              "invalid archive paths",
			expectedReference: flock.MustParseHexAddress("d4a822f936cce42ab13b0c2660631bad84d525f3cf0ff403bafa80f44645bb3c"),
			files: []f{
				{
					data:     []byte("<h1>Flock"),
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
					data:     []byte("<h1>Flock"),
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
					jsonhttptest.WithRequestHeader("Content-Type", file.header.Get("Content-Type")),
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
					jsonhttptest.WithRequestHeader(api.FlockPostageBatchIdHeader, batchOkStr),
					jsonhttptest.WithRequestBody(tarReader),
					jsonhttptest.WithRequestHeader(api.FlockCollectionHeader, "True"),
					jsonhttptest.WithRequestHeader("Content-Type", api.ContentTypeTar),
					jsonhttptest.WithUnmarshalJSONResponse(&resp),
				}
				if tc.indexFilenameOption != nil {
					options = append(options, tc.indexFilenameOption)
				}
				if tc.errorFilenameOption != nil {
					options = append(options, tc.errorFilenameOption)
				}
				if tc.encrypt {
					options = append(options, jsonhttptest.WithRequestHeader(api.FlockEncryptHeader, "true"))
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
						jsonhttptest.WithRequestHeader(api.FlockPostageBatchIdHeader, batchOkStr),
						jsonhttptest.WithRequestBody(mwReader),
						jsonhttptest.WithRequestHeader(api.FlockCollectionHeader, "True"),
						jsonhttptest.WithRequestHeader("Content-Type", fmt.Sprintf("multipart/form-data; boundary=%q", mwBoundary)),
						jsonhttptest.WithUnmarshalJSONResponse(&resp),
					}
					if tc.indexFilenameOption != nil {
						options = append(options, tc.indexFilenameOption)
					}
					if tc.errorFilenameOption != nil {
						options = append(options, tc.errorFilenameOption)
					}
					if tc.encrypt {
						options = append(options, jsonhttptest.WithRequestHeader(api.FlockEncryptHeader, "true"))
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

func multipartFiles(t *testing.T, files []f) (*bytes.Buffer, string) {
	t.Helper()

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	for _, file := range files {
		hdr := make(textproto.MIMEHeader)
		if file.name != "" {
			hdr.Set("Content-Disposition", fmt.Sprintf("form-data; name=%q", file.name))

		}
		contentType := file.header.Get("Content-Type")
		if contentType != "" {
			hdr.Set("Content-Type", contentType)

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
