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

func matchLine(line []byte, pattern string) (matchFound bool, err error, matchLength int) {

	// defer func() {
	// 	fmt.Println(string(line), pattern, matchFound)
	// }()

	p := parsePattern(pattern)

	if len(p.patternComponents) == 0 {
		return true, nil, 0
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
			p.matchFoundForEverySegment = true
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
				if p.matchBeginning {
					return false, nil, 0
				}
				if p.matchFoundForEverySegment && matchFound && ((!matchEnd) || i == len(line)-1) {
					return true, nil, i
				}
				p.currentIndex = 0
			}
		}

		if p.currentComponent().hasEnoughMatches() {
			p.currentIndex += 1
		}
		if p.currentIndex == len(p.patternComponents) {
			p.matchFoundForEverySegment = true
		}

		if p.matchFoundForEverySegment && matchFound && ((!matchEnd) || i == len(line)-1) {
			return true, nil, i
		}

		nextLineCharacter = i + patternLength
	}

	if p.matchFoundForEverySegment && matchFound {
		return true, nil, len(line)
	}

	return false, nil, 0
}

type fullPattern struct {
	patternComponents         []patternSegment
	currentIndex              int
	matchFoundForEverySegment bool
	matchBeginning bool 
}

func (v fullPattern) currentComponent() *patternSegment {
	if v.currentIndex > len(v.patternComponents)-1 {
		if v.lastComponent().isRepeated() {
			return v.lastComponent()
		}
		return &patternSegment{empty: true}
	}

	
	if v.patternComponents[v.currentIndex].backReference {
		v.patternComponents[v.currentIndex].subPatterns = map[string]bool{v.patternComponents[v.patternComponents[v.currentIndex].previousGroupIndex].m: false}
	}

	return &v.patternComponents[v.currentIndex] 
}

func (v fullPattern) previousRepeatedComponent() *patternSegment {
	if (v.currentIndex == 0) || (v.currentIndex > len(v.patternComponents)) {
		if v.lastComponent().isRepeated() {
			return v.lastComponent()
		}
		return &patternSegment{empty: true}
	}
	return &v.patternComponents[v.currentIndex-1]
}

func (v fullPattern) lastComponent() *patternSegment {
	if len(v.patternComponents) == 0 {
		return &patternSegment{empty: true}
	}
	return &v.patternComponents[len(v.patternComponents)-1]
}

func parsePattern(pattern string) fullPattern {
	output := make([]patternSegment, 0)
	currentCharacter := patternSegment{}
	inside := false
	var fp fullPattern
mainPatternLoop:
	for i := range pattern {
		if (i == 0) && (pattern[i] == '^') {
			fp.matchBeginning = true 
			continue 
		}

		if pattern[i] == ')' || pattern[i] == ']' {
			inside = false
			continue
		}
		if inside {
			continue
		}
		if len(output) > 0 && output[len(output)-1].isRepeated() && (output[len(output)-1].b == pattern[i]) {
			if output[len(output)-1].qualifier == repeated {
				output[len(output)-1].matchesRequired += 1
			}
			continue
		}
		switch pattern[i] {
		case byte('\\'):
			currentCharacter.escaped = true
		case byte(repeated), byte(zeroOrOne):
			if len(output) > 0 {
				output[len(output)-1].qualifier = qualifier(pattern[i])
				if pattern[i] == byte(zeroOrOne) {
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
			currentCharacter.subPatterns = make(map[string]bool)
			if idx != -1 {
				for _, subPat := range strings.Split(pattern[i+1:i+idx], "|") {
					p := parsePattern(subPat)
					currentCharacter.subPatterns[subPat] = p.lastComponent().isRepeated()
				}
			}
			currentCharacter.matchesRequired += 1
			inside = true
			output = append(output, currentCharacter)
			currentCharacter = patternSegment{}
		default:
			if currentCharacter.escaped {
				if pattern[i] == '1' {
					for j, pat := range output {
						if pat.subPatterns != nil {
							output = append(output, patternSegment{
								previousGroupIndex: j,
								backReference: true,
								matchesRequired: 1,
							})
							currentCharacter = patternSegment{}
							continue mainPatternLoop
						}
					}
				}
				if pattern[i] == '2' {
					var oneFound bool
					for j, pat := range output {
						if pat.subPatterns != nil {
							if oneFound {
								output = append(output, patternSegment{
									previousGroupIndex: j,
									backReference: true,
									matchesRequired: 1,
								})
								currentCharacter = patternSegment{}
								continue mainPatternLoop
							}
							oneFound = true
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
	fp.patternComponents = output
	return fp
}

type qualifier byte

const (
	repeated    qualifier = '+'
	zeroOrOne   qualifier = '?'
	noQualifier qualifier = 0
	not         qualifier = '^'
)

type patternSegment struct {
	s                  string
	b                  byte
	qualifier          qualifier
	subPatterns        map[string]bool
	bytes              []byte
	matchesRequired    int
	matchesFound       int
	negativeMatch      bool
	escaped            bool
	empty              bool
	m                  string
	previousGroupIndex int
	backReference      bool
}

func isInt(b byte) bool {
	return b >= '0' && b <= '9'
}

func isAlphanumeric(b byte) bool {
	return (b >= '0' && b <= '9') || (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z')
}

func (v *patternSegment) isMatch(line []byte) (matchFound bool, patternLength int) {
	if v.empty {
		return false, 1
	}

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
		case zeroOrOne:
			if v.matchesFound > 1 {
				return false, 1
			}
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
		for pat, _ := range v.subPatterns {
			if len(line) > 1 {
				pat = strings.Replace(pat, "$", "", 1)
			}
			ok, err, matchedLineLength := matchLine(line, "^"+pat)
			if err != nil {
				panic("an error")
			}
			if ok {
				v.m = v.m + string(line[:matchedLineLength+1])
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

	if v.subPatterns != nil {
		for _, ok := range v.subPatterns {
			if ok {
				return true
			}
		}
	}

	return v.qualifier == zeroOrOne || v.qualifier == repeated
}
