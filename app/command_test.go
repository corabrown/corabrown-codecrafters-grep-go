package main

import (
	"testing"
)

func TestOneRepeat(t *testing.T) {
	str := []byte("cat")
	pattern := "ca+t"
	result, err, _ := matchLine([]byte("cat"), "ca+t")
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
