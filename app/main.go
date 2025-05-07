package main

import (
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

	if patternComponents[0].b == '^' {
		patternComponents = patternComponents[1:]
		line = line[:len(patternComponents)]
	}
	if patternComponents[len(patternComponents)-1].b == '$' {
		patternComponents = patternComponents[:len(patternComponents)-1]
		startingIndex := len(line) - len(patternComponents)
		line = line[startingIndex:]
	}

	matchedPatternIndex := 0
	var currentCharacterToMatch patternByte
	var previousCharacterToMatch patternByte
	var nextLineCharacter int
	lineLoop: 
		for i, b := range line {
			if i < nextLineCharacter {
				continue
			}
			previousCharacterToMatch = patternByte{}
			if patternComponents[matchedPatternIndex].qualifier == zeroOrMore {
				matchedPatternIndex += 1
			}
			if matchedPatternIndex == len(patternComponents) && matchFound {
				return true, nil
			}

			currentCharacterToMatch = patternComponents[matchedPatternIndex]
			if (matchedPatternIndex != 0) && (patternComponents[matchedPatternIndex-1].isRepeated()) {
				previousCharacterToMatch = patternComponents[matchedPatternIndex-1]
				if (previousCharacterToMatch.b == currentCharacterToMatch.b) && (currentCharacterToMatch.qualifier == noQualifier) {
					currentCharacterToMatch.qualifier = previousCharacterToMatch.qualifier
					patternComponents[matchedPatternIndex] = currentCharacterToMatch
				}
			}

			// check for match with current character
			if currentCharacterToMatch.isMatch(b) {
				matchFound = true
				matchedPatternIndex += 1
			} else if previousCharacterToMatch.isMatch(b) {
				matchFound = true
			} else if currentCharacterToMatch.subPatterns != nil {
				for _, pat := range currentCharacterToMatch.subPatterns {
					if match, err := matchLine(line[i:], "^"+pat); err != nil {
						return false, err
					} else if match {
						matchFound = true
						matchedPatternIndex += 1
						nextLineCharacter = i + len(pat)
						if matchedPatternIndex == len(patternComponents) && matchFound {
							return true, nil
						}
						continue lineLoop
					}
				}
			} else {
				matchFound = false
				matchedPatternIndex = 0
			}
			if matchedPatternIndex == len(patternComponents) && matchFound {
				return true, nil
			}
			nextLineCharacter = i + 1
		}

	return false, nil
}

func parsePattern(pattern string) []patternByte {
	output := make([]patternByte, 0)
	currentCharacter := patternByte{}
	inside := false
	for i := range pattern {
		if pattern[i] == ')' || pattern[i] == ']' {
			inside = false
			continue
		}
		if inside {
			continue
		}
		switch pattern[i] {
		case byte(escaped):
			currentCharacter.qualifier = qualifier(pattern[i])
		case byte(repeated), byte(zeroOrMore):
			if len(output) > 0 {
				output[len(output)-1].qualifier = qualifier(pattern[i])
			}
		case '[':
			idx := strings.IndexByte(pattern[i:], ']')
			if idx != -1 {
				if pattern[i+1] == '^' {
					currentCharacter.qualifier = not
				}
				currentCharacter.bytes = []byte(pattern[i+1 : i+idx])
			}
			inside = true
			output = append(output, currentCharacter)
			currentCharacter = patternByte{}
		case '(':
			idx := strings.IndexByte(pattern[i:], ')')
			if idx != -1 {
				currentCharacter.subPatterns = strings.Split(pattern[i+1:i+idx], "|")
			}
			inside = true
			output = append(output, currentCharacter)
			currentCharacter = patternByte{}
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
	not         qualifier = '^'
)

type patternByte struct {
	b           byte
	qualifier   qualifier
	subPatterns []string
	bytes       []byte
}

func isInt(b byte) bool {
	return b >= '0' && b <= '9'
}

func isAlphanumeric(b byte) bool {
	return (b >= '0' && b <= '9') || (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z')
}

func (v patternByte) isMatch(b byte) (isMatch bool) {
	if v.b != 0 {
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
			if v.b == '.' {
				return true
			}
			return v.b == b
		}
	}
	if v.bytes != nil {
		byteEqual := false
		for _, c := range v.bytes {
			if c == b {
				byteEqual = true
			}
		}
		return (v.qualifier != not) == byteEqual
	}
	return false
}

func (v patternByte) isRepeated() bool {
	return v.qualifier == zeroOrMore || v.qualifier == repeated
}
