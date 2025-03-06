package parameters

import (
	"errors"
	randomDataTime "github.com/duktig-solutions/go-random-date-generator"
	"strings"
	"time"
)

// default code for the date parser
const DEFAULTDATECODE = NOW

// default date range for the RANDOM code of the date parser
const DEFAULTDATERANGE = "2001-01-01,2024-12-31"

// DateColumnParser is the struct responsible for processing parameters and creating queries for the Date type columns
type DateColumnParser struct {
}

// ParseColumn takes in a column and processes it in order to return a string value along with any errors that occur
func (p *DateColumnParser) ParseColumn(col column) (string, error) {
	code := col.Code
	if code == 0 {
		code = DEFAULTDATECODE
	}
	if code == RANDOM {
		return p.handleRandom(col)
	} else if code == STATIC {
		return p.handleStatic(col)
	} else if code == NOW {
		return p.handleNow()
	} else if code == NULL {
		return p.handleNull()
	} else {
		return "", errors.New("invalid code given")
	}
}

func (p *DateColumnParser) handleRandom(col column) (string, error) {
	r := col.Other
	r = strings.TrimSpace(r) // trim space
	if r == "" {
		r = DEFAULTDATERANGE
	}
	dates := strings.Split(r, ",")
	if len(dates) != 2 {
		return "", errors.New("malformed date range")
	}
	return randomDataTime.GenerateDate(dates[0], dates[1])
}

func (p *DateColumnParser) handleStatic(col column) (string, error) {
	r := col.Other
	r = strings.TrimSpace(r) // trim space
	if isDate(r) {
		return r, nil
	} else {
		return "", errors.New("invalid value given")
	}
}

func (p *DateColumnParser) handleNow() (string, error) {
	return time.Now().String()[0:19], nil
}

func (p *DateColumnParser) handleNull() (string, error) {
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
