package commands

import (
	"encoding/hex"
	"fmt"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	crypto "github.com/tepleton/go-crypto"

	"github.com/tepleton/tepleton-sdk/client/context"
	sdk "github.com/tepleton/tepleton-sdk/types"
	"github.com/tepleton/tepleton-sdk/wire" // XXX fix
	"github.com/tepleton/tepleton-sdk/x/stake"
)

//nolint
var (
	fsValAddr         = flag.NewFlagSet("", flag.ContinueOnError)
	fsDelAddr         = flag.NewFlagSet("", flag.ContinueOnError)
	FlagValidatorAddr = "address"
	FlagDelegatorAddr = "delegator-address"
)

func init() {
	//Add Flags
	fsValAddr.String(FlagValidatorAddr, "", "Address of the validator/candidate")
	fsDelAddr.String(FlagDelegatorAddr, "", "Delegator hex address")

}

// create command to query for all candidates
func GetCmdQueryCandidates(storeName string, cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "candidates",
		Short: "Query for the set of validator-candidates pubkeys",
		RunE: func(cmd *cobra.Command, args []string) error {

			key := stake.CandidatesKey

			ctx := context.NewCoreContextFromViper()
			res, err := ctx.Query(key, storeName)
			if err != nil {
				return err
			}

			// parse out the candidates
			candidates := new(stake.Candidates)
			err = cdc.UnmarshalBinary(res, candidates)
			if err != nil {
				return err
			}
			output, err := wire.MarshalJSONIndent(cdc, candidates)
			if err != nil {
				return err
			}
			fmt.Println(string(output))
			return nil

			// TODO output with proofs / machine parseable etc.
		},
	}

	cmd.Flags().AddFlagSet(fsDelAddr)
	return cmd
}

// get the command to query a candidate
func GetCmdQueryCandidate(storeName string, cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "candidate",
		Short: "Query a validator-candidate account",
		RunE: func(cmd *cobra.Command, args []string) error {

			addr, err := sdk.GetAddress(viper.GetString(FlagValidatorAddr))
			if err != nil {
				return err
			}

			key := stake.GetCandidateKey(addr)

			ctx := context.NewCoreContextFromViper()

			res, err := ctx.Query(key, storeName)
			if err != nil {
				return err
			}

			// parse out the candidate
			candidate := new(stake.Candidate)
			err = cdc.UnmarshalBinary(res, candidate)
			if err != nil {
				return err
			}
			output, err := wire.MarshalJSONIndent(cdc, candidate)
			if err != nil {
				return err
			}
			fmt.Println(string(output))
			return nil

			// TODO output with proofs / machine parseable etc.
		},
	}

	cmd.Flags().AddFlagSet(fsValAddr)
	return cmd
}

// get the command to query a single delegator bond
func GetCmdQueryDelegatorBond(storeName string, cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delegator-bond",
		Short: "Query a delegators bond based on address and candidate pubkey",
		RunE: func(cmd *cobra.Command, args []string) error {

			addr, err := sdk.GetAddress(viper.GetString(FlagValidatorAddr))
			if err != nil {
				return err
			}

			bz, err := hex.DecodeString(viper.GetString(FlagDelegatorAddr))
			if err != nil {
				return err
			}
			delegator := crypto.Address(bz)

			key := stake.GetDelegatorBondKey(delegator, addr, cdc)

			ctx := context.NewCoreContextFromViper()

			res, err := ctx.Query(key, storeName)
			if err != nil {
				return err
			}

			// parse out the bond
			var bond stake.DelegatorBond
			err = cdc.UnmarshalBinary(res, bond)
			if err != nil {
				return err
			}
			output, err := wire.MarshalJSONIndent(cdc, bond)
			if err != nil {
				return err
			}
			fmt.Println(string(output))
			return nil

			// TODO output with proofs / machine parseable etc.
		},
	}

	cmd.Flags().AddFlagSet(fsValAddr)
	cmd.Flags().AddFlagSet(fsDelAddr)
	return cmd
}

// get the command to query all the candidates bonded to a delegator
func GetCmdQueryDelegatorBonds(storeName string, cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delegator-candidates",
		Short: "Query all delegators candidates' pubkeys based on address",
		RunE: func(cmd *cobra.Command, args []string) error {

			bz, err := hex.DecodeString(viper.GetString(FlagDelegatorAddr))
			if err != nil {
				return err
			}
			delegator := crypto.Address(bz)

			key := stake.GetDelegatorBondsKey(delegator, cdc)

			ctx := context.NewCoreContextFromViper()

			res, err := ctx.Query(key, storeName)
			if err != nil {
				return err
			}

			// parse out the candidates list
			var candidates []crypto.PubKey
			err = cdc.UnmarshalBinary(res, candidates)
			if err != nil {
				return err
			}
			output, err := wire.MarshalJSONIndent(cdc, candidates)
			if err != nil {
				return err
			}
			fmt.Println(string(output))
			return nil

			// TODO output with proofs / machine parseable etc.
		},
	}
	cmd.Flags().AddFlagSet(fsDelAddr)
	return cmd
}
