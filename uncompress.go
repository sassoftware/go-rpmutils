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
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"

	"github.com/DataDog/zstd"
	"github.com/xi2/xz"
)

// Wrap RPM payload with uncompress reader, assumes that header has
// already been read.
func uncompressRpmPayloadReader(r io.Reader, hdr *RpmHeader) (io.Reader, error) {
	// Check to make sure payload format is a cpio archive. If the tag does
	// not exist, assume archive is cpio.
	if hdr.HasTag(PAYLOADFORMAT) {
		val, err := hdr.GetString(PAYLOADFORMAT)
		if err != nil {
			return nil, err
		}
		if val != "cpio" {
			return nil, fmt.Errorf("Unknown payload format %s", val)
		}
	}

	// Check to see how the payload was compressed. If the tag does not
	// exist, check if it is gzip, if not it is uncompressed.
	var compression string
	if hdr.HasTag(PAYLOADCOMPRESSOR) {
		val, err := hdr.GetString(PAYLOADCOMPRESSOR)
		if err != nil {
			return nil, err
		}
		compression = val
	} else {
		b := make([]byte, 4096)
		_, err := r.Read(b)
		if err != nil {
			return nil, err
		}
		if len(b) > 2 && b[0] == 0x1f && b[1] == 0x8b {
			compression = "gzip"
		} else {
			compression = "uncompressed"
		}
	}

	switch compression {
	case "gzip":
		return gzip.NewReader(r)
	case "bzip2":
		return bzip2.NewReader(r), nil
	case "lzma", "xz":
		return xz.NewReader(r, 0)
	case "zstd":
		return zstd.NewReader(r), nil
	case "uncompressed":
		return r, nil
	default:
		return nil, fmt.Errorf("Unknown compression type %s", compression)
	}
}
