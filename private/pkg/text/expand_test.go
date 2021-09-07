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
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_varInfo(t *testing.T) {
	for _, td := range []struct {
		input        string
		name         string
		defaultValue string
		length       int
		wantErr      bool
	}{
		{input: "foo}", name: "foo", length: 4},
		{input: "f}", name: "f", length: 2},
		{input: "}", wantErr: true},
		{input: "foo", wantErr: true},
		{input: "H}OME", name: "H", length: 2, wantErr: false},
		{input: "foo|bar}", name: "foo", defaultValue: "bar", length: 8},
		{input: "foo|hello|world}", name: "foo", defaultValue: "hello|world", length: 16},
		{input: "foo|}", name: "foo", length: 5},
	} {
		t.Run(td.input, func(t *testing.T) {
			name, defaultValue, length, err := varInfo([]byte(td.input))
			errCheck := assert.NoError
			if td.wantErr {
				errCheck = assert.Error
			}
			errCheck(t, err)
			assert.Equal(t, td.name, string(name))
			assert.Equal(t, td.defaultValue, string(defaultValue))
			assert.Equal(t, td.length, length)
		})
	}
}

func Test_readDefaultValue(t *testing.T) {
	for _, td := range []struct {
		input  string
		output string
		length int
		err    error
	}{
		{input: "", err: errUnterminated},
		{input: `asdf`, length: 4, err: errUnterminated},
		{input: `asdf}`, length: 5, output: `asdf`},
		{input: `as\}f}`, length: 6, output: `as}f`},
		{input: `as\}}f}`, length: 5, output: `as}`},
		{input: `as\\\}}f}`, length: 7, output: `as\}`},
		{input: `\\\}}f}`, length: 5, output: `\}`},
		{input: `\\\}}f}`, length: 5, output: `\}`},
		{input: `world|foo}`, length: 10, output: `world|foo`},
	} {
		t.Run(td.input, func(t *testing.T) {
			output, length, err := readDefaultValue([]byte(td.input))
			assert.Equal(t, td.err, err)
			assert.Equal(t, td.length, length)
			assert.Equal(t, td.output, string(output))
		})
	}
}

func Test_readVarName(t *testing.T) {
	for _, td := range []struct {
		input  string
		output string
		length int
		err    error
	}{
		{input: "", err: errUnterminated},
		{input: "|", err: errEmptyString},
		{input: "}asdf", err: errEmptyString},
		{input: "asdf", length: 0, err: errUnterminated},
		{input: "as*f}", length: 2, err: errInvalidCharacter},
		{input: "asdf}jkl;", length: 5, output: "asdf"},
		{input: "asdf}", length: 5, output: "asdf"},
		{input: "asdf|jkl;", length: 5, output: "asdf"},
		{input: "2asdf|jkl;", err: errInvalidStartingCharacter},
		{input: "{", err: errInvalidStartingCharacter},
	} {
		t.Run(td.input, func(t *testing.T) {
			output, length, err := readVarName([]byte(td.input))
			assert.Equal(t, td.err, err)
			assert.Equal(t, td.length, length)
			assert.Equal(t, td.output, string(output))
		})
	}
}

func TestExpand(t *testing.T) {
	env := map[string]string{
		"*":      "all the args",
		"#":      "NARGS",
		"$":      "PID",
		"1":      "ARGUMENT1",
		"HOME":   "/usr/gopher",
		"H":      "(Value of H)",
		"home_1": "/usr/foo",
		"_":      "underscore",
	}
	for _, test := range []struct {
		in  string
		out string
		err error
	}{
		{},
		{in: `$*`, out: `$*`},
		{in: `$1`, out: `$1`},
		{in: `${1}`, err: newInvalidSyntaxError(2, `${1}`, errInvalidStartingCharacter)},
		{in: `now is the time`, out: `now is the time`},
		{in: `${home_1}`, out: `/usr/foo`},
		{in: `${H}OME`, out: `(Value of H)OME`},
		{in: `a${H}run`, out: `a(Value of H)run`},
		{in: `start$+middle$^end$`, out: `start$+middle$^end$`},
		{in: `$`, out: `$`},
		{in: `$}`, out: `$}`},
		{in: `${`, err: newInvalidSyntaxError(2, `${`, errUnterminated)},
		{in: `${asdf`, err: newInvalidSyntaxError(2, `${asdf`, errUnterminated)},
		{in: `a$df${asdf`, err: newInvalidSyntaxError(2, `${asdf`, errUnterminated)},
		{in: `${}`, err: newInvalidSyntaxError(2, `${}`, errEmptyString)},
		{in: `abc${}`, err: newInvalidSyntaxError(2, `${}`, errEmptyString)},
		{in: `abc${hello|world|foo}`, out: `abcworld|foo`},
	} {
		t.Run(test.in, func(t *testing.T) {
			result, err := Expand([]byte(test.in), env)
			assert.Equal(t, test.err, err)
			assert.Equal(t, test.out, string(result))
		})
	}
}
