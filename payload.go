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
	"errors"
	"fmt"
	"io"

	"github.com/sassoftware/go-rpmutils/cpio"
)

type PayloadReader interface {
	Next() (FileInfo, error)
	Read([]byte) (int, error)
	IsLink() bool
}

type payloadReader struct {
	cr      *cpio.Reader
	files   []*fileInfo
	fileMap map[string]int
	isLink  []bool
	index   int
}

func newPayloadReader(r io.Reader, files []FileInfo) *payloadReader {
	pr := &payloadReader{
		files:   make([]*fileInfo, len(files)),
		fileMap: make(map[string]int, len(files)),
		isLink:  make([]bool, len(files)),
	}
	fileSizes := make([]int64, len(files))
	var lastInode uint64
	for i, info := range files {
		fileSt := info.(*fileInfo)
		pr.files[i] = fileSt
		pr.fileMap[fileSt.name] = i
		switch fileSt.fileType() {
		case cpio.S_ISREG:
			fileSizes[i] = fileSt.Size()
			// all but the last file in a link group will have no contents. flag
			// them so we don't try to read the nonexistent payload.
			ino := fileSt.inode64()
			if ino == lastInode && ino != 0 {
				pr.isLink[i-1] = true
				fileSizes[i-1] = 0
			}
			lastInode = ino
		case cpio.S_ISLNK:
			fileSizes[i] = int64(len(fileSt.linkName))
		}
	}
	pr.cr = cpio.NewReaderWithSizes(r, fileSizes)
	return pr
}

func (pr *payloadReader) Next() (FileInfo, error) {
	hdr, err := pr.cr.Next()
	if err != nil {
		return nil, err
	}
	var index int
	if hdr.IsStripped() {
		index = hdr.Index()
	} else {
		var ok bool
		name := hdr.Filename()
		if len(name) > 1 && name[0] == '.' && name[1] == '/' {
			name = name[1:]
		}
		index, ok = pr.fileMap[name]
		if !ok {
			return nil, fmt.Errorf("invalid file \"%s\" in payload", name)
		}
	}
	if index >= len(pr.files) {
		return nil, errors.New("invalid file index")
	}
	pr.index = index
	return pr.files[index], nil
}

func (pr *payloadReader) Read(d []byte) (int, error) {
	return pr.cr.Read(d)
}

func (pr *payloadReader) IsLink() bool {
	return pr.isLink[pr.index]
}
