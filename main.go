package main

import (
	"flag"

	"github.com/tepleton/basecoin/app"
	"github.com/tepleton/basecoin/types"
	. "github.com/tepleton/go-common"
	"github.com/tepleton/go-wire"
	eyes "github.com/tepleton/merkleeyes/client"
	"github.com/tepleton/wrsp/server"
)

func main() {

	addrPtr := flag.String("address", "tcp://0.0.0.0:46658", "Listen address")
	eyesPtr := flag.String("eyes", "tcp://0.0.0.0:46659", "MerkleEyes address")
	genPtr := flag.String("genesis", "genesis.json", "Genesis JSON file")
	flag.Parse()

	// Connect to MerkleEyes
	eyesCli, err := eyes.NewMerkleEyesClient(*eyesPtr)
	if err != nil {
		Exit("connect to MerkleEyes: " + err.Error())
	}

	// Create Basecoin app
	app := app.NewBasecoin(eyesCli)

	// Load GenesisState
	jsonBytes, err := ReadFile(*genPtr)
	if err != nil {
		Exit("read genesis: " + err.Error())
	}
	genesisState := types.GenesisState{}
	wire.ReadJSONPtr(&genesisState, jsonBytes, &err)
	if err != nil {
		Exit("parsing genesis JSON: " + err.Error())
	}
	for _, account := range genesisState.Accounts {
		// pubKeyBytes := wire.BinaryBytes(account.PubKey)
		pubKeyString := account.PubKey.KeyString()
		accBytes := wire.BinaryBytes(account.Account)
		err = eyesCli.SetSync([]byte(pubKeyString), accBytes)
		if err != nil {
			Exit("loading genesis accounts: " + err.Error())
		}
	}

	// Start the listener
	_, err = server.StartListener(*addrPtr, app)
	if err != nil {
		Exit("create listener: " + err.Error())
	}

	// Wait forever
	TrapSignal(func() {
		// Cleanup
	})

}
