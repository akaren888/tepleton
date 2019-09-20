package commands

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	lc "github.com/tepleton/light-client"

	sdk "github.com/tepleton/tepleton-sdk"
	"github.com/tepleton/tepleton-sdk/client/commands"
	"github.com/tepleton/tepleton-sdk/client/commands/query"
	"github.com/tepleton/tepleton-sdk/modules/nonce"
	"github.com/tepleton/tepleton-sdk/stack"
)

// NonceQueryCmd - command to query an nonce account
var NonceQueryCmd = &cobra.Command{
	Use:   "nonce [address]",
	Short: "Get details of a nonce sequence number, with proof",
	RunE:  commands.RequireInit(nonceQueryCmd),
}

func nonceQueryCmd(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return errors.New("Missing required argument [address]")
	}
	addr := strings.Join(args, ",")

	signers, err := commands.ParseActors(addr)
	if err != nil {
		return err
	}

	seq, height, err := doNonceQuery(signers)
	if err != nil {
		return err
	}

	return query.OutputProof(seq, height)
}

func doNonceQuery(signers []sdk.Actor) (sequence uint32, height uint64, err error) {
	key := stack.PrefixedKey(nonce.NameNonce, nonce.GetSeqKey(signers))
	prove := !viper.GetBool(commands.FlagTrustNode)
	height, err = query.GetParsed(key, &sequence, prove)
	if lc.IsNoDataErr(err) {
		// no data, return sequence 0
		return 0, 0, nil
	}
	return
}
