/*
 * Copyright (c) SAS Institute Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package rpmutils

import (
	"bytes"
	"crypto"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"time"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/packet"
)

var headerSigTags = []int{SIG_RSA, SIG_DSA}
var payloadSigTags = []int{
	SIG_PGP - _SIGHEADER_TAG_BASE,
	SIG_GPG - _SIGHEADER_TAG_BASE,
}

// Signature describes a PGP signature found within a RPM while verifying it.
type Signature struct {
	// Signer is the PGP identity that created the signature. It may be nil if
	// the public key is not available at verification time, but KeyId will
	// always be set.
	Signer *openpgp.Entity
	// Hash is the algorithm used to digest the signature contents
	Hash crypto.Hash
	// CreationTime is when the signature was created
	CreationTime time.Time
	// HeaderOnly is true for signatures that only cover the general RPM header,
	// and false for signatures that cover the general header plus the payload
	HeaderOnly bool
	// KeyId is the PGP key that created the signature.
	KeyId uint64

	packet packet.Packet
	hash   hash.Hash
}

// Verify the PGP signature over a RPM file. knownKeys should enumerate public
// keys to check against, otherwise the signature validity cannot be verified.
// If knownKeys is nil then digests will be checked but only the raw key ID will
// be available.
func Verify(stream io.Reader, knownKeys openpgp.EntityList) (header *RpmHeader, sigs []*Signature, err error) {
	lead, sigHeader, err := readSignatureHeader(stream)
	if err != nil {
		return nil, nil, err
	}
	// parse the general header
	headerDigestValue, headerDigestType := getHashAndType(sigHeader)
	genHeader, err := readHeader(stream, headerDigestValue, headerDigestType, sigHeader.isSource, false)
	if err != nil {
		return nil, nil, err
	}
	// setup digesters for PGP and payload digest
	sigs, err = digestAndVerify(sigHeader, genHeader, stream)
	if err != nil {
		return nil, nil, err
	}
	// verify PGP signatures
	if knownKeys != nil {
		for _, sig := range sigs {
			if err := checkSig(sig, knownKeys); err != nil {
				return nil, nil, err
			}
		}
	}
	hdr := &RpmHeader{
		lead:      lead,
		sigHeader: sigHeader,
		genHeader: genHeader,
		isSource:  sigHeader.isSource,
	}
	return hdr, sigs, nil
}

// Try to parse a PGP signature with the given tag and return its metadata and
// hash function.
func setupDigester(sigHeader *rpmHeader, tag int) (*Signature, error) {
	blob, err := sigHeader.GetBytes(tag)
	if _, ok := err.(NoSuchTagError); ok {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	packetReader := packet.NewReader(bytes.NewReader(blob))
	genpkt, err := packetReader.Next()
	if err != nil {
		return nil, err
	}
	var sig *Signature
	switch pkt := genpkt.(type) {
	case *packet.SignatureV3:
		sig = &Signature{
			Hash:         pkt.Hash,
			CreationTime: pkt.CreationTime,
			KeyId:        pkt.IssuerKeyId,
		}
	case *packet.Signature:
		if pkt.IssuerKeyId == nil {
			return nil, errors.New("missing keyId in signature")
		}
		sig = &Signature{
			Hash:         pkt.Hash,
			CreationTime: pkt.CreationTime,
			KeyId:        *pkt.IssuerKeyId,
		}
	default:
		return nil, fmt.Errorf("tag %d does not contain a PGP signature", tag)
	}
	_, err = packetReader.Next()
	if err != io.EOF {
		return nil, fmt.Errorf("trailing garbage after signature in tag %d", tag)
	}
	sig.packet = genpkt
	if !sig.Hash.Available() {
		return nil, errors.New("signature uses unknown digest")
	}
	sig.hash = sig.Hash.New()
	return sig, nil
}

// Parse signatures from the header and determine which hash functions are
// needed to digest the RPM. The caller must write the payload to the returned
// WriteCloser, then call Close to check if the payload digest matches.
func digestAndVerify(sigHeader, genHeader *rpmHeader, payloadReader io.Reader) ([]*Signature, error) {
	var sigs []*Signature
	// signatures over the general header alone
	for _, tag := range headerSigTags {
		sig, err := setupDigester(sigHeader, tag)
		if err != nil {
			return nil, err
		} else if sig == nil {
			continue
		}
		sig.HeaderOnly = true
		sig.hash.Write(genHeader.orig)
		sigs = append(sigs, sig)
	}
	// signatures over the general header + payload
	var payloadWriters []io.Writer
	for _, tag := range payloadSigTags {
		sig, err := setupDigester(sigHeader, tag)
		if err != nil {
			return nil, err
		} else if sig == nil {
			continue
		}
		sig.HeaderOnly = false
		sig.hash.Write(genHeader.orig)
		payloadWriters = append(payloadWriters, sig.hash)
		sigs = append(sigs, sig)
	}
	return sigs, digestPayload(sigHeader, genHeader, payloadReader, payloadWriters)
}

func digestPayload(sigHeader, genHeader *rpmHeader, payloadReader io.Reader, payloadWriters []io.Writer) error {
	// Also compute a digest over the payload for integrity checking purposes
	if payloadValue, payloadType := getPayloadDigest(genHeader); payloadType != 0 {
		if !payloadType.Available() {
			return fmt.Errorf("unknown payload digest %s", payloadType)
		}
		payloadHasher := payloadType.New()
		// hash payload only
		payloadWriters = append(payloadWriters, payloadHasher)
		if _, err := io.Copy(io.MultiWriter(payloadWriters...), payloadReader); err != nil {
			return err
		}
		calculated := hex.EncodeToString(payloadHasher.Sum(nil))
		if calculated != payloadValue {
			return fmt.Errorf("payload %s digest mismatch", payloadType)
		}
		return nil
	}
	// Check legacy MD5 digest in sig header as a last resort. This is the only
	// digest found in the signature header that covers the payload, so for some
	// old RPMs that don't have a payload digest in the general header this is
	// the only integrity check we can use unless we're verifying the PGP
	// signatures.
	if sigmd5, _ := sigHeader.GetBytes(SIG_MD5 - _SIGHEADER_TAG_BASE); len(sigmd5) != 0 {
		payloadHasher := md5.New()
		// hash header + payload
		payloadHasher.Write(genHeader.orig)
		payloadWriters = append(payloadWriters, payloadHasher)
		if _, err := io.Copy(io.MultiWriter(payloadWriters...), payloadReader); err != nil {
			return err
		}
		calculated := payloadHasher.Sum(nil)
		if !bytes.Equal(calculated, sigmd5) {
			return errors.New("md5 digest mismatch")
		}
		return nil
	}
	return errors.New("no usable payload digest found")
}

func checkSig(sig *Signature, knownKeys openpgp.EntityList) error {
	keys := knownKeys.KeysById(sig.KeyId)
	if keys == nil {
		return fmt.Errorf("keyid %x not found", sig.KeyId)
	}
	key := keys[0]
	sig.Signer = key.Entity
	var err error
	switch pkt := sig.packet.(type) {
	case *packet.Signature:
		err = key.PublicKey.VerifySignature(sig.hash, pkt)
	case *packet.SignatureV3:
		err = key.PublicKey.VerifySignatureV3(sig.hash, pkt)
	}
	if err != nil {
		return err
	}
	sig.packet = nil
	sig.hash = nil
	return nil
}
