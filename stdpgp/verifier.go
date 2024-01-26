package stdpgp

import (
	"bytes"
	"hash"
	"io"

	"github.com/sassoftware/go-rpmutils/internal"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/packet"
)

// Verifier uses the deprecated openpgp implementation from the Go standard library.
// Other implementations have dropped support for PGP signature V3 packets,
// which are still frequently found even in newer RPM-based distributions.
// Thus, this implementation is still recommended for applications
// that intend to validate RPMs from many sources.
type Verifier struct {
	// EntityList retrieves known public keys by ID
	EntityList openpgp.EntityList
	// ParseOnly disables signature verification.
	// Verifier parses and returns only the Key ID from each signature.
	// Keys will not be searched in EntityList and errors will not be raised for unknown keys.
	ParseOnly bool
}

func (v Verifier) ParseSignature(blob []byte) (*internal.Signature, error) {
	packetReader := packet.NewReader(bytes.NewReader(blob))
	genpkt, err := packetReader.Next()
	if err != nil {
		return nil, err
	}
	var sig *internal.Signature
	switch pkt := genpkt.(type) {
	case *packet.SignatureV3:
		sig = &internal.Signature{
			Hash:         pkt.Hash,
			CreationTime: pkt.CreationTime,
			KeyId:        pkt.IssuerKeyId,
		}
		sig.Validate = func(h hash.Hash) error {
			key, err := v.findKey(sig)
			if key != nil {
				err = key.VerifySignatureV3(h, pkt)
			}
			return err
		}
	case *packet.Signature:
		if pkt.IssuerKeyId == nil {
			return nil, internal.ErrNoPGPSignature
		}
		sig = &internal.Signature{
			Hash:         pkt.Hash,
			CreationTime: pkt.CreationTime,
			KeyId:        *pkt.IssuerKeyId,
		}
		sig.Validate = func(h hash.Hash) error {
			key, err := v.findKey(sig)
			if key != nil {
				err = key.VerifySignature(h, pkt)
			}
			return err
		}
	default:
		return nil, internal.ErrNoPGPSignature
	}
	_, err = packetReader.Next()
	if err != io.EOF {
		return nil, internal.ErrTrailingGarbage
	}
	return sig, nil
}

func (v Verifier) findKey(sig *internal.Signature) (*packet.PublicKey, error) {
	if v.ParseOnly {
		return nil, nil
	}
	keys := v.EntityList.KeysById(sig.KeyId)
	if len(keys) == 0 || keys[0].PublicKey == nil {
		return nil, internal.KeyNotFoundError{KeyID: sig.KeyId}
	}
	sig.Signer = keys[0].Entity
	return keys[0].PublicKey, nil
}

// ensure the interface is implemented
var _ internal.Verifier = Verifier{}
