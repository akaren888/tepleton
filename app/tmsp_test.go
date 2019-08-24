package app

import (
	"testing"

	cmn "github.com/tepleton/basecoin/common"
	"github.com/tepleton/basecoin/types"
	. "github.com/tepleton/go-common"
	"github.com/tepleton/go-wire"
	eyescli "github.com/tepleton/merkleeyes/client"
)

func TestSendTx(t *testing.T) {
	eyesCli := eyescli.NewLocalClient()
	chainID := "test_chain_id"
	bcApp := NewBasecoin(eyesCli)
	bcApp.SetOption("base/chainID", chainID)
	t.Log(bcApp.Info())

	test1PrivAcc := cmn.PrivAccountFromSecret("test1")
	test2PrivAcc := cmn.PrivAccountFromSecret("test2")

	// Seed Basecoin with account
	test1Acc := test1PrivAcc.Account
	test1Acc.Balance = types.Coins{{"", 1000}}
	t.Log(bcApp.SetOption("base/account", string(wire.JSONBytes(test1Acc))))

	res := bcApp.Commit()
	if res.IsErr() {
		Exit(Fmt("Failed Commit: %v", res.Error()))
	}

	// Construct a SendTx signature
	tx := &types.SendTx{
		Fee: 0,
		Gas: 0,
		Inputs: []types.TxInput{
			cmn.MakeInput(test1PrivAcc.Account.PubKey, types.Coins{{"", 1}}, 1),
		},
		Outputs: []types.TxOutput{
			types.TxOutput{
				Address: test2PrivAcc.Account.PubKey.Address(),
				Coins:   types.Coins{{"", 1}},
			},
		},
	}

	// Sign request
	signBytes := tx.SignBytes(chainID)
	t.Log("Sign bytes: %X\n", signBytes)
	sig := test1PrivAcc.PrivKey.Sign(signBytes)
	tx.Inputs[0].Signature = sig
	t.Log("Signed TX bytes: %X\n", wire.BinaryBytes(struct{ types.Tx }{tx}))

	// Write request
	txBytes := wire.BinaryBytes(struct{ types.Tx }{tx})
	res = bcApp.DeliverTx(txBytes)
	t.Log(res)
	if res.IsErr() {
		t.Errorf(Fmt("Failed: %v", res.Error()))
	}
}

func TestSequence(t *testing.T) {
	eyesCli := eyescli.NewLocalClient()
	chainID := "test_chain_id"
	bcApp := NewBasecoin(eyesCli)
	bcApp.SetOption("base/chainID", chainID)
	t.Log(bcApp.Info())

	// Get the test account
	test1PrivAcc := cmn.PrivAccountFromSecret("test1")
	test1Acc := test1PrivAcc.Account
	test1Acc.Balance = types.Coins{{"", 1 << 53}}
	t.Log(bcApp.SetOption("base/account", string(wire.JSONBytes(test1Acc))))

	res := bcApp.Commit()
	if res.IsErr() {
		t.Errorf(Fmt("Failed Commit: %v", res.Error()))
	}

	sequence := int(1)
	// Make a bunch of PrivAccounts
	privAccounts := cmn.RandAccounts(1000, 1000000, 0)
	privAccountSequences := make(map[string]int)
	// Send coins to each account

	for i := 0; i < len(privAccounts); i++ {
		privAccount := privAccounts[i]

		tx := &types.SendTx{
			Fee: 2,
			Gas: 2,
			Inputs: []types.TxInput{
				cmn.MakeInput(test1Acc.PubKey, types.Coins{{"", 1000002}}, sequence),
			},
			Outputs: []types.TxOutput{
				types.TxOutput{
					Address: privAccount.Account.PubKey.Address(),
					Coins:   types.Coins{{"", 1000000}},
				},
			},
		}
		sequence += 1

		// Sign request
		signBytes := tx.SignBytes(chainID)
		sig := test1PrivAcc.PrivKey.Sign(signBytes)
		tx.Inputs[0].Signature = sig
		// t.Log("ADDR: %X -> %X\n", tx.Inputs[0].Address, tx.Outputs[0].Address)

		// Write request
		txBytes := wire.BinaryBytes(struct{ types.Tx }{tx})
		res := bcApp.DeliverTx(txBytes)
		if res.IsErr() {
			t.Errorf("DeliverTx error: " + res.Error())
		}

	}

	t.Log("-------------------- RANDOM SENDS --------------------")

	res = bcApp.Commit()
	if res.IsErr() {
		t.Errorf(Fmt("Failed Commit: %v", res.Error()))
	}

	// Now send coins between these accounts
	for i := 0; i < 10000; i++ {
		randA := RandInt() % len(privAccounts)
		randB := RandInt() % len(privAccounts)
		if randA == randB {
			continue
		}

		privAccountA := privAccounts[randA]
		privAccountASequence := privAccountSequences[privAccountA.Account.PubKey.KeyString()]
		privAccountSequences[privAccountA.Account.PubKey.KeyString()] = privAccountASequence + 1
		privAccountB := privAccounts[randB]

		tx := &types.SendTx{
			Fee: 2,
			Gas: 2,
			Inputs: []types.TxInput{
				cmn.MakeInput(privAccountA.Account.PubKey, types.Coins{{"", 3}}, privAccountASequence+1),
			},
			Outputs: []types.TxOutput{
				types.TxOutput{
					Address: privAccountB.Account.PubKey.Address(),
					Coins:   types.Coins{{"", 1}},
				},
			},
		}

		// Sign request
		signBytes := tx.SignBytes(chainID)
		sig := privAccountA.PrivKey.Sign(signBytes)
		tx.Inputs[0].Signature = sig
		// t.Log("ADDR: %X -> %X\n", tx.Inputs[0].Address, tx.Outputs[0].Address)

		// Write request
		txBytes := wire.BinaryBytes(struct{ types.Tx }{tx})
		res := bcApp.DeliverTx(txBytes)
		if res.IsErr() {
			t.Errorf("DeliverTx error: " + res.Error())
		}
	}
}
