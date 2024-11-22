// Copyright 2024 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// Provenance-includes-location: https://github.com/golang/go/blob/f2d118fd5f7e872804a5825ce29797f81a28b0fa/src/strings/strings.go
// Provenance-includes-license: BSD-3-Clause
// Provenance-includes-copyright: Copyright The Go Authors.

package prometheus

import (
	"strings"
	"unicode/utf8"
)

// getTokensAndSeparators was originally a copy of strings.FieldsFunc from the Go standard library.
// We've made a few changes to it, making sure we return the separator between tokens. If
// UTF-8 isn't allowed, we replace all runes from the separator with an underscore.
func getTokensAndSeparators(s string, f func(rune) bool, allowUTF8 bool) ([]string, []string) {
	// A token is used to record a slice of s of the form s[start:end].
	// The start index is inclusive and the end index is exclusive.
	type token struct {
		start int
		end   int
	}
	tokens := make([]token, 0, 32)

	// Find the field start and end indices.
	// Doing this in a separate pass (rather than slicing the string s
	// and collecting the result substrings right away) is significantly
	// more efficient, possibly due to cache effects.
	start := -1 // valid span start if >= 0
	for end, rune := range s {
		if f(rune) {
			if start >= 0 {
				tokens = append(tokens, token{start, end})
				// Set start to a negative value.
				// Note: using -1 here consistently and reproducibly
				// slows down this code by a several percent on amd64.
				start = ^start
			}
		} else {
			if start < 0 {
				start = end
			}
		}
	}

	// Last field might end at EOF.
	if start >= 0 {
		tokens = append(tokens, token{start, len(s)})
	}

	// Create strings from recorded field indices.
	a := make([]string, len(tokens))
	// Separators are infered from the position gaps between tokens.
	separators := make([]string, 0, len(tokens)-1)
	a[0] = s[tokens[0].start:tokens[0].end]
	for i := 1; i < len(tokens); i++ {
		a[i] = s[tokens[i].start:tokens[i].end]
		if allowUTF8 {
			separators = append(separators, s[tokens[i-1].end:tokens[i].start])
		} else {
			sep := ""
			// We measure the rune count instead of using len because len returns amount of
			// bytes in the string, and utf-8 runes can be multiple bytes long.
			runeCount := utf8.RuneCountInString(s[tokens[i-1].end:tokens[i].start])
			for j := 0; j < runeCount; j++ {
				sep += "_"
			}
			separators = append(separators, sep)
		}
	}

	return a, separators
}

// join is a copy of strings.Join from the Go standard library,
// but it also accepts a slice of separators to join the elements with.
// If the slice of separators is shorter than the slice of elements, use a default value.
// We also don't check for integer overflow.
func join(elems []string, separators []string, def string) string {
	switch len(elems) {
	case 0:
		return ""
	case 1:
		return elems[0]
	}

	var n int
	var sep string
	sepLen := len(separators)
	for i, elem := range elems {
		if i >= sepLen {
			sep = def
		} else {
			sep = separators[i]
		}
		n += len(sep) + len(elem)
	}

	var b strings.Builder
	b.Grow(n)
	b.WriteString(elems[0])
	for i, s := range elems[1:] {
		if i >= sepLen {
			sep = def
		} else {
			sep = separators[i]
		}
		b.WriteString(sep)
		b.WriteString(s)
	}
	return b.String()
}
