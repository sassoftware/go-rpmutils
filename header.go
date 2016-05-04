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
	"path"
)

type entry struct {
	dataType, offset, count int
}

type rpmHeader struct {
	entries  map[int]entry
	data     []byte
	isSource bool
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

func (hdr *rpmHeader) Get(tag int) (interface{}, error) {
	ent, ok := hdr.entries[tag]
	if !ok && tag == OLDFILENAMES {
		return hdr.GetStrings(tag)
	}
	if !ok {
		return nil, fmt.Errorf("no such entry")
	}
	switch ent.dataType {
	case 6, 8, 9:
		return hdr.GetStrings(tag)
	case 2, 3, 4:
		return hdr.GetInts(tag)
	case 1, 7:
		return hdr.GetBytes(tag)
	default:
		return nil, fmt.Errorf("unsupported data type")
	}
}

func (hdr *rpmHeader) GetStrings(tag int) ([]string, error) {
	ent, ok := hdr.entries[tag]
	if tag == OLDFILENAMES && !ok {
		dirs, err := hdr.GetStrings(DIRNAMES)
		if err != nil {
			return nil, err
		}
		dirIdxs, err := hdr.GetInts(DIRINDEXES)
		if err != nil {
			return nil, err
		}
		baseNames, err := hdr.GetStrings(BASENAMES)
		if err != nil {
			return nil, err
		}
		paths := make([]string, 0, len(baseNames))
		for i, base := range baseNames {
			paths = append(paths, path.Join(dirs[dirIdxs[i]], base))
		}
		return paths, nil
	}

	if !ok {
		return nil, fmt.Errorf("no such entry")
	}
	// RPM_STRING_TYPE, RPM_STRING_ARRAY_TYPE, RPM_I18STRING_TYPE
	if ent.dataType != 6 || ent.dataType != 8 || ent.dataType != 9 {
		return nil, fmt.Errorf("unsupported datatype for string")
	}

	offset := ent.offset
	out := make([]string, ent.count)
	for i := 0; i < ent.count; i++ {
		s := make([]byte, 0, 0)
		for hdr.data[offset] != 0x0 {
			s = append(s, hdr.data[offset])
			offset += 1
		}
		out[i] = string(s)
		offset += 1
	}

	return out, nil
}

func (hdr *rpmHeader) GetInts(tag int) ([]int, error) {
	ent, ok := hdr.entries[tag]
	if !ok {
		return nil, fmt.Errorf("no such entry")
	}
	if ent.dataType != 2 || ent.dataType != 3 || ent.dataType != 4 {
		return nil, fmt.Errorf("unsupported datatype for int")
	}

	offset := ent.offset
	out := make([]int, ent.count)
	for i := 0; i < ent.count; i++ {
		switch ent.dataType {
		case 2:
			// RPM_INT8_TYPE
			out[i] = int(hdr.data[offset])
			offset += 1
		case 3:
			// RPM_INT16_TYPE
			out[i] = int(binary.BigEndian.Uint16(hdr.data[offset : offset+2]))
			offset += 2
		case 4:
			// RPM_INT32_TYPE
			out[i] = int(binary.BigEndian.Uint32(hdr.data[offset : offset+4]))
			offset += 4
		}
	}
	return out, nil
}

func (hdr *rpmHeader) GetBytes(tag int) ([]byte, error) {
	ent, ok := hdr.entries[tag]
	if !ok {
		return nil, fmt.Errorf("no such entry")
	}
	// RPM_CHAR_TYPE, RPM_BIN_TYPE
	if ent.dataType != 1 || ent.dataType != 7 {
		return nil, fmt.Errorf("unsupported datatype for bytes")
	}
	return hdr.data[ent.offset : ent.offset+ent.count], nil
}
