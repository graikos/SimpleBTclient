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
		return decodeBencodeString(bencodedString)
	} else if firstDigit == 'i' {
		return decodeBencodedInt(bencodedString)
	} else if firstDigit == 'l' {
		return decodeBencodeList(bencodedString)
	} else {
		return "", fmt.Errorf("unrecognized format")
	}
}

func decodeBencodeList(bencodedString string) ([]interface{}, error) {
	// check if last is 'e'
	l := len(bencodedString)
	if bencodedString[l-1] != 'e' {
		return nil, fmt.Errorf("invalid list format")
	}
	var result []interface{}
	currentIdx := 1
	var firstRune rune
	for (l-1)-currentIdx > 0 {
		firstRune = rune(bencodedString[currentIdx])
		if unicode.IsDigit(firstRune) {
			count, err := strconv.Atoi(string(firstRune))
			if err != nil {
				return nil, err
			}
			// decoding the string using the previous decoder
			res, err := decodeBencode(bencodedString[currentIdx : currentIdx+2+count])
			if err != nil {
				return nil, err
			}
			result = append(result, res)
			// updating the current idx to point after the string
			currentIdx += (2 + count)
		} else if firstRune == 'i' {
			// look through to find the ending
			foundIdx := currentIdx
			// don't take into account the last 'e'
			for foundIdx < l-1 {
				if bencodedString[foundIdx] == 'e' {
					break
				}
				foundIdx++
			}
			// this means it was not found
			if foundIdx == l-1 {
				return nil, fmt.Errorf("invalid integer format")
			}
			res, err := decodeBencode(bencodedString[currentIdx : foundIdx+1])
			if err != nil {
				return nil, err
			}
			result = append(result, res)
			currentIdx = foundIdx + 1
		}
	}
	return result, nil

}

func decodeBencodedInt(bencodedString string) (interface{}, error) {
	l := len(bencodedString)
	if bencodedString[l-1] != 'e' {
		return 0, fmt.Errorf("invalid integer format")
	}
	sign := bencodedString[1]
	numPart := bencodedString[1 : l-1]
	num, err := strconv.Atoi(numPart)
	if err != nil {
		return 0, err
	}
	if sign == '-' && num == 0 {
		return 0, fmt.Errorf("negative zero not allowed")
	} else if sign == '0' && num != 0 || num == 0 && l != 3 {
		// catching the leading zeros except for exactly '0'
		return 0, fmt.Errorf("leading zeros are not allowed")
	}
	return num, nil
}

func decodeBencodeString(bencodedString string) (interface{}, error) {
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
			return "", err
		}

		return bencodedString[firstColonIndex+1 : firstColonIndex+1+length], nil
	} else {
		return "", fmt.Errorf("Only strings are supported at the moment")
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
