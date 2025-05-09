package main

import (
	"testing"
)

func TestOneRepeat(t *testing.T) {
	str := []byte("cat")
	pattern := "ca+t"
	result, err, _ := matchLine(str, pattern)
	if err != nil {
		panic("error")
	}

	if !result {
		t.Fatalf("incorrect result for %v, %v", str, pattern)
	}
}

func TestEscaped(t *testing.T) {
	str := []byte("123")
	pattern := "\\d"
	result, err, _ := matchLine(str, pattern)
	if err != nil {
		panic("error")
	}

	if !result {
		t.Fatalf("incorrect result for %v, %v", str, pattern)
	}
}

func TestRepeatWithSameCharacter(t *testing.T) {
	str := []byte("caat")
	pattern := "ca+at"
	result, err, _ := matchLine(str, pattern)
	if err != nil {
		panic("error")
	}

	if !result {
		t.Fatalf("incorrect result for %v, %v", str, pattern)
	}
}

func TestSingleCharacterNonMatch(t *testing.T) {
	str := []byte("dog")
	pattern := "f"
	result, err, _ := matchLine(str, pattern)
	if err != nil {
		panic("error")
	}

	if result {
		t.Fatalf("incorrect result for %v, %v", str, pattern)
	}
}

func TestPositiveCharacterGroups(t *testing.T) {
	str := []byte("a")
	pattern := "[abcd]"
	result, err, _ := matchLine(str, pattern)
	if err != nil {
		panic("error")
	}

	if !result {
		t.Fatalf("incorrect result for %v, %v", str, pattern)
	}
}

func TestZeroOrMore(t *testing.T) {
	str := []byte("act")
	pattern := "ca?t"
	result, err, _ := matchLine(str, pattern)
	if err != nil {
		panic("error")
	}

	if !result {
		t.Fatalf("incorrect result for %v, %v", str, pattern)
	}
}

func TestAlternation(t *testing.T) {
	str := []byte("a cat")
	pattern := "a (cat|dog)"
	result, err, _ := matchLine(str, pattern)
	if err != nil {
		panic("error")
	}

	if !result {
		t.Fatalf("incorrect result for %v, %v", str, pattern)
	}
}

func TestSingleBackreferenceEasy(t *testing.T) {
	str := []byte("cat and cat")
	pattern := "(cat) and \\1"
	result, err, _ := matchLine(str, pattern)
	if err != nil {
		panic("error")
	}

	if !result {
		t.Fatalf("incorrect result for %v, %v", str, pattern)
	}
}

func TestSingleBackreference(t *testing.T) {
	str := []byte("grep 101 is doing grep 101 times")
	pattern := "(\\w\\w\\w\\w \\d\\d\\d) is doing \\1 times"
	result, err, _ := matchLine(str, pattern)
	if err != nil {
		panic("error")
	}

	if !result {
		t.Fatalf("incorrect result for %v, %v", str, pattern)
	}
}

func TestSingleBackreferenceHarder(t *testing.T) {
	str := []byte("abcd is abcd, not efg")
	pattern := "([abcd]+) is \\1, not [^xyz]+"
	result, err, _ := matchLine(str, pattern)
	if err != nil {
		panic("error")
	}

	if !result {
		t.Fatalf("incorrect result for %v, %v", str, pattern)
	}
}

func TestAnotherSingleBackreference(t *testing.T) {
	str := []byte("t with tf")
	pattern := "^(\\w+) with \\1$"
	result, err, _ := matchLine(str, pattern)
	if err != nil {
		panic("error")
	}

	if result {
		t.Fatalf("incorrect result for %v, %v", str, pattern)
	}
}

func TestEnd(t *testing.T) {
	str := []byte("cats")
	pattern := "cat$"
	result, err, _ := matchLine(str, pattern)
	if err != nil {
		panic("error")
	}

	if result {
		t.Fatalf("incorrect result for %v, %v", str, pattern)
	}
}

func TestAnotherBackreference(t *testing.T) {
	str := []byte("abcd is abcd, not efg")
	pattern := "([abcd]+) is \\1, not [^xyz]+"
	result, err, _ := matchLine(str, pattern)
	if err != nil {
		panic("error")
	}

	if !result {
		t.Fatalf("incorrect result for %v, %v", str, pattern)
	}
}

func TestMoreBackReferencing(t *testing.T) {
	str := []byte("this with this")
	pattern := "^(\\w+) with \\1$"
	result, err, _ := matchLine(str, pattern)
	if err != nil {
		panic("error")
	}

	if !result {
		t.Fatalf("incorrect result for %v, %v", string(str), pattern)
	}
}

func TestDoubleBackreference(t *testing.T) {
	str := []byte("3 red squares and 3 red circles")
	pattern := "(\\d+) (\\w+) squares and \\1 \\2 circles"
	result, err, _ := matchLine(str, pattern)
	if err != nil {
		panic("error")
	}

	if !result {
		t.Fatalf("incorrect result for %v, %v", str, pattern)
	}
}



func TestAllMatches(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		pattern  string
		expected bool
	}{
		{
			"one-repeat",
			"cat",
			"ca+t",
			true,
		},
		{
			"repeat-with-same-character",
			"caat",
			"ca+at",
			true,
		},
		{
			"single-character-non-match",
			"dog",
			"f",
			false,
		},
		{
			"positive-character-groups",
			"a",
			"[abcd]",
			true,
		},
		{
			"zero-or-more",
			"act",
			"ca?t",
			true,
		},
		{
			"alternation",
			"a cat",
			"a (cat|dog)",
			true,
		},
		{
			"single-backreference",
			"grep 101 is doing grep 101 times",
			"(\\w\\w\\w\\w \\d\\d\\d) is doing \\1 times",
			true,
		},
		{
			"single-backreference-harder",
			"abcd is abcd, not efg",
			"([abcd]+) is \\1, not [^xyz]+",
			true,
		},
		{
			"multiple-alternating-groups",
			"a dog and cats",
			"a (cat|dog) and (cat|dog)s",
			true,
		},
		{
			"another-single-backreference",
			"this starts and ends with this",
			"^(\\w+) starts and ends with \\1$",
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _, _ := matchLine([]byte(tt.input), tt.pattern)
			if result != tt.expected {
				t.Errorf("incorrect result for %v, %v", string(tt.input), tt.pattern)
			}
		})
	}

}
