# Defined Strategies

## Overview
Each `Strategy` is supported by a specific type and a code. In other words, those are the keys for the `Strategy` in the internal maps.
As such, each supported type will have a section and underneath their section will be a description of the Strategies supported for that type.
A special case is the `NULL` code. All types by default accept the `NULL` code.

## ALL
### NULL
#### Type: `Override Strategy`
#### Expected Value: N/A
#### Behavior: Returns null regardless of value each time the `Strategy` is executed
#### Valid Example:
```json
{
  "column": {
    "type": "FLOAT",
    "code": "NULL",
    "value": [1, 10]
  }
}
```
```json
{
  "column": {
    "type": "VARCHAR",
    "code": "NULL",
    "value": true
  }
}
```
#### Invalid Example:
N/A

## INT

### RANDOM
#### Type: `Required Strategy`
#### Expected Value: [2]int with int[0] < int[1]
#### Behavior: Outputs a number in the range of the two bounds [ int[0], int[1] ) each time the `Strategy` is executed
#### Valid Example:
```json
{
  "column": {
    "type": "INT",
    "code": "RANDOM",
    "value": [1, 10]
  }
}
```

#### Invalid Examples
```json
{
  "column": {
    "type": "INT",
    "code": "RANDOM",
    "value": [10, 5]
  }
}
```

```json
{
  "column": {
    "type": "INT",
    "code": "RANDOM",
    "value": "1-20"
  }
}
```

### STATIC
#### Type: `Required Strategy`
#### Expected Value: any integer
#### Behavior: Outputs the given integer each time the `Strategy` is executed
#### Valid Example:
```json
{
  "column": {
    "type": "INT",
    "code": "STATIC",
    "value": 5
  }
}
```

#### Invalid Example:
```json
{
  "column": {
    "type": "INT",
    "code": "STATIC",
    "value": "5.0"
  }
}
```

### SERIAL
#### Type: `Optional Strategy`
#### Expected Value: null (defaults to 1) or any integer >= 1
#### Behavior: Outputs the given integer and then increments the value by 1 for the next time the `Strategy` is executed
#### Valid Example:
```json
{
  "column": {
    "type": "INT",
    "code": "STATIC",
    "value": null
  }
}
```
```json
{
  "column": {
    "type": "INT",
    "code": "STATIC",
    "value": 5
  }
}
```

#### Invalid Example:
```json
{
  "column": {
    "type": "INT",
    "code": "STATIC",
    "value": 0
  }
}
```
```json
{
  "column": {
    "type": "INT",
    "code": "STATIC",
    "value": "6"
  }
}
```

## FLOAT
### RANDOM
#### Type: `Required Strategy`
#### Expected Value: [2]float with float[0] < float[1]
#### Behavior: Outputs a floating point number in the range of the two bounds [ float[0], float[1] ) each time the `Strategy` is executed
#### Valid Example:
```json
{
  "column": {
    "type": "FLOAT",
    "code": "RANDOM",
    "value": [1, 10]
  }
}
```

#### Invalid Examples
```json
{
  "column": {
    "type": "FLOAT",
    "code": "RANDOM",
    "value": [10, 5]
  }
}
```

```json
{
  "column": {
    "type": "FLOAT",
    "code": "RANDOM",
    "value": "1.75-10.80"
  }
}
```

### STATIC
#### Type: `Required Strategy`
#### Expected Value: Any floating point number
#### Behavior: Outputs the given number each time the `Strategy` is executed
#### Valid Example:
```json
{
  "column": {
    "type": "FLOAT",
    "code": "STATIC",
    "value": 5
  }
}
```
```json
{
  "column": {
    "type": "FLOAT",
    "code": "STATIC",
    "value": 20.34
  }
}
```
#### Invalid Example:
```json
{
  "column": {
    "type": "FLOAT",
    "code": "STATIC",
    "value": "54.76"
  }
}
```

## UUID
### UUID
#### Type: `Override Strategy`
#### Expected Value: N/A
#### Behavior: Outputs a new UUID every time the `Strategy` is executed
#### Valid Examples:
```json
{
  "column": {
    "type": "UUID",
    "code": "UUID",
    "value": [1, 20, 314]
  }
}
```
```json
{
  "column": {
    "type": "UUID",
    "code": "UUID",
    "value": true
  }
}
```
#### Invalid Examples:
N/A

## BOOL
### RANDOM
#### Type: `Override Strategy`
#### Expected Value: N/A
#### Behavior: Outputs either `true` or `false` randomly every time the `Strategy` is executed
#### Valid Example:
```json
{
  "column": {
    "type": "BOOL",
    "code": "RANDOM",
    "value": "some string"
  }
}
```
```json
{
  "column": {
    "type": "BOOL",
    "code": "RANDOM",
    "value": true
  }
}
```
#### Invalid Example:
N/A

