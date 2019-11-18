package slashing

import (
	"bytes"
	"encoding/binary"
	"fmt"

	sdk "github.com/tepleton/tepleton-sdk/types"
	wrsp "github.com/tepleton/wrsp/types"
	crypto "github.com/tepleton/go-crypto"
)

func NewBeginBlocker(sk Keeper) sdk.BeginBlocker {
	return func(ctx sdk.Context, req wrsp.RequestBeginBlock) wrsp.ResponseBeginBlock {
		// Tag the height
		heightBytes := make([]byte, 8)
		binary.LittleEndian.PutUint64(heightBytes, uint64(req.Header.Height))
		tags := sdk.NewTags("height", heightBytes)

		// Deal with any equivocation evidence
		for _, evidence := range req.ByzantineValidators {
			var pk crypto.PubKey
			sk.cdc.MustUnmarshalBinary(evidence.PubKey, &pk)
			switch {
			case bytes.Compare(evidence.Type, []byte("doubleSign")) == 0:
				sk.handleDoubleSign(ctx, evidence.Height, evidence.Time, pk)
			default:
				ctx.Logger().With("module", "x/slashing").Error(fmt.Sprintf("Ignored unknown evidence type: %s", string(evidence.Type)))
			}
		}

		// Figure out which validators were absent
		absent := make(map[string]bool)
		for _, pubkey := range req.AbsentValidators {
			var pk crypto.PubKey
			sk.cdc.MustUnmarshalBinary(pubkey, &pk)
			absent[string(pk.Bytes())] = true
		}

		// Iterate over all the validators which *should* have signed this block
		sk.stakeKeeper.IterateValidatorsBonded(ctx, func(_ int64, validator sdk.Validator) (stop bool) {
			pubkey := validator.GetPubKey()
			sk.handleValidatorSignature(ctx, pubkey, !absent[string(pubkey.Bytes())])
			return false
		})

		// Return the begin block response
		// TODO Return something composable, so other modules can also have BeginBlockers
		// TODO Add some more tags so clients can track slashing events
		return wrsp.ResponseBeginBlock{
			Tags: tags.ToKVPairs(),
		}
	}
}