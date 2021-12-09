package endorsement

import (
	"github.com/btcsuite/btcutil/base58"
)

// DecodeSignature -
func DecodeSignature(signature string) []byte {
	decoded := base58.Decode(signature)
	return decoded[3 : len(decoded)-4]
}

func decodeChainID(chainID string) []byte {
	decoded := base58.Decode(chainID)
	if len(decoded) < 3 {
		return []byte(chainID)
	}
	return decoded[3 : len(decoded)-4]
}

// DecodePublicKey -
func DecodePublicKey(key string) []byte {
	decoded := base58.Decode(key)
	return decoded[4 : len(decoded)-4]
}
