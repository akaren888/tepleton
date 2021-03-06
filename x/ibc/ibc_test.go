package ibc

import (
	"testing"

	"github.com/stretchr/testify/require"

	wrsp "github.com/tepleton/tepleton/wrsp/types"
	"github.com/tepleton/tepleton/crypto"
	dbm "github.com/tepleton/tmlibs/db"
	"github.com/tepleton/tmlibs/log"

	"github.com/tepleton/tepleton-sdk/store"
	sdk "github.com/tepleton/tepleton-sdk/types"
	"github.com/tepleton/tepleton-sdk/wire"
	"github.com/tepleton/tepleton-sdk/x/auth"
	"github.com/tepleton/tepleton-sdk/x/bank"
)

// AccountMapper(/Keeper) and IBCMapper should use different StoreKey later

func defaultContext(key sdk.StoreKey) sdk.Context {
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db)
	cms.MountStoreWithDB(key, sdk.StoreTypeIAVL, db)
	cms.LoadLatestVersion()
	ctx := sdk.NewContext(cms, wrsp.Header{}, false, log.NewNopLogger())
	return ctx
}

func newAddress() crypto.Address {
	return crypto.GenPrivKeyEd25519().PubKey().Address()
}

func getCoins(ck bank.Keeper, ctx sdk.Context, addr crypto.Address) (sdk.Coins, sdk.Error) {
	zero := sdk.Coins(nil)
	coins, _, err := ck.AddCoins(ctx, addr, zero)
	return coins, err
}

func makeCodec() *wire.Codec {
	var cdc = wire.NewCodec()

	// Register Msgs
	cdc.RegisterInterface((*sdk.Msg)(nil), nil)
	cdc.RegisterConcrete(bank.MsgSend{}, "test/ibc/Send", nil)
	cdc.RegisterConcrete(bank.MsgIssue{}, "test/ibc/Issue", nil)
	cdc.RegisterConcrete(IBCTransferMsg{}, "test/ibc/IBCTransferMsg", nil)
	cdc.RegisterConcrete(IBCReceiveMsg{}, "test/ibc/IBCReceiveMsg", nil)

	// Register AppAccount
	cdc.RegisterInterface((*auth.Account)(nil), nil)
	cdc.RegisterConcrete(&auth.BaseAccount{}, "test/ibc/Account", nil)
	wire.RegisterCrypto(cdc)

	return cdc
}

func TestIBC(t *testing.T) {
	cdc := makeCodec()

	key := sdk.NewKVStoreKey("ibc")
	ctx := defaultContext(key)

	am := auth.NewAccountMapper(cdc, key, &auth.BaseAccount{})
	ck := bank.NewKeeper(am)

	src := newAddress()
	dest := newAddress()
	chainid := "ibcchain"
	zero := sdk.Coins(nil)
	mycoins := sdk.Coins{sdk.NewCoin("mycoin", 10)}

	coins, _, err := ck.AddCoins(ctx, src, mycoins)
	require.Nil(t, err)
	require.Equal(t, mycoins, coins)

	ibcm := NewMapper(cdc, key, DefaultCodespace)
	h := NewHandler(ibcm, ck)
	packet := IBCPacket{
		SrcAddr:   src,
		DestAddr:  dest,
		Coins:     mycoins,
		SrcChain:  chainid,
		DestChain: chainid,
	}

	store := ctx.KVStore(key)

	var msg sdk.Msg
	var res sdk.Result
	var egl int64
	var igs int64

	egl = ibcm.getEgressLength(store, chainid)
	require.Equal(t, egl, int64(0))

	msg = IBCTransferMsg{
		IBCPacket: packet,
	}
	res = h(ctx, msg)
	require.True(t, res.IsOK())

	coins, err = getCoins(ck, ctx, src)
	require.Nil(t, err)
	require.Equal(t, zero, coins)

	egl = ibcm.getEgressLength(store, chainid)
	require.Equal(t, egl, int64(1))

	igs = ibcm.GetIngressSequence(ctx, chainid)
	require.Equal(t, igs, int64(0))

	msg = IBCReceiveMsg{
		IBCPacket: packet,
		Relayer:   src,
		Sequence:  0,
	}
	res = h(ctx, msg)
	require.True(t, res.IsOK())

	coins, err = getCoins(ck, ctx, dest)
	require.Nil(t, err)
	require.Equal(t, mycoins, coins)

	igs = ibcm.GetIngressSequence(ctx, chainid)
	require.Equal(t, igs, int64(1))

	res = h(ctx, msg)
	require.False(t, res.IsOK())

	igs = ibcm.GetIngressSequence(ctx, chainid)
	require.Equal(t, igs, int64(1))
}
