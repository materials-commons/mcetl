package spreadsheet

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type cellConverter struct {
	// intVal stores the value that isNumeric received from ParseInt. This
	// allows using that value without having to call ParseInt a second time
	// to access it.
	intVal int64

	// boolVal stores the value that isBool received from ParseBool. This
	// allows using that value without having to call ParseBool a second time
	// to access it.
	boolVal bool
}

func newCellConverter() *cellConverter {
	// explicitly initialize so we know what default values are
	return &cellConverter{intVal: 0, boolVal: false}
}

// cellToJSONMap will take a cell entry which is a string. It looks at the string to determine what type
// it is and then attempts to turn it into a json string that we can call json.Unmarshal() on in order to
// create a map of the JSON value. Because the user may not have stored the value in the cell as something
// we can turn into a particular bit of JSON, as a last resort we will treat it as a string and json.Unmarshal()
// that. As an example, imagine a cell that has the following in it:
//  [0,1], [2,3]
// This has two separate values and there isn't any easy way to determine what they are. Unmarshal will fail unless
// treat this as a string. Doing this still allows us to store the value in the database, and the user can see that
// value. Its just not represented as an object of arrays.
func (c *cellConverter) cellToJSONMap(cell string) (map[string]interface{}, error) {
	switch {
	case strings.HasPrefix(cell, "{") && strings.HasSuffix(cell, "}"):
		// object
		return c.cellToObject(cell)
	case strings.HasPrefix(cell, "[") && strings.HasSuffix(cell, "]"):
		// array
		return c.cellToArray(cell)
	case strings.Contains(cell, ".") && strings.Count(cell, ".") == 1:
		// float
		return c.cellToFloat(cell)
	case c.isNumeric(cell):
		// int
		return c.cellToInt(cell)
	case c.isBool(cell):
		// boolean
		return c.cellToBool(cell)
	default:
		// Store as string
		return c.cellToString(cell)
	}
}

// isNumeric will check if the cell is an integer. If it is it stores the converted
// value in c.intVal and returns true.
func (c *cellConverter) isNumeric(str string) bool {
	var err error
	c.intVal, err = strconv.ParseInt(str, 10, 64)
	return err == nil
}

// isBool will check if the cell is a boolean. If it is it stores the converted
// value in c.boolVal and returns true.
func (c *cellConverter) isBool(str string) bool {
	var err error
	c.boolVal, err = strconv.ParseBool(str)
	return err == nil
}

// cellToObject returns the value as a JSON object, if that fails return as a string.
func (c *cellConverter) cellToObject(cell string) (map[string]interface{}, error) {
	val := make(map[string]interface{})
	if err := json.Unmarshal([]byte(fmt.Sprintf(`{"value": %s}`, cell)), &val); err != nil {
		return c.cellToString(cell)
	}
	return val, nil
}

// cellToArray returns an array value. Underneath it just calls cellToObject since the logic
// is the same. There isn't any special formatting that needs to be done on the cell.
func (c *cellConverter) cellToArray(cell string) (map[string]interface{}, error) {
	return c.cellToObject(cell)
}

// cellToFloat will attempt to create json object with a float value. It uses ParseFloat to
// convert the string to a float. If that fails then it will return cellToString(). If ParseFloat
// succeeds then it will attempt to use json.Unmarshal to create the map. If that now fails
// it will then again default to cellToString()
func (c *cellConverter) cellToFloat(cell string) (map[string]interface{}, error) {
	val := make(map[string]interface{})

	floatVal, err := strconv.ParseFloat(cell, 64)
	if err != nil {
		// We thought it was a float, but its not so treat as a string
		return c.cellToString(cell)
	}

	if err := json.Unmarshal([]byte(fmt.Sprintf(`{"value": %f}`, floatVal)), &val); err == nil {
		return c.cellToString(cell)
	}
	return val, nil
}

// cellToInt returns a JSON value for an int, if that fails return as a string.
func (c *cellConverter) cellToInt(cell string) (map[string]interface{}, error) {
	val := make(map[string]interface{})

	if err := json.Unmarshal([]byte(fmt.Sprintf(`{"value": %d}`, c.intVal)), &val); err != nil {
		return c.cellToString(cell)
	}

	return val, nil
}

// cellToBool returns a JSON value for a bool, if that fails return as a string.
func (c *cellConverter) cellToBool(cell string) (map[string]interface{}, error) {
	val := make(map[string]interface{})

	if err := json.Unmarshal([]byte(fmt.Sprintf(`{"value": %t}`, c.boolVal)), &val); err != nil {
		return c.cellToString(cell)
	}

	return val, nil
}

// cellToString returns the JSON value as a string. It is the fallback case for the other
// cellToXxx calls, as it is a last ditch attempt at converting the cell value into some
// sort of JSON representation.
func (c *cellConverter) cellToString(cell string) (map[string]interface{}, error) {
	val := make(map[string]interface{})
	if err := json.Unmarshal([]byte(fmt.Sprintf(`{"value": "%s"}`, cell)), &val); err != nil {
		return nil, err
	}

	return val, nil
}
