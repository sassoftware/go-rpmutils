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
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"testing/iotest"
)

func TestExtract(t *testing.T) {
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

	hf := iotest.HalfReader(f)
	if err := Extract(hf, tmpdir); err != nil {
		t.Fatal(err)
	}

	if f, err = os.Open("../testdata/foo.cpio"); err != nil {
		t.Fatal(err)
	}

	hf = iotest.HalfReader(f)
	if err := Extract(hf, tmpdir); err != nil {
		t.Fatal(err)
	}
}

func TestExtractDotdot(t *testing.T) {
	f, err := os.Open("../testdata/dotdot.cpio")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	tmpdir, err := ioutil.TempDir("", "cpio")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)
	err = Extract(f, tmpdir)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(tmpdir, "aaaaaaaaa")); err != nil {
		t.Error("expected file with ../ to extract into top of destdir:", err)
	}
}
