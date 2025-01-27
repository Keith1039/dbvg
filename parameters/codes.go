package parameters

// ENUMS that will be used to generate SQL strings
const (
	RANDOM = iota + 1
	STATIC
	SEQ
	UUID
	REGEX
	EMAIL
	FIRSTNAME
	LASTNAME
	FULLNAME
	PHONE
	COUNTRY
	ADDRESS
	ZIPCODE
	CITY
	NULL
	NOW
)

var stringToEnum = map[string]int{
	"RANDOM":    RANDOM,
	"STATIC":    STATIC,
	"SEQ":       SEQ,
	"UUID":      UUID,
	"REGEX":     REGEX,
	"EMAIL":     EMAIL,
	"FIRSTNAME": FIRSTNAME,
	"LASTNAME":  LASTNAME,
	"FULLNAME":  FULLNAME,
	"PHONE":     PHONE,
	"COUNTRY":   COUNTRY,
	"ADDRESS":   ADDRESS,
	"ZIPCODE":   ZIPCODE,
	"CITY":      CITY,
	"NULL":      NULL,
	"NOW":       NOW,
}