### STATIC
#### Type: `Required Strategy`
#### Expected Value: either `true` or `false`
#### Behavior: Outputs the given value every time the `Strategy` is executed
#### Valid Example:
```json
{
  "column": {
    "type": "BOOL",
    "code": "STATIC",
    "value": true
  }
}
```
```json
{
  "column": {
    "type": "BOOL",
    "code": "STATIC",
    "value": false
  }
}
```
#### Invalid Example:
```json
{
  "column": {
    "type": "BOOL",
    "code": "STATIC",
    "value": "true"
  }
}
```
```json
{
  "column": {
    "type": "BOOL",
    "code": "STATIC",
    "value": null
  }
}
```

## DATE

### NOW
#### Type: `Override Strategy`
#### Expected Value: N/A
#### Behavior: Outputs the current time as a `time.Time` struct every time the `Strategy` is executed
#### Valid Example:
```json
{
  "column": {
    "type": "DATE",
    "code": "NOW",
    "value": null
  }
}
```
```json
{
  "column": {
    "type": "DATE",
    "code": "NOW",
    "value": 50
  }
}
```
#### Invalid Example:
N/A

### STATIC
#### Type: `Required Strategy`
#### Expected Value: valid date string
#### Behavior: Outputs a `time.Time` struct with using the given date string every time the `Strategy` is executed
#### Valid Example:
```json
{
  "column": {
    "type": "DATE",
    "code": "STATIC",
    "value": "2009-11-10 23:00:00"
  }
}
```
#### Invalid Example
```json
{
  "column": {
    "type": "DATE",
    "code": "STATIC",
    "value": 20221001
  }
}
```
```json
{
  "column": {
    "type": "DATE",
    "code": "STATIC",
    "value": "14-2020"
  }
}
```

### RANDOM
#### Type: `Required Strategy`
#### Expected Value: [2]string with 2 valid date strings where string[0] < string[1]
#### Behavior: Outputs a `time.Time` struct with a date that's in the range of [ string[0], string[1] ) every time the `Strategy` is executed
#### Valid Example:
```json
{
  "column": {
    "type": "DATE",
    "code": "RANDOM",
    "value": ["2000-01-02", "2026-01-01"]
  }
}
```
#### Invalid Example:
```json
{
  "column": {
    "type": "DATE",
    "code": "RANDOM",
    "value": true
  }
}
```
```json
{
  "column": {
    "type": "DATE",
    "code": "RANDOM",
    "value": ["2026-01-01", "2000-01-02"]
  }
}
```
## TIME

### NOW
#### Type: `Override Strategy`
#### Expected Value: N/A
#### Behavior: Outputs the current time as a `time.Time` struct every time the `Strategy` is executed
#### Valid Example:
```json
{
  "column": {
    "type": "TIME",
    "code": "NOW",
    "value": null
  }
}
```
```json
{
  "column": {
    "type": "TIME",
    "code": "NOW",
    "value": 50
  }
}
```
#### Invalid Example:
N/A

### STATIC
#### Type: `Required Strategy`
#### Expected Value: valid time string of format (HH:MM:SS)
#### Behavior: Outputs a `time.Time` struct with using the given time string every time the `Strategy` is executed
#### Valid Example:
```json
{
  "column": {
    "type": "TIME",
    "code": "STATIC",
    "value": "23:00:00"
  }
}
```
#### Invalid Example
```json
{
  "column": {
    "type": "TIME",
    "code": "STATIC",
    "value": 230000
  }
}
```
```json
{
  "column": {
    "type": "TIME",
    "code": "STATIC",
    "value": "2009-11-10 23:00:00"
  }
}
```

### RANDOM
#### Type: `Required Strategy`
#### Expected Value: [2]string with 2 valid time (HH:MM:SS) strings where string[0] < string[1]
#### Behavior: Outputs a `time.Time` struct with a time that's in the range of [ string[0], string[1] ) every time the `Strategy` is executed
#### Valid Example:
```json
{
  "column": {
    "type": "TIME",
    "code": "RANDOM",
    "value": ["00:00:00", "23:59:59"]
  }
}
```
#### Invalid Example
```json
{
  "column": {
    "type": "TIME",
    "code": "RANDOM",
    "value": ["23:59:59", "00:00:00"]
  }
}
```
```json
{
  "column": {
    "type": "TIME",
    "code": "RANDOM",
    "value": 500
  }
}
```

## VARCHAR

### REGEX
#### Type: `Required Strategy`
#### Expected Value: any valid regex string
#### Behavior: Outputs a string that fits the given Regex everytime the `Strategy` is executed
#### Valid Example:
```json
{
  "column": {
    "type": "VARCHAR",
    "code": "STATIC",
    "value": "gr(a|e)y" 
  }
}
```
#### Invalid Examples:
```json
{
  "column": {
    "type": "VARCHAR",
    "code": "STATIC",
    "value": "[aZ-Az]"
  }
}
```
```json
{
  "column": {
    "type": "VARCHAR",
    "code": "STATIC",
    "value": null
  }
}
```

