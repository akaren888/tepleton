package slashing

import (
	"encoding/binary"
	"fmt"

	sdk "github.com/tepleton/tepleton-sdk/types"
	wrsp "github.com/tepleton/tepleton/wrsp/types"
	tmtypes "github.com/tepleton/tepleton/types"
)

// slashing begin block functionality
func BeginBlocker(ctx sdk.Context, req wrsp.RequestBeginBlock, sk Keeper) (tags sdk.Tags) {
	// Tag the height
	heightBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(heightBytes, uint64(req.Header.Height))
	tags = sdk.NewTags("height", heightBytes)

	// Deal with any equivocation evidence
	for _, evidence := range req.ByzantineValidators {
		pk, err := tmtypes.PB2TM.PubKey(evidence.Validator.PubKey)
		if err != nil {
			panic(err)
		}
		switch evidence.Type {
		case tmtypes.WRSPEvidenceTypeDuplicateVote:
			sk.handleDoubleSign(ctx, evidence.Height, evidence.Time, pk)
		default:
			ctx.Logger().With("module", "x/slashing").Error(fmt.Sprintf("ignored unknown evidence type: %s", evidence.Type))
		}
	}

	// Iterate over all the validators  which *should* have signed this block
	for _, validator := range req.Validators {
		present := validator.SignedLastBlock
		pubkey, err := tmtypes.PB2TM.PubKey(validator.Validator.PubKey)
		if err != nil {
			panic(err)
		}
		sk.handleValidatorSignature(ctx, pubkey, present)
	}

	return
}
