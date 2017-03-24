package session

import (
	"encoding/base64"
	"os"

	"github.com/gorilla/securecookie"
	"github.com/pkg/errors"
)

// EnvCodecs retrieves the codecs from the environment variables HASH_KEY and BLOCK_KEY
func EnvCodecs() ([]securecookie.Codec, error) {
	hashStr := os.Getenv("HASH_KEY")
	blockStr := os.Getenv("BLOCK_KEY")

	if hashStr == "" || blockStr == "" {
		return []securecookie.Codec{}, nil
	}

	hashKey, err := base64.StdEncoding.DecodeString(hashStr)
	if err != nil {
		return nil, errors.Wrap(err, "app:codec:err:hash_key")
	}

	blockKey, err := base64.StdEncoding.DecodeString(blockStr)
	if err != nil {
		return nil, errors.Wrap(err, "app:codec:err:block_key")
	}

	return []securecookie.Codec{
		securecookie.New(hashKey, blockKey),
	}, nil
}

// GenerateCodec generates new keypairs suitable for gorilla sessions; blockSizes must be in pairs divisible by 32
// by default GenerateCodec will generate a codec with 64, 32
func GenerateKeyPairs(blockSizes ...int) [][]byte {
	if len(blockSizes) == 0 {
		blockSizes = []int{64, 32}
	}

	pairs := make([][]byte, 0, len(blockSizes))
	for _, size := range blockSizes {
		pairs = append(pairs, securecookie.GenerateRandomKey(size))
	}

	return pairs
}
