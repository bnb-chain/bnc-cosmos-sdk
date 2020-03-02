package client

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/tendermint/go-amino"
	abci "github.com/tendermint/tendermint/abci/types"
	cryptoAmino "github.com/tendermint/tendermint/crypto/encoding/amino"
	cmn "github.com/tendermint/tendermint/libs/common"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Codec = amino.Codec

var Cdc *Codec

func init() {
	cdc := amino.NewCodec()
	cryptoAmino.RegisterAmino(cdc)
	Cdc = cdc.Seal()
}

func SerializeBindPackage(bep2TokenSymbol string, bep2TokenOwner sdk.AccAddress, contractAddr []byte, totalSupply int64, peggyAmount int64, relayReward int64) ([]byte, error) {
	serializedBytes := make([]byte, 32+20+20+32+32+32)
	if len(bep2TokenSymbol) > 32 {
		return nil, fmt.Errorf("bep2 token symbol length should be no more than 32")
	}
	copy(serializedBytes[0:32], bep2TokenSymbol)
	copy(serializedBytes[32:52], bep2TokenOwner)

	if len(contractAddr) != 20 {
		return nil, fmt.Errorf("contract address length must be 20")
	}
	copy(serializedBytes[52:72], contractAddr)

	binary.BigEndian.PutUint64(serializedBytes[96:104], uint64(totalSupply))
	binary.BigEndian.PutUint64(serializedBytes[128:136], uint64(peggyAmount))
	binary.BigEndian.PutUint64(serializedBytes[160:168], uint64(relayReward))

	return serializedBytes, nil
}

func SerializeTimeoutPackage(refundAmount int64, contractAddr []byte, refundAddr []byte) ([]byte, error) {
	serializedBytes := make([]byte, 32+20+20)
	if len(contractAddr) != 20 || len(refundAddr) != 20 {
		return nil, fmt.Errorf("length of address must be 20")
	}
	binary.BigEndian.PutUint64(serializedBytes[24:32], uint64(refundAmount))
	copy(serializedBytes[32:52], contractAddr)
	copy(serializedBytes[52:], refundAddr)

	return serializedBytes, nil
}

func SerializeTransferPackage(bep2TokenSymbol string, contractAddr []byte, sender []byte, recipient []byte, amount int64, expireTime int64, relayReward int64) ([]byte, error) {
	serializedBytes := make([]byte, 32+20+20+20+32+32+32)
	if len(bep2TokenSymbol) > 32 {
		return nil, fmt.Errorf("bep2 token symbol length should be no more than 32")
	}
	copy(serializedBytes[0:32], bep2TokenSymbol)

	if len(contractAddr) != 20 || len(sender) != 20 || len(recipient) != 20 {
		return nil, fmt.Errorf("length of address must be 20")
	}
	copy(serializedBytes[32:52], contractAddr)
	copy(serializedBytes[52:72], sender)
	copy(serializedBytes[72:92], recipient)

	binary.BigEndian.PutUint64(serializedBytes[116:124], uint64(amount))
	binary.BigEndian.PutUint64(serializedBytes[148:156], uint64(expireTime))
	binary.BigEndian.PutUint64(serializedBytes[180:188], uint64(relayReward))

	return serializedBytes, nil
}

func isQueryStoreWithProof(path string) bool {
	if !strings.HasPrefix(path, "/") {
		return false
	}

	paths := strings.SplitN(path[1:], "/", 3)
	if len(paths) != 3 {
		return false
	}

	if store.RequireProof("/" + paths[2]) {
		return true
	}

	return false
}

func QueryStore(ctx context.CLIContext, path string, key cmn.HexBytes) (res abci.ResponseQuery, err error) {
	node, err := ctx.GetNode()
	if err != nil {
		return res, err
	}

	opts := rpcclient.ABCIQueryOptions{
		Height: ctx.Height,
		Prove:  !ctx.TrustNode,
	}

	result, err := node.ABCIQueryWithOptions(path, key, opts)
	if err != nil {
		return res, err
	}

	resp := result.Response
	if !resp.IsOK() {
		return res, errors.Errorf(resp.Log)
	}

	// data from trusted node or subspace query doesn't need verification
	if ctx.TrustNode || !isQueryStoreWithProof(path) {
		return resp, nil
	}

	header, err := queryTendermintHeader(node, resp.Height+1)
	if err != nil {
		return res, err
	}
	headerBytes, err := header.serialize()
	if err != nil {
		return res, err
	}

	proofBytes, _ := resp.Proof.Marshal()

	fmt.Println(fmt.Sprintf("key: %s", hex.EncodeToString(resp.Key)))
	fmt.Println(fmt.Sprintf("value: %s", hex.EncodeToString(resp.Value)))
	fmt.Println(fmt.Sprintf("proof: %s", hex.EncodeToString(proofBytes)))
	fmt.Println(fmt.Sprintf("height: %d", resp.Height))
	fmt.Println(fmt.Sprintf("header: %s", hex.EncodeToString(headerBytes)))
	fmt.Println(fmt.Sprintf("header height: %d", header.Height))
	fmt.Println(fmt.Sprintf("appHash: %s", header.AppHash.String()))

	return resp, nil
}

type Header struct {
	tmtypes.SignedHeader
	ValidatorSet     *tmtypes.ValidatorSet `json:"validator_set" yaml:"validator_set"`
	NextValidatorSet *tmtypes.ValidatorSet `json:"next_validator_set" yaml:"next_validator_set"`
}

func (h *Header) validate(chainID string) error {
	if err := h.SignedHeader.ValidateBasic(chainID); err != nil {
		return err
	}
	if h.ValidatorSet == nil {
		return fmt.Errorf("validator set is nil")
	}
	if h.NextValidatorSet == nil {
		return fmt.Errorf("next validator set is nil")
	}
	return nil
}

func (h Header) serialize() ([]byte, error) {
	bz, err := Cdc.MarshalBinaryLengthPrefixed(h)
	if err != nil {
		return nil, err
	}
	return bz, nil
}

func DecodeHeader(input []byte) (Header, error) {
	var header Header
	err := Cdc.UnmarshalBinaryLengthPrefixed(input, &header)
	if err != nil {
		return Header{}, err
	}
	return header, nil
}

func queryTendermintHeader(node rpcclient.Client, height int64) (Header, error) {
	prevheight := height - 1

	commit, err := node.Commit(&height)
	if err != nil {
		return Header{}, err
	}

	validators, err := node.Validators(&prevheight)
	if err != nil {
		return Header{}, err
	}

	nextvalidators, err := node.Validators(&height)
	if err != nil {
		return Header{}, err
	}

	header := Header{
		SignedHeader:     commit.SignedHeader,
		ValidatorSet:     tmtypes.NewValidatorSet(validators.Validators),
		NextValidatorSet: tmtypes.NewValidatorSet(nextvalidators.Validators),
	}

	return header, nil
}
