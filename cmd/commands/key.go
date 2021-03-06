package commands

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	//"github.com/pkg/errors"

	"github.com/spf13/viper"

	"github.com/tepleton/go-crypto"
	"github.com/tepleton/tmlibs/cli"
)

//---------------------------------------------
// simple implementation of a key

type Address [20]byte

func (a Address) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%x"`, a[:])), nil
}

func (a *Address) UnmarshalJSON(addrHex []byte) error {
	addr, err := hex.DecodeString(strings.Trim(string(addrHex), `"`))
	if err != nil {
		return err
	}
	copy(a[:], addr)
	return nil
}

type Key struct {
	Address Address        `json:"address"`
	PubKey  crypto.PubKey  `json:"pub_key"`
	PrivKey crypto.PrivKey `json:"priv_key"`
}

// Implements Signer
func (k *Key) Sign(msg []byte) crypto.Signature {
	return k.PrivKey.Sign(msg)
}

func LoadKey(keyFile string) (*Key, error) {
	filePath := keyFile

	if !strings.HasPrefix(keyFile, "/") && !strings.HasPrefix(keyFile, ".") {
		rootDir := viper.GetString(cli.HomeFlag)
		filePath = path.Join(rootDir, keyFile)
	}

	keyJSONBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	key := new(Key)
	err = json.Unmarshal(keyJSONBytes, key)
	if err != nil {
		return nil, fmt.Errorf("Error reading key from %v: %v\n", filePath, err) //never stack trace
	}

	return key, nil
}
