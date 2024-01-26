//go:build sigv3

package pmpgp

import (
	"hash"

	"github.com/ProtonMail/go-crypto/openpgp/packet"
	"github.com/sassoftware/go-rpmutils/internal"
)

// Using a fork of PM that does parse V3 signatures, so check if that's what it is
func (v Verifier) parseSignatureMaybeV3(genpkt packet.Packet) (*internal.Signature, error) {
	pkt, ok := genpkt.(*packet.SignatureV3)
	if !ok {
		return v.parseSignatureV4(genpkt)
	}
	sig := &internal.Signature{
		Hash:         pkt.Hash,
		CreationTime: pkt.CreationTime,
		KeyId:        pkt.IssuerKeyId,
	}
	sig.Validate = func(h hash.Hash) error {
		if v.ParseOnly {
			return nil
		}
		keys := v.EntityList.KeysById(pkt.IssuerKeyId)
		if len(keys) == 0 || keys[0].PublicKey == nil {
			return internal.KeyNotFoundError{KeyID: pkt.IssuerKeyId}
		}
		sig.Signer = keys[0].Entity
		return keys[0].PublicKey.VerifySignatureV3(h, pkt)
	}
	return sig, nil
}
