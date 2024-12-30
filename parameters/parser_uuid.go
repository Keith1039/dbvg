package parameters

import (
	"errors"
	"github.com/google/uuid"
)

const DEFAULTUUIDCODE = UUID

type UUIDParser struct {
	Column column
}

func (p *UUIDParser) ParseColumn() (string, error) {
	code := p.Column.Code
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

func (p *UUIDParser) handleUUID() (string, error) {
	return uuid.New().String(), nil
}

func (p *UUIDParser) handleNull() (string, error) {
	return uuid.Nil.String(), nil
}
