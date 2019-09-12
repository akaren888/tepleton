package stack

import (
	"fmt"

	"github.com/tepleton/tmlibs/log"

	"github.com/tepleton/basecoin"
	"github.com/tepleton/basecoin/errors"
	"github.com/tepleton/basecoin/state"
)

// nolint
const (
	NameRecovery = "rcvr"
)

// Recovery catches any panics and returns them as errors instead
type Recovery struct{}

// Name of the module - fulfills Middleware interface
func (Recovery) Name() string {
	return NameRecovery
}

var _ Middleware = Recovery{}

// CheckTx catches any panic and converts to error - fulfills Middlware interface
func (Recovery) CheckTx(ctx basecoin.Context, store state.KVStore, tx basecoin.Tx, next basecoin.Checker) (res basecoin.Result, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = normalizePanic(r)
		}
	}()
	return next.CheckTx(ctx, store, tx)
}

// DeliverTx catches any panic and converts to error - fulfills Middlware interface
func (Recovery) DeliverTx(ctx basecoin.Context, store state.KVStore, tx basecoin.Tx, next basecoin.Deliver) (res basecoin.Result, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = normalizePanic(r)
		}
	}()
	return next.DeliverTx(ctx, store, tx)
}

// SetOption catches any panic and converts to error - fulfills Middlware interface
func (Recovery) SetOption(l log.Logger, store state.KVStore, module, key, value string, next basecoin.SetOptioner) (log string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = normalizePanic(r)
		}
	}()
	return next.SetOption(l, store, module, key, value)
}

// normalizePanic makes sure we can get a nice TMError (with stack) out of it
func normalizePanic(p interface{}) error {
	if err, isErr := p.(error); isErr {
		return errors.Wrap(err)
	}
	msg := fmt.Sprintf("%v", p)
	return errors.ErrInternal(msg)
}