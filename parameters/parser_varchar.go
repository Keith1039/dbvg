package parameters

import (
	"errors"
	"fmt"
	"github.com/brianvoe/gofakeit/v7"
	regen "github.com/zach-klippenstein/goregen"
)

// default value for the REGEX code for the varchar parser
const DEFAULTEXPR = "[a-zA-Z]{10}"

// default code for the varchar parser
const DEFAULVARCHARCODE = REGEX

// VarcharColumnParser is the struct responsible for processing parameters and creating queries for Varchar type columns
type VarcharColumnParser struct {
}

func (p *VarcharColumnParser) ParseColumn(col column) (string, error) {
	code := col.Code
	if code == 0 {
		code = DEFAULVARCHARCODE
	}

	if code == REGEX {
		return p.handleRegex(col)
	} else if code == EMAIL {
		return p.handleEmail()
	} else if code == FIRSTNAME {
		return p.handleFirstName()
	} else if code == LASTNAME {
		return p.handleLastName()
	} else if code == FULLNAME {
		return p.handleFullName()
	} else if code == PHONE {
		return p.handlePhone()
	} else if code == COUNTRY {
		return p.handleCountry()
	} else if code == ADDRESS {
		return p.handleAddress()
	} else if code == ZIPCODE {
		return p.handleZipCode()
	} else if code == CITY {
		return p.handleCity()
	} else if code == STATIC {
		return p.handleStatic(col)
	} else if code == NULL {
		return p.handleNull()
	} else {
		return "", errors.New("invalid code given")
	}
}

func (p *VarcharColumnParser) handleRegex(col column) (string, error) {
	expression := col.Other // regen doesn't care about whitespace so there's no need to trim
	if expression == "" {
		length, _ := col.ColumnDetails.Length() // no point in checking since all types mapped to varchar give a valid length
		if length != -5 && length < 10 {        // check if it's not default VARCHAR or BPCHAR
			expression = fmt.Sprintf("[a-zA-Z]{%d}", length) // change the default expression to fit the container
		} else {
			expression = DEFAULTEXPR
		}
	}
	genString, err := regen.Generate(expression)
	return genString, err
}

func (p *VarcharColumnParser) handleEmail() (string, error) {
	return gofakeit.Email(), nil
}

func (p *VarcharColumnParser) handleFirstName() (string, error) {
	return gofakeit.Person().FirstName, nil
}

func (p *VarcharColumnParser) handleLastName() (string, error) {
	return gofakeit.Person().LastName, nil
}

func (p *VarcharColumnParser) handleFullName() (string, error) {
	return gofakeit.Name(), nil
}

func (p *VarcharColumnParser) handlePhone() (string, error) {
	return gofakeit.Phone(), nil
}

func (p *VarcharColumnParser) handleCountry() (string, error) {
	return gofakeit.Country(), nil
}

func (p *VarcharColumnParser) handleAddress() (string, error) {
	return gofakeit.Address().Address, nil
}

func (p *VarcharColumnParser) handleZipCode() (string, error) {
	return gofakeit.Zip(), nil
}

func (p *VarcharColumnParser) handleCity() (string, error) {
	return gofakeit.City(), nil
}

func (p *VarcharColumnParser) handleStatic(col column) (string, error) {
	return col.Other, nil
}

func (p *VarcharColumnParser) handleNull() (string, error) {
	return "NULL", nil
}
