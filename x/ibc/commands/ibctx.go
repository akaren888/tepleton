package commands

import (
	"encoding/hex"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tepleton/tepleton-sdk/client"
	"github.com/tepleton/tepleton-sdk/client/builder"

	sdk "github.com/tepleton/tepleton-sdk/types"
	wire "github.com/tepleton/tepleton-sdk/wire"

	"github.com/tepleton/tepleton-sdk/x/ibc"
)

const (
	flagTo     = "to"
	flagAmount = "amount"
	flagChain  = "chain"
)

func IBCTransferCmd(cdc *wire.Codec) *cobra.Command {
	cmdr := sendCommander{cdc}
	cmd := &cobra.Command{
		Use:  "send",
		RunE: cmdr.runIBCTransfer,
	}
	cmd.Flags().String(flagTo, "", "Address to send coins")
	cmd.Flags().String(flagAmount, "", "Amount of coins to send")
	cmd.Flags().String(flagChain, "", "Destination chain to send coins")
	return cmd
}

type sendCommander struct {
	cdc *wire.Codec
}

func (c commander) sendIBCTransfer(cmd *cobra.Command, args []string) error {
	from, err := builder.GetFromAddress()
	if err != nil {
		return err
	}

	msg, err := buildMsg(from)
	if err != nil {
		return err
	}

	res, err := builder.SignBuildBroadcast(msg, c.cdc)
	if err != nil {
		return err
	}

	fmt.Printf("Committed at block %d. Hash: %s\n", res.Height, res.Hash.String())
	return nil
}

func buildMsg(from sdk.Address) (sdk.Msg, error) {
	amount := viper.GetString(flagAmount)
	coins, err := sdk.ParseCoins(amount)
	if err != nil {
		return nil, err
	}

	dest := viper.GetString(flagTo)
	bz, err := hex.DecodeString(dest)
	if err != nil {
		return nil, err
	}

	to := sdk.Address(bz)

	msg := ibc.NewIBCPacket(from, to, coins, viper.GetString(client.FlagNode),
		viper.GetString(flagChain))
	return msg, nil
}