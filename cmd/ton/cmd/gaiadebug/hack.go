package main

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"
	wrsp "github.com/tepleton/tepleton/wrsp/types"
	"github.com/tepleton/tepleton/crypto"
	cmn "github.com/tepleton/tmlibs/common"
	dbm "github.com/tepleton/tmlibs/db"
	"github.com/tepleton/tmlibs/log"

	bam "github.com/tepleton/tepleton-sdk/baseapp"
	sdk "github.com/tepleton/tepleton-sdk/types"

	"github.com/tepleton/tepleton-sdk/wire"
	"github.com/tepleton/tepleton-sdk/x/auth"
	"github.com/tepleton/tepleton-sdk/x/bank"
	"github.com/tepleton/tepleton-sdk/x/ibc"
	"github.com/tepleton/tepleton-sdk/x/slashing"
	"github.com/tepleton/tepleton-sdk/x/stake"

	ton "github.com/tepleton/tepleton-sdk/cmd/ton/app"
)

func runHackCmd(cmd *cobra.Command, args []string) error {

	if len(args) != 1 {
		return fmt.Errorf("Expected 1 arg")
	}

	// ".tond"
	dataDir := args[0]
	dataDir = path.Join(dataDir, "data")

	// load the app
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	db, err := dbm.NewGoLevelDB("ton", dataDir)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	app := NewGaiaApp(logger, db)

	// print some info
	id := app.LastCommitID()
	lastBlockHeight := app.LastBlockHeight()
	fmt.Println("ID", id)
	fmt.Println("LastBlockHeight", lastBlockHeight)

	//----------------------------------------------------
	// XXX: start hacking!
	//----------------------------------------------------
	// eg. ton-6001 testnet bug
	// We paniced when iterating through the "bypower" keys.
	// The following powerKey was there, but the corresponding "trouble" validator did not exist.
	// So here we do a binary search on the past states to find when the powerKey first showed up ...

	// owner of the validator the bonds, gets revoked, later unbonds, and then later is still found in the bypower store
	trouble := hexToBytes("D3DC0FF59F7C3B548B7AFA365561B87FD0208AF8")
	// this is his "bypower" key
	powerKey := hexToBytes("05303030303030303030303033FFFFFFFFFFFF4C0C0000FFFED3DC0FF59F7C3B548B7AFA365561B87FD0208AF8")

	topHeight := lastBlockHeight
	bottomHeight := int64(0)
	checkHeight := topHeight
	for {
		// load the given version of the state
		err = app.LoadVersion(checkHeight, app.keyMain)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		ctx := app.NewContext(true, wrsp.Header{})

		// check for the powerkey and the validator from the store
		store := ctx.KVStore(app.keyStake)
		res := store.Get(powerKey)
		val, _ := app.stakeKeeper.GetValidator(ctx, trouble)
		fmt.Println("checking height", checkHeight, res, val)
		if res == nil {
			bottomHeight = checkHeight
		} else {
			topHeight = checkHeight
		}
		checkHeight = (topHeight + bottomHeight) / 2
	}
}

func base64ToPub(b64 string) crypto.PubKeyEd25519 {
	data, _ := base64.StdEncoding.DecodeString(b64)
	var pubKey crypto.PubKeyEd25519
	copy(pubKey[:], data)
	return pubKey

}

func hexToBytes(h string) []byte {
	trouble, _ := hex.DecodeString(h)
	return trouble

}

//--------------------------------------------------------------------------------
// NOTE: This is all copied from ton/app/app.go
// so we can access internal fields!

const (
	appName = "GaiaApp"
)

// default home directories for expected binaries
var (
	DefaultCLIHome  = os.ExpandEnv("$HOME/.toncli")
	DefaultNodeHome = os.ExpandEnv("$HOME/.tond")
)

// Extended WRSP application
type GaiaApp struct {
	*bam.BaseApp
	cdc *wire.Codec

	// keys to access the substores
	keyMain     *sdk.KVStoreKey
	keyAccount  *sdk.KVStoreKey
	keyIBC      *sdk.KVStoreKey
	keyStake    *sdk.KVStoreKey
	keySlashing *sdk.KVStoreKey

	// Manage getting and setting accounts
	accountMapper       auth.AccountMapper
	feeCollectionKeeper auth.FeeCollectionKeeper
	coinKeeper          bank.Keeper
	ibcMapper           ibc.Mapper
	stakeKeeper         stake.Keeper
	slashingKeeper      slashing.Keeper
}

