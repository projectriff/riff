/*
 * Copyright 2019 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cli

import (
	"github.com/fatih/color"
)

var (
	FaintColor   = color.New(color.Faint)
	InfoColor    = color.New(color.FgCyan)
	SuccessColor = color.New(color.FgGreen)
	WarnColor    = color.New(color.FgYellow)
	ErrorColor   = color.New(color.FgRed)
)

func Sfaintf(format string, a ...interface{}) string {
	return FaintColor.Sprintf(format, a...)
}

func Sinfof(format string, a ...interface{}) string {
	return InfoColor.Sprintf(format, a...)
}

func Ssuccessf(format string, a ...interface{}) string {
	return SuccessColor.Sprintf(format, a...)
}

func Swarnf(format string, a ...interface{}) string {
	return WarnColor.Sprintf(format, a...)
}

func Serrorf(format string, a ...interface{}) string {
	return ErrorColor.Sprintf(format, a...)
}
