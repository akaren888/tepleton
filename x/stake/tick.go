package stake

import (
	sdk "github.com/tepleton/tepleton-sdk/types"

	wrsp "github.com/tepleton/wrsp/types"
	"github.com/tepleton/tmlibs/rational"
)

// Tick - called at the end of every block
func Tick(ctx sdk.Context, store types.KVStore) (change []*wrsp.Validator, err error) {

	// retrieve params
	params := loadParams(store)
	gs := loadGlobalState(store)
	height := ctx.BlockHeight()

	// Process Validator Provisions
	// XXX right now just process every 5 blocks, in new SDK make hourly
	if gs.InflationLastTime+5 <= height {
		gs.InflationLastTime = height
		processProvisions(store, gs, params)
	}

	return UpdateValidatorSet(store, gs, params)
}

var hrsPerYr = rational.New(8766) // as defined by a julian year of 365.25 days

// process provisions for an hour period
func processProvisions(store types.KVStore, gs *GlobalState, params Params) {

	gs.Inflation = nextInflation(gs, params).Round(1000000000)

	// Because the validators hold a relative bonded share (`GlobalStakeShare`), when
	// more bonded tokens are added proportionally to all validators the only term
	// which needs to be updated is the `BondedPool`. So for each previsions cycle:

	provisions := gs.Inflation.Mul(rational.New(gs.TotalSupply)).Quo(hrsPerYr).Evaluate()
	gs.BondedPool += provisions
	gs.TotalSupply += provisions

	// XXX XXX XXX XXX XXX XXX XXX XXX XXX
	// XXX Mint them to the hold account
	// XXX XXX XXX XXX XXX XXX XXX XXX XXX

	// save the params
	saveGlobalState(store, gs)
}

// get the next inflation rate for the hour
func nextInflation(gs *GlobalState, params Params) (inflation rational.Rat) {

	// The target annual inflation rate is recalculated for each previsions cycle. The
	// inflation is also subject to a rate change (positive of negative) depending or
	// the distance from the desired ratio (67%). The maximum rate change possible is
	// defined to be 13% per year, however the annual inflation is capped as between
	// 7% and 20%.

	// (1 - bondedRatio/GoalBonded) * InflationRateChange
	inflationRateChangePerYear := rational.One.Sub(gs.bondedRatio().Quo(params.GoalBonded)).Mul(params.InflationRateChange)
	inflationRateChange := inflationRateChangePerYear.Quo(hrsPerYr)

	// increase the new annual inflation for this next cycle
	inflation = gs.Inflation.Add(inflationRateChange)
	if inflation.GT(params.InflationMax) {
		inflation = params.InflationMax
	}
	if inflation.LT(params.InflationMin) {
		inflation = params.InflationMin
	}

	return
}