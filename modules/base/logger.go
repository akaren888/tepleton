package base

import (
	"time"

	"github.com/tepleton/tmlibs/log"

	"github.com/tepleton/basecoin"
	"github.com/tepleton/basecoin/stack"
	"github.com/tepleton/basecoin/state"
)

// nolint
const (
	NameLogger = "lggr"
)

// Logger catches any panics and returns them as errors instead
type Logger struct{}

// Name of the module - fulfills Middleware interface
func (Logger) Name() string {
	return NameLogger
}

var _ stack.Middleware = Logger{}

// CheckTx logs time and result - fulfills Middlware interface
func (Logger) CheckTx(ctx basecoin.Context, store state.SimpleDB, tx basecoin.Tx, next basecoin.Checker) (res basecoin.CheckResult, err error) {
	start := time.Now()
	res, err = next.CheckTx(ctx, store, tx)
	delta := time.Now().Sub(start)
	// TODO: log some info on the tx itself?
	l := ctx.With("duration", micros(delta))
	if err == nil {
		l.Debug("CheckTx", "log", res.Log)
	} else {
		l.Info("CheckTx", "err", err)
	}
	return
}

// DeliverTx logs time and result - fulfills Middlware interface
func (Logger) DeliverTx(ctx basecoin.Context, store state.SimpleDB, tx basecoin.Tx, next basecoin.Deliver) (res basecoin.DeliverResult, err error) {
	start := time.Now()
	res, err = next.DeliverTx(ctx, store, tx)
	delta := time.Now().Sub(start)
	// TODO: log some info on the tx itself?
	l := ctx.With("duration", micros(delta))
	if err == nil {
		l.Info("DeliverTx", "log", res.Log)
	} else {
		l.Error("DeliverTx", "err", err)
	}
	return
}

// SetOption logs time and result - fulfills Middlware interface
func (Logger) SetOption(l log.Logger, store state.SimpleDB, module, key, value string, next basecoin.SetOptioner) (string, error) {
	start := time.Now()
	res, err := next.SetOption(l, store, module, key, value)
	delta := time.Now().Sub(start)
	// TODO: log the value being set also?
	l = l.With("duration", micros(delta)).With("mod", module).With("key", key)
	if err == nil {
		l.Info("SetOption", "log", res)
	} else {
		l.Error("SetOption", "err", err)
	}
	return res, err
}

// micros returns how many microseconds passed in a call
func micros(d time.Duration) int {
	return int(d.Seconds() * 1000000)
}
