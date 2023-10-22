package bencode

import (
	"encoding/json"
	"fmt"
	"strconv"
	"unicode"
)

// Example:
// - 5:hello -> hello
// - 10:hello12345 -> hello12345
func DecodeBencode(bencodedString string) (interface{}, error) {
	firstDigit := rune(bencodedString[0])

	var result interface{}
	var err error

	if unicode.IsDigit(firstDigit) {
		result, _, err = decodeBencodedString(bencodedString)
	} else {
		switch firstDigit {
		case 'i':
			result, _, err = decodeBencodedInt(bencodedString)
		case 'l':
			result, _, err = decodeBencodedList(bencodedString, true)
		case 'd':
			result, _, err = decodeBencodedDict(bencodedString, true)
		default:
			return "", fmt.Errorf("unrecognized format")
		}
	}

	return result, err
}

func decodeBencodedDict(bencodedString string, isTopLevel bool) (interface{}, int, error) {
	result := make(map[string]interface{})
	keyMode := true

	l := len(bencodedString)
	currentIdx := 1
	var firstRune rune
	var innerCount int
	var innerRes interface{}
	var recentKey string
	// var previousKey string
	var err error
LOOP:
	for currentIdx < l {
		firstRune = rune(bencodedString[currentIdx])

		if keyMode && (firstRune == 'i' || firstRune == 'l' || firstRune == 'd') {
			return nil, 0, fmt.Errorf("invalid key type")
		}

		if unicode.IsDigit(firstRune) {
			innerRes, innerCount, err = decodeBencodedString(bencodedString[currentIdx:l])
		} else {
			switch firstRune {
			case 'i':
				innerRes, innerCount, err = decodeBencodedInt(bencodedString[currentIdx:l])
			case 'l':
				// return nested list and its count
				innerRes, innerCount, err = decodeBencodedList(bencodedString[currentIdx:l], false)
			case 'd':
				innerRes, innerCount, err = decodeBencodedDict(bencodedString[currentIdx:l], false)
			default:
				// if anything else is found, the provided string is not an exact match of the element
				// so stop the parsing here
				break LOOP
			}
		}

		if err != nil {
			return nil, 0, err
		}
		// result = append(result, innerRes)
		// currentIdx += innerCount
		if keyMode {
			// previousKey = recentKey
			recentKey = innerRes.(string)
			// Removed the lexicographical sorting requirement because tracker did not always respond like that
			// if previousKey > recentKey {
			// 	return nil, 0, fmt.Errorf("keys not lexicographically sorted")
			// }
			keyMode = false
		} else {
			result[recentKey] = innerRes
			keyMode = true
		}
		currentIdx += innerCount
	}

	// check if list ends with an 'e'
	if bencodedString[currentIdx] != 'e' {
		return nil, 0, fmt.Errorf("invalid dict format")
	}

	// if the dict is top level, the length must be checked to avoid extra trailing chars
	if isTopLevel && currentIdx+1 != l {
		return nil, 0, fmt.Errorf("invalid dict format (extra chars present)")
	}

	// currentIdx+1 will show the true length of the string-encoded list just parsed
	return result, currentIdx + 1, nil
}

func decodeBencodedList(bencodedString string, isTopLevel bool) (interface{}, int, error) {
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
LOOP:
	for currentIdx < l {
		firstRune = rune(bencodedString[currentIdx])
		if unicode.IsDigit(firstRune) {
			innerRes, innerCount, err = decodeBencodedString(bencodedString[currentIdx:l])
		} else {
			// if firstRune == 'i' {
			// 	innerRes, innerCount, err = decodeBencodedInt(bencodedString[currentIdx:l])
			// } else if firstRune == 'l' {
			// 	// return nested list and its count
			// 	innerRes, innerCount, err = decodeBencodedList(bencodedString[currentIdx:l], false)
			// } else if firstRune == 'd' {
			// 	innerRes, innerCount, err = decodeBencodedDict(bencodedString[currentIdx:l], false)
			// } else {
			// 	// if anything else is found, the provided string is not an exact match of the element
			// 	// so stop the parsing here
			// 	break
			// }
			switch firstRune {
			case 'i':
				innerRes, innerCount, err = decodeBencodedInt(bencodedString[currentIdx:l])
			case 'l':
				// return nested list and its count
				innerRes, innerCount, err = decodeBencodedList(bencodedString[currentIdx:l], false)
			case 'd':
				innerRes, innerCount, err = decodeBencodedDict(bencodedString[currentIdx:l], false)
			default:
				// if anything else is found, the provided string is not an exact match of the element
				// so stop the parsing here
				// break with label used to break from the loop, not the auto break from the case
				break LOOP
			}
		}

		if err != nil {
			return nil, 0, err
		}

		result = append(result, innerRes)
		currentIdx += innerCount
	}

	// check if list ends with an 'e'
	if bencodedString[currentIdx] != 'e' {
		return nil, 0, fmt.Errorf("invalid list format")
	}

	// if the list is top level, the length must be checked to avoid extra trailing chars
	if isTopLevel && currentIdx+1 != l {
		return nil, 0, fmt.Errorf("invalid list format (extra chars present)")
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
		fmt.Println(bencodedString)
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

func DecodeBencodeToJSON(bencodedString string) (string, error) {
	decoded, err := DecodeBencode(bencodedString)
	if err != nil {
		return "", err
	}

	jsonOutput, err := json.Marshal(decoded)
	return string(jsonOutput), err
}
