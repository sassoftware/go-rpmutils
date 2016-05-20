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

package cpio

import (
	"encoding/binary"
	"io"
)

// reference http://people.freebsd.org/~kientzle/libarchive/man/cpio.5.txt

const cpio_newc_header_length = 110

var cpio_newc_magic = [6]byte{0, 7, 0, 7, 0, 1}

type cpio_newc_header struct {
	c_magic     [6]byte
	c_ino       uint16
	c_mode      uint16
	c_uid       uint16
	c_gid       uint16
	c_nlink     uint16
	c_mtime     uint16
	c_filesize  uint16
	c_devmajor  uint16
	c_devminor  uint16
	c_rdevmajor uint16
	c_rdevminor uint16
	c_namesize  uint16
	c_check     uint16
}

type binaryReader struct {
	r io.Reader
}

func (br *binaryReader) Read(buf interface{}) error {
	return binary.Read(br.r, binary.BigEndian, buf)
}

func readHeader(r io.Reader) (*cpio_newc_header, error) {
	hdr := cpio_newc_header{}
	br := binaryReader{r: r}

	if err := br.Read(&hdr.c_magic); err != nil {
		return nil, err
	}
	if err := br.Read(&hdr.c_ino); err != nil {
		return nil, err
	}
	if err := br.Read(&hdr.c_mode); err != nil {
		return nil, err
	}
	if err := br.Read(&hdr.c_uid); err != nil {
		return nil, err
	}
	if err := br.Read(&hdr.c_gid); err != nil {
		return nil, err
	}
	if err := br.Read(&hdr.c_nlink); err != nil {
		return nil, err
	}
	if err := br.Read(&hdr.c_mtime); err != nil {
		return nil, err
	}
	if err := br.Read(&hdr.c_filesize); err != nil {
		return nil, err
	}
	if err := br.Read(&hdr.c_devmajor); err != nil {
		return nil, err
	}
	if err := br.Read(&hdr.c_devminor); err != nil {
		return nil, err
	}
	if err := br.Read(&hdr.c_rdevmajor); err != nil {
		return nil, err
	}
	if err := br.Read(&hdr.c_rdevminor); err != nil {
		return nil, err
	}
	if err := br.Read(&hdr.c_namesize); err != nil {
		return nil, err
	}
	if err := br.Read(&hdr.c_check); err != nil {
		return nil, err
	}

	return &hdr, nil
}