func NewGaiaApp(logger log.Logger, db dbm.DB) *GaiaApp {
	cdc := MakeCodec()

	// create your application object
	var app = &GaiaApp{
		BaseApp:     bam.NewBaseApp(appName, cdc, logger, db),
		cdc:         cdc,
		keyMain:     sdk.NewKVStoreKey("main"),
		keyAccount:  sdk.NewKVStoreKey("acc"),
		keyIBC:      sdk.NewKVStoreKey("ibc"),
		keyStake:    sdk.NewKVStoreKey("stake"),
		keySlashing: sdk.NewKVStoreKey("slashing"),
	}

	// define the accountMapper
	app.accountMapper = auth.NewAccountMapper(
		app.cdc,
		app.keyAccount,      // target store
		&auth.BaseAccount{}, // prototype
	)

	// add handlers
	app.coinKeeper = bank.NewKeeper(app.accountMapper)
	app.ibcMapper = ibc.NewMapper(app.cdc, app.keyIBC, app.RegisterCodespace(ibc.DefaultCodespace))
	app.stakeKeeper = stake.NewKeeper(app.cdc, app.keyStake, app.coinKeeper, app.RegisterCodespace(stake.DefaultCodespace))
	app.slashingKeeper = slashing.NewKeeper(app.cdc, app.keySlashing, app.stakeKeeper, app.RegisterCodespace(slashing.DefaultCodespace))

	// register message routes
	app.Router().
		AddRoute("bank", bank.NewHandler(app.coinKeeper)).
		AddRoute("ibc", ibc.NewHandler(app.ibcMapper, app.coinKeeper)).
		AddRoute("stake", stake.NewHandler(app.stakeKeeper))

	// initialize BaseApp
	app.SetInitChainer(app.initChainer)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetEndBlocker(app.EndBlocker)
	app.SetAnteHandler(auth.NewAnteHandler(app.accountMapper, app.feeCollectionKeeper))
	app.MountStoresIAVL(app.keyMain, app.keyAccount, app.keyIBC, app.keyStake, app.keySlashing)
	err := app.LoadLatestVersion(app.keyMain)
	if err != nil {
		cmn.Exit(err.Error())
	}

	return app
}

// custom tx codec
func MakeCodec() *wire.Codec {
	var cdc = wire.NewCodec()
	ibc.RegisterWire(cdc)
	bank.RegisterWire(cdc)
	stake.RegisterWire(cdc)
	slashing.RegisterWire(cdc)
	auth.RegisterWire(cdc)
	sdk.RegisterWire(cdc)
	wire.RegisterCrypto(cdc)
	return cdc
}

// application updates every end block
func (app *GaiaApp) BeginBlocker(ctx sdk.Context, req wrsp.RequestBeginBlock) wrsp.ResponseBeginBlock {
	tags := slashing.BeginBlocker(ctx, req, app.slashingKeeper)

	return wrsp.ResponseBeginBlock{
		Tags: tags.ToKVPairs(),
	}
}

// application updates every end block
// nolint: unparam
func (app *GaiaApp) EndBlocker(ctx sdk.Context, req wrsp.RequestEndBlock) wrsp.ResponseEndBlock {
	validatorUpdates := stake.EndBlocker(ctx, app.stakeKeeper)

	return wrsp.ResponseEndBlock{
		ValidatorUpdates: validatorUpdates,
	}
}

// custom logic for ton initialization
func (app *GaiaApp) initChainer(ctx sdk.Context, req wrsp.RequestInitChain) wrsp.ResponseInitChain {
	stateJSON := req.AppStateBytes
	// TODO is this now the whole genesis file?

	var genesisState ton.GenesisState
	err := app.cdc.UnmarshalJSON(stateJSON, &genesisState)
	if err != nil {
		panic(err) // TODO https://github.com/tepleton/tepleton-sdk/issues/468
		// return sdk.ErrGenesisParse("").TraceCause(err, "")
	}

	// load the accounts
	for _, gacc := range genesisState.Accounts {
		acc := gacc.ToAccount()
		app.accountMapper.SetAccount(ctx, acc)
	}

	// load the initial stake information
	stake.InitGenesis(ctx, app.stakeKeeper, genesisState.StakeData)
	return wrsp.ResponseInitChain{}

}
