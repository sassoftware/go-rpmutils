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

// Standard set of permission bit masks.
const (
	S_ISUID  = 04000   // Set uid
	S_ISGID  = 02000   // Set gid
	S_ISVTX  = 01000   // Save text (sticky bit)
	S_ISDIR  = 040000  // Directory
	S_ISFIFO = 010000  // FIFO
	S_ISREG  = 0100000 // Regular file
	S_ISLNK  = 0120000 // Symbolic link
	S_ISBLK  = 060000  // Block special file
	S_ISCHR  = 020000  // Character special file
	S_ISSOCK = 0140000 // Socket
)

func Extract(rs io.Reader, dest string) error {
	linkMap := make(map[int][]string)

	stream := NewCpioStream(rs)

	for {
		entry, err := stream.ReadNextEntry()
		if err != nil {
			return err
		}

		if entry.Header.filename == TRAILER {
			break
		}

		target := path.Join(dest, entry.Header.filename)
		parent := path.Dir(target)

		// Create the parent directory if it doesn't exist.
		if _, err := os.Stat(parent); os.IsNotExist(err) {
			if err := os.MkdirAll(parent, 0755); err != nil {
				return err
			}
		}

		// FIXME: Need a makedev implementation in go.

		switch entry.Header.Mode() &^ 07777 {
		case S_ISCHR:
			log.Debug("unpacking char device")
			// FIXME: skipping due to lack of makedev.
			continue
		case S_ISBLK:
			log.Debug("unpacking block device")
			// FIXME: skipping due to lack of makedev.
			continue
		case S_ISDIR:
			log.Debug("unpacking dir")
			m := os.FileMode(entry.Header.Mode()).Perm()
			if err := os.Mkdir(target, m); err != nil {
				return err
			}
		case S_ISFIFO:
			log.Debug("unpacking named pipe")
			if err := syscall.Mkfifo(target, uint32(entry.Header.Mode())); err != nil {
				return err
			}
		case S_ISLNK:
			log.Debug("unpacking symlink")
			buf := make([]byte, entry.Header.c_filesize)
			if _, err := entry.payload.Read(buf); err != nil {
				return err
			}
			if err := os.Symlink(string(buf), target); err != nil {
				return err
			}
		case S_ISREG:
			log.Debug("unpacking regular file")
			// save hardlinks until after the taget is written
			if entry.Header.c_nlink > 1 && entry.Header.c_filesize == 0 {
				log.Debug("regular file is a hard link")
				l, ok := linkMap[entry.Header.c_ino]
				if !ok {
					l = make([]string, 0)
				}
				l = append(l, target)
				linkMap[entry.Header.c_ino] = l
				continue
			}

			// FIXME: Set permissions on files when creating.
			f, err := os.Create(target)
			if err != nil {
				return err
			}
			written, err := io.Copy(f, entry.payload)
			if err != nil {
				return err
			}
			if written != int64(entry.Header.c_filesize) {
				log.Debugf("written: %d, filesize: %d", written, entry.Header.c_filesize)
				return fmt.Errorf("short write")
			}
			if err := f.Close(); err != nil {
				return err
			}

			// Create hardlinks after the file content is written.
			if entry.Header.c_nlink > 1 && entry.Header.c_filesize > 0 {
				l, ok := linkMap[entry.Header.c_ino]
				if !ok {
					return fmt.Errorf("hardlinks missing")
				}

				for _, t := range l {
					if err := os.Link(target, t); err != nil {
						return err
					}
				}
			}
		default:
			return fmt.Errorf("unknown file mode 0%o for %s",
				entry.Header.c_mode, entry.Header.filename)
		}
	}

	return nil
}
