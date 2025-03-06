package parameters

import (
	"errors"
	"github.com/Keith1039/dbvg/utils"
	"math/rand"
)

// default code for the boolean parser
const DEFAULTBOOLCODE = RANDOM

// BooleanColumnParser is the struct responsible for processing parameters and creating queries for the Boolean type columns
type BooleanColumnParser struct {
}

// ParseColumn takes in a column and processes it in order to return a string value along with any errors that occur
func (p *BooleanColumnParser) ParseColumn(col column) (string, error) {
	code := col.Code
	if code == 0 {
		code = DEFAULTBOOLCODE
	}
	if code == RANDOM {
		return p.handleRandom()
	} else if code == STATIC {
		return p.handleStatic(col)
	} else if code == NULL {
		return p.handleNull()
	} else {
		return "", errors.New("invalid code given")
	}

}

func (p *BooleanColumnParser) handleRandom() (string, error) {
	num := rand.Intn(100)
	if num < 50 {
		return "true", nil
	} else {
		return "false", nil
	}
}

func (p *BooleanColumnParser) handleStatic(col column) (string, error) {
	val := col.Other
	val = utils.TrimAndLowerString(val) // trim and lower string
	if val != "true" && val != "false" {
		return "", errors.New("invalid Value given")
	} else {
		return val, nil
	}
}

func (p *BooleanColumnParser) handleNull() (string, error) {
	return "NULL", nil
}
