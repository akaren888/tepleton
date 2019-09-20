package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	wrsp "github.com/tepleton/wrsp/types"
	wire "github.com/tepleton/go-wire"
	cmn "github.com/tepleton/tmlibs/common"
	"github.com/tepleton/tmlibs/log"

	"github.com/tepleton/basecoin"
	"github.com/tepleton/basecoin/errors"
	"github.com/tepleton/basecoin/state"
)

//--------------------------------
// Setup tx and handler for validation test cases

const (
	ValName       = "val"
	TypeValChange = ValName + "/change"
	ByteValChange = 0xfe
)

func init() {
	basecoin.TxMapper.RegisterImplementation(ValChangeTx{}, TypeValChange, ByteValChange)
}

type ValSetHandler struct {
	basecoin.NopCheck
	basecoin.NopInitState
	basecoin.NopInitValidate
}

var _ basecoin.Handler = ValSetHandler{}

func (ValSetHandler) Name() string {
	return ValName
}

func (ValSetHandler) DeliverTx(ctx basecoin.Context, store state.SimpleDB,
	tx basecoin.Tx) (res basecoin.DeliverResult, err error) {
	change, ok := tx.Unwrap().(ValChangeTx)
	if !ok {
		return res, errors.ErrUnknownTxType(tx)
	}
	res.Diff = change.Diff
	return
}

type ValChangeTx struct {
	Diff []*wrsp.Validator
}

func (v ValChangeTx) Wrap() basecoin.Tx {
	return basecoin.Tx{v}
}

func (v ValChangeTx) ValidateBasic() error { return nil }

//-----------------------------------
// Test cases start here

func power() uint64 {
	// % can return negative numbers, so this ensures result is positive
	return uint64(cmn.RandInt()%50 + 60)
}

func makeVal() *wrsp.Validator {
	return &wrsp.Validator{
		PubKey: cmn.RandBytes(10),
		Power:  power(),
	}
}

// newPower returns a copy of the validator with a different power
func newPower(val *wrsp.Validator) *wrsp.Validator {
	res := *val
	res.Power = power()
	if res.Power == val.Power {
		panic("no no")
	}
	return &res
}

func TestEndBlock(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	logger := log.NewNopLogger()
	store := MockStore()
	handler := ValSetHandler{}
	app := NewBasecoin(handler, store, logger)

	val1 := makeVal()
	val2 := makeVal()
	val3 := makeVal()
	val1a := newPower(val1)
	val2a := newPower(val2)

	cases := [...]struct {
		changes  [][]*wrsp.Validator
		expected []*wrsp.Validator
	}{
		// Nothing in, nothing out, no crash
		0: {},
		// One in, one out, no problem
		1: {
			changes:  [][]*wrsp.Validator{{val1}},
			expected: []*wrsp.Validator{val1},
		},
		// Combine a few ones
		2: {
			changes:  [][]*wrsp.Validator{{val1}, {val2, val3}},
			expected: []*wrsp.Validator{val1, val2, val3},
		},
		// Make sure changes all to one validators are squished into one diff
		3: {
			changes:  [][]*wrsp.Validator{{val1}, {val2, val1a}, {val2a, val3}},
			expected: []*wrsp.Validator{val1a, val2a, val3},
		},
	}

	for i, tc := range cases {
		app.BeginBlock(nil, nil)
		for _, c := range tc.changes {
			tx := ValChangeTx{c}.Wrap()
			txBytes := wire.BinaryBytes(tx)
			res := app.DeliverTx(txBytes)
			require.True(res.IsOK(), "%#v", res)
		}
		diff := app.EndBlock(app.height)
		// TODO: don't care about order here...
		assert.Equal(tc.expected, diff.Diffs, "%d", i)
	}
}
