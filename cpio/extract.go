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
	"os"
	"path"
	"syscall"
)

func Extract(rs io.ReadSeeker, dest string) error {
	linkMap := make(map[int][]string)

	stream := NewCpioStream(rs)

	entry, err := stream.ReadNextEntry()
	if err != nil {
		return err
	}

	for entry.filename != TRAILER {
		target := path.Join(dest, entry.filename)
		parent := path.Dir(target)

		// Create the parent directory if it doesn't exist.
		if _, err := os.Stat(parent); os.IsNotExist(err) {
			if err := os.MkdirAll(parent, 0755); err != nil {
				return err
			}
		}

		// FIXME: Need a makedev implementation in go.

		mode := os.FileMode(entry.header.c_mode)
		if mode&os.ModeCharDevice != 0 {
			// FIXME: skipping due to lack of makedev.
			continue
		} else if mode&os.ModeDevice != 0 {
			// FIXME: skipping due to lack of makedev.
			continue
		} else if mode&os.ModeDir != 0 {
			if err := os.Mkdir(target, mode); err != nil {
				return err
			}
		} else if mode&os.ModeNamedPipe != 0 {
			if err := syscall.Mkfifo(target, uint32(mode)); err != nil {
				return err
			}
		} else if mode&os.ModeSymlink != 0 {
			buf := make([]byte, entry.header.c_filesize)
			if _, err := entry.payload.Read(buf); err != nil {
				return err
			}
			if err := os.Symlink(string(buf), target); err != nil {
				return err
			}
		} else if mode&os.ModeType == 0 {
			// save hardlinks until after the taget is written
			if entry.header.c_nlink > 1 && entry.header.c_filesize == 0 {
				l, ok := linkMap[entry.header.c_ino]
				if !ok {
					l = make([]string, 0)
				}
				l = append(l, target)
				linkMap[entry.header.c_ino] = l
				continue
			}

			f, err := os.Create(target)
			if err != nil {
				return err
			}
			written, err := io.Copy(f, entry.payload)
			if err != nil {
				return err
			}
			if written != int64(entry.header.c_filesize) {
				return fmt.Errorf("short write")
			}
			if err := f.Close(); err != nil {
				return err
			}

			// Create hardlinks after the file content is written.
			if entry.header.c_nlink > 1 && entry.header.c_filesize > 0 {
				l, ok := linkMap[entry.header.c_ino]
				if !ok {
					return fmt.Errorf("hardlinks missing")
				}

				for _, t := range l {
					if err := os.Link(target, t); err != nil {
						return err
					}
				}
			}

		} else {
			return fmt.Errorf("unknown file mode 0%o for %s",
				entry.header.c_mode, entry.filename)
		}
	}

	return nil
}
