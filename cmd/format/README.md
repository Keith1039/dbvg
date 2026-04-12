# Templates

Templates are used for data generation in order to provide control over what data is generated.
JSON templates are created and used to store/define certain fields which are referenced when generating data.

The following is a template generated for a given table "products":

```
{
  "products": {
    "price": {
      "code": "",
      "type": "FLOAT",
      "value": nil
    }
    ... 
}
```
the first key, `products`, is the table name. The second key, `price`, is the column name. Mapped to the column name are the fields
used for data generation. These fields are `code`, `type` and `value`.


## code
This is a string field that determines the type of data that will be generated. These codes include `RANDOM`, `STATIC`, `NULL`
etc. By using these values, you can influence how the data for that specific column is generated. 
For more information regarding the codes, please consult the [code guide](#code-guide). The value for this field is
case-insensitive so putting "random" would be the same as putting "RANDOM".

## type
This is a string field representing the perceived type of the column, in other words, how the program will interpret the column's type.
This key's value is assigned during the template's creation. It serves as a reference for the user when using the [code guide](#code-guide). 
This value should not be edited by users. For more information about accepted types as well as how the program interprets column type,
please consult the [`database` package's README](../../db/README.md).

## value
This is the field meant to be used with the [code](#code) field to help generate data. The fields of `code` and `type` are
used to select the `Strategy` that will use this field's value. This field is of type `any` and thus can be anything, 
however it should be noted that the value needed will depend on the linked `Strategy`. 
Please consult the [Code Guide](#code-guide) for more information regarding the accepted data type for the `value` field.

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

**INT**: `RANDOM` | `STATIC` | `SERIAL` | `NULL`

**FLOAT**: `RANDOM` | `STATIC` | `NULL`

**UUID**: `UUID` | `NULL`

**VARCHAR**: `STATIC` | `REGEX` | `EMAIL` | `FIRSTNAME` | `LASTNAME` | `FULLNAME` | `PHONE` | `COUNTRY` | `ADDRESS` | `ZIPCODE` | `CITY` | `NULL`

## Code Type Defaults
**BOOL**: `RANDOM` randomly selects between true and false as value

**DATE**: `NOW` uses the current date

**INT**: `SERIAL` uses an auto incrementing integer for value

**FLOAT**: `RANDOM` with the default range being a random number between 1.0 - 10.0

**UUID**: `UUID`  generates a new UUID

**VARCHAR**: `REGEX` with default value: `[a-zA-Z]{10}` however, this default will scale down to match the column's length if it is lower than 10 characters.
For example, if a column of type `VARCHAR` only allows for 6 characters, the regex will become `[a-zA-Z]{6}`

For more details on `Strategy` please consult the [Strategy README](../../strategy/README.md)
For the definitive list of defined Strategies, please consult the [Existing_Strategies markdown file](../../strategy/Existing_Strategies.md)

