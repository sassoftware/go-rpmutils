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
	"os"
	"testing"
	"testing/iotest"
)

func TestReadHeader(t *testing.T) {
	f, err := os.Open("../testdata/foo.cpio")
	if err != nil {
		t.Fatal(err)
	}

	hdr, err := readHeader(iotest.HalfReader(f))
	if err != nil {
		t.Fatal(err)
	}

	if hdr.c_magic != cpio_newc_magic {
		t.Fatal("bad magic")
	}
	if hdr.c_ino != 512785 {
		t.Fatal("bad inode")
	}
	if hdr.c_mode != 33188 {
		t.Fatal("bad mode")
	}
	if hdr.c_uid != 0 {
		t.Fatal("incorrect uid")
	}
	if hdr.c_gid != 0 {
		t.Fatal("incorrect gid")
	}
	if hdr.c_nlink != 1 {
		t.Fatal("incorrect nlink")
	}
	if hdr.c_mtime != 1263588698 {
		t.Fatal("incorrect mtime")
	}
	if hdr.c_filesize != 7 {
		t.Fatal("incorrect filesize")
	}
	if hdr.c_devmajor != 8 {
		t.Fatal("incorrect devmajor")
	}
	if hdr.c_devminor != 6 {
		t.Fatal("incorrect devminor")
	}
	if hdr.c_rdevmajor != 0 {
		t.Fatal("incorrect rdevmajor")
	}
	if hdr.c_rdevminor != 0 {
		t.Fatal("incorrect rdevminor")
	}
	if hdr.c_namesize != 9 {
		t.Fatal("incorrect namesize")
	}
	if hdr.c_check != 0 {
		t.Fatal("incorrect check")
	}
}
