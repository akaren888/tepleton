package auth

import (
	"testing"

	sdk "github.com/tepleton/tepleton-sdk/types"
	"github.com/stretchr/testify/assert"
	crypto "github.com/tepleton/go-crypto"
	wire "github.com/tepleton/go-wire"
)

func TestBaseAccount(t *testing.T) {
	key := crypto.GenPrivKeyEd25519()
	pub := key.PubKey()
	addr := pub.Address()
	someCoins := sdk.Coins{{"atom", 123}, {"eth", 246}}
	seq := int64(7)

	acc := NewBaseAccountWithAddress(addr)

	err := acc.SetPubKey(pub)
	assert.Nil(t, err)
	assert.Equal(t, pub, acc.GetPubKey())

	assert.Equal(t, addr, acc.GetAddress())

	err = acc.SetCoins(someCoins)
	assert.Nil(t, err)
	assert.Equal(t, someCoins, acc.GetCoins())

	err = acc.SetSequence(seq)
	assert.Nil(t, err)
	assert.Equal(t, seq, acc.GetSequence())

	b, err := wire.MarshalBinary(acc)
	assert.Nil(t, err)

	var acc2 BaseAccount
	err = wire.UnmarshalBinary(b, &acc2)
	assert.Nil(t, err)
	assert.Equal(t, acc, acc2)

	acc2 = BaseAccount{}
	err = wire.UnmarshalBinary(b[:len(b)/2], &acc2)
	assert.NotNil(t, err)
}
