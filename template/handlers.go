package template

import (
	"errors"
	"fmt"
	"regexp"
	"time"
)

func handleNumerics(code string, columnType string, val any) error {
	switch t := val.(type) {
	case int: // good case (unreachable code considering how parsing JSON works)
		return nil

	case float64:
		if columnType == "INT" {
			// log warning about rounding
		}
		return nil
	default:
		return unsupportedTypeError{typeStr: fmt.Sprintf("%T", t), code: code, supportedTypes: []string{"int", "float64"}}
	}
}

func handleRandomNumerics(code string, columnType string, val any) error {
	switch t := val.(type) {
	case []int:
		if len(t) == 2 {
			if t[0] > t[1] { // check to see if the bound works
				return RandomBoundError{lowerBound: t[0], upperBound: t[1]}
			}
			return nil
		} else {
			return NewUnexpectedArrayLengthError(2, len(t))
		}
	case []float64:
		if len(t) == 2 {
			if columnType == "INT" {
				// log warning about rounding if this isn't strict
			}
			return nil
		} else {
			return NewUnexpectedArrayLengthError(2, len(t))
		}

	default:
		return unsupportedTypeError{typeStr: fmt.Sprintf("%T", t), code: code, supportedTypes: []string{"[]int", "[]float"}}
	}
}

func handleIntRequired(colType string, code string, val any) error {
	switch code {
	case "RANDOM":
		return handleRandomNumerics(code, colType, val)
	case "STATIC":
		return handleNumerics(code, colType, val)
	default:
		return unsupportedCodeError{code: code, columnType: colType}
	}
}

func handleIntOptional(colType string, code string, val any) error {
	switch code {
	case "SEQ":
		switch t := val.(type) {
		case int:
			return nil
		case float64:
			if colType == "INT" {
				// warn about rounding down
			}
			return nil
		case nil:
			return nil
		default:
			return unsupportedTypeError{typeStr: fmt.Sprintf("%T", t), code: code, supportedTypes: []string{"int", "float64"}}
		}
	default:
		return unsupportedCodeError{code: code, columnType: colType}
	}
}

func handleFloatRequired(colType string, code string, val any) error {
	switch code {
	case "RANDOM":
		return handleRandomNumerics(code, colType, val)
	case "STATIC":
		return handleNumerics(code, colType, val)
	default:
		return unsupportedCodeError{code: code, columnType: colType}
	}
}

func handleBoolRequired(colType string, code string, val any) error {
	switch code {
	case "STATIC":
		switch t := val.(type) {
		case bool:
			return nil
		default:
			return UnexpectedTypeError{expectedType: "bool", actualType: fmt.Sprintf("%T", t)}
		}
	default:
		return unsupportedCodeError{code: code, columnType: colType}
	}
}

func handleVarcharRequired(colType string, code string, val any) error {
	switch code {
	case "STATIC":
		switch t := val.(type) {
		case string:
			return nil
		default:
			return UnexpectedTypeError{expectedType: "string", actualType: fmt.Sprintf("%T", t)}
		}
	case "REGEX":
		switch t := val.(type) {
		case string:
			_, err := regexp.Compile(t)
			if err != nil {
				return InvalidRegexError{regexStr: t}
			} else {
				return nil
			}
		default:
			return UnexpectedTypeError{expectedType: "string", actualType: fmt.Sprintf("%T", t)}
		}
	default:
		return unsupportedCodeError{code: code, columnType: colType}
	}
}

func handleDateRequired(colType string, code string, val any) error {
	switch code {
	case "RANDOM":
		switch t := val.(type) {
		case []string:
			if len(t) == 2 {
				time1, err1 := time.Parse(time.RFC3339, t[0])
				time2, err2 := time.Parse(time.RFC3339, t[1])
				if err1 != nil || err2 != nil {
					return errors.New(fmt.Sprintf("date(s) in string '%s' do not follow RFC3339 convention", t))
				} else {
					if time1.After(time2) {
						return RandomBoundError{lowerBound: time1, upperBound: time2}
					} else {
						return nil
					}
				}
			} else {
				return NewUnexpectedArrayLengthError(2, len(t))
			}
		default:
			return unsupportedTypeError{
				typeStr:        fmt.Sprintf("%T", t),
				code:           code,
				supportedTypes: []string{"[]string"},
			}
		}
	case "STATIC":
		switch t := val.(type) {
		case string:
			_, err := time.Parse(time.RFC3339, t)
			if err != nil {
				return ImproperDateStringFormatError{dateStr: t}
			} else {
				return nil
			}
		default:
			return UnexpectedTypeError{expectedType: "string", actualType: fmt.Sprintf("%T", t)}
		}
	default:
		return unsupportedCodeError{code: code, columnType: colType}
	}
}
