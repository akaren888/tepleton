package keeper

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tepleton/tepleton-sdk/x/stake/types"
)

func TestParams(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 0)
	expParams := types.DefaultParams()

	//check that the empty keeper loads the default
	resParams := keeper.GetParams(ctx)
	assert.True(t, expParams.Equal(resParams))

	//modify a params, save, and retrieve
	expParams.MaxValidators = 777
	keeper.SetParams(ctx, expParams)
	resParams = keeper.GetParams(ctx)
	assert.True(t, expParams.Equal(resParams))
}

func TestPool(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 0)
	expPool := types.InitialPool()

	//check that the empty keeper loads the default
	resPool := keeper.GetPool(ctx)
	assert.True(t, expPool.Equal(resPool))

	//modify a params, save, and retrieve
	expPool.BondedTokens = 777
	keeper.SetPool(ctx, expPool)
	resPool = keeper.GetPool(ctx)
	assert.True(t, expPool.Equal(resPool))
}