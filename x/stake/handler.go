package stake

import (
	"bytes"
	"fmt"

	sdk "github.com/tepleton/tepleton-sdk/types"
	wrsp "github.com/tepleton/wrsp/types"
)

//nolint
const (
	GasDeclareCandidacy int64 = 20
	GasEditCandidacy    int64 = 20
	GasDelegate         int64 = 20
	GasUnbond           int64 = 20
)

func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		// NOTE msg already has validate basic run
		switch msg := msg.(type) {
		case MsgDeclareCandidacy:
			return handleMsgDeclareCandidacy(ctx, msg, k)
		case MsgEditCandidacy:
			return handleMsgEditCandidacy(ctx, msg, k)
		case MsgDelegate:
			return handleMsgDelegate(ctx, msg, k)
		case MsgUnbond:
			return handleMsgUnbond(ctx, msg, k)
		default:
			return sdk.ErrTxDecode("invalid message parse in staking module").Result()
		}
	}
}

// NewEndBlocker generates sdk.EndBlocker
// Performs tick functionality
func NewEndBlocker(k Keeper) sdk.EndBlocker {
	return func(ctx sdk.Context, req wrsp.RequestEndBlock) (res wrsp.ResponseEndBlock) {
		res.ValidatorUpdates = k.Tick(ctx)
		return
	}
}

//_____________________________________________________________________

// These functions assume everything has been authenticated,
// now we just perform action and save

func handleMsgDeclareCandidacy(ctx sdk.Context, msg MsgDeclareCandidacy, k Keeper) sdk.Result {

	// check to see if the pubkey or sender has been registered before
	_, found := k.GetValidator(ctx, msg.ValidatorAddr)
	if found {
		return ErrValidatorExistsAddr(k.codespace).Result()
	}
	if msg.Bond.Denom != k.GetParams(ctx).BondDenom {
		return ErrBadBondingDenom(k.codespace).Result()
	}
	if ctx.IsCheckTx() {
		return sdk.Result{
			GasUsed: GasDeclareCandidacy,
		}
	}

	validator := NewValidator(msg.ValidatorAddr, msg.PubKey, msg.Description)
	validator = k.setValidator(ctx, validator)
	tags := sdk.NewTags(
		"action", []byte("declareCandidacy"),
		"validator", msg.ValidatorAddr.Bytes(),
		"moniker", []byte(msg.Description.Moniker),
		"identity", []byte(msg.Description.Identity),
	)

	// move coins from the msg.Address account to a (self-bond) delegator account
	// the validator account and global shares are updated within here
	delegateTags, err := delegate(ctx, k, msg.ValidatorAddr, msg.Bond, validator)
	if err != nil {
		return err.Result()
	}
	tags = tags.AppendTags(delegateTags)
	return sdk.Result{
		Tags: tags,
	}
}

func handleMsgEditCandidacy(ctx sdk.Context, msg MsgEditCandidacy, k Keeper) sdk.Result {

	// validator must already be registered
	validator, found := k.GetValidator(ctx, msg.ValidatorAddr)
	if !found {
		return ErrBadValidatorAddr(k.codespace).Result()
	}
	if ctx.IsCheckTx() {
		return sdk.Result{
			GasUsed: GasEditCandidacy,
		}
	}

	// XXX move to types
	// replace all editable fields (clients should autofill existing values)
	validator.Description.Moniker = msg.Description.Moniker
	validator.Description.Identity = msg.Description.Identity
	validator.Description.Website = msg.Description.Website
	validator.Description.Details = msg.Description.Details

	k.setValidator(ctx, validator)
	tags := sdk.NewTags(
		"action", []byte("editCandidacy"),
		"validator", msg.ValidatorAddr.Bytes(),
		"moniker", []byte(msg.Description.Moniker),
		"identity", []byte(msg.Description.Identity),
	)
	return sdk.Result{
		Tags: tags,
	}
}

func handleMsgDelegate(ctx sdk.Context, msg MsgDelegate, k Keeper) sdk.Result {
	fmt.Println("wackydebugoutput handleMsgDelegate 0")

	validator, found := k.GetValidator(ctx, msg.ValidatorAddr)
	if !found {
		fmt.Println("wackydebugoutput handleMsgDelegate 1")
		return ErrBadValidatorAddr(k.codespace).Result()
	}
	fmt.Println("wackydebugoutput handleMsgDelegate 2")
	if msg.Bond.Denom != k.GetParams(ctx).BondDenom {
		fmt.Println("wackydebugoutput handleMsgDelegate 3")
		return ErrBadBondingDenom(k.codespace).Result()
	}
	fmt.Println("wackydebugoutput handleMsgDelegate 4")
	if validator.Status == sdk.Revoked {
		fmt.Println("wackydebugoutput handleMsgDelegate 5")
		return ErrValidatorRevoked(k.codespace).Result()
	}
	fmt.Println("wackydebugoutput handleMsgDelegate 6")
	if ctx.IsCheckTx() {
		fmt.Println("wackydebugoutput handleMsgDelegate 7")
		return sdk.Result{
			GasUsed: GasDelegate,
		}
		fmt.Println("wackydebugoutput handleMsgDelegate 9")
	}
	fmt.Println("wackydebugoutput handleMsgDelegate 10")
	tags, err := delegate(ctx, k, msg.DelegatorAddr, msg.Bond, validator)
	if err != nil {
		fmt.Println("wackydebugoutput handleMsgDelegate 11")
		return err.Result()
	}
	fmt.Println("wackydebugoutput handleMsgDelegate 12")
	return sdk.Result{
		Tags: tags,
	}
}

