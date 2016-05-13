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
)

type RpmHeader struct {
	sigHeader *rpmHeader
	genHeader *rpmHeader
	isSource  bool
}

func ReadHeader(f io.Reader) (*RpmHeader, error) {
	sigHeader, err := readSignatureHeader(f)
	if err != nil {
		return nil, err
	}

	sha1 := "" // need to read this from the sig header.

	genHeader, err := readHeader(f, sha1, sigHeader.isSource, false)
	if err != nil {
		return nil, err
	}

	return &RpmHeader{
		sigHeader: sigHeader,
		genHeader: genHeader,
		isSource:  sigHeader.isSource,
	}, nil
}

func readSignatureHeader(f io.Reader) (*rpmHeader, error) {
	// Read signature header
	lead := make([]byte, 96)
	s, err := f.Read(lead)
	if s != 96 {
		return nil, fmt.Errorf("short sig header, got %d bytes, expected 96", s)
	}
	if err != nil {
		return nil, err
	}

	// Check file magic
	magic := binary.BigEndian.Uint32(lead[0:4])
	if magic&0xffffffff != 0xedabeedb {
		return nil, fmt.Errorf("file is not an RPM")
	}

	// Check source flag
	isSource := binary.BigEndian.Uint16(lead[6:8]) == 1

	// Return signature header
	return readHeader(f, "", isSource, true)
}

func (hdr *RpmHeader) Get(tag int) (interface{}, error) {
	h, t := hdr.getHeader(tag)
	return h.Get(t)
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
