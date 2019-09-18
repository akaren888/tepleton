package abi

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tepleton/light-client/certifiers"
	"github.com/tepleton/tmlibs/log"

	"github.com/tepleton/basecoin"
	"github.com/tepleton/basecoin/errors"
	"github.com/tepleton/basecoin/stack"
	"github.com/tepleton/basecoin/state"
)

type checkErr func(error) bool

func noErr(err error) bool {
	return err == nil
}

func genEmptySeed(keys certifiers.ValKeys, chain string, h int,
	appHash []byte, count int) certifiers.Seed {

	vals := keys.ToValidators(10, 0)
	cp := keys.GenCheckpoint(chain, h, nil, vals, appHash, 0, count)
	return certifiers.Seed{cp, vals}
}

// this tests registration without registrar permissions
func TestABIRegister(t *testing.T) {
	assert := assert.New(t)

	// the validators we use to make seeds
	keys := certifiers.GenValKeys(5)
	keys2 := certifiers.GenValKeys(7)
	appHash := []byte{0, 4, 7, 23}
	appHash2 := []byte{12, 34, 56, 78}

	// badSeed doesn't validate
	badSeed := genEmptySeed(keys2, "chain-2", 123, appHash, len(keys2))
	badSeed.Header.AppHash = appHash2

	cases := []struct {
		seed    certifiers.Seed
		checker checkErr
	}{
		{
			genEmptySeed(keys, "chain-1", 100, appHash, len(keys)),
			noErr,
		},
		{
			genEmptySeed(keys, "chain-1", 200, appHash, len(keys)),
			IsAlreadyRegisteredErr,
		},
		{
			badSeed,
			IsInvalidCommitErr,
		},
		{
			genEmptySeed(keys2, "chain-2", 123, appHash2, 5),
			noErr,
		},
	}

	ctx := stack.MockContext("hub", 50)
	store := state.NewMemKVStore()
	app := stack.New().Dispatch(stack.WrapHandler(NewHandler()))

	for i, tc := range cases {
		tx := RegisterChainTx{tc.seed}.Wrap()
		_, err := app.DeliverTx(ctx, store, tx)
		assert.True(tc.checker(err), "%d: %+v", i, err)
	}
}

// this tests registration without registrar permissions
func TestABIRegisterPermissions(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	// the validators we use to make seeds
	keys := certifiers.GenValKeys(4)
	appHash := []byte{0x17, 0x21, 0x5, 0x1e}

	foobar := basecoin.Actor{App: "foo", Address: []byte("bar")}
	baz := basecoin.Actor{App: "baz", Address: []byte("bar")}
	foobaz := basecoin.Actor{App: "foo", Address: []byte("baz")}

	cases := []struct {
		seed      certifiers.Seed
		registrar basecoin.Actor
		signer    basecoin.Actor
		checker   checkErr
	}{
		// no sig, no registrar
		{
			seed:    genEmptySeed(keys, "chain-1", 100, appHash, len(keys)),
			checker: noErr,
		},
		// sig, no registrar
		{
			seed:    genEmptySeed(keys, "chain-2", 100, appHash, len(keys)),
			signer:  foobaz,
			checker: noErr,
		},
		// registrar, no sig
		{
			seed:      genEmptySeed(keys, "chain-3", 100, appHash, len(keys)),
			registrar: foobar,
			checker:   errors.IsUnauthorizedErr,
		},
		// registrar, wrong sig
		{
			seed:      genEmptySeed(keys, "chain-4", 100, appHash, len(keys)),
			signer:    foobaz,
			registrar: foobar,
			checker:   errors.IsUnauthorizedErr,
		},
		// registrar, wrong sig
		{
			seed:      genEmptySeed(keys, "chain-5", 100, appHash, len(keys)),
			signer:    baz,
			registrar: foobar,
			checker:   errors.IsUnauthorizedErr,
		},
		// registrar, proper sig
		{
			seed:      genEmptySeed(keys, "chain-6", 100, appHash, len(keys)),
			signer:    foobar,
			registrar: foobar,
			checker:   noErr,
		},
	}

	store := state.NewMemKVStore()
	app := stack.New().Dispatch(stack.WrapHandler(NewHandler()))

	for i, tc := range cases {
		// set option specifies the registrar
		msg, err := json.Marshal(tc.registrar)
		require.Nil(err, "%+v", err)
		_, err = app.SetOption(log.NewNopLogger(), store,
			NameABI, OptionRegistrar, string(msg))
		require.Nil(err, "%+v", err)

		// add permissions to the context
		ctx := stack.MockContext("hub", 50).WithPermissions(tc.signer)
		tx := RegisterChainTx{tc.seed}.Wrap()
		_, err = app.DeliverTx(ctx, store, tx)
		assert.True(tc.checker(err), "%d: %+v", i, err)
	}
}

func TestABIUpdate(t *testing.T) {

}

func TestABICreatePacket(t *testing.T) {

}

func TestABIPostPacket(t *testing.T) {

}

func TestABISendTx(t *testing.T) {

}
