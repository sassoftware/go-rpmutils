package internal

import (
	"crypto"
	"errors"
	"fmt"
	"hash"
	"time"
)

// Signature holds a parsed PGP signature from an RPM
type Signature struct {
	// Signer is the PGP identity that created the signature. It may be nil if
	// the public key is not available at verification time, but KeyId will
	// always be set.
	Signer any
	// Hash is the algorithm used to digest the signature contents
	Hash crypto.Hash
	// CreationTime is when the signature was created
	CreationTime time.Time
	// HeaderOnly is true for signatures that only cover the general RPM header,
	// and false for signatures that cover the general header plus the payload
	HeaderOnly bool
	// KeyId is the PGP key that created the signature.
	KeyId uint64
	// KeyFingerprint is the fingerprint of the public key that created the
	// signature, if available.
	KeyFingerprint []byte
	// Validate the signature against a hash of the
	// payload. The provided Hash will be mutated by this call.
	Validate func(hash.Hash) error
}

// Verifier implements parsing and verifying PGP signatures
type Verifier interface {
	ParseSignature([]byte) (*Signature, error)
}

// Signer implements signing messages with a PGP key
type Signer interface {
	Sign(hash.Hash, crypto.Hash, time.Time) ([]byte, error)
}

var (
	ErrNoPGPSignature  = errors.New("no supported PGP signature packet found")
	ErrTrailingGarbage = errors.New("trailing garbage after PGP signature packet")
)

type KeyNotFoundError struct {
	KeyID       uint64
	Fingerprint []byte
}

func (e KeyNotFoundError) Error() string {
	if e.KeyID == 0 && len(e.Fingerprint) > 0 {
		return fmt.Sprintf("key with fingerprint %x not found", e.Fingerprint)
	}
	return fmt.Sprintf("keyid %08x not found", e.KeyID)
}
