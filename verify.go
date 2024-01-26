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
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
)

var headerSigTags = []int{SIG_RSA, SIG_DSA}
var payloadSigTags = []int{
	SIG_PGP - _SIGHEADER_TAG_BASE,
	SIG_GPG - _SIGHEADER_TAG_BASE,
}

// VerifyStream verifies a RPM file stream using the supplied PGP implementation.
func VerifyStream(stream io.Reader, verifier Verifier) (*RpmHeader, []*Signature, error) {
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
	sigs, hashes, err := digestAndVerify(sigHeader, genHeader, stream, verifier)
	if err != nil {
		return nil, nil, err
	}
	// verify PGP signatures
	for i, sig := range sigs {
		h := hashes[i]
		if err := sig.Validate(h); err != nil {
			return nil, nil, err
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
func setupDigester(sigHeader *rpmHeader, tag int, verifier Verifier) (*Signature, error) {
	blob, err := sigHeader.GetBytes(tag)
	if _, ok := err.(NoSuchTagError); ok {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return verifier.ParseSignature(blob)
}

// Parse signatures from the header and determine which hash functions are
// needed to digest the RPM. The caller must write the payload to the returned
// WriteCloser, then call Close to check if the payload digest matches.
func digestAndVerify(sigHeader, genHeader *rpmHeader, payloadReader io.Reader, verifier Verifier) ([]*Signature, []hash.Hash, error) {
	var sigs []*Signature
	var hashes []hash.Hash
	// signatures over the general header alone
	for _, tag := range headerSigTags {
		sig, err := setupDigester(sigHeader, tag, verifier)
		if err != nil {
			return nil, nil, err
		} else if sig == nil {
			continue
		}
		sig.HeaderOnly = true
		h := sig.Hash.New()
		h.Write(genHeader.orig)
		sigs = append(sigs, sig)
		hashes = append(hashes, h)
	}
	// signatures over the general header + payload
	var payloadWriters []io.Writer
	for _, tag := range payloadSigTags {
		sig, err := setupDigester(sigHeader, tag, verifier)
		if err != nil {
			return nil, nil, err
		} else if sig == nil {
			continue
		}
		sig.HeaderOnly = false
		h := sig.Hash.New()
		h.Write(genHeader.orig)
		payloadWriters = append(payloadWriters, h)
		sigs = append(sigs, sig)
		hashes = append(hashes, h)
	}
	err := digestPayload(sigHeader, genHeader, payloadReader, payloadWriters)
	return sigs, hashes, err
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
