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
	"time"
)

// Extract the contents of a cpio stream from and writes it as a tar file into the provided writer
func Tar(rs io.Reader, tarfile *tar.Writer) error {
	hardLinks := map[int][]*tar.Header{}
	inodes := map[int]string{}

	stream := NewCpioStream(rs)

	for {
		entry, err := stream.ReadNextEntry()
		if err != nil {
			return err
		}

		if entry.Header.filename == TRAILER {
			break
		}

		tarHeader := &tar.Header{
			Name:     entry.Header.filename,
			Size:     entry.Header.Filesize64(),
			Mode:     int64(entry.Header.Mode()),
			Uid:      entry.Header.Uid(),
			Gid:      entry.Header.Gid(),
			ModTime:  time.Unix(int64(entry.Header.Mtime()), 0),
			Devmajor: int64(entry.Header.Devmajor()),
			Devminor: int64(entry.Header.Devminor()),
		}

		var payload io.Reader
		switch entry.Header.Mode() &^ 07777 {
		case S_ISCHR:
			tarHeader.Typeflag = tar.TypeChar
		case S_ISBLK:
			tarHeader.Typeflag = tar.TypeBlock
		case S_ISDIR:
			tarHeader.Typeflag = tar.TypeDir
		case S_ISFIFO:
			tarHeader.Typeflag = tar.TypeFifo
		case S_ISLNK:
			tarHeader.Typeflag = tar.TypeSymlink
			buf := make([]byte, entry.Header.c_filesize)
			if _, err := entry.payload.Read(buf); err != nil {
				return err
			}
			tarHeader.Linkname = string(buf)
		case S_ISREG:
			if entry.Header.c_nlink > 1 && entry.Header.c_filesize == 0 {
				tarHeader.Typeflag = tar.TypeLink
				hardLinks[entry.Header.c_ino] = append(hardLinks[entry.Header.c_ino], tarHeader)
				continue
			}
			tarHeader.Typeflag = tar.TypeReg
			payload = entry.payload
			inodes[entry.Header.c_ino] = entry.Header.filename
		default:
			return fmt.Errorf("unknown file mode 0%o for %s",
				entry.Header.c_mode, entry.Header.filename)
		}
		if err := tarfile.WriteHeader(tarHeader); err != nil {
			return fmt.Errorf("could not write tar header for %v: %v", tarHeader.Name, err)
		}
		if payload != nil {
			written, err := io.Copy(tarfile, entry.payload)
			if err != nil {
				return fmt.Errorf("could not write body for %v: %v", tarHeader.Name, err)
			}
			if written != int64(entry.Header.c_filesize) {
				return fmt.Errorf("short write body for %v", tarHeader.Name)
			}
		}
	}
	// write hardlinks
	for node, links := range hardLinks {
		target := inodes[node]
		if target == "" {
			return fmt.Errorf("no target file for inode %v found", node)
		}
		for _, tarHeader := range links {
			tarHeader.Linkname = target
			if err := tarfile.WriteHeader(tarHeader); err != nil {
				return fmt.Errorf("could not write tar header for %v", tarHeader.Name)
			}
		}
	}

	return nil
}
