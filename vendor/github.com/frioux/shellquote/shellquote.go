/*
shellquote quotes strings for shell scripts.

Sometimes you get strings from the internet and need to quote them for security,
other times you'll need to quote your own strings because doing it by hand is
just too much work.
*/
package shellquote

import (
	"errors"
	"regexp"
	"strings"
)

// ErrNull will be returned from Quote if any of the strings contains a null
// byte.
var ErrNull = errors.New("No way to quote string containing null bytes")

// Quote will return a shell quoted string for the passed tokens.
func Quote(in []string) (string, error) {
	tmp := make([]string, len(in))
	var sawNonEqual bool
	for i, x := range in {
		if x == "" {
			tmp[i] = `''`
			continue
		}
		if strings.Contains(x, "\x00") {
			return "", ErrNull
		}

		var escape bool
		hasEqual := strings.Contains(x, "=")
		if hasEqual {
			if !sawNonEqual {
				escape = true
			}
		} else {
			sawNonEqual = true
		}

		toEsc := regexp.MustCompile(`[^\w!%+,\-./:=@^]`)
		if !escape && toEsc.MatchString(x) {
			escape = true
		}

		if escape || (!sawNonEqual && hasEqual) {
			y := strings.Replace(x, `'`, `'\''`, -1)

			simplifyRe := regexp.MustCompile(`(?:'\\''){2,}`)
			y = simplifyRe.ReplaceAllStringFunc(y, func(str string) string {
				var inner string
				for i := 0; i < len(str)/4; i++ {
					inner += "'"
				}
				return `'"` + inner + `"'`
			})

			y = `'` + y + `'`
			y = strings.TrimSuffix(y, `''`)
			y = strings.TrimPrefix(y, `''`)

			tmp[i] = y
			continue
		}
		tmp[i] = x
	}
	return strings.Join(tmp, " "), nil
}

/*
Copyright 2018 Arthur Axel fREW Schmidt

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software

distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
