package pmpgp

import (
	"bytes"
	"crypto"
	"hash"
	"time"

	"github.com/ProtonMail/go-crypto/openpgp/packet"
	"github.com/sassoftware/go-rpmutils/internal"
)

type Signer struct {
	PrivateKey *packet.PrivateKey
}

func (s Signer) Sign(h hash.Hash, hashType crypto.Hash, creationTime time.Time) ([]byte, error) {
	sig := &packet.Signature{
		Version:           s.PrivateKey.Version,
		SigType:           packet.SigTypeBinary,
		CreationTime:      creationTime,
		PubKeyAlgo:        s.PrivateKey.PublicKey.PubKeyAlgo,
		Hash:              hashType,
		IssuerKeyId:       &s.PrivateKey.KeyId,
		IssuerFingerprint: s.PrivateKey.PublicKey.Fingerprint,
	}
	err := sig.Sign(h, s.PrivateKey, nil)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	err = sig.Serialize(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ensure the interface is implemented
var _ internal.Signer = Signer{}
