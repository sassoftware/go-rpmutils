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
	Header  *Cpio_newc_header
	payload *file_stream
}

type CpioStream struct {
	stream   *countingReader
	next_pos int64
}

type countingReader struct {
	stream   io.Reader
	curr_pos int64
}

func NewCpioStream(stream io.Reader) *CpioStream {
	return &CpioStream{
		stream: &countingReader{
			stream:   stream,
			curr_pos: 0,
		},
		next_pos: 0,
	}
}

func (cs *CpioStream) ReadNextEntry() (*CpioEntry, error) {
	if cs.next_pos != cs.stream.curr_pos {
		log.Debugf("seeking %d, curr_pos: %d, next_pos: %d", cs.next_pos-cs.stream.curr_pos, cs.stream.curr_pos, cs.next_pos)
		_, err := cs.stream.Seek(cs.next_pos-cs.stream.curr_pos, 1)
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
		log.Errorf("short read, got %d, expected %d", n, len(buf))
		log.Debugf("namesize: %d", hdr.c_namesize)
		return nil, fmt.Errorf("short read")
	}

	filename := string(buf[:len(buf)-1])
	log.Debugf("filename: %s", filename)

	offset := pad(cpio_newc_header_length+int(hdr.c_namesize)) - cpio_newc_header_length - int(hdr.c_namesize)

	if offset > 0 {
		_, err := cs.stream.Seek(int64(offset), 1)
		if err != nil {
			return nil, err
		}
	}

	// Find the next entry
	cs.next_pos = pad64(cs.stream.curr_pos + int64(hdr.c_filesize))

	// Find the payload
	payload, err := newFileStream(cs.stream, int64(hdr.c_filesize))
	if err != nil {
		return nil, err
	}

	// Create then entry
	hdr.filename = filename
	entry := CpioEntry{
		Header:  hdr,
		payload: payload,
	}

	return &entry, nil
}

func (cr *countingReader) Read(p []byte) (n int, err error) {
	n, err = cr.stream.Read(p)
	cr.curr_pos += int64(n)
	return
}

func (cr *countingReader) Seek(offset int64, whence int) (int64, error) {
	if whence != 1 {
		return 0, fmt.Errorf("only seeking from current location supported")
	}
	if offset == 0 {
		return cr.curr_pos, nil
	}
	log.Debugf("offset: %d, curr_pos: %d", offset, cr.curr_pos)
	b := make([]byte, offset)
	n, err := cr.Read(b)
	if err != nil && err != io.EOF {
		return 0, err
	}
	if int64(n) != offset {
		return int64(n), fmt.Errorf("short seek")
	}
	return int64(n), nil
}

func pad(num int) int {
	return num + 3 - (num+3)%4
}

func pad64(num int64) int64 {
	return num + 3 - (num+3)%4
}
