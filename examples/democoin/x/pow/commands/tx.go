package commands

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/tepleton/tepleton-sdk/client/context"

	"github.com/tepleton/tepleton-sdk/examples/democoin/x/pow"
	"github.com/tepleton/tepleton-sdk/wire"
	authcmd "github.com/tepleton/tepleton-sdk/x/auth/commands"
)

func MineCmd(cdc *wire.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "mine [difficulty] [count] [nonce] [solution]",
		Short: "Mine some coins with proof-of-work!",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 4 {
				return errors.New("You must provide a difficulty, a count, a solution, and a nonce (in that order)")
			}

			// get from address and parse arguments

			ctx := context.NewCoreContextFromViper().WithDecoder(authcmd.GetAccountDecoder(cdc))

			from, err := ctx.GetFromAddress()
			if err != nil {
				return err
			}

			difficulty, err := strconv.ParseUint(args[0], 0, 64)
			if err != nil {
				return err
			}

			count, err := strconv.ParseUint(args[1], 0, 64)
			if err != nil {
				return err
			}

			nonce, err := strconv.ParseUint(args[2], 0, 64)
			if err != nil {
				return err
			}

			solution := []byte(args[3])

			msg := pow.NewMineMsg(from, difficulty, count, nonce, solution)

			// get account name
			name := ctx.FromAddressName

			// default to next sequence number if none provided
			ctx, err = context.EnsureSequence(ctx)
			if err != nil {
				return err
			}

			// build and sign the transaction, then broadcast to Tendermint
			res, err := ctx.SignBuildBroadcast(name, msg, cdc)
			if err != nil {
				return err
			}

			fmt.Printf("Committed at block %d. Hash: %s\n", res.Height, res.Hash.String())
			return nil
		},
	}
}
