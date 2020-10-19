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
	"archive/tar"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestTar(t *testing.T) {
	f, err := os.Open("../testdata/foo.cpio")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	tmpdir, err := ioutil.TempDir("", "cpio")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)
	tarWriter, err := os.Create(filepath.Join(tmpdir, "test.tar"))
	if err != nil {
		t.Fatal(err)
	}
	defer tarWriter.Close()
	if err := Tar(f, tar.NewWriter(tarWriter)); err != nil {
		t.Fatal(err)
	}
	tarWriter.Close()

	tarReader, err := os.Open(filepath.Join(tmpdir, "test.tar"))
	if err != nil {
		t.Fatal(err)
	}
	tarBall := tar.NewReader(tarReader)
	headers := map[string]*tar.Header{
		"./config": {Name: "./config", Size: 7, Mode: 33188},
		"./dir":    {Name: "./dir", Size: 0, Mode: 16877},
		"./normal": {Name: "./normal", Size: 7, Mode: 33188},
	}
	for {
		header, err := tarBall.Next()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				t.Fatal(err)
			}
		}
		fmt.Println(header)
		if headers[header.Name].Name != header.Name {
			t.Fatal(fmt.Sprintf("Expected name %v, got %v", headers[header.Name].Name, header.Name))
		}
		if headers[header.Name].Size != header.Size {
			t.Fatal(fmt.Sprintf("Expected size %v, got %v", headers[header.Name].Size, header.Size))
		}
		if headers[header.Name].Mode != header.Mode {
			t.Fatal(fmt.Sprintf("Expected mode %v, got %v", headers[header.Name].Mode, header.Mode))
		}
		if headers[header.Name].Uid != header.Uid {
			t.Fatal(fmt.Sprintf("Expected uid %v, got %v", headers[header.Name].Uid, header.Uid))
		}
		if headers[header.Name].Gid != header.Gid {
			t.Fatal(fmt.Sprintf("Expected gid %v, got %v", headers[header.Name].Gid, header.Gid))
		}
	}
}
