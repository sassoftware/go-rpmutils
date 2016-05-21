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
	"fmt"
	"io"
	"strconv"
)

// reference http://people.freebsd.org/~kientzle/libarchive/man/cpio.5.txt

const cpio_newc_header_length = 110

const cpio_newc_magic = "070701"

type cpio_newc_header struct {
	c_magic     string
	c_ino       int
	c_mode      int
	c_uid       int
	c_gid       int
	c_nlink     int
	c_mtime     int
	c_filesize  int
	c_devmajor  int
	c_devminor  int
	c_rdevmajor int
	c_rdevminor int
	c_namesize  int
	c_check     int
}

type binaryReader struct {
	r io.Reader
}

func (br *binaryReader) Read(buf interface{}) error {
	return binary.Read(br.r, binary.BigEndian, buf)
}

func (br *binaryReader) Read16(buf *int) error {
	b := make([]byte, 8)
	if err := br.Read(&b); err != nil {
		return err
	}
	i, err := strconv.ParseInt(string(b), 16, 0)
	if err != nil {
		return err
	}
	*buf = int(i)
	return nil
}

func readHeader(r io.Reader) (*cpio_newc_header, error) {
	hdr := cpio_newc_header{}
	br := binaryReader{r: r}

	magic := make([]byte, 6)
	if _, err := r.Read(magic); err != nil {
		return nil, err
	}
	if string(magic) != cpio_newc_magic {
		return nil, fmt.Errorf("bad magic")
	}
	hdr.c_magic = cpio_newc_magic

	if err := br.Read16(&hdr.c_ino); err != nil {
		return nil, err
	}
	if err := br.Read16(&hdr.c_mode); err != nil {
		return nil, err
	}
	if err := br.Read16(&hdr.c_uid); err != nil {
		return nil, err
	}
	if err := br.Read16(&hdr.c_gid); err != nil {
		return nil, err
	}
	if err := br.Read16(&hdr.c_nlink); err != nil {
		return nil, err
	}
	if err := br.Read16(&hdr.c_mtime); err != nil {
		return nil, err
	}
	if err := br.Read16(&hdr.c_filesize); err != nil {
		return nil, err
	}
	if err := br.Read16(&hdr.c_devmajor); err != nil {
		return nil, err
	}
	if err := br.Read16(&hdr.c_devminor); err != nil {
		return nil, err
	}
	if err := br.Read16(&hdr.c_rdevmajor); err != nil {
		return nil, err
	}
	if err := br.Read16(&hdr.c_rdevminor); err != nil {
		return nil, err
	}
	if err := br.Read16(&hdr.c_namesize); err != nil {
		return nil, err
	}
	if err := br.Read16(&hdr.c_check); err != nil {
		return nil, err
	}
	dumpHeader(&hdr)

	return &hdr, nil
}

func dumpHeader(hdr *cpio_newc_header) {
	log.Debugf("header %v", hdr)
}
