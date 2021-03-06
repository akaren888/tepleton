package vote

import (
	"fmt"
	"testing"

	"github.com/tepleton/basecoin/app"
	cmn "github.com/tepleton/basecoin/common"
	"github.com/tepleton/basecoin/types"
	. "github.com/tepleton/go-common"
	"github.com/tepleton/go-wire"
	eyescli "github.com/tepleton/merkleeyes/client"
)

const PluginNameVote = "vote"

func TestVote(t *testing.T) {
	//base initialization
	eyesCli := eyescli.NewLocalClient()
	chainID := "test_chain_id"
	bcApp := app.NewBasecoin(eyesCli)
	bcApp.SetOption("base/chainID", chainID)
	fmt.Println(bcApp.Info())

	//account initialization
	test1PrivAcc := cmn.PrivAccountFromSecret("test1")

	// Seed Basecoin with account
	test1Acc := test1PrivAcc.Account
	test1Acc.Balance = types.Coins{{"", 1000}}
	fmt.Println(bcApp.SetOption("base/account", string(wire.JSONBytes(test1Acc))))

	//vote initialization
	votePlugin := NewVoteInstance("humanRights")
	bcApp.RegisterPlugin(
		PluginNameVote,
		votePlugin,
	)

	//commit
	res := bcApp.Commit()
	if res.IsErr() {
		Exit(Fmt("Failed Commit: %v", res.Error()))
	}

	//transaction sequence number
	seqNum := 1

	//Construct, Sign, Write function variable
	CSW := func(fees, sendCoins int64) {
		// Construct an AppTx signature
		tx := &types.AppTx{
			Fee:   fees,
			Gas:   0,
			Name:  PluginNameVote,
			Input: cmn.MakeInput(test1Acc.PubKey, types.Coins{{"", sendCoins}}, seqNum),
			Data:  wire.BinaryBytes(struct{ Tx }{Tx{voteYes: true}}), //a vote for human rights
		}

		// Sign request
		signBytes := tx.SignBytes(chainID)
		fmt.Printf("Sign bytes: %X\n", signBytes)
		sig := test1PrivAcc.PrivKey.Sign(signBytes)
		tx.Input.Signature = sig
		fmt.Printf("Signed TX bytes: %X\n", wire.BinaryBytes(struct{ types.Tx }{tx}))

		// Write request
		txBytes := wire.BinaryBytes(struct{ types.Tx }{tx})
		res = bcApp.DeliverTx(txBytes)
		fmt.Println(res)

		if res.IsOK() {
			seqNum += 1
		}
	}

	//Test a basic send, no fees
	CSW(0, 1)
	if res.IsErr() {
		Exit(Fmt("Failed: %v", res.Error()))
	}

	//Test fee prevented transaction
	CSW(2, 1)
	if res.IsOK() {
		Exit(Fmt("expected bad transaction"))
	}

	//Test equal fees
	CSW(2, 2)
	if res.IsErr() {
		Exit(Fmt("Failed: %v", res.Error()))
	}

	//Test more send coins than fees
	CSW(2, 3)
	if res.IsErr() {
		Exit(Fmt("Failed: %v", res.Error()))
	}
}
