package handlers

import (
	"bytes"
	"math/rand"

	"github.com/pkg/errors"
	"github.com/tepleton/basecoin"
	"github.com/tepleton/basecoin/types"
)

// store nonce as it's own type so no one can even try to fake it
type nonce int64

type secureContext struct {
	id    nonce
	app   string
	perms []basecoin.Permission
}

func NewContext() basecoin.Context {
	return secureContext{
		id: nonce(rand.Int63()),
	}
}

var _ basecoin.Context = secureContext{}

// WithPermissions will panic if they try to set permission without the proper app
func (c secureContext) WithPermissions(perms ...basecoin.Permission) basecoin.Context {
	// the guard makes sure you only set permissions for the app you are inside
	for _, p := range perms {
		if p.App != c.app {
			err := errors.Errorf("Cannot set permission for %s from %s", c.app, p.App)
			panic(err)
		}
	}

	return secureContext{
		id:    c.id,
		perms: append(c.perms, perms...),
	}
}

func (c secureContext) HasPermission(perm basecoin.Permission) bool {
	for _, p := range c.perms {
		if perm.App == p.App && bytes.Equal(perm.Address, p.Address) {
			return true
		}
	}
	return false
}

// IsParent ensures that this is derived from the given secureClient
func (c secureContext) IsParent(other basecoin.Context) bool {
	so, ok := other.(secureContext)
	if !ok {
		return false
	}
	return c.id == so.id
}

// Reset should give a fresh context,
// but carry on knowledge that this is a child
func (c secureContext) Reset() basecoin.Context {
	return secureContext{
		id:  c.id,
		app: c.app,
	}
}

// withApp is a private method that we can use to properly set the
// app controls in the middleware
func withApp(ctx basecoin.Context, app string) basecoin.Context {
	sc, ok := ctx.(secureContext)
	if !ok {
		return ctx
	}
	return secureContext{
		id:    sc.id,
		app:   app,
		perms: sc.perms,
	}
}

func secureCheck(h basecoin.Checker, parent basecoin.Context) basecoin.Checker {
	next := func(ctx basecoin.Context, store types.KVStore, tx basecoin.Tx) (res basecoin.Result, err error) {
		if !parent.IsParent(ctx) {
			return res, errors.New("Passing in non-child Context")
		}
		return h.CheckTx(ctx, store, tx)
	}
	return basecoin.CheckerFunc(next)
}

func secureDeliver(h basecoin.Deliver, parent basecoin.Context) basecoin.Deliver {
	next := func(ctx basecoin.Context, store types.KVStore, tx basecoin.Tx) (res basecoin.Result, err error) {
		if !parent.IsParent(ctx) {
			return res, errors.New("Passing in non-child Context")
		}
		return h.DeliverTx(ctx, store, tx)
	}
	return basecoin.DeliverFunc(next)
}
