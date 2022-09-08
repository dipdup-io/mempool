package endorsement

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"math/big"

	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/ed25519"

	"github.com/pkg/errors"
	"github.com/ubiq/go-ubiq/crypto/secp256k1"
)

var watermarks = map[string][]byte{}

// Hash -
func Hash(chainID string, msg []byte) [32]byte {
	return blake2b.Sum256(append(getWatermark(chainID), msg...))
}

// CheckKey -
func CheckKey(prefix string, key, signature []byte, hash [32]byte) bool {
	switch prefix {
	case "edpk":
		return ed25519.Verify(key, hash[:], signature)
	case "sppk":
		return secp256k1.VerifySignature(key, hash[:], signature)
	case "p2pk":
		return verifyP256(key, hash[:], signature)
	default:
		return false
	}
}

func getWatermark(chainID string) []byte {
	watermark, ok := watermarks[chainID]
	if !ok {
		watermark = append([]byte{2}, decodeChainID(chainID)...)
		watermarks[chainID] = watermark
	}
	return watermark
}

func verifyP256(key, message, signature []byte) bool {
	x, y, err := getCoordinates(key)
	if err != nil {
		return false
	}

	pubKey := ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     x,
		Y:     y,
	}

	sign1 := new(big.Int).SetBytes(signature[:32])
	sign2 := new(big.Int).SetBytes(signature[32:])

	return ecdsa.Verify(&pubKey, message, sign1, sign2)
}

// https://stackoverflow.com/a/46289709
func getCoordinates(data []byte) (*big.Int, *big.Int, error) {
	// Split the sign byte from the rest
	signByte := uint(data[0])
	xBytes := data[1:]

	// The params for P256
	c := elliptic.P256().Params()

	// Convert to big Int.
	x := new(big.Int).SetBytes(xBytes)
	x3 := new(big.Int).Mul(x, x)
	x3 = x3.Mul(x3, x)

	threeX := new(big.Int).Lsh(x, 1)
	threeX.Add(threeX, x)

	y2 := x3.Sub(x3, threeX)
	y2 = y2.Add(y2, c.B)
	y2 = y2.Mod(y2, c.P)

	y := y2.ModSqrt(y2, c.P)
	if y == nil {
		// If this happens then you're dealing with an invalid point.
		// Panic, return an error, whatever you want.
		return new(big.Int), new(big.Int), errors.New("Invalid point")
	}

	// Finally, check if you have the correct root. If not you want
	// -y mod P
	if y.Bit(0) != signByte&1 {
		y.Neg(y)
		y.Mod(y, c.P)
	}

	return x, y, nil
}
