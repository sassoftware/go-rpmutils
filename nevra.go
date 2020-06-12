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
	"fmt"
	"sort"
)

type NEVRA struct {
	Name    string
	Epoch   string
	Version string
	Release string
	Arch    string
}

// TODO: in v2 change epoch to an int

func (nevra *NEVRA) String() string {
	return fmt.Sprintf("%s-%s:%s-%s.%s.rpm", nevra.Name, nevra.Epoch, nevra.Version, nevra.Release, nevra.Arch)
}

func NEVRAcmp(a NEVRA, b NEVRA) int {
	if res := Vercmp(a.Epoch, b.Epoch); res != 0 {
		return res
	}
	if res := Vercmp(a.Version, b.Version); res != 0 {
		return res
	}
	if res := Vercmp(a.Release, b.Release); res != 0 {
		return res
	}
	return 0
}

type NEVRASlice []NEVRA

func (s NEVRASlice) Len() int {
	return len(s)
}

func (s NEVRASlice) Less(i, j int) bool {
	return NEVRAcmp(s[i], s[j]) == -1
}

func (s NEVRASlice) Swap(i, j int) {
	n := s[i]
	s[i] = s[j]
	s[j] = n
}

func (s NEVRASlice) Sort() {
	sort.Sort(s)
}
