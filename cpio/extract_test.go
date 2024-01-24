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
	"io"
	"os"
	"path/filepath"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/require"
)

func TestExtract(t *testing.T) {
	f, err := os.Open("../testdata/foo.cpio")
	require.NoError(t, err)
	defer f.Close()

	tmpdir := t.TempDir()
	hf := iotest.HalfReader(f)
	require.NoError(t, Extract(hf, tmpdir), "failed to extract cpio")

	// extract again and ensure it overwrites successfully
	_, err = f.Seek(0, io.SeekStart)
	require.NoError(t, err)
	require.NoError(t, Extract(hf, tmpdir), "failed to overwrite extracted cpio")
}

func TestExtractDotdot(t *testing.T) {
	f, err := os.Open("../testdata/dotdot.cpio")
	require.NoError(t, err)
	defer f.Close()

	tmpdir := t.TempDir()
	require.NoError(t, Extract(f, tmpdir))

	if _, err := os.Stat(filepath.Join(tmpdir, "aaaaaaaaa")); err != nil {
		t.Error("expected file with ../ to extract into top of destdir:", err)
	}
}
