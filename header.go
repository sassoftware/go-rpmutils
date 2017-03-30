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
	"strings"
)

const introMagic = 0x8eade801

type entry struct {
	dataType, count int32
	contents        []byte
}

type rpmHeader struct {
	entries  map[int]entry
	isSource bool
	origSize int
}

type headerIntro struct {
	Magic, Reserved, Entries, Size uint32
}

type headerTag struct {
	Tag, DataType, Offset, Count int32
}

var typeAlign = map[int32]int{
	RPM_INT16_TYPE: 2,
	RPM_INT32_TYPE: 4,
	RPM_INT64_TYPE: 8,
}

var typeSizes = map[int32]int{
	RPM_NULL_TYPE:  0,
	RPM_CHAR_TYPE:  1,
	RPM_INT8_TYPE:  1,
	RPM_INT16_TYPE: 2,
	RPM_INT32_TYPE: 4,
	RPM_INT64_TYPE: 8,
	RPM_BIN_TYPE:   1,
}

func readExact(f io.Reader, n int) ([]byte, error) {
	buf := make([]byte, n)
	_, err := io.ReadFull(f, buf)
	return buf, err
}

func readHeader(f io.Reader, hash string, isSource bool, sigBlock bool) (*rpmHeader, error) {
	var intro headerIntro
	if err := binary.Read(f, binary.BigEndian, &intro); err != nil {
		return nil, fmt.Errorf("error reading RPM header: %s", err.Error())
	}
	if intro.Magic != introMagic {
		return nil, fmt.Errorf("bad magic for header")
	}
	entryTable, err := readExact(f, int(intro.Entries*16))
	if err != nil {
		return nil, fmt.Errorf("error reading RPM header table: %s", err.Error())
	}

	size := intro.Size
	if sigBlock {
		// signature block is padded to 8 byte alignment
		size = (size + 7) / 8 * 8
	}
	data, err := readExact(f, int(size))
	if err != nil {
		return nil, fmt.Errorf("error reading RPM header data: %s", err.Error())
	}

	// Check sha1 if it was specified
	if len(hash) > 1 {
		h := sha1.New()
		binary.Write(h, binary.BigEndian, &intro)
		h.Write(entryTable)
		h.Write(data)
		if fmt.Sprintf("%x", h.Sum(nil)) != hash {
			return nil, fmt.Errorf("bad header sha1")
		}
	}

	ents := make(map[int]entry)
	buf := bytes.NewReader(entryTable)
	for i := 0; i < int(intro.Entries); i++ {
		var tag headerTag
		if err := binary.Read(buf, binary.BigEndian, &tag); err != nil {
			return nil, err
		}
		typeSize, ok := typeSizes[tag.DataType]
		var end int
		if ok {
			end = int(tag.Offset) + typeSize*int(tag.Count)
		} else {
			// String types are null-terminated
			end = int(tag.Offset)
			for i := 0; i < int(tag.Count); i++ {
				next := bytes.IndexByte(data[end:], 0)
				if next < 0 {
					return nil, fmt.Errorf("tag %d is truncated", tag.Tag)
				}
				end += next + 1
			}
		}
		ents[int(tag.Tag)] = entry{
			dataType: tag.DataType,
			count:    tag.Count,
			contents: data[tag.Offset:end],
		}
	}

	return &rpmHeader{
		entries:  ents,
		isSource: isSource,
		origSize: 16 + len(entryTable) + len(data),
	}, nil
}

func (hdr *rpmHeader) HasTag(tag int) bool {
	_, ok := hdr.entries[tag]
	return ok
}

func (hdr *rpmHeader) Get(tag int) (interface{}, error) {
	ent, ok := hdr.entries[tag]
	if !ok && tag == OLDFILENAMES {
		return hdr.GetStrings(tag)
	}
	if !ok {
		return nil, NewNoSuchTagError(tag)
	}
	switch ent.dataType {
	case RPM_STRING_TYPE, RPM_STRING_ARRAY_TYPE, RPM_I18NSTRING_TYPE:
		return hdr.GetStrings(tag)
	case RPM_INT8_TYPE, RPM_INT16_TYPE, RPM_INT32_TYPE:
		return hdr.GetInts(tag)
	case RPM_CHAR_TYPE, RPM_BIN_TYPE:
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
		return nil, NewNoSuchTagError(tag)
	}
	if ent.dataType != RPM_STRING_TYPE && ent.dataType != RPM_STRING_ARRAY_TYPE && ent.dataType != RPM_I18NSTRING_TYPE {
		return nil, fmt.Errorf("unsupported datatype for string: %d, tag: %d", ent.dataType, tag)
	}
	strs := strings.Split(string(ent.contents), "\x00")
	return strs[:ent.count], nil
}

