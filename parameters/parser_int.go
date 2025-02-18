package parameters

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
)

// default range for the RANGE code of the integer parser
const DEFAULTRANGE = "0,100"

// default value for the STATIC code of the integer parser
const DEFAULTSTATIC = "0"

// default code for the integer parser
const DEFAULTINTCODE = SEQ

// IntColumnParser is the struct responsible for processing parameters and creating queries for Integer type columns
type IntColumnParser struct {
	latest int
}

// ParseColumn takes in a column and processes it in order to return a string value along with any errors that occur
func (p *IntColumnParser) ParseColumn(col column) (string, error) {
	code := col.Code
	if code == 0 {
		code = DEFAULTINTCODE
	}
	if code == RANDOM {
		return p.handleRandomCode(col)
	} else if code == STATIC {
		return p.handleStatic(col)
	} else if code == SEQ {
		return p.handleSeq()
	} else if code == NULL {
		return p.handleNull()
	} else {
		err := errors.New("invalid code given")
		return "", err
	}
}

func (p *IntColumnParser) handleRandomCode(col column) (string, error) {
	var value string
	var err error
	r := col.Other
	if r == "" {
		r = DEFAULTRANGE
	}
	ranges := strings.Split(r, ",")
	if len(ranges) != 2 {
		err = errors.New("malformed range")
		return "", err
	}
	lowerBound, boundErr := strconv.Atoi(strings.TrimSpace(ranges[0]))
	if boundErr != nil {
		return "", boundErr
	}
	upperBound, boundErr2 := strconv.Atoi(strings.TrimSpace(ranges[1]))
	if boundErr2 != nil {
		return "", boundErr
	}
	if lowerBound > upperBound {
		err = errors.New("lower bound is greater than upper bound")
		return "", err
	}
	value = strconv.Itoa(rand.Intn(upperBound-lowerBound) + lowerBound)
	return value, err
}

func (p *IntColumnParser) handleStatic(col column) (string, error) {
	var value string
	var err error
	r := col.Other
	fmt.Println("r")
	if r == "" {
		r = DEFAULTSTATIC
	}
	_, err = strconv.Atoi(r)
	if err != nil {
		return "", err
	} else {
		value = r
	}
	return value, err
}

func (p *IntColumnParser) handleSeq() (string, error) {
	l := p.latest // get the latest
	p.latest++    // increment
	return strconv.Itoa(l), nil
}

func (p *IntColumnParser) handleNull() (string, error) {
	return "NULL", nil
}
