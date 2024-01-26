package rpmutils

import "github.com/sassoftware/go-rpmutils/internal"

// Signature holds a parsed PGP signature from an RPM
type Signature = internal.Signature

// Verifier implements parsing and verifying PGP signatures
type Verifier = internal.Verifier

// Signer implements signing messages with a PGP key
type Signer = internal.Signer

// KeyNotFoundError is returned when the public key to validate a signature cannot be retrieved
type KeyNotFoundError = internal.KeyNotFoundError

var (
	ErrNoPGPSignature  = internal.ErrNoPGPSignature
	ErrTrailingGarbage = internal.ErrTrailingGarbage
)

// DigestOnlyVerifier checks the unencrypted digests found in the RPM but does not attempt any PGP parsing.
type DigestOnlyVerifier struct{}

func (DigestOnlyVerifier) ParseSignature([]byte) (*Signature, error) {
	return nil, nil
}

var _ Verifier = DigestOnlyVerifier{}
