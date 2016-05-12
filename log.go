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
	"io"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("rpmutils")

var _format = logging.MustStringFormatter(
	`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
)

func SetupLogging(cmdOut io.Writer, logFile io.Writer, debug bool, cmddebug bool) {
	if cmdOut != nil {
		cmdBackend := logging.NewLogBackend(cmdOut, "", 0)
		cmdLevel := logging.AddModuleLevel(cmdBackend)
		if !cmddebug {
			cmdLevel.SetLevel(logging.INFO, "")
		} else {
			cmdLevel.SetLevel(logging.DEBUG, "")
		}
		logging.SetBackend(cmdLevel)
	}

	if logFile != nil {
		logBackend := logging.NewLogBackend(logFile, "", 0)
		logFormatter := logging.NewBackendFormatter(logBackend, _format)
		logLevel := logging.AddModuleLevel(logFormatter)
		if !debug {
			logLevel.SetLevel(logging.INFO, "")
		} else {
			logLevel.SetLevel(logging.DEBUG, "")
		}
		logging.SetBackend(logLevel)
	}
}
