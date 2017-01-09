/*
 * Copyright (c) SAS Institute, Inc.
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
	"encoding/binary"
	"fmt"
	"io"

	"github.com/sassoftware/go-rpmutils/cpio"
)

type Rpm struct {
	Header *RpmHeader
	f      io.Reader
}

type RpmHeader struct {
	lead      []byte
	sigHeader *rpmHeader
	genHeader *rpmHeader
	isSource  bool
}

func ReadRpm(f io.Reader) (*Rpm, error) {
	hdr, err := ReadHeader(f)
	if err != nil {
		return nil, err
	}
	return &Rpm{
		Header: hdr,
		f:      f,
	}, nil
}

func (rpm *Rpm) ExpandPayload(dest string) error {
	pld, err := uncompressRpmPayloadReader(rpm.f, rpm.Header)
	if err != nil {
		return err
	}
	return cpio.Extract(pld, dest)
}

func (rpm *Rpm) PayloadReader() (*cpio.Reader, error) {
	pld, err := uncompressRpmPayloadReader(rpm.f, rpm.Header)
	if err != nil {
		return nil, err
	}
	return cpio.NewReader(pld), nil
}

func ReadHeader(f io.Reader) (*RpmHeader, error) {
	lead, sigHeader, err := readSignatureHeader(f)
	if err != nil {
		return nil, err
	}

	sha1 := "" // need to read this from the sig header.

	genHeader, err := readHeader(f, sha1, sigHeader.isSource, false)
	if err != nil {
		return nil, err
	}

	return &RpmHeader{
		lead:      lead,
		sigHeader: sigHeader,
		genHeader: genHeader,
		isSource:  sigHeader.isSource,
	}, nil
}

func readSignatureHeader(f io.Reader) ([]byte, *rpmHeader, error) {
	// Read signature header
	lead, err := readExact(f, 96)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading RPM lead: %s", err.Error())
	}

	// Check file magic
	magic := binary.BigEndian.Uint32(lead[0:4])
	if magic&0xffffffff != 0xedabeedb {
		return nil, nil, fmt.Errorf("file is not an RPM")
	}

	// Check source flag
	isSource := binary.BigEndian.Uint16(lead[6:8]) == 1

	// Return signature header
	hdr, err := readHeader(f, "", isSource, true)
	return lead, hdr, err
}

func (hdr *RpmHeader) HasTag(tag int) bool {
	h, t := hdr.getHeader(tag)
	return h.HasTag(t)
}

func (hdr *RpmHeader) Get(tag int) (interface{}, error) {
	h, t := hdr.getHeader(tag)
	return h.Get(t)
}

func (hdr *RpmHeader) GetString(tag int) (string, error) {
	vals, err := hdr.GetStrings(tag)
	if err != nil {
		return "", err
	}
	if len(vals) != 1 {
		return "", fmt.Errorf("incorrect number of values")
	}
	return vals[0], nil
}

func (hdr *RpmHeader) GetStrings(tag int) ([]string, error) {
	h, t := hdr.getHeader(tag)
	return h.GetStrings(t)
}

func (hdr *RpmHeader) GetInts(tag int) ([]int, error) {
	h, t := hdr.getHeader(tag)
	return h.GetInts(t)
}

func (hdr *RpmHeader) GetBytes(tag int) ([]byte, error) {
	h, t := hdr.getHeader(tag)
	return h.GetBytes(t)
}

func (hdr *RpmHeader) getHeader(tag int) (*rpmHeader, int) {
	if tag > _SIGHEADER_TAG_BASE {
		return hdr.sigHeader, tag - _SIGHEADER_TAG_BASE
	}
	if tag < _GENERAL_TAG_BASE {
		return hdr.sigHeader, tag
	}
	return hdr.genHeader, tag
}

func (hdr *RpmHeader) GetNEVRA() (*NEVRA, error) {
	return hdr.genHeader.GetNEVRA()
}

func (hdr *RpmHeader) GetFiles() ([]FileInfo, error) {
	return hdr.genHeader.GetFiles()
}