func (hdr *rpmHeader) GetInts(tag int) ([]int, error) {
	ent, ok := hdr.entries[tag]
	if !ok {
		return nil, NewNoSuchTagError(tag)
	}
	out := make([]int, ent.count)
	content := ent.contents
	for i := int32(0); i < ent.count; i++ {
		switch ent.dataType {
		case RPM_INT8_TYPE:
			out[i] = int(content[0])
			content = content[1:]
		case RPM_INT16_TYPE:
			out[i] = int(binary.BigEndian.Uint16(content[:2]))
			content = content[2:]
		case RPM_INT32_TYPE:
			out[i] = int(binary.BigEndian.Uint32(content[:4]))
			content = content[4:]
		}
	}
	return out, nil
}

func (hdr *rpmHeader) GetBytes(tag int) ([]byte, error) {
	ent, ok := hdr.entries[tag]
	if !ok {
		return nil, NewNoSuchTagError(tag)
	}
	if ent.dataType != RPM_CHAR_TYPE && ent.dataType != RPM_BIN_TYPE {
		return nil, fmt.Errorf("unsupported datatype for bytes: %d, tag: %d", ent.dataType, tag)
	}
	return ent.contents, nil
}

func (hdr *rpmHeader) GetNEVRA() (*NEVRA, error) {
	name, err := hdr.GetStrings(NAME)
	if err != nil {
		return nil, err
	}
	epoch, err := hdr.GetStrings(EPOCH)
	// Special case epoch, if it doesn't exist, it should be 0 or None
	if epoch == nil {
		switch err.(type) {
		case NoSuchTagError:
			epoch = []string{"0"}
		default:
			return nil, err
		}
	}
	version, err := hdr.GetStrings(VERSION)
	if err != nil {
		return nil, err
	}
	release, err := hdr.GetStrings(RELEASE)
	if err != nil {
		return nil, err
	}
	arch, err := hdr.GetStrings(ARCH)
	if err != nil {
		return nil, err
	}
	return &NEVRA{
		Name:    name[0],
		Epoch:   epoch[0],
		Version: version[0],
		Release: release[0],
		Arch:    arch[0],
	}, nil
}

func (hdr *rpmHeader) GetFiles() ([]FileInfo, error) {
	paths, err := hdr.GetStrings(OLDFILENAMES)
	if err != nil {
		return nil, err
	}
	fileSizes, err := hdr.GetInts(FILESIZES)
	if err != nil {
		return nil, err
	}
	fileUserName, err := hdr.GetStrings(FILEUSERNAME)
	if err != nil {
		return nil, err
	}
	fileGroupName, err := hdr.GetStrings(FILEGROUPNAME)
	if err != nil {
		return nil, err
	}
	fileFlags, err := hdr.GetInts(FILEFLAGS)
	if err != nil {
		return nil, err
	}
	fileMtimes, err := hdr.GetInts(FILEMTIMES)
	if err != nil {
		return nil, err
	}
	fileDigests, err := hdr.GetStrings(FILEDIGESTS)
	if err != nil {
		return nil, err
	}
	fileModes, err := hdr.GetInts(FILEMODES)
	if err != nil {
		return nil, err
	}
	linkTos, err := hdr.GetStrings(FILELINKTOS)
	if err != nil {
		return nil, err
	}

	files := make([]FileInfo, len(paths))
	for i := 0; i < len(paths); i++ {
		files[i] = &fileInfo{
			name:      paths[i],
			size:      int64(fileSizes[i]),
			userName:  fileUserName[i],
			groupName: fileGroupName[i],
			flags:     fileFlags[i],
			mtime:     fileMtimes[i],
			digest:    fileDigests[i],
			mode:      fileModes[i],
			linkName:  linkTos[i],
		}
	}

	return files, nil
}
