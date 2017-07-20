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

type FileInfo interface {
	Name() string
	Size() int64
	UserName() string
	GroupName() string
	Flags() int
	Mtime() int
	Digest() string
	Mode() int
	Linkname() string
	Device() int
	Inode() int
}

type fileInfo struct {
	name      string
	size      uint64
	userName  string
	groupName string
	flags     uint32
	mtime     uint32
	digest    string
	mode      uint32
	linkName  string
	device    uint32
	inode     uint32
}

func (fi *fileInfo) Name() string {
	return fi.name
}

func (fi *fileInfo) Size() int64 {
	return int64(fi.size)
}

func (fi *fileInfo) UserName() string {
	return fi.userName
}

func (fi *fileInfo) GroupName() string {
	return fi.groupName
}

func (fi *fileInfo) Flags() int {
	return int(fi.flags)
}

func (fi *fileInfo) Mtime() int {
	return int(fi.mtime)
}

func (fi *fileInfo) Digest() string {
	return fi.digest
}

func (fi *fileInfo) Mode() int {
	return int(fi.mode)
}

func (fi *fileInfo) Linkname() string {
	return fi.linkName
}

func (fi *fileInfo) Device() int {
	return int(fi.device)
}

func (fi *fileInfo) Inode() int {
	return int(fi.inode)
}

func (fi *fileInfo) fileType() uint32 {
	return fi.mode &^ 07777
}

func (fi *fileInfo) inode64() uint64 {
	return (uint64(fi.device) << 32) | uint64(fi.inode)
}
