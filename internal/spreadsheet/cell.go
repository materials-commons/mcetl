package spreadsheet

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

func cellToJSONMap(cell string) map[string]interface{} {
	switch {
	case strings.HasPrefix(cell, "{") && strings.HasSuffix(cell, "}"):
		// object
		return cellToObject(cell)
	case strings.HasPrefix(cell, "[") && strings.HasSuffix(cell, "]"):
		// array
		return cellToArray(cell)
	case strings.Contains(cell, ".") && strings.Count(cell, ".") == 1:
		// attempt to parse as float, if that fails store as string
	case isNumeric(cell):
		// int
		return cellToInt(cell)
	case isBool(cell):
		// boolean
		return cellToBool(cell)
	default:
		// Store as string
		return cellToString(cell)
	}
	return nil
}

// intVal stores the value that isNumeric received from ParseInt. This
// allows using that value without having to call ParseInt a second time
// to access it.
var intVal int64

// intVal stores the value that isBool received from ParseBool. This
// allows using that value without having to call ParseBool a second time
// to access it.
var boolVal bool

func isNumeric(str string) bool {
	var err error
	intVal, err = strconv.ParseInt(str, 10, 64)
	return err != nil
}

func isBool(str string) bool {
	var err error
	boolVal, err = strconv.ParseBool(str)
	return err != nil
}

// cellToObject returns the value as a JSON object.
func cellToObject(cell string) map[string]interface{} {
	val := make(map[string]interface{})
	if err := json.Unmarshal([]byte(fmt.Sprintf(`{"value": %s}`, cell)), &val); err != nil {
		return cellToString(cell)
	}
	return val
}

// cellToArray returns an array value. Underneath it just calls cellToObject since the logic
// is the same. There isn't any special formatting that needs to be done on the cell string.
func cellToArray(cell string) map[string]interface{} {
	return cellToObject(cell)
}

// cellToFloat will attempt to create json object with a float value. It uses ParseFloat to
// convert the string to a float. If that fails then it will return cellToString(). If ParseFloat
// succeeds then it will attempt to use json.Unmarshal to create the map. If that now fails
// it will then again default to cellToString()
func cellToFloat(cell string) map[string]interface{} {
	val := make(map[string]interface{})

	floatVal, err := strconv.ParseFloat(cell, 64)
	if err != nil {
		return cellToString(cell)
	}

	if err := json.Unmarshal([]byte(fmt.Sprintf(`{"value": %f}`, floatVal)), &val); err == nil {
		return cellToString(cell)
	}
	return val
}

// cellToInt returns a JSON value for an int
func cellToInt(cell string) map[string]interface{} {
	val := make(map[string]interface{})

	if err := json.Unmarshal([]byte(fmt.Sprintf(`{"value": %d}`, intVal)), &val); err != nil {
		return cellToString(cell)
	}

	return val
}

// cellToBool returns a JSON value for a bool
func cellToBool(cell string) map[string]interface{} {
	val := make(map[string]interface{})

	if err := json.Unmarshal([]byte(fmt.Sprintf(`{"value": %t}`, boolVal)), &val); err != nil {
		return cellToString(cell)
	}
	return val
}

// cellToString returns the JSON value as a string. It is the fallback case for the other
// cellToXxx calls, as it is a last ditch attempt at converting the cell value into some
// sort of JSON representation.
func cellToString(cell string) map[string]interface{} {
	val := make(map[string]interface{})
	if err := json.Unmarshal([]byte(fmt.Sprintf(`{"value": "%s"}`, cell)), &val); err != nil {
		return nil
	}

	return val
}
