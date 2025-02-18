package parameters

import (
	"errors"
	"github.com/google/uuid"
)

// default code for the UUID parser
const DEFAULTUUIDCODE = UUID

// UUIDColumnParser is the struct responsible for processing parameters and creating queries for the 'UUID' type columns
type UUIDColumnParser struct {
}

// ParseColumn takes in a column and processes it in order to return a string value along with any errors that occur
func (p *UUIDColumnParser) ParseColumn(col column) (string, error) {
	code := col.Code
	if code == 0 {
		code = DEFAULTUUIDCODE
	}
	if code == UUID {
		return p.handleUUID()
	} else if code == NULL {
		return p.handleNull()
	} else {
		return "", errors.New("invalid code given")
	}

}

func (p *UUIDColumnParser) handleUUID() (string, error) {
	return uuid.New().String(), nil
}

func (p *UUIDColumnParser) handleNull() (string, error) {
	return uuid.Nil.String(), nil
}
