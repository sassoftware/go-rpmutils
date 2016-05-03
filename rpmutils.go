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
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"io"
)

type entry struct {
	dataType, offset, count int
}

type RpmHeader struct {
	sigHeader *rpmHeader
	genHeader *rpmHeader
	isSource  bool
}

type rpmHeader struct {
	entries  map[int]entry
	data     []byte
	isSource bool
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

func readHeader(f io.Reader, hash string, isSource bool, sigBlock bool) (*rpmHeader, error) {
	intro := make([]byte, 16)
	s, err := f.Read(intro)
	if s != 16 {
		return nil, fmt.Errorf("short intro, got %d bytes, expected 16", s)
	}
	if err != nil {
		return nil, err
	}

	if intro[0] != 0x8e || intro[1] != 0xad || intro[2] != 0xe8 || intro[3] != 01 {
		return nil, fmt.Errorf("bad magic for header")
	}

	//reserved := binary.BigEndian.Uint32(intro[4:8])
	entries := binary.BigEndian.Uint32(intro[8:12])
	size := binary.BigEndian.Uint32(intro[12:16])

	entryTable := make([]byte, entries*16)
	s, err = f.Read(entryTable)
	if s != int(entries)*16 {
		return nil, fmt.Errorf("short read on entry table")
	}
	if err != nil {
		return nil, err
	}

	data := make([]byte, size)
	s, err = f.Read(data)
	if s != int(size) {
		return nil, fmt.Errorf("short read on data")
	}
	if err != nil {
		return nil, err
	}

	// Check sha1 if it was specified
	if len(hash) > 1 {
		h := sha1.New()
		h.Write(intro)
		h.Write(entryTable)
		h.Write(data)
		if fmt.Sprintf("%x", h.Sum(nil)) != hash {
			return nil, fmt.Errorf("bad header sha1")
		}
	}

	ents := make(map[int]entry)
	buf := bytes.NewReader(entryTable)
	var tag, dataType, offset, count int32
	for i := 0; i < int(entries); i++ {
		if err := binary.Read(buf, binary.BigEndian, &tag); err != nil {
			return nil, err
		}
		if err := binary.Read(buf, binary.BigEndian, &dataType); err != nil {
			return nil, err
		}
		if err := binary.Read(buf, binary.BigEndian, &offset); err != nil {
			return nil, err
		}
		if err := binary.Read(buf, binary.BigEndian, &count); err != nil {
			return nil, err
		}

		ents[int(tag)] = entry{
			dataType: int(dataType),
			offset:   int(offset),
			count:    int(count),
		}
	}

	if sigBlock {
		// We need to align to an 8-byte boundary.
		// So far we read the intro (which is 16 bytes) and the entry table
		// (which is a multiple of 16 bytes). So we only have to worry about
		// the actual header data not being aligned.
		alignement := size % 8
		if alignement > 0 {
			f.Read(make([]byte, 8-alignement))
		}
	}

	return &rpmHeader{
		entries:  ents,
		data:     data,
		isSource: isSource,
	}, nil
}
