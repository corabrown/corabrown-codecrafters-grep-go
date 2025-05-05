package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
)

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

	matchFound := false
	patternComponents := parsePattern(pattern)

	var matchBeginningOfString bool
	if patternComponents[0].b == '^' {
		matchBeginningOfString = true
		patternComponents = patternComponents[1:]
	}
	if patternComponents[len(patternComponents)-1].b == '$' {
		patternComponents = patternComponents[:len(patternComponents)-1]
		startingIndex := len(line) - len(patternComponents)
		line = line[startingIndex:]
	}

	matchedPatternIndex := 0
	possibleMatches := make(map[patternByte]struct{})

	for _, b := range line {
		possibleMatches[patternComponents[matchedPatternIndex]] = struct{}{}
		if patternComponents[matchedPatternIndex].isRepeated() {
			if len(patternComponents) > matchedPatternIndex + 1 {
				possibleMatches[patternComponents[matchedPatternIndex + 1]] = struct{}{}
			} else if patternComponents[matchedPatternIndex].qualifier == zeroOrMore {
				return true, nil 
			}
		}

		byteMatched := false 
		for patternByte := range possibleMatches {
			if patternByte.isMatch(b) {
				byteMatched = true 
				if !patternByte.isRepeated() {
					matchedPatternIndex += 1 
				}
				if patternByte.qualifier != repeated {
					delete(possibleMatches, patternByte)
				}
				break
			} else if patternByte.qualifier == repeated {
				matchedPatternIndex += 1
				delete(possibleMatches, patternByte)
			}

		}
		if byteMatched {
			matchFound = true 
		} else {
			if matchBeginningOfString {
				return false, nil 
			}
			matchFound = false 
			matchedPatternIndex = 0 
			possibleMatches = make(map[patternByte]struct{})
		}
		if matchedPatternIndex == len(patternComponents) && matchFound {
			return true, nil 
		}
	}

	if strings.HasPrefix(pattern, "[^") && strings.HasSuffix(pattern, "]") {
		for _, b := range line {
			if !bytes.ContainsAny([]byte{b}, pattern[2:len(pattern)-1]) {
				return true, nil
			}
		}
		return false, nil
	}
	if strings.HasPrefix(pattern, "[") && strings.HasSuffix(pattern, "]") {
		for _, r := range pattern[1 : len(pattern)-1] {
			if ok := bytes.ContainsAny(line, string(r)); ok {
				return ok, nil
			}
		}
	}

	return false, nil
}

func parsePattern(pattern string) []patternByte {
	output := make([]patternByte, 0)
	currentCharacter := patternByte{}
	for i := range pattern {
		switch pattern[i] {
		case byte(escaped):
			currentCharacter.qualifier = qualifier(pattern[i])
		case byte(repeated), byte(zeroOrMore):
			if len(output) > 0 {
				output[len(output)-1].qualifier = qualifier(pattern[i])
			}
		default:
			if currentCharacter.qualifier == escaped {
				if pattern[i] != 'd' && pattern[i] != 'w' {
					output = append(output, patternByte{b: '\\'})
					currentCharacter.qualifier = noQualifier
				}
			}
			currentCharacter.b = pattern[i]
			output = append(output, currentCharacter)
			currentCharacter = patternByte{}
		}
	}
	return output
}

type qualifier byte

const (
	escaped     qualifier = '\\'
	repeated    qualifier = '+'
	zeroOrMore  qualifier = '?'
	noQualifier qualifier = 0
)

type patternByte struct {
	b         byte
	qualifier qualifier
}

func isInt(b byte) bool {
	return b >= '0' && b <= '9'
}

func isAlphanumeric(b byte) bool {
	return (b >= '0' && b <= '9') || (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z')
}

func (v patternByte) isMatch(b byte) (isMatch bool) {
	switch v.qualifier {
	case escaped:
		switch v.b {
		case 'd':
			return isInt(b)
		case 'w':
			return isAlphanumeric(b)
		default:
		}
	case zeroOrMore:
		return true 
	default:
		return v.b == b
	}
	return false
}

func (v patternByte) isRepeated() bool {
	return v.qualifier == zeroOrMore || v.qualifier == repeated
}
