package rpmutils

import (
	"io"
	"os"

	"github.com/sassoftware/go-rpmutils/stdpgp"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/packet"
)

// Verify the PGP signature over a RPM file. knownKeys should enumerate public
// keys to check against, otherwise the signature validity cannot be verified.
// If knownKeys is nil then digests will be checked but only the raw key ID will
// be available.
//
// Deprecated: Use VerifyStream and supply either a pmpgp.Verifier or stdpgp.Verifier instead.
func Verify(stream io.Reader, knownKeys openpgp.EntityList) (header *RpmHeader, sigs []*Signature, err error) {
	verifier := stdpgp.Verifier{EntityList: knownKeys}
	if knownKeys == nil {
		verifier.ParseOnly = true
	}
	return VerifyStream(stream, verifier)
}

// SignRpmStream reads an RPM and signs it, returning the set of headers updated with the new signature.
//
// Deprecated: Use SignStream and supply either a pmpgp.Signer or stdpgp.Signer instead.
func SignRpmStream(stream io.Reader, key *packet.PrivateKey, opts *SignatureOptions) (header *RpmHeader, err error) {
	signer := stdpgp.Signer{PrivateKey: key}
	return SignStream(stream, signer, opts)
}

// SignRpmFile signs infile and writes it to outpath, which may be the same file
//
// Deprecated: Use SignFile and supply either a pmpgp.Signer or stdpgp.Signer instead.
func SignRpmFile(infile *os.File, outpath string, key *packet.PrivateKey, opts *SignatureOptions) (header *RpmHeader, err error) {
	header, err = SignRpmStream(infile, key, opts)
	if err != nil {
		return
	}
	return header, rewriteRpm(infile, outpath, header)
}
