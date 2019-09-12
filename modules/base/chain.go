package base

import (
	"github.com/tepleton/basecoin"
	"github.com/tepleton/basecoin/errors"
	"github.com/tepleton/basecoin/stack"
	"github.com/tepleton/basecoin/state"
)

//nolint
const (
	NameChain = "chain"
)

// Chain enforces that this tx was bound to the named chain
type Chain struct {
	stack.PassOption
}

// Name of the module - fulfills Middleware interface
func (Chain) Name() string {
	return NameChain
}

var _ stack.Middleware = Chain{}

// CheckTx makes sure we are on the proper chain - fulfills Middlware interface
func (c Chain) CheckTx(ctx basecoin.Context, store state.KVStore, tx basecoin.Tx, next basecoin.Checker) (res basecoin.Result, err error) {
	stx, err := c.checkChain(ctx.ChainID(), tx)
	if err != nil {
		return res, err
	}
	return next.CheckTx(ctx, store, stx)
}

// DeliverTx makes sure we are on the proper chain - fulfills Middlware interface
func (c Chain) DeliverTx(ctx basecoin.Context, store state.KVStore, tx basecoin.Tx, next basecoin.Deliver) (res basecoin.Result, err error) {
	stx, err := c.checkChain(ctx.ChainID(), tx)
	if err != nil {
		return res, err
	}
	return next.DeliverTx(ctx, store, stx)
}

// checkChain makes sure the tx is a Chain Tx and is on the proper chain
func (c Chain) checkChain(chainID string, tx basecoin.Tx) (basecoin.Tx, error) {
	ctx, ok := tx.Unwrap().(ChainTx)
	if !ok {
		return tx, errors.ErrNoChain()
	}
	if ctx.ChainID != chainID {
		return tx, errors.ErrWrongChain(ctx.ChainID)
	}
	return ctx.Tx, nil
}