package app

import (
	"encoding/json"

	wrsp "github.com/tepleton/wrsp/types"
	cmn "github.com/tepleton/tmlibs/common"
	dbm "github.com/tepleton/tmlibs/db"
	"github.com/tepleton/tmlibs/log"

	bam "github.com/tepleton/tepleton-sdk/baseapp"
	sdk "github.com/tepleton/tepleton-sdk/types"
	"github.com/tepleton/tepleton-sdk/wire"
	"github.com/tepleton/tepleton-sdk/x/auth"
	"github.com/tepleton/tepleton-sdk/x/bank"
	"github.com/tepleton/tepleton-sdk/x/ibc"
	"github.com/tepleton/tepleton-sdk/x/simplestake"

	"github.com/tepleton/tepleton-sdk/examples/basecoin/types"
)

const (
	appName = "BasecoinApp"
)

// Extended WRSP application
type BasecoinApp struct {
	*bam.BaseApp
	cdc *wire.Codec

	// keys to access the substores
	capKeyMainStore    *sdk.KVStoreKey
	capKeyAccountStore *sdk.KVStoreKey
	capKeyIBCStore     *sdk.KVStoreKey
	capKeyStakingStore *sdk.KVStoreKey

	// Manage getting and setting accounts
	accountMapper sdk.AccountMapper

	// Handle fees
	feeHandler sdk.FeeHandler
}

func NewBasecoinApp(logger log.Logger, db dbm.DB) *BasecoinApp {

	// Create app-level codec for txs and accounts.
	var cdc = MakeCodec()

	// Create your application object.
	var app = &BasecoinApp{
		BaseApp:            bam.NewBaseApp(appName, logger, db),
		cdc:                cdc,
		capKeyMainStore:    sdk.NewKVStoreKey("main"),
		capKeyAccountStore: sdk.NewKVStoreKey("acc"),
		capKeyIBCStore:     sdk.NewKVStoreKey("ibc"),
		capKeyStakingStore: sdk.NewKVStoreKey("stake"),
	}

	// Define the accountMapper.
	app.accountMapper = auth.NewAccountMapper(
		cdc,
		app.capKeyMainStore, // target store
		&types.AppAccount{}, // prototype
	).Seal()

	// Add handlers.
	coinKeeper := bank.NewCoinKeeper(app.accountMapper)
	ibcMapper := ibc.NewIBCMapper(app.cdc, app.capKeyIBCStore)
	stakeKeeper := simplestake.NewKeeper(app.capKeyStakingStore, coinKeeper)
	app.Router().
		AddRoute("bank", bank.NewHandler(coinKeeper)).
		AddRoute("ibc", ibc.NewHandler(ibcMapper, coinKeeper)).
		AddRoute("simplestake", simplestake.NewHandler(stakeKeeper))

	// Define the feeHandler.
	app.feeHandler = func(ctx sdk.Context, tx sdk.Tx, fees sdk.Coins) {}

	// Initialize BaseApp.
	app.SetTxDecoder(app.txDecoder)
	app.SetInitChainer(app.initChainer)
	app.MountStoresIAVL(app.capKeyMainStore, app.capKeyAccountStore, app.capKeyIBCStore, app.capKeyStakingStore)
	app.SetAnteHandler(auth.NewAnteHandler(app.accountMapper, app.feeHandler))
	err := app.LoadLatestVersion(app.capKeyMainStore)
	if err != nil {
		cmn.Exit(err.Error())
	}

	return app
}

// Custom tx codec
func MakeCodec() *wire.Codec {
	var cdc = wire.NewCodec()

	// Register Msgs
	cdc.RegisterInterface((*sdk.Msg)(nil), nil)
	cdc.RegisterConcrete(bank.SendMsg{}, "basecoin/Send", nil)
	cdc.RegisterConcrete(bank.IssueMsg{}, "basecoin/Issue", nil)
	cdc.RegisterConcrete(ibc.IBCTransferMsg{}, "basecoin/IBCTransferMsg", nil)
	cdc.RegisterConcrete(ibc.IBCReceiveMsg{}, "basecoin/IBCReceiveMsg", nil)
	cdc.RegisterConcrete(simplestake.BondMsg{}, "basecoin/BondMsg", nil)
	cdc.RegisterConcrete(simplestake.UnbondMsg{}, "basecoin/UnbondMsg", nil)

	// Register AppAccount
	cdc.RegisterInterface((*sdk.Account)(nil), nil)
	cdc.RegisterConcrete(&types.AppAccount{}, "basecoin/Account", nil)

	// Register crypto.
	wire.RegisterCrypto(cdc)

	return cdc
}

// Custom logic for transaction decoding
func (app *BasecoinApp) txDecoder(txBytes []byte) (sdk.Tx, sdk.Error) {
	var tx = sdk.StdTx{}

	if len(txBytes) == 0 {
		return nil, sdk.ErrTxDecode("txBytes are empty")
	}

	// StdTx.Msg is an interface. The concrete types
	// are registered by MakeTxCodec in bank.RegisterAmino.
	err := app.cdc.UnmarshalBinary(txBytes, &tx)
	if err != nil {
		return nil, sdk.ErrTxDecode("").TraceCause(err, "")
	}
	return tx, nil
}

// Custom logic for basecoin initialization
func (app *BasecoinApp) initChainer(ctx sdk.Context, req wrsp.RequestInitChain) wrsp.ResponseInitChain {
	stateJSON := req.AppStateBytes

	genesisState := new(types.GenesisState)
	err := json.Unmarshal(stateJSON, genesisState)
	if err != nil {
		panic(err) // TODO https://github.com/tepleton/tepleton-sdk/issues/468
		// return sdk.ErrGenesisParse("").TraceCause(err, "")
	}

	for _, gacc := range genesisState.Accounts {
		acc, err := gacc.ToAppAccount()
		if err != nil {
			panic(err) // TODO https://github.com/tepleton/tepleton-sdk/issues/468
			//	return sdk.ErrGenesisParse("").TraceCause(err, "")
		}
		app.accountMapper.SetAccount(ctx, acc)
	}
	return wrsp.ResponseInitChain{}
}
