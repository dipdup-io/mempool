package endorsement

import (
	"testing"

	"github.com/dipdup-net/go-lib/node"
	"github.com/dipdup-net/go-lib/tools/forge"
	"github.com/dipdup-net/mempool/cmd/mempool/models"
)

func TestCheckKey(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		signature   string
		chainID     string
		endorsement models.Endorsement
		want        bool
	}{
		{
			name:      "ed25519",
			key:       "edpkuEhzJqdFBCWMw6TU3deADRK2fq3GuwWFUphwyH7ero1Na4oGFP",
			signature: "siggEYDRoz7tiECt2fc1M75ieJNeVAP6MGHLhpyPpPue8EU3QYjYSJLnDoDPgxkmrjr6R33qGrAxLASwkyQqa1r3tc5mGPwT",
			chainID:   "NetXdQprcVkpaWU",
			endorsement: models.Endorsement{
				Level: 751292,
				MempoolOperation: models.MempoolOperation{
					Branch:    "BMbpxQAU7Jat7g9ZnKrP3brgqFX6r2VX8PPXCxNbFZeURA6DbEF",
					Signature: "siggEYDRoz7tiECt2fc1M75ieJNeVAP6MGHLhpyPpPue8EU3QYjYSJLnDoDPgxkmrjr6R33qGrAxLASwkyQqa1r3tc5mGPwT",
				},
			},
			want: true,
		}, {
			name:      "secp256k1",
			key:       "sppk7bMuoa8w2LSKz3XEuPsKx1WavsMLCWgbWG9CZNAsJg9eTmkXRPd",
			signature: "sigdoarpkht5iMuEHcUPnr9KWMuScWgsMru1amLHzyV4cDxXVCVxQw8dXDcKrvVPd9XbkK2aDzN4Dbo8EiP3ythKZNKCCwFf",
			chainID:   "NetXdQprcVkpaWU",
			endorsement: models.Endorsement{
				Level: 751179,
				MempoolOperation: models.MempoolOperation{
					Branch:    "BM2JkusQmT885mqjKiJXfMJrgQXZTwoEsCM8tkuvGyJLPrSw2ih",
					Signature: "sigdoarpkht5iMuEHcUPnr9KWMuScWgsMru1amLHzyV4cDxXVCVxQw8dXDcKrvVPd9XbkK2aDzN4Dbo8EiP3ythKZNKCCwFf",
				},
			},
			want: true,
		}, {
			name:      "p256",
			key:       "p2pk66iTZwLmRPshQgUr2HE3RUzSFwAN5MNaBQ5rfduT1dGKXd25pNN",
			signature: "sigTtssYjbVCgJR8WpiDgyg3dSuGF1STSjQLXxTSFiWzTGdS2CdFUBKZUbDwK42L1gsYYN8fqryXQAfuZVk1wknBaKhsNTZP",
			chainID:   "NetXdQprcVkpaWU",
			endorsement: models.Endorsement{
				Level: 751447,
				MempoolOperation: models.MempoolOperation{
					Branch:    "BLp1dxsyPLc58x4cSMKGVevdQfgo9VBHy46kqnhsJSrNgteDPex",
					Signature: "sigTtssYjbVCgJR8WpiDgyg3dSuGF1STSjQLXxTSFiWzTGdS2CdFUBKZUbDwK42L1gsYYN8fqryXQAfuZVk1wknBaKhsNTZP",
				},
			},
			want: true,
		}, {
			name:      "ed25519: PosDog",
			key:       "edpku4Jnsyp9geSL3W4xEwGhtTDjbM89Q7RyG43fftxzR3Cs4YY6K7",
			signature: "signGt3ExC4ELXEvzuPwpBFmfyo4CEqKktawD9ZH6a7D9Tn7zg2Tt9L7HYAiWb8u9Yjs8MnUpNNzixpD2y9TYQxqYnQzbp2s",
			chainID:   "NetXdQprcVkpaWU",
			endorsement: models.Endorsement{
				Level: 1479809,
				MempoolOperation: models.MempoolOperation{
					Branch:    "BL38RNz32eAVhgvV5bUMWUxGMq2v2wkst9UaN7CbW7hLC6THCQ6",
					Signature: "signGt3ExC4ELXEvzuPwpBFmfyo4CEqKktawD9ZH6a7D9Tn7zg2Tt9L7HYAiWb8u9Yjs8MnUpNNzixpD2y9TYQxqYnQzbp2s",
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := forge.Endorsement(node.Endorsement{Level: tt.endorsement.Level}, tt.endorsement.Branch)
			if err != nil {
				t.Errorf("forge.Endorsement() err = %s", err.Error())
				return
			}
			if got := CheckKey(tt.key[:4], DecodePublicKey(tt.key), DecodeSignature(tt.signature), Hash(tt.chainID, data)); got != tt.want {
				t.Errorf("CheckKey() = %v, want %v", got, tt.want)
			}
		})
	}
}
