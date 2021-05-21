package models

import (
	"encoding/binary"

	"github.com/btcsuite/btcutil/base58"
)

func forgeString(s string) []byte {
	decoded := base58.Decode(s)
	return decoded[2 : len(decoded)-4]
}

func forgeInt(n int64) []byte {
	bs := make([]byte, 4)
	binary.BigEndian.PutUint32(bs, uint32(n))
	return bs
}

func forgeNat(n uint64) (out []byte) {
	more := true
	for more {
		b := byte(n & 0x7F)
		n >>= 7
		if n == 0 {
			more = false
		} else {
			b |= 0x80
		}
		out = append(out, b)
	}
	return
}
