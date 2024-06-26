//go:build !windows
// +build !windows

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

package fileutil

import (
	"os"
	"golang.org/x/sys/unix"
)

// HasLinks returns true if the given file has Nlink > 1
func HasLinks(info os.FileInfo) bool {
	stat, ok := info.Sys().(*unix.Stat_t)
	if !ok {
		return false
	}
	return stat.Nlink != 1
}

// Mkfifo creates a named pipe with the specified path and permissions
func Mkfifo(path string, mode uint32) error {
	return unix.Mkfifo(path, mode)
}
