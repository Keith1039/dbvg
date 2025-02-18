package parameters

import (
	"errors"
	"math/rand"
	"strings"
)

const DEFAULTBOOLCODE = RANDOM

type BooleanParser struct {
}

func (p *BooleanParser) ParseColumn(col column) (string, error) {
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

func (p *BooleanParser) handleRandom() (string, error) {
	num := rand.Intn(100)
	if num < 50 {
		return "true", nil
	} else {
		return "false", nil
	}
}

func (p *BooleanParser) handleStatic(col column) (string, error) {
	val := col.Other
	val = strings.Trim(strings.ToLower(val), " ")
	if val != "true" && val != "false" {
		return "", errors.New("invalid Value given")
	} else {
		return val, nil
	}
}

func (p *BooleanParser) handleNull() (string, error) {
	return "NULL", nil
}
