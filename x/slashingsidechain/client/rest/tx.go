package rest

import (
	"bytes"
	"encoding/json"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtxb "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"
	"github.com/cosmos/cosmos-sdk/x/slashingsidechain"
	"github.com/cosmos/cosmos-sdk/x/slashingsidechain/sidechain"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
)

func registerTxRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *codec.Codec, kb keys.Keybase) {
	r.HandleFunc(
		"/slashsc/evidence/submit",
		evidenceSubmitRequestHandlerFn(cdc, kb, cliCtx),
	).Methods("POST")
}

type EvidenceSubmitReq struct {
	BaseReq     utils.BaseReq        `json:"base_req"`
	Submitter   string               `json:"submitter"` // in bech 32
	SideChainId string               `json:"side_chain_id"`
	Headers     []*sidechain.Header  `json:"headers"`
}

func evidenceSubmitRequestHandlerFn(cdc *codec.Codec, kb keys.Keybase, cliCtx context.CLIContext) func(http.ResponseWriter, *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		var req EvidenceSubmitReq
		err = json.Unmarshal(body, &req)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		baseReq := req.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w) {
			return
		}

		info, err := kb.Get(baseReq.Name)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, err.Error())
			return
		}

		submitter, err := sdk.AccAddressFromBech32(req.Submitter)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		if !bytes.Equal(info.GetPubKey().Address(), submitter) {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, "Must use own submitter address")
			return
		}

		if req.Headers == nil || len(req.Headers) != 2 {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, "Must have 2 headers exactly")
			return
		}

		var headers [2]*sidechain.Header
		copy(headers[:],req.Headers[:])
		msg := slashingsidechain.NewMsgSubmitEvidence(submitter, req.SideChainId, headers)

		txBldr := authtxb.TxBuilder{
			Codec:   cdc,
			ChainID: baseReq.ChainID,
		}
		txBldr = txBldr.WithAccountNumber(baseReq.AccountNumber).WithSequence(baseReq.Sequence)
		baseReq.Sequence++

		if utils.HasDryRunArg(r) {
			// Todo return something
			return
		}

		if utils.HasGenerateOnlyArg(r) {
			utils.WriteGenerateStdTxResponse(w, txBldr, []sdk.Msg{msg})
			return
		}

		txBytes, err := txBldr.BuildAndSign(baseReq.Name, baseReq.Password, []sdk.Msg{msg})
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, err.Error())
			return
		}

		res, err := cliCtx.BroadcastTx(txBytes)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		utils.PostProcessResponse(w, cdc, res, cliCtx.Indent)
	}

}
