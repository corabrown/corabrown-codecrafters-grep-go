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
	if patternComponents[0][0] == '^'{
		matchBeginningOfString = true 
		patternComponents = patternComponents[1:]
	}



	reset := false 
	matchedPatternIndex := 0 
	for _, b := range line {
		switch string(patternComponents[matchedPatternIndex]) {
		case "\\d":
			if (b >= '0' && b <= '9') {
				matchFound = true 
				matchedPatternIndex += 1 
			} else {
				reset = true 
			}
		case "\\w":
			if (b >= '0' && b <= '9') || (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z') {
				matchFound = true 
				matchedPatternIndex += 1 
			} else {
				reset = true 
			}
		default: 
			if b == patternComponents[matchedPatternIndex][0] {
				matchFound = true 
				matchedPatternIndex += 1 
			} else {
				reset = true
			}
		}	
		if (matchedPatternIndex == len(patternComponents)) && matchFound {
			return true, nil 
		}
		
		if reset {
			if matchBeginningOfString {
				return false, nil 
			}
			matchFound = false
			matchedPatternIndex = 0 
			reset = false 
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
		for _, r := range pattern[1:len(pattern)-1] {
			if ok := bytes.ContainsAny(line, string(r)); ok {
				return ok, nil 
			}
		}
	}
	
	return false, nil
}

func parsePattern(pattern string) [][]byte {
	output := make([][]byte, 0)
	currentCharacter := ""
	for i := range pattern {
		if pattern[i] == byte('\\') {
			currentCharacter = "\\"
			continue 
		}
		if currentCharacter != "" {
			if (pattern[i] == 'd') || (pattern[i] == 'w') {
				output = append(output, []byte(fmt.Sprintf("\\%v", string(pattern[i]))))
			} else {
				output = append(output, []byte{'\\'}, []byte{pattern[i]})
			}
			currentCharacter = ""
		} else {
			output = append(output, []byte{pattern[i]})
		}
	}
	return output 
}