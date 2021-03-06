package app

import (
	"github.com/tepleton/tepleton-sdk/baseapp"
	sdk "github.com/tepleton/tepleton-sdk/types"
	"github.com/tepleton/tepleton-sdk/x/auth"
)

// initBaseApp() happens after initCapKeys() and initStores().
// initBaseApp() happens before initRoutes().
func (app *BasecoinApp) initBaseApp() {
	app.BaseApp = baseapp.NewBaseApp(appName, app.multiStore)
	app.initBaseAppTxDecoder()
	app.initBaseAppAnteHandler()
}

func (app *BasecoinApp) initBaseAppTxDecoder() {
	cdc := makeTxCodec()
	app.BaseApp.SetTxDecoder(func(txBytes []byte) (sdk.Tx, error) {
		var tx = sdk.StdTx{}
		err := cdc.UnmarshalBinary(txBytes, &tx)
		return tx, err
	})
}

func (app *BasecoinApp) initBaseAppAnteHandler() {
	var authAnteHandler = auth.NewAnteHandler(app.accStore)
	app.BaseApp.SetDefaultAnteHandler(authAnteHandler)
}
