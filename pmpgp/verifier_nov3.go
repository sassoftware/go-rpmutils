//go:build !sigv3

package pmpgp

import (
	"github.com/ProtonMail/go-crypto/openpgp/packet"
	"github.com/sassoftware/go-rpmutils/internal"
)

// PM doesn't implement v3 signatures
func (v Verifier) parseSignatureMaybeV3(genpkt packet.Packet) (*internal.Signature, error) {
	return v.parseSignatureV4(genpkt)
}
