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
	} else {
		return "", fmt.Errorf("unrecognized format")
	}
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
