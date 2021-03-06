package commands

import (
	"encoding/hex"
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	wrsp "github.com/tepleton/wrsp/types"
	wire "github.com/tepleton/go-wire"

	"github.com/tepleton/basecoin/types"

	client "github.com/tepleton/tepleton/rpc/client"
	tmtypes "github.com/tepleton/tepleton/types"
)

//Quickly registering flags can be quickly achieved through using the utility functions
//RegisterFlags, and RegisterPersistentFlags. Ex:
//	flags := []Flag2Register{
//		{&myStringFlag, "mystringflag", "foobar", "description of what this flag does"},
//		{&myBoolFlag, "myboolflag", false, "description of what this flag does"},
//		{&myInt64Flag, "myintflag", 333, "description of what this flag does"},
//	}
//	RegisterFlags(MyCobraCmd, flags)
type Flag2Register struct {
	Pointer interface{}
	Use     string
	Value   interface{}
	Desc    string
}

//register flag utils
func RegisterFlags(c *cobra.Command, flags []Flag2Register) {
	registerFlags(c, flags, false)
}

func RegisterPersistentFlags(c *cobra.Command, flags []Flag2Register) {
	registerFlags(c, flags, true)
}

func registerFlags(c *cobra.Command, flags []Flag2Register, persistent bool) {

	var flagset *pflag.FlagSet
	if persistent {
		flagset = c.PersistentFlags()
	} else {
		flagset = c.Flags()
	}

	for _, f := range flags {

		ok := false

		switch f.Value.(type) {
		case string:
			if _, ok = f.Pointer.(*string); ok {
				flagset.StringVar(f.Pointer.(*string), f.Use, f.Value.(string), f.Desc)
			}
		case int:
			if _, ok = f.Pointer.(*int); ok {
				flagset.IntVar(f.Pointer.(*int), f.Use, f.Value.(int), f.Desc)
			}
		case uint64:
			if _, ok = f.Pointer.(*uint64); ok {
				flagset.Uint64Var(f.Pointer.(*uint64), f.Use, f.Value.(uint64), f.Desc)
			}
		case bool:
			if _, ok = f.Pointer.(*bool); ok {
				flagset.BoolVar(f.Pointer.(*bool), f.Use, f.Value.(bool), f.Desc)
			}
		}

		if !ok {
			panic("improper use of RegisterFlags")
		}
	}
}

// Returns true for non-empty hex-string prefixed with "0x"
func isHex(s string) bool {
	if len(s) > 2 && s[:2] == "0x" {
		_, err := hex.DecodeString(s[2:])
		if err != nil {
			return false
		}
		return true
	}
	return false
}

func StripHex(s string) string {
	if isHex(s) {
		return s[2:]
	}
	return s
}

func Query(tmAddr string, key []byte) (*wrsp.ResultQuery, error) {
	httpClient := client.NewHTTP(tmAddr, "/websocket")
	return queryWithClient(httpClient, key)
}

func queryWithClient(httpClient *client.HTTP, key []byte) (*wrsp.ResultQuery, error) {
	res, err := httpClient.WRSPQuery("/key", key, true)
	if err != nil {
		return nil, errors.Errorf("Error calling /wrsp_query: %v", err)
	}
	if !res.Code.IsOK() {
		return nil, errors.Errorf("Query got non-zero exit code: %v. %s", res.Code, res.Log)
	}
	return res.ResultQuery, nil
}

// fetch the account by querying the app
func getAccWithClient(httpClient *client.HTTP, address []byte) (*types.Account, error) {

	key := types.AccountKey(address)
	response, err := queryWithClient(httpClient, key)
	if err != nil {
		return nil, err
	}

	accountBytes := response.Value

	if len(accountBytes) == 0 {
		return nil, fmt.Errorf("Account bytes are empty for address: %X ", address) //never stack trace
	}

	var acc *types.Account
	err = wire.ReadBinaryBytes(accountBytes, &acc)
	if err != nil {
		return nil, errors.Errorf("Error reading account %X error: %v",
			accountBytes, err.Error())
	}

	return acc, nil
}

func getHeaderAndCommit(tmAddr string, height int) (*tmtypes.Header, *tmtypes.Commit, error) {
	httpClient := client.NewHTTP(tmAddr, "/websocket")
	res, err := httpClient.Commit(height)
	if err != nil {
		return nil, nil, errors.Errorf("Error on commit: %v", err)
	}
	header := res.Header
	commit := res.Commit

	return header, commit, nil
}

func waitForBlock(httpClient *client.HTTP) error {
	res, err := httpClient.Status()
	if err != nil {
		return err
	}

	lastHeight := res.LatestBlockHeight
	for {
		res, err := httpClient.Status()
		if err != nil {
			return err
		}
		if res.LatestBlockHeight > lastHeight {
			break
		}

	}
	return nil
}
