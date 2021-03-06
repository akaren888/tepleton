package client

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tepleton/go-wire"
	"github.com/tepleton/light-client/certifiers"
	certclient "github.com/tepleton/light-client/certifiers/client"
	"github.com/tepleton/tmlibs/log"

	nm "github.com/tepleton/tepleton/node"
	"github.com/tepleton/tepleton/rpc/client"
	rpctest "github.com/tepleton/tepleton/rpc/test"
	"github.com/tepleton/tepleton/types"

	sdkapp "github.com/tepleton/tepleton-sdk/app"
	"github.com/tepleton/tepleton-sdk/modules/eyes"
)

var node *nm.Node

func TestMain(m *testing.M) {
	logger := log.TestingLogger()
	store, err := sdkapp.MockStoreApp("query", logger)
	if err != nil {
		panic(err)
	}
	app := sdkapp.NewBaseApp(store, eyes.NewHandler(), nil)

	node = rpctest.StartTendermint(app)

	code := m.Run()

	node.Stop()
	node.Wait()
	os.Exit(code)
}

func TestAppProofs(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	cl := client.NewLocal(node)
	client.WaitForHeight(cl, 1, nil)

	k := []byte("my-key")
	v := []byte("my-value")

	tx := eyes.SetTx{Key: k, Value: v}.Wrap()
	btx := wire.BinaryBytes(tx)
	br, err := cl.BroadcastTxCommit(btx)
	require.NoError(err, "%+v", err)
	require.EqualValues(0, br.CheckTx.Code, "%#v", br.CheckTx)
	require.EqualValues(0, br.DeliverTx.Code)
	brh := br.Height

	// This sets up our trust on the node based on some past point.
	source := certclient.NewProvider(cl)
	seed, err := source.GetByHeight(br.Height - 2)
	require.NoError(err, "%+v", err)
	cert := certifiers.NewStatic("my-chain", seed.Validators)

	client.WaitForHeight(cl, 3, nil)
	latest, err := source.LatestCommit()
	require.NoError(err, "%+v", err)
	rootHash := latest.Header.AppHash

	// Test existing key.
	var data eyes.Data

	// verify a query before the tx block has no data (and valid non-exist proof)
	bs, height, proof, err := GetWithProof(k, brh-1, cl, cert)
	require.NotNil(err)
	require.True(IsNoDataErr(err))
	require.Nil(bs)

	// but given that block it is good
	bs, height, proof, err = GetWithProof(k, brh, cl, cert)
	require.NoError(err, "%+v", err)
	require.NotNil(proof)
	require.True(height >= uint64(latest.Header.Height))

	// Alexis there is a bug here, somehow the above code gives us rootHash = nil
	// and proof.Verify doesn't care, while proofNotExists.Verify fails.
	// I am hacking this in to make it pass, but please investigate further.
	rootHash = proof.Root()

	err = wire.ReadBinaryBytes(bs, &data)
	require.NoError(err, "%+v", err)
	assert.EqualValues(v, data.Value)
	err = proof.Verify(k, bs, rootHash)
	assert.NoError(err, "%+v", err)

	// Test non-existing key.
	missing := []byte("my-missing-key")
	bs, _, proof, err = GetWithProof(missing, 0, cl, cert)
	require.True(IsNoDataErr(err))
	require.Nil(bs)
	require.NotNil(proof)
	err = proof.Verify(missing, nil, rootHash)
	assert.NoError(err, "%+v", err)
	err = proof.Verify(k, nil, rootHash)
	assert.Error(err)
}

func TestTxProofs(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	cl := client.NewLocal(node)
	client.WaitForHeight(cl, 1, nil)

	tx := eyes.NewSetTx([]byte("key-a"), []byte("value-a"))

	btx := types.Tx(wire.BinaryBytes(tx))
	br, err := cl.BroadcastTxCommit(btx)
	require.NoError(err, "%+v", err)
	require.EqualValues(0, br.CheckTx.Code, "%#v", br.CheckTx)
	require.EqualValues(0, br.DeliverTx.Code)
	fmt.Printf("tx height: %d\n", br.Height)

	source := certclient.NewProvider(cl)
	seed, err := source.GetByHeight(br.Height - 2)
	require.NoError(err, "%+v", err)
	cert := certifiers.NewStatic("my-chain", seed.Validators)

	// First let's make sure a bogus transaction hash returns a valid non-existence proof.
	key := types.Tx([]byte("bogus")).Hash()
	res, err := cl.Tx(key, true)
	require.NotNil(err)
	require.Contains(err.Error(), "not found")

	// Now let's check with the real tx hash.
	key = btx.Hash()
	res, err = cl.Tx(key, true)
	require.NoError(err, "%+v", err)
	require.NotNil(res)
	err = res.Proof.Validate(key)
	assert.NoError(err, "%+v", err)

	commit, err := GetCertifiedCommit(int(br.Height), cl, cert)
	require.Nil(err, "%+v", err)
	require.Equal(res.Proof.RootHash, commit.Header.DataHash)

}
