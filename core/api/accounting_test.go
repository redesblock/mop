package api_test

import (
	"errors"
	"math/big"
	"net/http"
	"reflect"
	"testing"

	"github.com/redesblock/mop/core/api"
	"github.com/redesblock/mop/core/api/jsonhttp"
	"github.com/redesblock/mop/core/api/jsonhttp/jsonhttptest"
	"github.com/redesblock/mop/core/incentives/bookkeeper"
	"github.com/redesblock/mop/core/incentives/bookkeeper/mock"
	"github.com/redesblock/mop/core/util/bigint"
)

func TestAccountingInfo(t *testing.T) {
	accountingFunc := func() (map[string]bookkeeper.PeerInfo, error) {
		ret := make(map[string]bookkeeper.PeerInfo)
		ret["MOPF"] = bookkeeper.PeerInfo{
			Balance:               big.NewInt(25),
			ThresholdReceived:     big.NewInt(37),
			ThresholdGiven:        big.NewInt(49),
			SurplusBalance:        big.NewInt(74),
			ReservedBalance:       big.NewInt(85),
			ShadowReservedBalance: big.NewInt(92),
			GhostBalance:          big.NewInt(94),
		}
		ret["M33F"] = bookkeeper.PeerInfo{
			Balance:               big.NewInt(26),
			ThresholdReceived:     big.NewInt(38),
			ThresholdGiven:        big.NewInt(50),
			SurplusBalance:        big.NewInt(75),
			ReservedBalance:       big.NewInt(86),
			ShadowReservedBalance: big.NewInt(93),
			GhostBalance:          big.NewInt(95),
		}
		ret["ME3F"] = bookkeeper.PeerInfo{
			Balance:               big.NewInt(27),
			ThresholdReceived:     big.NewInt(39),
			ThresholdGiven:        big.NewInt(51),
			SurplusBalance:        big.NewInt(76),
			ReservedBalance:       big.NewInt(87),
			ShadowReservedBalance: big.NewInt(94),
			GhostBalance:          big.NewInt(96),
		}

		return ret, nil
	}

	testServer, _, _, _ := newTestServer(t, testServerOptions{
		DebugAPI:       true,
		AccountingOpts: []mock.Option{mock.WithPeerAccountingFunc(accountingFunc)},
	})

	expected := &api.PeerData{
		InfoResponse: map[string]api.PeerDataResponse{
			"MOPF": {
				Balance:               bigint.Wrap(big.NewInt(25)),
				ThresholdReceived:     bigint.Wrap(big.NewInt(37)),
				ThresholdGiven:        bigint.Wrap(big.NewInt(49)),
				SurplusBalance:        bigint.Wrap(big.NewInt(74)),
				ReservedBalance:       bigint.Wrap(big.NewInt(85)),
				ShadowReservedBalance: bigint.Wrap(big.NewInt(92)),
				GhostBalance:          bigint.Wrap(big.NewInt(94)),
			},
			"M33F": {
				Balance:               bigint.Wrap(big.NewInt(26)),
				ThresholdReceived:     bigint.Wrap(big.NewInt(38)),
				ThresholdGiven:        bigint.Wrap(big.NewInt(50)),
				SurplusBalance:        bigint.Wrap(big.NewInt(75)),
				ReservedBalance:       bigint.Wrap(big.NewInt(86)),
				ShadowReservedBalance: bigint.Wrap(big.NewInt(93)),
				GhostBalance:          bigint.Wrap(big.NewInt(95)),
			},
			"ME3F": {
				Balance:               bigint.Wrap(big.NewInt(27)),
				ThresholdReceived:     bigint.Wrap(big.NewInt(39)),
				ThresholdGiven:        bigint.Wrap(big.NewInt(51)),
				SurplusBalance:        bigint.Wrap(big.NewInt(76)),
				ReservedBalance:       bigint.Wrap(big.NewInt(87)),
				ShadowReservedBalance: bigint.Wrap(big.NewInt(94)),
				GhostBalance:          bigint.Wrap(big.NewInt(96)),
			},
		},
	}

	// We expect a list of items unordered by peer:
	var got *api.PeerData
	jsonhttptest.Request(t, testServer, http.MethodGet, "/bookkeeper", http.StatusOK,
		jsonhttptest.WithUnmarshalJSONResponse(&got),
	)

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("got bookkeeper: %v, expected: %v", got, expected)
	}

}

func TestAccountingInfoError(t *testing.T) {
	wantErr := errors.New("ASDF")
	accountingFunc := func() (map[string]bookkeeper.PeerInfo, error) {
		return nil, wantErr
	}
	testServer, _, _, _ := newTestServer(t, testServerOptions{
		DebugAPI:       true,
		AccountingOpts: []mock.Option{mock.WithPeerAccountingFunc(accountingFunc)},
	})

	jsonhttptest.Request(t, testServer, http.MethodGet, "/bookkeeper", http.StatusInternalServerError,
		jsonhttptest.WithExpectedJSONResponse(jsonhttp.StatusResponse{
			Message: api.HttpErrGetAccountingInfo,
			Code:    http.StatusInternalServerError,
		}),
	)
}
