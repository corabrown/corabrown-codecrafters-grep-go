package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"unicode/utf8"
	"strings"
)

// Ensures gofmt doesn't remove the "bytes" import above (feel free to remove this!)
var _ = bytes.ContainsAny

// Usage: echo <input_text> | your_program.sh -E <pattern>
func main() {
	if len(os.Args) < 3 || os.Args[1] != "-E" {
		fmt.Fprintf(os.Stderr, "usage: mygrep -E <pattern>\n")
		os.Exit(2) // 1 means no lines were selected, >1 means error
	}

	pattern := os.Args[2]

	line, err := io.ReadAll(os.Stdin) // assume we're only dealing with a single line
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: read input text: %v\n", err)
		os.Exit(2)
	}

	ok, err := matchLine(line, pattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}

	if !ok {
		os.Exit(1)
		fmt.Println("not found")
	}
	fmt.Println("found")
}

func matchLine(line []byte, pattern string) (bool, error) {

	switch pattern {
	case "\\d":
		if ok := matchDigit(line); ok {
			return ok, nil
		}
	case "\\w": 
		if ok := matchAlphaNumeric(line); ok {
			return ok, nil 
		}
	}
	if strings.HasPrefix(pattern, "[") && strings.HasSuffix(pattern, "]") {
		for _, r := range pattern[1:len(pattern)-1] {
			if ok := bytes.ContainsAny(line, string(r)); ok {
				return ok, nil 
			}
		}
	}

	if utf8.RuneCountInString(pattern) == 1 {
		if ok := bytes.ContainsAny(line, pattern); ok {
			return ok, nil 
		}
	}
	
	return false, nil
}


func matchDigit(line []byte) bool {
	for _, b := range line {
		if (b >= '0' && b <= '9') {
			return true 
		 }
	}
	return false 
}

func matchAlphaNumeric(line []byte) bool {
	for _, b := range line {
		if (b >= '0' && b <= '9') || (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z') {
			return true 
		 }
	}
	return false 
}