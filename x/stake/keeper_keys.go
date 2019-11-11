package stake

import (
	"encoding/binary"

	sdk "github.com/tepleton/tepleton-sdk/types"
	"github.com/tepleton/tepleton-sdk/wire"
	crypto "github.com/tepleton/go-crypto"
)

// TODO remove some of these prefixes once have working multistore

//nolint
var (
	// Keys for store prefixes
	ParamKey               = []byte{0x00} // key for global parameters relating to staking
	PoolKey                = []byte{0x01} // key for global parameters relating to staking
	CandidatesKey          = []byte{0x02} // prefix for each key to a candidate
	ValidatorsKey          = []byte{0x03} // prefix for each key to a validator
	AccUpdateValidatorsKey = []byte{0x04} // prefix for each key to a validator which is being updated
	CurrentValidatorsKey   = []byte{0x05} // prefix for each key to the last updated validator group
	ToKickOutValidatorsKey = []byte{0x06} // prefix for each key to the last updated validator group
	DelegationKeyPrefix    = []byte{0x07} // prefix for each key to a delegator's bond
	IntraTxCounterKey      = []byte{0x08} // key for block-local tx index
)

const maxDigitsForAccount = 12 // ~220,000,000 atoms created at launch

// get the key for the candidate with address
func GetCandidateKey(addr sdk.Address) []byte {
	return append(CandidatesKey, addr.Bytes()...)
}

// get the key for the validator used in the power-store
func GetValidatorKey(validator Validator) []byte {
	powerBytes := []byte(validator.Power.ToLeftPadded(maxDigitsForAccount)) // power big-endian (more powerful validators first)

	// TODO ensure that the key will be a readable string.. probably should add seperators and have
	// heightBytes and counterBytes represent strings like powerBytes does
	heightBytes := make([]byte, binary.MaxVarintLen64)
	binary.BigEndian.PutUint64(heightBytes, ^uint64(validator.Height)) // invert height (older validators first)
	counterBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(counterBytes, ^uint16(validator.Counter)) // invert counter (first txns have priority)
	return append(ValidatorsKey,
		append(powerBytes,
			append(heightBytes,
				append(counterBytes, validator.Address.Bytes()...)...)...)...)
}

// get the key for the accumulated update validators
func GetAccUpdateValidatorKey(addr sdk.Address) []byte {
	return append(AccUpdateValidatorsKey, addr.Bytes()...)
}

// get the key for the current validator group, ordered like tepleton
func GetCurrentValidatorsKey(pk crypto.PubKey) []byte {
	addr := pk.Address()
	return append(CurrentValidatorsKey, addr.Bytes()...)
}

// get the key for the accumulated update validators
func GetToKickOutValidatorKey(addr sdk.Address) []byte {
	return append(ToKickOutValidatorsKey, addr.Bytes()...)
}

// get the key for delegator bond with candidate
func GetDelegationKey(delegatorAddr, candidateAddr sdk.Address, cdc *wire.Codec) []byte {
	return append(GetDelegationsKey(delegatorAddr, cdc), candidateAddr.Bytes()...)
}

// get the prefix for a delegator for all candidates
func GetDelegationsKey(delegatorAddr sdk.Address, cdc *wire.Codec) []byte {
	res, err := cdc.MarshalBinary(&delegatorAddr)
	if err != nil {
		panic(err)
	}
	return append(DelegationKeyPrefix, res...)
}

//______________________________________________________________

// remove the prefix byte from a key, possibly revealing and address
func AddrFromKey(key []byte) sdk.Address {
	return key[1:]
}
