package basecoin

import (
	wrsp "github.com/tepleton/wrsp/types"
	"github.com/tepleton/go-wire/data"
	"github.com/tepleton/tmlibs/log"

	"github.com/tepleton/basecoin/types"
)

// Handler is anything that processes a transaction
type Handler interface {
	Checker
	Deliver
	SetOptioner
	Named
	// TODO: flesh these out as well
	// InitChain(store types.KVStore, vals []*wrsp.Validator)
	// BeginBlock(store types.KVStore, hash []byte, header *wrsp.Header)
	// EndBlock(store types.KVStore, height uint64) wrsp.ResponseEndBlock
}

type Named interface {
	Name() string
}

type Checker interface {
	CheckTx(ctx Context, store types.KVStore, tx Tx) (Result, error)
}

// CheckerFunc (like http.HandlerFunc) is a shortcut for making wrapers
type CheckerFunc func(Context, types.KVStore, Tx) (Result, error)

func (c CheckerFunc) CheckTx(ctx Context, store types.KVStore, tx Tx) (Result, error) {
	return c(ctx, store, tx)
}

type Deliver interface {
	DeliverTx(ctx Context, store types.KVStore, tx Tx) (Result, error)
}

// DeliverFunc (like http.HandlerFunc) is a shortcut for making wrapers
type DeliverFunc func(Context, types.KVStore, Tx) (Result, error)

func (c DeliverFunc) DeliverTx(ctx Context, store types.KVStore, tx Tx) (Result, error) {
	return c(ctx, store, tx)
}

type SetOptioner interface {
	SetOption(l log.Logger, store types.KVStore, key, value string) (string, error)
}

// SetOptionFunc (like http.HandlerFunc) is a shortcut for making wrapers
type SetOptionFunc func(log.Logger, types.KVStore, string, string) (string, error)

func (c SetOptionFunc) SetOption(l log.Logger, store types.KVStore, key, value string) (string, error) {
	return c(l, store, key, value)
}

// Result captures any non-error wrsp result
// to make sure people use error for error cases
type Result struct {
	Data data.Bytes
	Log  string
}

func (r Result) ToWRSP() wrsp.Result {
	return wrsp.Result{
		Data: r.Data,
		Log:  r.Log,
	}
}

// placeholders
// holders
type NopCheck struct{}

func (_ NopCheck) CheckTx(Context, types.KVStore, Tx) (r Result, e error) { return }

type NopDeliver struct{}

func (_ NopDeliver) DeliverTx(Context, types.KVStore, Tx) (r Result, e error) { return }

type NopOption struct{}

func (_ NopOption) SetOption(log.Logger, types.KVStore, string, string) (string, error) {
	return "", nil
}
