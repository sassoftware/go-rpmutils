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
	"fmt"
	"io"
)

const TRAILER = "TRAILER!!!"

type CpioEntry struct {
	header   *cpio_newc_header
	filename string
	payload  *file_stream
}

type CpioStream struct {
	stream   io.ReadSeeker
	curr_pos int64
	next_pos int64
}

func NewCpioStream(stream io.ReadSeeker) *CpioStream {
	return &CpioStream{
		stream:   stream,
		curr_pos: 0,
		next_pos: 0,
	}
}

func (cs *CpioStream) ReadNextEntry() (*CpioEntry, error) {
	if cs.next_pos != cs.curr_pos {
		n, err := cs.stream.Seek(cs.next_pos-cs.curr_pos, 1)
		cs.curr_pos += n
		if err != nil {
			return nil, err
		}
	}

	// Read header
	hdr, err := readHeader(cs.stream)
	if err != nil {
		return nil, err
	}

	// Read filename
	buf := make([]byte, hdr.c_namesize)
	n, err := cs.stream.Read(buf)
	if err != nil {
		return nil, err
	}
	if n != len(buf) {
		return nil, fmt.Errorf("short read")
	}

	filename := string(buf[:len(buf)-1])

	offset := pad(cpio_newc_header_length+int(hdr.c_namesize)) - cpio_newc_header_length - int(hdr.c_namesize)

	if offset > 0 {
		_, err := cs.stream.Seek(int64(offset), 1)
		if err != nil {
			return nil, err
		}
	}

	// Find the next entry
	cs.next_pos = pad64(cs.curr_pos + int64(hdr.c_filesize))

	// Find the payload
	payload, err := newFileStream(cs.stream, int64(hdr.c_filesize))
	if err != nil {
		return nil, err
	}

	// Create then entry
	entry := CpioEntry{
		header:   hdr,
		filename: filename,
		payload:  payload,
	}

	return &entry, nil
}

func pad(num int) int {
	return num + 3 - (num+3)%4
}

func pad64(num int64) int64 {
	return num + 3 - (num+3)%4
}