### STATIC
#### Type: `Required Strategy`
#### Expected Value: any string
#### Behavior: Outputs the given string every time the `Strategy` is executed
#### Valid Example:
```json
{
  "column": {
    "type": "VARCHAR",
    "code": "STATIC",
    "value": "some string"
  }
}
```
#### Invalid Examples:
```json
{
  "column": {
    "type": "VARCHAR",
    "code": "STATIC",
    "value": true
  }
}
```
```json
{
  "column": {
    "type": "VARCHAR",
    "code": "STATIC",
    "value": null
  }
}
```

### EMAIL
#### Type: `Override Strategy`
#### Expected Value: N/A
#### Behavior: Uses the `gofakeit` library to output a random email as string every time the `Strategy` is executed
#### Valid Example:
```json
{
  "column": {
    "type": "VARCHAR",
    "code": "EMAIL",
    "value": null
  }
}
```
```json
{
  "column": {
    "type": "VARCHAR",
    "code": "EMAIL",
    "value": true
  }
}
```
#### Invalid Example
N/A

### FIRSTNAME
#### Type: `Override Strategy`
#### Expected Value: N/A
#### Behavior: Uses the `gofakeit` library to output a random first name as a string every time the `Strategy` is executed
#### Valid Example:
```json
{
  "column": {
    "type": "VARCHAR",
    "code": "FIRSTNAME",
    "value": null
  }
}
```
```json
{
  "column": {
    "type": "VARCHAR",
    "code": "FIRSTNAME",
    "value": 20
  }
}
```
#### Invalid Example
N/A


### LASTNAME
#### Type: `Override Strategy`
#### Expected Value: N/A
#### Behavior: Uses the `gofakeit` library to output a random last name as a string every time the `Strategy` is executed
#### Valid Example:
```json
{
  "column": {
    "type": "VARCHAR",
    "code": "LASTNAME",
    "value": null
  }
}
```
```json
{
  "column": {
    "type": "VARCHAR",
    "code": "LASTNAME",
    "value": [1, ""]
  }
}
```
#### Invalid Example
N/A

### FULLNAME
#### Type: `Override Strategy`
#### Expected Value: N/A
#### Behavior: Uses the `gofakeit` library to output a random full name as a string everytime the `Strategy` is executed
#### Valid Example:
```json
{
  "column": {
    "type": "VARCHAR",
    "code": "FULLNAME",
    "value": null
  }
}
```
```json
{
  "column": {
    "type": "VARCHAR",
    "code": "FULLNAME",
    "value": [1,"", 1204]
  }
}
```
#### Invalid Example
N/A

### PHONE
#### Type: `Override Strategy`
#### Expected Value: N/A
#### Behavior: Uses the `gofakeit` library to output a random phone number as a string every time the `Strategy` is executed
#### Valid Example:
```json
{
  "column": {
    "type": "VARCHAR",
    "code": "PHONE",
    "value": null
  }
}
```
```json
{
  "column": {
    "type": "VARCHAR",
    "code": "PHONE",
    "value": "2026-01"
  }
}
```
#### Invalid Example
N/A


### COUNTRY
#### Type: `Override Strategy`
#### Expected Value: N/A
#### Behavior: Uses the `gofakeit` library to output a random country's name as a string every time the `Strategy` is executed
#### Valid Example:
```json
{
  "column": {
    "type": "VARCHAR",
    "code": "COUNTRY",
    "value": null
  }
}
```
```json
{
  "column": {
    "type": "VARCHAR",
    "code": "COUNTRY",
    "value": "Random Country Name"
  }
}
```
#### Invalid Example
N/A

### ADDRESS
#### Type: `Override Strategy`
#### Expected Value: N/A
#### Behavior: Uses the `gofakeit` library to output a random address as a string every time the `Strategy` is executed
#### Valid Example:
```json
{
  "column": {
    "type": "VARCHAR",
    "code": "ADDRESS",
    "value": null
  }
}
```
```json
{
  "column": {
    "type": "VARCHAR",
    "code": "ADDRESS",
    "value": "900 some address ave"
  }
}
```
#### Invalid Example
N/A

### ZIPCODE
#### Type: `Override Strategy`
#### Expected Value: N/A
#### Behavior: Uses the `gofakeit` library to output a random zip code as a string every time the `Strategy` is executed
#### Valid Example:
```json
{
  "column": {
    "type": "VARCHAR",
    "code": "ZIPCODE",
    "value": null
  }
}
```
```json
{
  "column": {
    "type": "VARCHAR",
    "code": "ZIPCODE",
    "value": "1 + 2"
  }
}
```
#### Invalid Example
N/A


### CITY
#### Type: `Override Strategy`
#### Expected Value: N/A
#### Behavior: Uses the `gofakeit` library to output a random city's name as a string every time the `Strategy` is executed
#### Valid Example:
```json
{
  "column": {
    "type": "VARCHAR",
    "code": "CITY",
    "value": null
  }
}
```
```json
{
  "column": {
    "type": "VARCHAR",
    "code": "CITY",
    "value": "Atl"
  }
}
```
#### Invalid Example
N/A
