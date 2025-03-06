package parameters

import (
	"errors"
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
	r = strings.TrimSpace(r) // trim space
	if r == "" {
		r = DEFAULTRANGE
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
	return value, err
}

func (p *FloatColumnParser) handleNull() (string, error) {
	return "NULL", nil
}
