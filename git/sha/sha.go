// Copyright 2023 Harness, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sha

import (
	"bytes"
	"encoding/hex"
	"errors"
	"strings"
)

// Nil defines empty git SHA.
const Nil = "0000000000000000000000000000000000000000"

// EmptyTree is the SHA of an empty tree.
const EmptyTree = "4b825dc642cb6eb9a060e54bf8d69288fbee4904"

var (
	ErrTypeNotSupported = errors.New("type not supported")
)

// SHA a git commit name.
type SHA struct {
	bytes []byte

	str string
}

// String returns string (hex) representation of the SHA.
func (s SHA) String() string {
	if s.str == "" {
		s.str = hex.EncodeToString(s.bytes)
	}
	return s.str
}

// IsZero returns whether this SHA1 is all zeroes.
func (s SHA) IsZero() bool {
	return len(s.bytes) == 0
}

// Equal returns true if val has the same SHA as s. It supports
// string, []byte, and SHA.
func (s SHA) Equal(val any) bool {
	switch v := val.(type) {
	case string:
		return v == s.String()
	case []byte:
		return bytes.Equal(v, s.bytes)
	case SHA:
		return bytes.Equal(v.bytes, s.bytes)
	default:
		return false
	}
}

// New creates a new SHA from a value T.
func New[T interface {
	~string | ~[]byte
}](value T) (SHA, error) {
	switch arg := any(value).(type) {
	case string:
		s := strings.TrimSpace(arg)
		b, err := hex.DecodeString(s)
		if err != nil {
			return SHA{}, err
		}
		return SHA{bytes: b}, nil
	case []byte:
		return SHA{bytes: arg}, nil
	default:
		return SHA{}, ErrTypeNotSupported
	}
}

func MustNew(s string) SHA {
	sha, _ := New(s)
	return sha
}