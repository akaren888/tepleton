package state

// TODO: something similar in the merkle package...

// import (
// 	"bytes"
// 	"testing"

// 	eyes "github.com/tepleton/merkleeyes/client"
// 	"github.com/tepleton/tmlibs/log"

// 	"github.com/stretchr/testify/assert"
// )

// func TestState(t *testing.T) {
// 	assert := assert.New(t)

// 	//States and Stores for tests
// 	store := NewMemKVStore()
// 	state := NewState(store, log.TestingLogger())
// 	cache := state.CacheWrap()
// 	eyesCli := eyes.NewLocalClient("", 0)

// 	//reset the store/state/cache
// 	reset := func() {
// 		store = NewMemKVStore()
// 		state = NewState(store, log.TestingLogger())
// 		cache = state.CacheWrap()
// 	}

// 	//set the state to using the eyesCli instead of MemKVStore
// 	useEyesCli := func() {
// 		state = NewState(eyesCli, log.TestingLogger())
// 		cache = state.CacheWrap()
// 	}

// 	//key value pairs to be tested within the system
// 	keyvalue := []struct {
// 		key   string
// 		value string
// 	}{
// 		{"foo", "snake"},
// 		{"bar", "mouse"},
// 	}

// 	//set the kvc to have all the key value pairs
// 	setRecords := func(kv KVStore) {
// 		for _, n := range keyvalue {
// 			kv.Set([]byte(n.key), []byte(n.value))
// 		}
// 	}

// 	//store has all the key value pairs
// 	storeHasAll := func(kv KVStore) bool {
// 		for _, n := range keyvalue {
// 			if !bytes.Equal(kv.Get([]byte(n.key)), []byte(n.value)) {
// 				return false
// 			}
// 		}
// 		return true
// 	}

// 	//test chainID
// 	state.SetChainID("testchain")
// 	assert.Equal(state.GetChainID(), "testchain", "ChainID is improperly stored")

// 	//test basic retrieve
// 	setRecords(state)
// 	assert.True(storeHasAll(state), "state doesn't retrieve after Set")

// 	//Test CacheWrap with local mem store
// 	reset()
// 	setRecords(cache)
// 	assert.False(storeHasAll(store), "store retrieving before CacheSync")
// 	cache.CacheSync()
// 	assert.True(storeHasAll(store), "store doesn't retrieve after CacheSync")

// 	//Test Commit on state with non-merkle store
// 	assert.True(state.Commit().IsErr(), "Commit shouldn't work with non-merkle store")

// 	//Test CacheWrap with merkleeyes client store
// 	useEyesCli()
// 	setRecords(cache)
// 	assert.False(storeHasAll(eyesCli), "eyesCli retrieving before Commit")
// 	cache.CacheSync()
// 	assert.True(state.Commit().IsOK(), "Bad Commit")
// 	assert.True(storeHasAll(eyesCli), "eyesCli doesn't retrieve after Commit")
// }
