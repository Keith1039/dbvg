package parameters

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
)

// default range for the RANGE code of the integer parser
const DEFAULTFLOATRANGE = "0, 100"

// default value for the STATIC code of the integer parser
const DEFAULTFLOATSTATIC = "0.0"

// default code for the integer parser
const DEFAULTFLOATCODE = RANDOM

// FloatColumnParser is the struct responsible for processing parameters and creating queries for Integer type columns
type FloatColumnParser struct {
	latest int
}

// ParseColumn takes in a column and processes it in order to return a string value along with any errors that occur
func (p *FloatColumnParser) ParseColumn(col column) (string, error) {
	code := col.Code
	if code == 0 {
		code = DEFAULTFLOATCODE
	}
	if code == RANDOM {
		return p.handleRandomCode(col)
	} else if code == STATIC {
		return p.handleStatic(col)
	} else if code == NULL {
		return p.handleNull()
	} else {
		err := errors.New("invalid code given")
		return "", err
	}
}

func (p *FloatColumnParser) handleRandomCode(col column) (string, error) {
	var value string
	var err error
	r := col.Other
	r = strings.TrimSpace(r)                                // trim space
	precision, scale, ok := col.ColumnDetails.DecimalSize() // get the precision and scale
	if r == "" {
		difference := precision - scale
		// precision - scale is the amount of digits we have before the decimal (in this case, if this value is less than 3 then the range of 1-100 doesn't work)
		if difference < 3 && ok {
			if difference <= 0 {
				log.Fatalf("Difference between precision and scale for column %s is less than or equal to zero. Please verify your schema.", col.ColumnName)
			} else {
				r = fmt.Sprintf("1, %d", int(difference*10-1)) // create the new range
			}
		} else {
			r = DEFAULTFLOATRANGE
		}
	}
	ranges := strings.Split(r, ",")
	if len(ranges) != 2 {
		err = errors.New("malformed range")
		return "", err
	}
	lowerBound, boundErr := strconv.ParseFloat(strings.TrimSpace(ranges[0]), 64)
	if boundErr != nil {
		return "", boundErr
	}
	upperBound, boundErr2 := strconv.ParseFloat(strings.TrimSpace(ranges[1]), 64)
	if boundErr2 != nil {
		return "", boundErr
	}
	if lowerBound > upperBound {
		err = errors.New("lower bound is greater than upper bound")
		return "", err
	}
	value = strconv.FormatFloat(rand.Float64()*(upperBound-lowerBound)+lowerBound, 'f', -1, 64) // format float
	databaseType := col.ColumnDetails.DatabaseTypeName()                                        // get the underlying type
	// conditions to perform type casting. Although this is done when inserting, since we need the exact value to delete, we apply the casting manually
	if ok {
		value = fmt.Sprintf("%s::%s(%d, %d)", value, databaseType, precision, scale)
	} else {
		value = fmt.Sprintf("%s::%s", value, databaseType)
	}
	return value, err
}

func (p *FloatColumnParser) handleStatic(col column) (string, error) {
	var value string
	var err error
	r := col.Other
	r = strings.TrimSpace(r) // trim the input
	if r == "" {
		r = DEFAULTFLOATSTATIC
	}
	_, err = strconv.ParseFloat(r, 64) // check if it's a float
	if err != nil {
		return "", err
	} else {
		value = r
	}
	precision, scale, ok := col.ColumnDetails.DecimalSize() // get the precision and scale
	databaseType := col.ColumnDetails.DatabaseTypeName()    // get the type
	// cast type appropriately
	if ok {
		value = fmt.Sprintf("%s::%s(%d, %d)", value, databaseType, precision, scale)
	} else {
		value = fmt.Sprintf("%s::%s", value, databaseType)
	}
	return value, err
}

func (p *FloatColumnParser) handleNull() (string, error) {
	return "NULL", nil
}
