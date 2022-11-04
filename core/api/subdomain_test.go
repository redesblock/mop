package api_test

import (
	"fmt"
	"net/http"
	"path"
	"testing"

	"github.com/redesblock/mop/core/api"
	"github.com/redesblock/mop/core/api/jsonhttp/jsonhttptest"
	"github.com/redesblock/mop/core/cluster"
	mockpost "github.com/redesblock/mop/core/incentives/voucher/mock"
	"github.com/redesblock/mop/core/log"
	resolverMock "github.com/redesblock/mop/core/resolver/mock"
	statestore "github.com/redesblock/mop/core/storer/statestore/mock"
	"github.com/redesblock/mop/core/storer/storage/mock"
	"github.com/redesblock/mop/core/tags"
)

func TestSubdomains(t *testing.T) {

	for _, tc := range []struct {
		name                string
		files               []f
		expectedReference   cluster.Address
		wantIndexFilename   string
		wantErrorFilename   string
		indexFilenameOption jsonhttptest.Option
		errorFilenameOption jsonhttptest.Option
	}{
		{
			name:              "nested files with extension",
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
			name:                "explicit index and error filename",
			expectedReference:   cluster.MustParseHexAddress("86ddb8ed182dd5729766d2988187333e35bedd62637a2b9bf901b003e1cf9f16"),
			wantIndexFilename:   "index.html",
			wantErrorFilename:   "error.html",
			indexFilenameOption: jsonhttptest.WithRequestHeader(api.ClusterIndexDocumentHeader, "index.html"),
			errorFilenameOption: jsonhttptest.WithRequestHeader(api.ClusterErrorDocumentHeader, "error.html"),
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
	} {
		t.Run(tc.name, func(t *testing.T) {
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
					Resolver: resolverMock.NewResolver(
						resolverMock.WithResolveFunc(
							func(string) (cluster.Address, error) {
								return tc.expectedReference, nil
							},
						),
					),
				})
			)

			validateAltPath := func(t *testing.T, fromPath, toPath string) {
				t.Helper()

				var respBytes []byte

				jsonhttptest.Request(t, client, http.MethodGet,
					fmt.Sprintf("http://test.eth.cluster.localhost/%s", toPath), http.StatusOK,
					jsonhttptest.WithPutResponseBody(&respBytes),
				)

				jsonhttptest.Request(t, client, http.MethodGet,
					fmt.Sprintf("http://test.eth.cluster.localhost/%s", fromPath), http.StatusOK,
					jsonhttptest.WithExpectedResponse(respBytes),
				)
			}

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

			jsonhttptest.Request(t, client, http.MethodPost, dirUploadResource, http.StatusCreated, options...)

			if resp.Reference.String() == "" {
				t.Fatalf("expected file reference, did not got any")
			}

			if tc.expectedReference.String() != resp.Reference.String() {
				t.Fatalf("got unexpected reference exp %s got %s", tc.expectedReference.String(), resp.Reference.String())
			}

			for _, f := range tc.files {
				jsonhttptest.Request(
					t, client, http.MethodGet,
					fmt.Sprintf("http://test.eth.cluster.localhost/%s", path.Join(f.dir, f.name)),
					http.StatusOK,
					jsonhttptest.WithExpectedResponse(f.data),
				)
			}

			if tc.wantIndexFilename != "" {
				validateAltPath(t, "", tc.wantIndexFilename)
			}
			if tc.wantErrorFilename != "" {
				validateAltPath(t, "_non_existent_file_path_", tc.wantErrorFilename)
			}
		})
	}
}