// common functionality between handlers
func delegate(ctx sdk.Context, k Keeper, delegatorAddr sdk.Address,
	bondAmt sdk.Coin, validator Validator) (sdk.Tags, sdk.Error) {
	fmt.Println("wackydebugoutput delegate 0")

	// Get or create the delegator bond
	bond, found := k.GetDelegation(ctx, delegatorAddr, validator.Address)
	if !found {
		fmt.Println("wackydebugoutput delegate 1")
		bond = Delegation{
			DelegatorAddr: delegatorAddr,
			ValidatorAddr: validator.Address,
			Shares:        sdk.ZeroRat(),
		}
		fmt.Println("wackydebugoutput delegate 3")
	}
	fmt.Println("wackydebugoutput delegate 4")

	// Account new shares, save
	pool := k.GetPool(ctx)
	_, _, err := k.coinKeeper.SubtractCoins(ctx, bond.DelegatorAddr, sdk.Coins{bondAmt})
	fmt.Println("wackydebugoutput delegate 5")
	if err != nil {
		fmt.Println("wackydebugoutput delegate 6")
		return nil, err
	}
	fmt.Println("wackydebugoutput delegate 7")
	validator, pool, newShares := validator.addTokensFromDel(pool, bondAmt.Amount)
	fmt.Printf("debug newShares: %v\n", newShares)
	bond.Shares = bond.Shares.Add(newShares)

	// Update bond height
	bond.Height = ctx.BlockHeight()

	k.setDelegation(ctx, bond)
	k.setValidator(ctx, validator)
	k.setPool(ctx, pool)
	tags := sdk.NewTags("action", []byte("delegate"), "delegator", delegatorAddr.Bytes(), "validator", validator.Address.Bytes())
	return tags, nil
}

func handleMsgUnbond(ctx sdk.Context, msg MsgUnbond, k Keeper) sdk.Result {

	// check if bond has any shares in it unbond
	bond, found := k.GetDelegation(ctx, msg.DelegatorAddr, msg.ValidatorAddr)
	if !found {
		return ErrNoDelegatorForAddress(k.codespace).Result()
	}
	if !bond.Shares.GT(sdk.ZeroRat()) { // bond shares < msg shares
		return ErrInsufficientFunds(k.codespace).Result()
	}

	var delShares sdk.Rat

	// test that there are enough shares to unbond
	if msg.Shares == "MAX" {
		if !bond.Shares.GT(sdk.ZeroRat()) {
			return ErrNotEnoughBondShares(k.codespace, msg.Shares).Result()
		}
	} else {
		var err sdk.Error
		delShares, err = sdk.NewRatFromDecimal(msg.Shares)
		if err != nil {
			return err.Result()
		}
		if bond.Shares.LT(delShares) {
			return ErrNotEnoughBondShares(k.codespace, msg.Shares).Result()
		}
	}

	// get validator
	validator, found := k.GetValidator(ctx, msg.ValidatorAddr)
	if !found {
		return ErrNoValidatorForAddress(k.codespace).Result()
	}

	if ctx.IsCheckTx() {
		return sdk.Result{
			GasUsed: GasUnbond,
		}
	}

	// retrieve the amount of bonds to remove (TODO remove redundancy already serialized)
	if msg.Shares == "MAX" {
		delShares = bond.Shares
	}

	// subtract bond tokens from delegator bond
	bond.Shares = bond.Shares.Sub(delShares)

	// remove the bond
	revokeCandidacy := false
	if bond.Shares.IsZero() {

		// if the bond is the owner of the validator then
		// trigger a revoke candidacy
		if bytes.Equal(bond.DelegatorAddr, validator.Address) &&
			validator.Status != sdk.Revoked {
			revokeCandidacy = true
		}

		k.removeDelegation(ctx, bond)
	} else {
		// Update bond height
		bond.Height = ctx.BlockHeight()
		k.setDelegation(ctx, bond)
	}

	// Add the coins
	p := k.GetPool(ctx)
	validator, p, returnAmount := validator.removeDelShares(p, delShares)
	returnCoins := sdk.Coins{{k.GetParams(ctx).BondDenom, returnAmount}}
	k.coinKeeper.AddCoins(ctx, bond.DelegatorAddr, returnCoins)

	/////////////////////////////////////

	// revoke validator if necessary
	if revokeCandidacy {

		// change the share types to unbonded if they were not already
		if validator.Status == sdk.Bonded {
			validator.Status = sdk.Unbonded
			validator, p = validator.UpdateSharesLocation(p)
		}

		// lastly update the status
		validator.Status = sdk.Revoked
	}

	// deduct shares from the validator
	if validator.DelegatorShares.IsZero() {
		k.removeValidator(ctx, validator.Address)
	} else {
		k.setValidator(ctx, validator)
	}
	k.setPool(ctx, p)
	tags := sdk.NewTags("action", []byte("unbond"), "delegator", msg.DelegatorAddr.Bytes(), "validator", msg.ValidatorAddr.Bytes())
	return sdk.Result{
		Tags: tags,
	}
}
