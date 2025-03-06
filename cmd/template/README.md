# Templates

Templates are used for data generation in order to provide some amount of control to the
data generated. JSON templates are created and used to store/define certain
fields which are referenced when generating data.

The following is a template generated for a given table "products":

```
{
  "products": {
    "price": {
      "Code": "",
      "Type": "FLOAT",
      "Value": ""
    }
    ... 
}
```
the first key, `products`, is the table name. The second key, `price`, is the column name. Mapped to the column name are the fields
used for data generation. These fields are `Code`, `Type` and `Value`.


## Code
This is the field that determines the type of data that will be generated. These codes include `RANDOM`, `STATIC`, `NULL`
etc. By using these values, you can influence how the data for that specific column is generated. 
For more information regarding the codes, please consult the [Code Guide](#code-guide). The value for this field is
case-insensitive so putting "random" would be the same as putting "RANDOM".

## Type
This is the perceived type of the column, in other words, how the program will interpret the column's type.
This key's value is assigned during the template's creation. It serves as a reference for the user 
when using the template and code guide. Any changes made to this field's value is irrelevant.

## Value
This is the field meant to be used with the Code field to help generate data.
While this value is always a string, different codes will require different formats.
Please consult the [Code Guide](#code-guide) for more information regarding the accepted format for the "Value" field.

## Warning
Template's can get fairly large, after all they include the source table and all it's dependencies.
As such, the golden rule of templates is simple, if you don't need to modify a column's output... don't. 
All parsers have default values in the case that they don't receive a code. As such, only assign codes and values to 
columns whose outputs you **NEED** to tamper with.

# Code Guide

## Support Map
A quick look at the supported codes for each column type

**BOOL**: `RANDOM` | `STATIC` | `NULL`

**DATE**: `RANDOM` | `STATIC` | `NOW` | `NULL`

**INT**: `RANDOM` | `STATIC` | `SEQ` | `NULL`

**FLOAT**: `RANDOM` | `STATIC` | `NULL`

**UUID**: `UUID` | `NULL`

**VARCHAR**: `STATIC` | `REGEX` | `EMAIL` | `FIRSTNAME` | `LASTNAME` | `FULLNAME` | `PHONE` | `COUNTRY` | `ADDRESS` | `ZIPCODE` | `CITY` | `NULL`

## Code Type Defaults
**BOOL**: `RANDOM`

**DATE**: `NOW`

**INT**: `SEQ`

**FLOAT**: `RANDOM`

**UUID**: `UUID`

**VARCHAR**: `REGEX` with default value: "[a-zA-Z]{10}"


## RANDOM
Generates a random value based on the `Value` field in the template. If no value is given, i.e.  `"Value": ""`, a default will be 
used. This code is supported only by `INT`, `FLOAT` and `Date` type columns. The supported value format is a range between two values of the
column type.

The following are examples of the supported format for the `RANDOM` code.

INT: "2,6"

FLOAT: "2, 6" or "2.0,6.0"

DATE: "2024-02-23, 2025-01-1"

## STATIC
Generates a static value based on the `Value` field in the template. In the case of `INT` type columns, if no value is
given, `"Value": ""`, the default of "0" is used instead. This code is supported by ALL column types 
(`BOOL`, `DATE`, `INT`, `FLOAT`, `UUID`, `VARCHAR`). The supported format for this code is simply a string
value for the given column type. Integer for INT, Boolean for Bool etc.

## SEQ
Generates a sequential integer value starting from 1. Behaves how a serial integer would behave. 
This code does not use the `Value` field. This code is supported by `INT` type columns.

## UUID
Generates a UUID for the column. This code does not use the `Value` field. This code is supported by `UUID` type
columns.

## REGEX
Generates a string value based on the REGEX given by the `Value` field. 
if no value is given, `"Value": ""`, the default of "[a-zA-Z]{10}" is used instead. The supported format for this code is
a valid regex string. This code is supported by `VARCHAR` type columns.

## EMAIL
Generates a random email string value using `gofakeit` for the column. This code does not use the `Value` field. 
This code is supported by `VARCHAR` type columns.

## FIRSTNAME
Generates a random first name string value using `gofakeit` for the column. This code does not use the `Value` field. 
This code is supported by `VARCHAR` type columns.

## LASTNAME
Generates a random last name string value using `gofakeit` for the column. This code does not use the `Value` field. 
This code is supported by `VARCHAR` type columns.

## FULLNAME
Generates a random full name string value using `gofakeit` for the column. This code does not use the `Value` field. 
This code is supported by `VARCHAR` type columns.

## PHONE
Generates a random phone number string value using `gofakeit` for the column. This code does not use the `Value` field. 
This code is supported by `VARCHAR` type columns.

## COUNTRY
Generates a random country string value using `gofakeit` for the column. This code does not use the `Value` field. 
This code is supported by `VARCHAR` type columns.

## ADDRESS
Generates a random address string value using `gofakeit` for the column. This code does not use the `Value` field. 
This code is supported by `VARCHAR` type columns.

## ZIPCODE
Generates a random zipcode string value using `gofakeit` for the column. This code does not use the `Value` field. 
This code is supported by `VARCHAR` type columns.

## CITY
Generates a random city string value using `gofakeit` for the column. This code does not use the `Value` field. 
This code is supported by `VARCHAR` type columns.

## NULL
Generate a NULL value for the given column. This code does not use the `Value` field. This code is supported by ALL
column types (`BOOL`, `DATE`, `INT`, `UUID`, `VARCHAR`).

## NOW
Generate a timestamp based on the current time and uses it as the value for the column. This code does not use the `Value` field. 
This code is supported by `DATE` type columns.


