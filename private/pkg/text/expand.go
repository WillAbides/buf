// Copyright 2020-2021 Buf Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package text

import (
	"fmt"
)

// Expand replaces variables formatted like ${var} or ${var|default value} in the data based on the env map.
// When a default value is set like ${var|foo} and there is no mapped value for "var", it will be replaced with "foo".
// In a default value, the character "}" must be escaped with "\}" and the character "\" must be escaped with "\\".
// Variable names must start with [a-zA-Z]. Subsequent characters must be [a-zA-Z0-9_].
func Expand(data []byte, env map[string]string) ([]byte, error) {
	var buf []byte
	i := 0
	dollar := false
	for j := 0; j < len(data); j++ {
		switch data[j] {
		case '$':
			dollar = true
		case '{':
			if !dollar {
				break
			}
			if buf == nil {
				buf = make([]byte, 0, 2*len(data))
			}
			buf = append(buf, data[i:j-1]...)
			name, defaultValue, w, err := varInfo(data[j+1:])
			if err != nil {
				errStringEnd := j + w + 5
				if errStringEnd > len(data) {
					errStringEnd = len(data)
				}
				return nil, newInvalidSyntaxError(w+2, string(data[j-1:errStringEnd]), err)
			}
			val, ok := env[string(name)]
			if ok {
				buf = append(buf, val...)
			} else {
				buf = append(buf, defaultValue...)
			}
			j += w
			i = j + 1
			dollar = false
		default:
			dollar = false
		}
	}
	if buf == nil {
		return data, nil
	}
	return append(buf, data[i:]...), nil
}

// varInfo returns information about a variable to be expanded.
// data is the remainder of a string after "${"
// name is the variable name
// defaultValue is the default value (the portion after a | pipe) or "" if no pipe is found
// n is the position in data after "}", or in case of an error, it's the position where the syntax becomes invalid
func varInfo(data []byte) (name, defaultValue []byte, n int, _ error) {
	var err error
	var nameLen int
	name, nameLen, err = readVarName(data)
	if err != nil {
		return nil, nil, nameLen, err
	}
	// readVarName shouldn't let this happen
	if nameLen > len(name)+1 {
		return nil, nil, nameLen, err
	}
	end := data[nameLen-1]
	if end == '}' {
		return name, nil, nameLen, nil
	}
	// readVarName shouldn't let this happen either
	if end != '|' {
		return nil, nil, nameLen, errInvalidCharacter
	}
	var valLen int
	defaultValue, valLen, err = readDefaultValue(data[nameLen:])
	if err != nil {
		return nil, nil, nameLen + valLen, err
	}
	return name, defaultValue, nameLen + valLen, nil
}

func newInvalidSyntaxError(
	position int,
	value string,
	err error,
) *invalidSyntaxErr {
	return &invalidSyntaxErr{
		position: position,
		value:    value,
		err:      err,
	}
}

type invalidSyntaxErr struct {
	position int
	value    string
	err      error
}

func (e *invalidSyntaxErr) Error() string {
	return fmt.Sprintf("invalid syntax at position %d of %q: %v",
		e.position,
		e.value,
		e.err,
	)
}

var (
	errInvalidCharacter         = fmt.Errorf("invalid character")
	errInvalidStartingCharacter = fmt.Errorf("invalid starting character")
	errUnterminated             = fmt.Errorf("unterminated")
	errEmptyString              = fmt.Errorf("empty string")
)

func isAlpha(c uint8) bool {
	return 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z'
}

func isAlphaNum(c uint8) bool {
	return isAlpha(c) || c == '_' || '0' <= c && c <= '9'
}

func readVarName(s []byte) ([]byte, int, error) {
	if len(s) == 0 {
		return nil, 0, errUnterminated
	}
	if s[0] == '}' || s[0] == '|' {
		return nil, 0, errEmptyString
	}
	if !isAlpha(s[0]) {
		return nil, 0, errInvalidStartingCharacter
	}
	i := 1
	for ; i < len(s); i++ {
		c := s[i]
		if c == '}' || c == '|' {
			return s[:i], i + 1, nil
		}
		if !isAlphaNum(c) {
			return nil, i, errInvalidCharacter
		}
	}
	return nil, 0, errUnterminated
}

func readDefaultValue(s []byte) ([]byte, int, error) {
	if len(s) == 0 {
		return nil, 0, errUnterminated
	}
	buf := make([]byte, 0, len(s))
	var escaped, foundClose bool
	i := 0
	for ; i < len(s); i++ {
		switch s[i] {
		case '\\':
			if escaped {
				buf = append(buf, s[i])
			}
			escaped = !escaped
			continue
		case '}':
			if !escaped {
				foundClose = true
				break
			}
		}
		if foundClose {
			break
		}
		buf = append(buf, s[i])
		escaped = false
	}
	if !foundClose {
		return nil, i, errUnterminated
	}
	return buf, i + 1, nil
}
