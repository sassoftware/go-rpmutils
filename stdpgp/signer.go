package stdpgp

import (
	"bytes"
	"crypto"
	"hash"
	"time"

	"github.com/sassoftware/go-rpmutils/internal"
	"golang.org/x/crypto/openpgp/packet"
)

type Signer struct {
	PrivateKey *packet.PrivateKey
}

func (s Signer) Sign(h hash.Hash, hashType crypto.Hash, creationTime time.Time) ([]byte, error) {
	sig := &packet.Signature{
		SigType:      packet.SigTypeBinary,
		CreationTime: creationTime,
		PubKeyAlgo:   s.PrivateKey.PublicKey.PubKeyAlgo,
		Hash:         hashType,
		IssuerKeyId:  &s.PrivateKey.KeyId,
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
