package main

import (
	"flag"

	"github.com/tepleton/wrsp/server"
	"github.com/tepleton/basecoin/app"
	"github.com/tepleton/basecoin/plugins/counter"
	cmn "github.com/tepleton/go-common"
	eyes "github.com/tepleton/merkleeyes/client"
)

func main() {
	addrPtr := flag.String("address", "tcp://0.0.0.0:46658", "Listen address")
	eyesPtr := flag.String("eyes", "local", "MerkleEyes address, or 'local' for embedded")
	genFilePath := flag.String("genesis", "", "Genesis file, if any")
	flag.Parse()

	// Connect to MerkleEyes
	eyesCli, err := eyes.NewClient(*eyesPtr)
	if err != nil {
		cmn.Exit("connect to MerkleEyes: " + err.Error())
	}

	// Create Basecoin app
	app := app.NewBasecoin(eyesCli)

	// add plugins
	// TODO: add some more, like the cool voting app
	counter := counter.New("counter")
	app.RegisterPlugin(counter)

	// If genesis file was specified, set key-value options
	if *genFilePath != "" {
		err := app.LoadGenesis(*genFilePath)
		if err != nil {
			cmn.Exit(cmn.Fmt("%+v", err))
		}
	}

	// Start the listener
	svr, err := server.NewServer(*addrPtr, "socket", app)
	if err != nil {
		cmn.Exit("create listener: " + err.Error())
	}

	// Wait forever
	cmn.TrapSignal(func() {
		// Cleanup
		svr.Stop()
	})

}
