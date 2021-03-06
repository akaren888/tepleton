package merkle

import (
	"bytes"
	"fmt"
	"path"
	"path/filepath"
	"strings"

	wrsp "github.com/tepleton/wrsp/types"
	"github.com/tepleton/go-wire"
	cmn "github.com/tepleton/tmlibs/common"
	dbm "github.com/tepleton/tmlibs/db"
	"github.com/tepleton/tmlibs/log"

	"github.com/tepleton/merkleeyes/iavl"
)

// Store contains the merkle tree, and all info to handle wrsp requests
type Store struct {
	State
	height    uint64
	hash      []byte
	persisted bool

	logger log.Logger
}

var stateKey = []byte("merkle:state") // Database key for merkle tree save value db values

// ChainState contains the latest Merkle root hash and the number of times `Commit` has been called
type ChainState struct {
	Hash   []byte
	Height uint64
}

// NewStore initializes an in-memory IAVLTree, or attempts to load a persistant
// tree from disk
func NewStore(dbName string, cacheSize int, logger log.Logger) *Store {
	// start at 1 so the height returned by query is for the
	// next block, ie. the one that includes the AppHash for our current state
	initialHeight := uint64(1)

	// Non-persistent case
	if dbName == "" {
		tree := iavl.NewIAVLTree(
			0,
			nil,
		)
		return &Store{
			State:  NewState(tree, false),
			height: initialHeight,
			logger: logger,
		}
	}

	// Expand the path fully
	dbPath, err := filepath.Abs(dbName)
	if err != nil {
		panic(fmt.Sprintf("Invalid Database Name: %s", dbName))
	}

	// Some external calls accidently add a ".db", which is now removed
	dbPath = strings.TrimSuffix(dbPath, path.Ext(dbPath))

	// Split the database name into it's components (dir, name)
	dir := path.Dir(dbPath)
	name := path.Base(dbPath)

	// Make sure the path exists
	empty, _ := cmn.IsDirEmpty(dbPath + ".db")

	// Open database called "dir/name.db", if it doesn't exist it will be created
	db := dbm.NewDB(name, dbm.LevelDBBackendStr, dir)
	tree := iavl.NewIAVLTree(cacheSize, db)

	var chainState ChainState
	if empty {
		logger.Info("no existing db, creating new db")
		chainState = ChainState{
			Hash:   tree.Save(),
			Height: initialHeight,
		}
		db.Set(stateKey, wire.BinaryBytes(chainState))
	} else {
		logger.Info("loading existing db")
		eyesStateBytes := db.Get(stateKey)
		err = wire.ReadBinaryBytes(eyesStateBytes, &chainState)
		if err != nil {
			logger.Error("error reading MerkleEyesState", "err", err)
			panic(err)
		}
		tree.Load(chainState.Hash)
	}

	return &Store{
		State:     NewState(tree, true),
		height:    chainState.Height,
		hash:      chainState.Hash,
		persisted: true,
		logger:    logger,
	}
}

// CloseDB closes the database
// func (s *Store) CloseDB() {
// 	if s.db != nil {
// 		s.db.Close()
// 	}
// }

// Info implements wrsp.Application. It returns the height, hash and size (in the data).
// The height is the block that holds the transactions, not the apphash itself.
func (s *Store) Info() wrsp.ResponseInfo {
	s.logger.Info("Info synced",
		"height", s.height,
		"hash", fmt.Sprintf("%X", s.hash))
	return wrsp.ResponseInfo{
		Data:             cmn.Fmt("size:%v", s.State.Committed().Size()),
		LastBlockHeight:  s.height - 1,
		LastBlockAppHash: s.hash,
	}
}

// Commit implements wrsp.Application
func (s *Store) Commit() wrsp.Result {
	s.hash = s.State.Hash()
	s.height++
	s.logger.Debug("Commit synced",
		"height", s.height,
		"hash", fmt.Sprintf("%X", s.hash))

	s.State.BatchSet(stateKey, wire.BinaryBytes(ChainState{
		Hash:   s.hash,
		Height: s.height,
	}))

	hash := s.State.Commit()
	if !bytes.Equal(hash, s.hash) {
		panic("AppHash is incorrect")
	}

	if s.State.Committed().Size() == 0 {
		return wrsp.NewResultOK(nil, "Empty hash for empty tree")
	}
	return wrsp.NewResultOK(s.hash, "")
}

// Query implements wrsp.Application
func (s *Store) Query(reqQuery wrsp.RequestQuery) (resQuery wrsp.ResponseQuery) {

	if reqQuery.Height != 0 {
		// TODO: support older commits
		resQuery.Code = wrsp.CodeType_InternalError
		resQuery.Log = "merkleeyes only supports queries on latest commit"
		return
	}

	// set the query response height to current
	resQuery.Height = s.height

	tree := s.State.Committed()

	switch reqQuery.Path {
	case "/store", "/key": // Get by key
		key := reqQuery.Data // Data holds the key bytes
		resQuery.Key = key
		if reqQuery.Prove {
			value, proof, exists := tree.Proof(key)
			if !exists {
				resQuery.Log = "Key not found"
			}
			resQuery.Value = value
			resQuery.Proof = proof
		} else {
			value := tree.Get(key)
			resQuery.Value = value
		}

	default:
		resQuery.Code = wrsp.CodeType_UnknownRequest
		resQuery.Log = cmn.Fmt("Unexpected Query path: %v", reqQuery.Path)
	}
	return
}
