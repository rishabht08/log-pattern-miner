package utils

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

func Tokenize(log string, regexMap map[string]string) []string {

	patterns := make([]string, 0, len(regexMap))
	for pattern := range regexMap {
		patterns = append(patterns, pattern)
	}

	// Iterate through all patterns and replace them
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		log = re.ReplaceAllString(log, regexMap[pattern]) // Remove matched pattern
	}

	// Split remaining log into words
	tokens := TrimToFirst30(strings.Fields(log))

	return tokens

}

func Sha1Hex(input string) string {
	hash := sha1.Sum([]byte(input))
	return hex.EncodeToString(hash[:])
}

// ---------------- MAIN ----------------

func RemoveKeywords(input []string) []string {
	keywords := map[string]bool{
		"Ts[T]": true,
		"time":  true,
		"date":  true,
	}

	var result []string
	for _, word := range input {
		if !keywords[word] {
			result = append(result, word)
		}
	}
	return result
}

func TrimToFirst30(arr []string) []string {
	if len(arr) > 50 {
		return arr[:50]
	}
	return arr
}

func PrintJSon(data interface{}) {
	b, bError := json.Marshal(data)
	if bError != nil {
		fmt.Printf("error marshalling data: %v\n", bError)
		return
	}
	fmt.Println(string(b))
}
