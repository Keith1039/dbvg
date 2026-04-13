# DB

## Overview
This package is responsible for obtaining useful schema information from the database. This package also defines which types
the program allows and what in-system type they get mapped to.

**Note:** One of the underlying libraries I'm using does provide schema support however, for the current purposes of the project
I assume that the user hasn't partitioned their database using schemas. As such, the program will always use the current schema.

## Supported Types
While `dbvg` will hopefully support all postgres data types in the future, due to limited time only the following types are supported:
- `INT2`
- `INT4`
- `INT8`
- `NUMERIC`
- `MONEY`
- `FLOAT4`
- `FLOAT8`
- `UUID`
- `VARCHAR`
- `BPCHAR`
- `TEXT`
- `BOOL`
- `DATE`
- `TIME`
- `TIMETZ`
- `TIMESTAMP`
- `TIMESTAMPTZ`

## Type Mapping
Below is how the individual types are mapped in the system. The reason for this is that the restrictions around certain types
are identical. Rather than make repeated code fragments for each column type, grouping them was more efficient. Of course, the nuances of each
need to be handled by the user (i.e. sizes, allowed data, constraints etc.). These mapped types are also what the system uses for `template`
and `Strategy` support.
- `INT2` --> `INT`
- `INT4` --> `INT`
- `INT8` --> `INT`
- `NUMERIC` --> `FLOAT`
- `MONEY` --> `FLOAT`
- `FLOAT4` --> `FLOAT`
- `FLOAT8` --> `FLOAT`
- `UUID` --> `UUID`
- `VARCHAR` --> `VARCHAR`
- `BPCHAR` --> `VARCHAR`
- `TEXT` --> `VARCHAR`
- `BOOL` --> `BOOL`
- `DATE` --> `DATE`
- `TIME` --> `TIME`
- `TIMETZ` --> `TIME`
- `TIMESTAMP` --> `DATE`
- `TIMESTAMPTZ` --> `DATE`

