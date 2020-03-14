// SPDX-License-Identifier: MIT
// SPDX-License-Identifier: Unlicense
package string

import (
	"strings"
	"unicode"
)

// Simple helper method that removes duplicates from
// any given string slice and then returns a nice
// duplicate free string slice
func RemoveStringDuplicates(elements []string) []string {
	encountered := map[string]bool{}
	result := []string{}

	for v := range elements {
		if !encountered[elements[v]] == true {
			encountered[elements[v]] = true
			result = append(result, elements[v])
		}
	}

	return result
}

func Contains(elements []string, needle string) bool {
	for _, v := range elements {
		if needle == v {
			return true
		}
	}

	return false
}

// Given a string returns a slice containing all possible case permutations
// of that string such that input of foo will return
// foo Foo fOo FOo foO FoO fOO FOO
// Note that very long inputs can produce an enormous amount of
// results in the returned slice OR result in an overflow and return nothing
func PermuteCase(input string) []string {
	l := len(input)
	max := 1 << l

	combinations := []string{}

	for i := 0; i < max; i++ {
		s := ""
		for idx, ch := range input {
			if (i & (1 << idx)) == 0 {
				s += strings.ToUpper(string(ch))
			} else {
				s += strings.ToLower(string(ch))
			}
		}

		combinations = append(combinations, s)
	}

	return RemoveStringDuplicates(combinations)
}

// Given a string returns a slice containing all possible case permutations
// with characters being folded such that S will return S s ſ
func PermuteCaseFolding(input string) []string {
	combinations := PermuteCase(input)
	combos := []string{}

	for _, combo := range combinations {
		for index, runeValue := range combo {
			for _, p := range AllSimpleFold(runeValue) {
				combos = append(combos, combo[:index]+string(p)+combo[index+len(string(runeValue)):])
			}
		}
	}

	return RemoveStringDuplicates(combos)
}

// Given an input rune return a rune slice containing
// all of the possible simple fold
func AllSimpleFold(input rune) []rune {
	res := []rune{}

	// This works for getting all folded representations
	// but feels totally wrong due to the bailout break.
	// That said its simpler than a while with checks
	// TODO https://github.com/golang/go/blob/master/src/regexp/syntax/prog.go#L215 possible way to implement
	for i := 0; i < 255; i++ {
		input = unicode.SimpleFold(input)
		if containsRune(res, input) {
			break
		}
		res = append(res, input)
	}

	return res
}

func containsRune(elements []rune, needle rune) bool {
	for _, v := range elements {
		if needle == v {
			return true
		}
	}

	return false
}

// Bytes MUST be UTF-8 encoded
// List of spaces detected (same as unicode.IsSpace):
// '\t', '\n', '\v', '\f', '\r', ' ', U+0085 (NEL), U+00A0 (NBSP).
// N.B only two bytes are required for these cases.  If we decided
// to support spaces like '，' then we'll need more bytes.
func IsSpace(firstByte, nextByte byte) bool {
    switch {
    case (9 <= firstByte) && (firstByte <= 13):  // \t, \n, \f, \r
        return true
    case firstByte == 32:  // SPACE
        return true
    case firstByte == 194:
        if nextByte == 133 {  // NEL
            return true
        } else if nextByte == 160 {  // NBSP
            return true
        }
    }
    return false
}

func StartOfRune(b byte) bool {
    if (b < (1 << 7)) || ((3 << 6) < b) {
        return true
    }
    // Else continuation bit
    return false
}
