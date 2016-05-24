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
	"bytes"
	"strings"
	"testing"
)

func TestCmdLogging(t *testing.T) {
	buf := new(bytes.Buffer)
	setupLogging(buf, nil, false, false)
	log.Info("foo")
	if string(buf.Bytes()) != "foo\n" {
		t.Fatalf("got the wrong output: %s", string(buf.Bytes()))
	}
	buf.Reset()

	log.Debug("bar")
	if string(buf.Bytes()) != "" {
		t.Fatalf("got the wrong output: %s", string(buf.Bytes()))
	}

	setupLogging(buf, nil, false, true)
	buf.Reset()
	log.Debug("bar")
	if string(buf.Bytes()) != "bar\n" {
		t.Fatalf("got the wrong output: %s", string(buf.Bytes()))
	}
}

func TestLogFileLogging(t *testing.T) {
	buf := new(bytes.Buffer)
	setupLogging(nil, buf, false, false)
	log.Info("foo")
	if !strings.HasSuffix(string(buf.Bytes()), "foo\n") {
		t.Fatal("wrong suffix: %s", string(buf.Bytes()))
	}
	buf.Reset()

	log.Debug("bar")
	if string(buf.Bytes()) != "" {
		t.Fatal("got wrong output: %s", string(buf.Bytes()))
	}
	buf.Reset()

	setupLogging(nil, buf, true, false)
	log.Debug("bar")
	if !strings.HasSuffix(string(buf.Bytes()), "bar\n") {
		t.Fatal("wrong suffix: %s", string(buf.Bytes()))
	}
}

func TestBothLogging(t *testing.T) {
	buf1 := new(bytes.Buffer)
	buf2 := new(bytes.Buffer)

	setupLogging(buf1, buf2, false, false)
	log.Info("foo")
	if string(buf1.Bytes()) != "foo\n" {
		t.Fatalf("got the wrong output: \"%s\"", string(buf1.Bytes()))
	}
	if !strings.HasSuffix(string(buf2.Bytes()), "foo\n") {
		t.Fatal("wrong suffix: %s", string(buf2.Bytes()))
	}
	buf1.Reset()
	buf2.Reset()

	log.Debug("bar")
	if string(buf1.Bytes()) != "" {
		t.Fatalf("got the wrong output: \"%s\"", string(buf1.Bytes()))
	}
	if string(buf2.Bytes()) != "" {
		t.Fatal("got wrong output: %s", string(buf2.Bytes()))
	}
	buf1.Reset()
	buf2.Reset()

	setupLogging(buf1, buf2, true, true)
	log.Debug("bar")
	if string(buf1.Bytes()) != "bar\n" {
		t.Fatalf("got the wrong output: %s", string(buf1.Bytes()))
	}
	if !strings.HasSuffix(string(buf2.Bytes()), "bar\n") {
		t.Fatal("wrong suffix: %s", string(buf2.Bytes()))
	}
}
