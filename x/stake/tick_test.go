package stake

import (
	"testing"

	sdk "github.com/tepleton/tepleton-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetInflation(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 0)
	pool := keeper.GetPool(ctx)
	params := keeper.GetParams(ctx)
	hrsPerYrRat := sdk.NewRat(hrsPerYr)

	// Governing Mechanism:
	//    bondedRatio = BondedTokens / TotalSupply
	//    inflationRateChangePerYear = (1- bondedRatio/ GoalBonded) * MaxInflationRateChange

	tests := []struct {
		name                          string
		setBondedPool, setTotalSupply int64
		setInflation, expectedChange  sdk.Rat
	}{
		// with 0% bonded atom supply the inflation should increase by InflationRateChange
		{"test 1", 0, 0, sdk.NewRat(7, 100), params.InflationRateChange.Quo(hrsPerYrRat).Round(precision)},

		// 100% bonded, starting at 20% inflation and being reduced
		// (1 - (1/0.67))*(0.13/8667)
		{"test 2", 1, 1, sdk.NewRat(20, 100),
			sdk.OneRat().Sub(sdk.OneRat().Quo(params.GoalBonded)).Mul(params.InflationRateChange).Quo(hrsPerYrRat).Round(precision)},

		// 50% bonded, starting at 10% inflation and being increased
		{"test 3", 1, 2, sdk.NewRat(10, 100),
			sdk.OneRat().Sub(sdk.NewRat(1, 2).Quo(params.GoalBonded)).Mul(params.InflationRateChange).Quo(hrsPerYrRat).Round(precision)},

		// test 7% minimum stop (testing with 100% bonded)
		{"test 4", 1, 1, sdk.NewRat(7, 100), sdk.ZeroRat()},
		{"test 5", 1, 1, sdk.NewRat(70001, 1000000), sdk.NewRat(-1, 1000000).Round(precision)},

		// test 20% maximum stop (testing with 0% bonded)
		{"test 6", 0, 0, sdk.NewRat(20, 100), sdk.ZeroRat()},
		{"test 7", 0, 0, sdk.NewRat(199999, 1000000), sdk.NewRat(1, 1000000).Round(precision)},

		// perfect balance shouldn't change inflation
		{"test 8", 67, 100, sdk.NewRat(15, 100), sdk.ZeroRat()},
	}
	for _, tc := range tests {
		pool.BondedTokens, pool.TotalSupply = tc.setBondedPool, tc.setTotalSupply
		pool.Inflation = tc.setInflation
		keeper.setPool(ctx, pool)

		inflation := keeper.nextInflation(ctx)
		diffInflation := inflation.Sub(tc.setInflation)

		assert.True(t, diffInflation.Equal(tc.expectedChange),
			"Name: %v\nDiff:  %v\nExpected: %v\n", tc.name, diffInflation, tc.expectedChange)
	}
}

func TestProcessProvisions(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 0)
	params := defaultParams()
	keeper.setParams(ctx, params)
	pool := keeper.GetPool(ctx)

	// create some validators some bonded, some unbonded
	validators := make([]Validator, 10)
	for i := 0; i < 10; i++ {
		v := Validator{
			Status:          sdk.Unbonded,
			PubKey:          pks[i],
			Address:         addrs[i],
			PShares:         NewUnbondedShares(sdk.NewRat(0)),
			DelegatorShares: sdk.NewRat(0),
		}
		if i < 5 {
			v.Status = sdk.Bonded
			v.PShares.Kind = ShareBonded
		}
		mintedTokens := int64((i + 1) * 10000000)
		pool.TotalSupply += mintedTokens
		v, pool, _ = v.addTokensFromDel(pool, mintedTokens)

		keeper.setValidator(ctx, v)
		validators[i] = v
	}
	keeper.setPool(ctx, pool)
	var totalSupply int64 = 550000000
	var bondedShares int64 = 150000000
	var unbondedShares int64 = 400000000
	assert.Equal(t, totalSupply, pool.TotalSupply)
	assert.Equal(t, bondedShares, pool.BondedTokens)
	assert.Equal(t, unbondedShares, pool.UnbondedTokens)

	// initial bonded ratio ~ 27%
	assert.True(t, pool.bondedRatio().Equal(sdk.NewRat(bondedShares, totalSupply)), "%v", pool.bondedRatio())

	// test the value of validator shares
	assert.True(t, pool.bondedShareExRate().Equal(sdk.OneRat()), "%v", pool.bondedShareExRate())

	initialSupply := pool.TotalSupply
	initialUnbonded := pool.TotalSupply - pool.BondedTokens

	// process the provisions a year
	for hr := 0; hr < 8766; hr++ {
		pool := keeper.GetPool(ctx)
		expInflation := keeper.nextInflation(ctx).Round(1000000000)
		expProvisions := (expInflation.Mul(sdk.NewRat(pool.TotalSupply)).Quo(hrsPerYrRat)).Evaluate()
		startBondedPool := pool.BondedTokens
		startTotalSupply := pool.TotalSupply
		pool = keeper.processProvisions(ctx)
		keeper.setPool(ctx, pool)
		//fmt.Printf("hr %v, startBondedPool %v, expProvisions %v, pool.BondedTokens %v\n", hr, startBondedPool, expProvisions, pool.BondedTokens)
		require.Equal(t, startBondedPool+expProvisions, pool.BondedTokens, "hr %v", hr)
		require.Equal(t, startTotalSupply+expProvisions, pool.TotalSupply)
	}
	pool = keeper.GetPool(ctx)
	assert.NotEqual(t, initialSupply, pool.TotalSupply)
	assert.Equal(t, initialUnbonded, pool.UnbondedTokens)
	//panic(fmt.Sprintf("debug total %v, bonded  %v, diff %v\n", p.TotalSupply, p.BondedTokens, pool.TotalSupply-pool.BondedTokens))

	// initial bonded ratio ~ from 27% to 40% increase for bonded holders ownership of total supply
	assert.True(t, pool.bondedRatio().Equal(sdk.NewRat(211813022, 611813022)), "%v", pool.bondedRatio())

	// global supply
	assert.Equal(t, int64(611813022), pool.TotalSupply)
	assert.Equal(t, int64(211813022), pool.BondedTokens)
	assert.Equal(t, unbondedShares, pool.UnbondedTokens)

	// test the value of validator shares
	assert.True(t, pool.bondedShareExRate().Mul(sdk.NewRat(bondedShares)).Equal(sdk.NewRat(211813022)), "%v", pool.bondedShareExRate())
}
