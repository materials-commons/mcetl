package spreadsheet

/*
 * keywords contains the keyword identifier for different attributes. A keyword
 * is added to a header cell to identify the attribute type. For example:
 *    process:Grain Size
 * In the above example the process: is the keyword and Grain Size is the Attribute.
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

// columnAttribyteTypeFromKeyword takes a cell, checks if it has a keyword
// in it and if so returns the keyword type.
func columnAttributeTypeFromKeyword(cell string) ColumnAttributeType {
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

// hasSampleAttributeKeyword return true if the cell contains a keyword
// from the SampleAttributeKeywords.
func hasSampleAttributeKeyword(cell string) bool {
	return hasKeywordInCell(cell, SampleAttributeKeywords)
}

// hasProcessAttributeKeyword returns true if the cell contains
// a keyword from the ProcessAttributesKeywords.
func hasProcessAttributeKeyword(cell string) bool {
	return hasKeywordInCell(cell, ProcessAttributeKeywords)
}

// hasFileAttributeKeyword returns true if the cell contains
// a keyword from the FileAttributesKeywords.
func hasFileAttributeKeyword(cell string) bool {
	return hasKeywordInCell(cell, FileAttributeKeywords)
}

// hasKeywordInCell checks if the cell contains a keyword from the
// given keyword map.
func hasKeywordInCell(cell string, keywords map[string]bool) bool {
	cell = strings.ToLower(cell)
	i := strings.Index(cell, ":")
	if i == -1 {
		return false
	}

	keyword := cell[:i]
	_, ok := keywords[keyword]
	return ok
}

// AddSampleKeyword adds a new keyword to the SampleAttributeKeywords map.
func AddSampleKeyword(keyword string) {
	SampleAttributeKeywords[keyword] = true
}

// SetProcessKeywords overrides the current ProcessAttributeKeywords with the
// new set of keywords. It clears the current set of keywords before
// setting the new set.
func SetSampleKeywords(keywords ...string) {
	// Clear SampleAttributeKeywords
	SampleAttributeKeywords = make(map[string]bool)

	// Add new set of keywords
	for _, keyword := range keywords {
		SampleAttributeKeywords[keyword] = true
	}
}

// AddProcessKeyword adds a new keyword to the ProcessAttributeKeywords map.
func AddProcessKeyword(keyword string) {
	ProcessAttributeKeywords[keyword] = true
}

// SetProcessKeywords overrides the current ProcessAttributeKeywords with the
// new set of keywords. It clears the current set of keywords before
// setting the new set.
func SetProcessKeywords(keywords ...string) {
	// Clear ProcessAttributeKeywords
	ProcessAttributeKeywords = make(map[string]bool)

	// Add new set of keywords
	for _, keyword := range keywords {
		ProcessAttributeKeywords[keyword] = true
	}
}

// AddFileKeyword adds a new keyword to the FileAttributeKeywords map.
func AddFileKeyword(keyword string) {
	FileAttributeKeywords[keyword] = true
}

// SetFileKeywords overrides the current FileAttributeKeywords with the
// new set of keywords. It clears the current set of keywords before
// setting the new set.
func SetFileKeywords(keywords ...string) {
	// Clear FileAttributeKeywords
	FileAttributeKeywords = make(map[string]bool)

	// Add new set of keywords
	for _, keyword := range keywords {
		FileAttributeKeywords[keyword] = true
	}
}

// ValidateKeywords goes through the ProcessAttributeKeywords, SampleAttributeKeywords,
// and FileAttributeKeywords
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
