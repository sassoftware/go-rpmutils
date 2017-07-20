/*
 * Copyright (c) SAS Institute Inc.
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
	"io"
	"os"
	"testing"
	"testing/iotest"
)

func TestReadStripped(t *testing.T) {
	f, err := os.Open("../testdata/stripped.cpio")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	r := NewReaderWithSizes(iotest.HalfReader(f), []int64{3, 6})
	h, err := r.Next()
	if err != nil {
		t.Fatal(err)
	}
	if !h.IsStripped() || h.Index() != 0 {
		t.Fatalf("wrong header: %#v", h)
	}
	h, err = r.Next()
	if err != nil {
		t.Fatal(err)
	}
	if !h.IsStripped() || h.Index() != 1 {
		t.Fatalf("wrong header: %#v", h)
	}
	_, err = r.Next()
	if err != io.EOF {
		t.Fatalf("wrong error: %v", err)
	}
}
