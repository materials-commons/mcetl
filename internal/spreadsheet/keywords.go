package spreadsheet

/*
 * keywords contains the keyword identifier for different attributes. A keyword
 * is added to a header cell to identify the attribute type. For example:
 *    file:Measurements
 * In the above example the file: is the keyword and Measurements tells the user
 * more about what the file contains.
 */

import (
	"fmt"
	"strings"
)

// Default set of keywords for sample attributes
var SampleAttributeKeywords = map[string]bool{
	"s":                true,
	"sample":           true,
	"sample attribute": true,
}

// Default set of keywords for process attributes
var ProcessAttributeKeywords = map[string]bool{
	"p":       true,
	"process": true,
}

// Default set of keywords for file attributes
var FileAttributeKeywords = map[string]bool{
	"f":     true,
	"file":  true,
	"files": true,
}

func hasSampleAttributeKeyword(cell string) bool {
	return findIn(cell, SampleAttributeKeywords)
}

func hasProcessAttributeKeyword(cell string) bool {
	return findIn(cell, ProcessAttributeKeywords)
}

func hasFileAttributeKeyword(cell string) bool {
	return findIn(cell, FileAttributeKeywords)
}

func columeAttributeTypeFromKeyword(cell string) ColumnAttributeType {
	// If you add a new Attribute Keyword then don't forget to update
	// processHeaderRow() and processSampleRow() case statements in
	// loader.go to handle those new keywords.
	if hasProcessAttributeKeyword(cell) {
		return ProcessAttributeColumn
	}

	if hasSampleAttributeKeyword(cell) {
		return SampleAttributeColumn
	}

	if hasFileAttributeKeyword(cell) {
		return FileAttributeColumn
	}

	return UnknownAttributeColumn
}

func findIn(cell string, keywords map[string]bool) bool {
	cell = strings.ToLower(cell)
	i := strings.Index(cell, ":")
	if i == -1 {
		return false
	}

	keyword := cell[:i]
	_, ok := keywords[keyword]
	return ok
}

func AddSampleKeyword(keyword string) {
	SampleAttributeKeywords[keyword] = true
}

func SetSampleKeywords(keywords ...string) {
	// Clear SampleAttributeKeywords
	SampleAttributeKeywords = make(map[string]bool)

	// Add new set of keywords
	for _, keyword := range keywords {
		SampleAttributeKeywords[keyword] = true
	}
}

func AddProcessKeyword(keyword string) {
	ProcessAttributeKeywords[keyword] = true
}

func SetProcessKeywords(keywords ...string) {
	// Clear ProcessAttributeKeywords
	ProcessAttributeKeywords = make(map[string]bool)

	// Add new set of keywords
	for _, keyword := range keywords {
		ProcessAttributeKeywords[keyword] = true
	}
}

func AddFileKeyword(keyword string) {
	FileAttributeKeywords[keyword] = true
}

func SetFileKeywords(keywords ...string) {
	// Clear FileAttributeKeywords
	FileAttributeKeywords = make(map[string]bool)

	// Add new set of keywords
	for _, keyword := range keywords {
		FileAttributeKeywords[keyword] = true
	}
}

func ValidateKeywords() error {
	switch {
	case len(ProcessAttributeKeywords) == 0:
		return fmt.Errorf("there must be at least 1 process keyword")
	case len(SampleAttributeKeywords) == 0:
		return fmt.Errorf("there must be at least 1 sample keyword")
	case len(FileAttributeKeywords) == 0:
		return fmt.Errorf("there must be at least 1 file keyword")
	case overlappingKeywords():
		return fmt.Errorf("there are overlapping keywords")
	}
	return nil
}

// overlappingKeywords returns true if a keyword occurs in more than one attribute keywords list.
func overlappingKeywords() bool {
	keywordCounts := make(map[string]int)

	// Load count of keywords for each of the attribute keyword lists

	for key := range ProcessAttributeKeywords {
		if count, ok := keywordCounts[key]; !ok {
			keywordCounts[key] = 1
		} else {
			count++
			keywordCounts[key] = count
		}
	}

	for key := range SampleAttributeKeywords {
		if count, ok := keywordCounts[key]; !ok {
			keywordCounts[key] = 1
		} else {
			count++
			keywordCounts[key] = count
		}
	}

	for key := range FileAttributeKeywords {
		if count, ok := keywordCounts[key]; !ok {
			keywordCounts[key] = 1
		} else {
			count++
			keywordCounts[key] = count
		}
	}

	// Check if any of the keyword counts is greater than 1. If it is then that
	// keyword is used in multiple lists.
	foundError := false
	for key := range keywordCounts {
		count, _ := keywordCounts[key]
		if count != 1 {
			fmt.Printf("Keyword '%s' repeated in multiple attribute keyword identifiers\n", key)
			foundError = true
		}
	}

	return foundError
}
