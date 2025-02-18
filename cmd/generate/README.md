# Templates

Templates are designed to be used for data generation purposes in order to provide some amount of control to the
data generation aspect of this project. In order to do this, JSON templates are created and used to store/define certain aspects
of the project. This information can then be used by the CLI, namely the entry command, to generate custom data.

The following is a template generated for a given table "a":

```
{
  "a": {
    "akey": {
      "Code": "",
      "Type": "INT",
      "Value": ""
    }
    ... 
}
```
the first key `a` is the table name, the second key `akey` is the column and mapped to the column is the information
that actually matters. This being the following keys `Code`, `Type` and `Value`.


## Code
This is the actual value that determines the type of data generation that will be done. These codes include RANDOM, STATIC, NULL
etc. By using these values, you can influence how the data for that specific column entry is generated. 
For more information regarding the codes please consult the [Code Guide](#code-guide). The value for this field is
case-insensitive so putting "random" in that field would still work.

## Type
This is the perceived type of the column, in other words, how the program will interpret the column's type.
This key always has an assigned value. It is mostly here as a visual aid for the template user to help them in 
using the template and code guide. Any changes made to it do not matter.

## Value
This is the key-value pair meant to be used in conjunction with Code to help generate data

## Warning
Template's can get fairly large, after all they include EVERY table needed to generate data for the table you actually care about.
As such the golden rule of templates is simple. If you don't need to modify a column's output... don't. 
All parsers have default values in the case that they don't receive a code. As such, only assign codes and values to 
columns whose outputs you NEED to tamper with.

# Code Guide

## Support Map
A quick look at what column types support what codes

### BOOL: RANDOM | STATIC | NULL
### DATE: RANDOM | STATIC | NOW | NULL
### INT: RANDOM | STATIC | SEQ | NULL
### UUID: UUID | NULL
### VARCHAR: STATIC | REGEX | EMAIL | FIRSTNAME | LASTNAME | FULLNAME | PHONE | COUNTRY | ADDRESS | ZIPCODE | CITY | NULL

## Code Type Defaults
### BOOL: RANDOM
### DATE: NOW
### INT: SEQ
### UUID: UUID
### VARCHAR: REGEX, DEFAULT VALUE: "[a-zA-Z]+"


## RANDOM
Generates a random value based on the `Value` key in the template. If no `Value` value is given,  `"Value": ""`, a default will be 
used. This code is supported only by `INT` and `Date` type columns.
The following are examples of what values can be used in conjunction with the RANDOM code "2,6" for INT types and 
"2024-02-23, 2025-01-1" for DATE types.

## STATIC
Generates a static value based on the `Value` key in the template. In the case of `INT` type columns, if no `Value` is
given, `"Value": ""`, the default of "0" is used instead. This code is supported by ALL column types 
(`BOOL`, `DATE`, `INT`, `UUID`, `VARCHAR`).

## SEQ
Generates a sequential integer value starting from 0. For example, if the last entry in that column was 1, this time the 
value will be 2 and so on. This code does not use the `Value` key value. This code is supported by `INT` type columns.

## UUID
Generates a UUID for the column. This code does not use the `Value` key value. This code is supported by `UUID` type
columns.

## REGEX
Generates a string value based on the REGEX given by the `Value` key value. This code is supported by `VARCHAR` type
columns.

## EMAIL
Generates a random email string value using gofakeit for the column. This code does not use the `Value` key value. 
This code is supported by `VARCHAR` type columns.

## FIRSTNAME
Generates a random first name string value using gofakeit for the column. This code does not use the `Value` key value. 
This code is supported by `VARCHAR` type columns.

## LASTNAME
Generates a random last name string value using gofakeit for the column. This code does not use the `Value` key value. 
This code is supported by `VARCHAR` type columns.

## FULLNAME
Generates a random full name string value using gofakeit for the column. This code does not use the `Value` key value. This code is supported
by `VARCHAR` type columns.

## PHONE
Generates a random phone number string value using gofakeit for the column. This code does not use the `Value` key value. 
This code is supported by `VARCHAR` type columns.

## COUNTRY
Generates a random country string value using gofakeit for the column. This code does not use the `Value` key value. 
This code is supported by `VARCHAR` type columns.

## ADDRESS
Generates a random address string value using gofakeit for the column. This code does not use the `Value` key value. 
This code is supported by `VARCHAR` type columns.

## ZIPCODE
Generates a random zipcode string value using gofakeit for the column. This code does not use the `Value` key value. 
This code is supported by `VARCHAR` type columns.

## CITY
Generates a random city string value using gofakeit for the column. This code does not use the `Value` key value. 
This code is supported by `VARCHAR` type columns.

## NULL
Generate a NULL value for the given column. This code does not use the `Value` key value. This code is supported by ALL
column types (`BOOL`, `DATE`, `INT`, `UUID`, `VARCHAR`).

## NOW
Generate a timestamp based on the current time and use it as the value for the column. This code does not use the `Value` key
value. This code is supported by `DATE` type columns.


