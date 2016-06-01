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
}

type fileInfo struct {
	name      string
	size      int64
	userName  string
	groupName string
	flags     int
	mtime     int
	digest    string
	mode      int
}

func (fi *fileInfo) Name() string {
	return fi.name
}

func (fi *fileInfo) Size() int64 {
	return fi.size
}

func (fi *fileInfo) UserName() string {
	return fi.userName
}

func (fi *fileInfo) GroupName() string {
	return fi.groupName
}

func (fi *fileInfo) Flags() int {
	return fi.flags
}

func (fi *fileInfo) Mtime() int {
	return fi.mtime
}

func (fi *fileInfo) Digest() string {
	return fi.digest
}

func (fi *fileInfo) Mode() int {
	return fi.mode
}
