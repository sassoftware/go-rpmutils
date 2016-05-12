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
	"os"
	"testing"
)

func TestReadHeader(t *testing.T) {
	//SetupLogging(os.Stderr, os.Stderr, true, true)
	f, err := os.Open("./testdata/simple-1.0.1-1.i386.rpm")
	if err != nil {
		t.Fatal(err)
	}

	hdr, err := ReadHeader(f)
	if err != nil {
		t.Fatal(err)
	}

	if hdr == nil {
		t.Fatal("no header found")
	}

	nevra, err := hdr.GetNEVRA()
	if err != nil {
		t.Fatal(err)
	}

	if nevra.Epoch != "0" || nevra.Name != "simple" || nevra.Version != "1.0.1" || nevra.Release != "1" || nevra.Arch != "i386" {
		t.Fatalf("incorrect nevra: %s-%s:%s-%s.%s", nevra.Name, nevra.Epoch, nevra.Version, nevra.Release, nevra.Arch)
	}
}
