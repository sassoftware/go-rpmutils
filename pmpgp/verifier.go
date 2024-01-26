package pmpgp

import (
	"bytes"
	"hash"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/packet"
	"github.com/sassoftware/go-rpmutils/internal"
)

// Verifier uses ProtonMail's forked openpgp implementation.
// This implementation has removed many obsolete features, including signature V3 packets
// which are still frequently found even in newer RPM-based distributions.
type Verifier struct {
	// EntityList retrieves known public keys by ID
	EntityList openpgp.EntityList
	// ParseOnly disables signature verification.
	// Verifier parses and returns only the Key ID from each signature.
	// Keys will not be searched in EntityList and errors will not be raised for unknown keys.
	ParseOnly bool
}

func (v Verifier) ParseSignature(blob []byte) (*internal.Signature, error) {
	reader := bytes.NewReader(blob)
	genpkt, err := packet.Read(reader)
	if err != nil {
		return nil, err
	} else if reader.Len() > 0 {
		return nil, internal.ErrTrailingGarbage
	}
	return v.parseSignatureMaybeV3(genpkt)
}

func (v Verifier) parseSignatureV4(genpkt packet.Packet) (*internal.Signature, error) {
	pkt, ok := genpkt.(*packet.Signature)
	if !ok || (pkt.IssuerKeyId == nil && len(pkt.IssuerFingerprint) == 0) {
		return nil, internal.ErrNoPGPSignature
	}
	sig := &internal.Signature{
		Hash:           pkt.Hash,
		CreationTime:   pkt.CreationTime,
		KeyId:          *pkt.IssuerKeyId,
		KeyFingerprint: pkt.IssuerFingerprint,
	}
	sig.Validate = func(h hash.Hash) error {
		if v.ParseOnly {
			return nil
		}
		if entity, key := v.findKey(pkt); key != nil {
			sig.Signer = entity
			return key.VerifySignature(h, pkt)
		}
		return internal.KeyNotFoundError{
			KeyID:       sig.KeyId,
			Fingerprint: sig.KeyFingerprint,
		}
	}
	return sig, nil
}

// find a key by its fingerprint (if available) or key ID
func (v Verifier) findKey(sig *packet.Signature) (*openpgp.Entity, *packet.PublicKey) {
	for _, entity := range v.EntityList {
		if sig.CheckKeyIdOrFingerprint(entity.PrimaryKey) {
			return entity, entity.PrimaryKey
		}
		for _, sub := range entity.Subkeys {
			if sig.CheckKeyIdOrFingerprint(sub.PublicKey) {
				return entity, sub.PublicKey
			}
		}
	}
	return nil, nil
}

// ensure the interface is implemented
var _ internal.Verifier = Verifier{}
