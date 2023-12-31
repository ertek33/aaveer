package rpc

import (
	"context"

	"github.com/Conflux-Chain/go-conflux-sdk/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// txPoolAPI provides core space txPool RPC proxy API.
type txPoolAPI struct{}

func (api *txPoolAPI) NextNonce(ctx context.Context, address types.Address) (*hexutil.Big, error) {
	return GetCfxClientFromContext(ctx).TxPool().NextNonce(address)
}
