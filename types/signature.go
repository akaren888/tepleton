package types

import crypto "github.com/tepleton/go-crypto"

// Standard Signature
type StdSignature struct {
	crypto.PubKey // optional
	crypto.Signature
	Sequence int64
}
