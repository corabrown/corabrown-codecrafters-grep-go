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

	ok, err, _ := matchLine(line, pattern)
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

func matchLine(line []byte, pattern string) (bool, error, int) {

	matchFound := false
	p := parsePattern(pattern)

	if len(p.patternComponents) == 0 {
		return true, nil, 0 
	}

	matchBeginning := false 
	if p.currentComponent().b == '^' {
		p.patternComponents = p.patternComponents[1:]
		matchBeginning = true 
	}
	matchEnd := false 
	if p.lastComponent().b == '$' {
		p.patternComponents = p.patternComponents[:len(p.patternComponents)-1]
		matchEnd = true 
	}

	var nextLineCharacter int
	for i := range line {
		l := string(line[i])
		_ = l 
		if i < nextLineCharacter {
			continue
		}
		if p.currentIndex == len(p.patternComponents) {
			if matchEnd && i != len(line) - 1 {
				return false, nil, 1 
			}
			return matchFound, nil, i 
		}
		if p.currentComponent().hasEnoughMatches() {
			p.currentIndex += 1 
		}

		var patternLength int 
		var ok bool 
		matchFound, patternLength = p.currentComponent().isMatch(line[i:])  
		if !matchFound {
			if ok, patternLength = p.previousRepeatedComponent().isMatch(line[i:]); ok {
				matchFound = true 
			} else {
				if matchBeginning {
					return false, nil, 0 
				}
				p.currentIndex = 0 
			}
		}

		if p.currentComponent().hasEnoughMatches() {
			p.currentIndex += 1 
		}
		if p.currentIndex == len(p.patternComponents) {
			return matchFound, nil, i
		}
		nextLineCharacter = i + patternLength
	}

	return false, nil, 0
}


type fullPattern struct {
	patternComponents []patternSegment
	currentIndex int 
}

func (v fullPattern) currentComponent() *patternSegment {
	if v.currentIndex > len(v.patternComponents) - 1 {
		return nil 
	}
	return &v.patternComponents[v.currentIndex]
}

func (v fullPattern) previousRepeatedComponent() *patternSegment {
	if v.currentIndex == 0 {
		return nil 
	}
	return &v.patternComponents[v.currentIndex-1]
}

func (v fullPattern) lastComponent() *patternSegment {
	if len(v.patternComponents) == 0 {
		return nil 
	}
	return &v.patternComponents[len(v.patternComponents)-1]
}


func parsePattern(pattern string) fullPattern {
	output := make([]patternSegment, 0)
	currentCharacter := patternSegment{}
	inside := false
mainPatternLoop:
	for i := range pattern {
		if pattern[i] == ')' || pattern[i] == ']' {
			inside = false
			continue
		}
		if inside {
			continue
		}
		if len(output) > 0 && output[len(output)-1].isRepeated() && (output[len(output)-1].b == pattern[i]) {
			if output[len(output) - 1].qualifier == repeated {
				output[len(output) - 1].matchesRequired += 1
			} 
			continue 
		}
		switch pattern[i] {
		case byte('\\'):
			currentCharacter.escaped = true 
		case byte(repeated), byte(zeroOrMore):
			if len(output) > 0 {
				output[len(output)-1].qualifier = qualifier(pattern[i])
				if pattern[i] == byte(zeroOrMore) {
					output[len(output)-1].matchesRequired = 0 
				}
			} 
		case '[':
			idx := strings.IndexByte(pattern[i:], ']')
			if idx != -1 {
				if pattern[i+1] == '^' {
					currentCharacter.negativeMatch = true
				}
				currentCharacter.bytes = []byte(pattern[i+1 : i+idx])
			}
			inside = true
			currentCharacter.matchesRequired += 1
			output = append(output, currentCharacter)
			currentCharacter = patternSegment{}
		case '(':
			idx := strings.IndexByte(pattern[i:], ')')
			if idx != -1 {
				currentCharacter.subPatterns = strings.Split(pattern[i+1:i+idx], "|")
			}
			currentCharacter.matchesRequired += 1
			inside = true
			output = append(output, currentCharacter)
			currentCharacter = patternSegment{}
		default:
			if currentCharacter.escaped {
				if pattern[i] == '1' {
					for _, pat := range output {
						if pat.subPatterns != nil {
							output = append(output, pat)
							currentCharacter = patternSegment{}
							continue mainPatternLoop
						}
					}
				}
				if pattern[i] != 'd' && pattern[i] != 'w' {
					output = append(output, patternSegment{b: '\\'})
					currentCharacter.qualifier = noQualifier
				}
			}
			currentCharacter.b = pattern[i]
			currentCharacter.s = string(pattern[i])
			currentCharacter.matchesRequired += 1 
			output = append(output, currentCharacter)
			currentCharacter = patternSegment{}
		}
	}
	return fullPattern{patternComponents: output, currentIndex: 0}
}

type qualifier byte

const (
	repeated    qualifier = '+'
	zeroOrMore  qualifier = '?'
	noQualifier qualifier = 0
	not         qualifier = '^'
)

type patternSegment struct {
	s           string
	b           byte
	qualifier   qualifier
	subPatterns []string
	bytes       []byte
	matchesRequired int 
	matchesFound int 
	negativeMatch bool 
	escaped bool 
}

func isInt(b byte) bool {
	return b >= '0' && b <= '9'
}

func isAlphanumeric(b byte) bool {
	return (b >= '0' && b <= '9') || (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z')
}


func (v *patternSegment) isMatch(line []byte) (matchFound bool, patternLength int) {
	if v == nil {
		return false, 1  
	}
	m, i := v.match(line)
	if m {
		v.matchesFound += 1 
	}

	return m, i
}

func (v *patternSegment) hasEnoughMatches() bool {
	return v.matchesFound == v.matchesRequired
}

func (v *patternSegment) match(line []byte) (matchFound bool, patternLength int) {
	if v == nil {
		return false, 1
	}
	if len(line) == 0 {
		return true, 1
	}
	b := line[0] 

	if v.b != 0 {

		if v.escaped {
			switch v.b {
			case 'd':
				return isInt(b), 1
			case 'w':
				return isAlphanumeric(b), 1
			default:
			}
		}

		switch v.qualifier {
		case zeroOrMore:
			if v.b == b {
				return true, 1
			}
			return false, 1
		default:
			if v.b == '.' {
				return true, 1
			}
			return v.b == b, 1  
		}
	}
	if v.bytes != nil {
		byteEqual := false
		for _, c := range v.bytes {
			if c == b {
				byteEqual = true
				break
			}
		}
		return v.negativeMatch != byteEqual, 1 
	}
	if v.subPatterns != nil {
		for _, pat := range v.subPatterns {
			if len(line) > 1 {
				pat = strings.Replace(pat, "$", "", 1)
			}
			ok, err, matchedLineLength := matchLine(line, "^" + pat)
			if err != nil {
				panic("an error")
			}
			if ok {
				return true, matchedLineLength + 1
			} 
		}
		return false, 1 
	}
	return false, 1 
}

func (v *patternSegment) isRepeated() bool {
	if v == nil {
		return false 
	}
	return v.qualifier == zeroOrMore || v.qualifier == repeated
}

