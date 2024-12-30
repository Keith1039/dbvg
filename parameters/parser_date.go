package parameters

import (
	"errors"
	randomDataTime "github.com/duktig-solutions/go-random-date-generator"
	"strings"
	"time"
)

const DEFAULTDATECODE = NOW
const DEFAULTDATERANGE = "2001-01-01,2024-12-31"

type DateParser struct {
	Column column
}

func (p *DateParser) ParseColumn() (string, error) {
	code := p.Column.Code
	if code == 0 {
		code = DEFAULTDATECODE
	}
	if code == RANDOM {
		return p.handleRandom()
	} else if code == STATIC {
		return p.handleStatic()
	} else if code == NOW {
		return p.handleNow()
	} else if code == NULL {
		return p.handleNull()
	} else {
		return "", errors.New("invalid code given")
	}
}

func (p *DateParser) handleRandom() (string, error) {
	r, _ := p.Column.Other["Range"]
	if r == "" {
		r = DEFAULTDATERANGE
	}
	dates := strings.Split(r, ",")
	if len(dates) != 2 {
		return "", errors.New("malformed date range")
	}
	return randomDataTime.GenerateDate(dates[0], dates[1])
}

func (p *DateParser) handleStatic() (string, error) {
	r, _ := p.Column.Other["Value"]
	if isDate(r) {
		return r, nil
	} else {
		return "", errors.New("invalid value given")
	}
}

func (p *DateParser) handleNow() (string, error) {
	return time.Now().String(), nil
}

func (p *DateParser) handleNull() (string, error) {
	return "NULL", nil
}

func isDate(dateString string) bool {
	var err error
	_, err = time.Parse(time.Layout, dateString)
	if err == nil {
		return true
	}
	_, err = time.Parse(time.ANSIC, dateString)
	if err == nil {
		return true
	}
	_, err = time.Parse(time.UnixDate, dateString)
	if err == nil {
		return true
	}
	_, err = time.Parse(time.RubyDate, dateString)
	if err == nil {
		return true
	}
	_, err = time.Parse(time.RFC822, dateString)
	if err == nil {
		return true
	}
	_, err = time.Parse(time.RFC822Z, dateString)
	if err == nil {
		return true
	}
	_, err = time.Parse(time.RFC850, dateString)
	if err == nil {
		return true
	}
	_, err = time.Parse(time.RFC1123, dateString)
	if err == nil {
		return true
	}
	_, err = time.Parse(time.RFC1123Z, dateString)
	if err == nil {
		return true
	}
	_, err = time.Parse(time.RFC3339, dateString)
	if err == nil {
		return true
	}
	_, err = time.Parse(time.RFC3339Nano, dateString)
	if err == nil {
		return true
	}
	_, err = time.Parse(time.Kitchen, dateString)
	if err == nil {
		return true
	}
	_, err = time.Parse(time.Stamp, dateString)
	if err == nil {
		return true
	}
	_, err = time.Parse(time.StampMilli, dateString)
	if err == nil {
		return true
	}
	_, err = time.Parse(time.StampMicro, dateString)
	if err == nil {
		return true
	}
	_, err = time.Parse(time.StampNano, dateString)
	if err == nil {
		return true
	}
	_, err = time.Parse(time.DateTime, dateString)
	if err == nil {
		return true
	}
	_, err = time.Parse(time.DateOnly, dateString)
	if err == nil {
		return true
	}
	_, err = time.Parse(time.TimeOnly, dateString)
	if err == nil {
		return true
	}
	return false
}
