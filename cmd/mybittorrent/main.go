package main

import (
	// Uncomment this line to pass the first stage
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"unicode"
	// bencode "github.com/jackpal/bencode-go" // Available if you need it!
)

// Example:
// - 5:hello -> hello
// - 10:hello12345 -> hello12345
func decodeBencode(bencodedString string) (interface{}, error) {
	firstDigit := rune(bencodedString[0])

	if unicode.IsDigit(firstDigit) {
		result, _, err := decodeBencodedString(bencodedString)
		return result, err
	} else if firstDigit == 'i' {
		result, _, err := decodeBencodedInt(bencodedString)
		return result, err
	} else if firstDigit == 'l' {
		result, _, err := decodeBencodeList(bencodedString)
		return result, err
	} else if firstDigit == 'd' {
		result, _, err := decodeBencodeDict(bencodedString)
		return result, err
	} else {
		return "", fmt.Errorf("unrecognized format")
	}
}

func decodeBencodeDict(bencodedString string) (interface{}, int, error) {
	result := make(map[string]interface{})
	keyMode := true

	l := len(bencodedString)
	currentIdx := 1
	var firstRune rune
	var innerCount int
	var innerRes interface{}
	var recentKey string
	var previousKey string
	var err error
	for currentIdx < l {
		firstRune = rune(bencodedString[currentIdx])
		if unicode.IsDigit(firstRune) {
			innerRes, innerCount, err = decodeBencodedString(bencodedString[currentIdx:l])
			if err != nil {
				return nil, 0, err
			}
		} else if firstRune == 'i' {
			if keyMode {
				return nil, 0, fmt.Errorf("invalid key type")
			}
			innerRes, innerCount, err = decodeBencodedInt(bencodedString[currentIdx:l])
			if err != nil {
				return nil, 0, err
			}
		} else if firstRune == 'l' {
			if keyMode {
				return nil, 0, fmt.Errorf("invalid key type")
			}
			// return nested list and its count
			innerRes, innerCount, err = decodeBencodeList(bencodedString[currentIdx:l])
			if err != nil {
				return nil, 0, err
			}
		} else if firstRune == 'd' {
			if keyMode {
				return nil, 0, fmt.Errorf("invalid key type")
			}
			innerRes, innerCount, err = decodeBencodeDict(bencodedString[currentIdx:l])
			if err != nil {
				return nil, 0, err
			}
		} else {
			// if anything else is found, the provided string is not an exact match of the element
			// so stop the parsing here
			break
		}
		// result = append(result, innerRes)
		// currentIdx += innerCount
		if keyMode {
			previousKey = recentKey
			recentKey = innerRes.(string)
			if previousKey > recentKey {
				return nil, 0, fmt.Errorf("keys not lexicographically sorted")
			}
			keyMode = false
		} else {
			result[recentKey] = innerRes
			keyMode = true
		}
		currentIdx += innerCount
	}

	// check if list ends with an 'e'
	if bencodedString[currentIdx] != 'e' {
		return nil, 0, fmt.Errorf("invalid list format")
	}

	// currentIdx+1 will show the true length of the string-encoded list just parsed
	return result, currentIdx + 1, nil
}

func decodeBencodeList(bencodedString string) (interface{}, int, error) {
	// parse integer or string
	// move carret
	// check location afterwards to make sure it's an 'e'
	// return length of list string
	result := make([]interface{}, 0)

	l := len(bencodedString)
	currentIdx := 1

	var firstRune rune
	var innerCount int
	var innerRes interface{}
	var err error
	for currentIdx < l {
		firstRune = rune(bencodedString[currentIdx])
		if unicode.IsDigit(firstRune) {
			innerRes, innerCount, err = decodeBencodedString(bencodedString[currentIdx:l])
			if err != nil {
				return nil, 0, err
			}
		} else if firstRune == 'i' {
			innerRes, innerCount, err = decodeBencodedInt(bencodedString[currentIdx:l])
			if err != nil {
				return nil, 0, err
			}
		} else if firstRune == 'l' {
			// return nested list and its count
			innerRes, innerCount, err = decodeBencodeList(bencodedString[currentIdx:l])
			if err != nil {
				return nil, 0, err
			}
		} else if firstRune == 'd' {
			innerRes, innerCount, err = decodeBencodeDict(bencodedString[currentIdx:l])
			if err != nil {
				return nil, 0, err
			}
		} else {
			// if anything else is found, the provided string is not an exact match of the element
			// so stop the parsing here
			break
		}
		result = append(result, innerRes)
		currentIdx += innerCount
	}

	// check if list ends with an 'e'
	if bencodedString[currentIdx] != 'e' {
		return nil, 0, fmt.Errorf("invalid list format")
	}

	// currentIdx+1 will show the true length of the string-encoded list just parsed
	return result, currentIdx + 1, nil
}

func decodeBencodedInt(bencodedString string) (interface{}, int, error) {
	l := len(bencodedString)
	// find the ending e
	foundIdx := 0
	for foundIdx < l {
		if bencodedString[foundIdx] == 'e' {
			break
		}
		foundIdx++
	}
	if foundIdx == l {
		return 0, 0, fmt.Errorf("invaid integer format")
	}
	first := bencodedString[1]
	numPart := bencodedString[1:foundIdx]
	num, err := strconv.Atoi(numPart)
	if err != nil {
		return 0, 0, err
	}
	if first == '-' {
		if num == 0 {
			return 0, 0, fmt.Errorf("negative zero not allowed")
		}
		first = bencodedString[2]
	}
	if first == '0' && num != 0 || num == 0 && l != 3 {
		// catching the leading zeros except for exactly '0'
		return 0, 0, fmt.Errorf("leading zeros are not allowed")
	}
	return num, foundIdx + 1, nil
}

func decodeBencodedString(bencodedString string) (interface{}, int, error) {
	if unicode.IsDigit(rune(bencodedString[0])) {
		var firstColonIndex int

		for i := 0; i < len(bencodedString); i++ {
			if bencodedString[i] == ':' {
				firstColonIndex = i
				break
			}
		}

		lengthStr := bencodedString[:firstColonIndex]

		length, err := strconv.Atoi(lengthStr)
		if err != nil {
			return "", 0, err
		}

		if (firstColonIndex + 1 + length) > len(bencodedString) {
			return "", 0, fmt.Errorf("provided length mismatch")
		}
		return bencodedString[firstColonIndex+1 : firstColonIndex+1+length], length + 1 + firstColonIndex, nil
	} else {
		return "", 0, fmt.Errorf("only strings are supported at the moment")
	}
}

func main() {

	command := os.Args[1]

	if command == "decode" {
		// Uncomment this block to pass the first stage

		bencodedValue := os.Args[2]

		decoded, err := decodeBencode(bencodedValue)
		if err != nil {
			fmt.Println(err)
			return
		}

		jsonOutput, _ := json.Marshal(decoded)
		fmt.Println(string(jsonOutput))
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
