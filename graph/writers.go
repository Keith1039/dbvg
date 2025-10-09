package graph

import (
	"database/sql"
	"fmt"
	"strings"
)

func getColumnQuery(pk string, newPK string, colMap map[string]*sql.ColumnType) string {
	var queryString string
	precision, scale, ok := colMap[pk].DecimalSize() // check to see if it's a variable float type
	databaseType := colMap[pk].DatabaseTypeName()
	// pretty unnecessary but oh well
	if databaseType == "FLOAT4" || databaseType == "FLOAT8" {
		switch databaseType {
		case "FLOAT4":
			queryString = fmt.Sprintf("\n\t %s %s,", newPK, "REAL")
		case "FLOAT8":
			queryString = fmt.Sprintf("\n\t %s %s,", newPK, "DOUBLE PRECISION")
		}
	} else {
		if ok {
			// NUMERIC types default is (65535, 65531) so we avoid that
			if precision != 65535 && scale != 65531 {
				queryString = fmt.Sprintf("\n\t %s %s(%d, %d),", newPK, databaseType, precision, scale) // add the precision to the new query
			} else {
				queryString = fmt.Sprintf("\n\t %s %s,", newPK, databaseType) // solely for NUMERIC
			}
		} else {
			length, isVarchar := colMap[pk].Length() // get the size of the column
			if isVarchar && databaseType != "TEXT" { // text is excluded from this
				if databaseType == "VARCHAR" && length != -5 { // exclude default varchar
					queryString = fmt.Sprintf("\n\t %s %s(%d),", newPK, databaseType, length)
				} else if databaseType == "BPCHAR" {
					if length != -5 && length != 1 { // exclude default BPCHAR and CHAR
						queryString = fmt.Sprintf("\n\t %s %s(%d),", newPK, databaseType, length)
					} else if length == 1 {
						queryString = fmt.Sprintf("\n\t %s %s,", newPK, "CHAR")
					} else {
						queryString = fmt.Sprintf("\n\t %s %s,", newPK, databaseType)
					}
				}
			} else {
				queryString = fmt.Sprintf("\n\t %s %s,", newPK, databaseType) // in case all other conditions fail
			}
		}
	}
	return queryString
}

func appendBuilder(builder *strings.Builder, pks []string, table string, colMap map[string]*sql.ColumnType, newTablePKs []string, slider *int) {
	// appends the new column name and the datatype
	for _, pk := range pks {
		newPK := fmt.Sprintf("%s_%s", table, pk)
		queryString := getColumnQuery(pk, newPK, colMap)
		builder.WriteString(queryString)
		newTablePKs[*slider] = newPK // assign the new pk to the array
		*slider++                    // increment slider
	}
}

func appendDropBuilder(builder *strings.Builder, refTable string, problemTable string, allRelations map[string]map[string]map[string]string) {
	relations := allRelations[problemTable]
	for column, relation := range relations {
		if relation["Table"] == refTable { // check to see if the fk matches
			builder.WriteString(fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s;\n", problemTable, column))
		}
	}
}

func appendColumnBuilder(builder *strings.Builder, relations map[string]map[string]map[string]string, newTablePks []string, refTable string, problemTable string, problemTableKeys []string, refTablePKs []string) {
	// builds the join query that moves existing data over to the new translation table
	var conditionBuilder strings.Builder  // builder for our join condition
	var selectBuilder strings.Builder     // builder for our select statement
	tableRelations := relations[refTable] // map of table relationships
	newTableName := fmt.Sprintf("%s_%s", problemTable, refTable)
	builder.WriteString(fmt.Sprintf("INSERT INTO %s(%s)\n", newTableName, strings.Join(newTablePks, ", ")))
	if problemTable != refTable {
		for _, key := range problemTableKeys {
			if selectBuilder.String() == "" {
				selectBuilder.WriteString(fmt.Sprintf("%s.%s", problemTable, key))
			} else {
				selectBuilder.WriteString(fmt.Sprintf(", %s.%s", problemTable, key))
			}
		}

		for _, key := range refTablePKs {
			selectBuilder.WriteString(fmt.Sprintf(", %s.%s", refTable, key))
		}
		for column, relation := range tableRelations {
			if relation["Table"] == problemTable {
				if conditionBuilder.String() == "" { // check to see if it's the first condition
					conditionBuilder.WriteString(fmt.Sprintf("%s.%s = %s.%s", refTable, column, problemTable, relation["Column"]))
				} else {
					conditionBuilder.WriteString(fmt.Sprintf(" AND %s.%s = %s.%s", refTable, column, problemTable, relation["Column"]))
				}
			}
		}
		// form the SELECT query for INNER-JOIN
		builder.WriteString(fmt.Sprintf("SELECT %s\n", selectBuilder.String()))
		builder.WriteString(fmt.Sprintf("FROM %s\n", refTable))
		builder.WriteString(fmt.Sprintf("INNER JOIN %s\n", problemTable))
		builder.WriteString(fmt.Sprintf("ON %s;", conditionBuilder.String()))
	} else {
		// since it's the same table we can just take from T1 table
		for _, key := range append(problemTableKeys, refTablePKs...) {
			if selectBuilder.String() == "" {
				selectBuilder.WriteString(fmt.Sprintf("T1.%s", key))
			} else {
				selectBuilder.WriteString(fmt.Sprintf(", T1.%s", key))
			}
		}
		for column, relation := range tableRelations {
			if relation["Table"] == problemTable {
				if conditionBuilder.String() == "" { // check to see if it's the first condition
					conditionBuilder.WriteString(fmt.Sprintf("T1.%s = T2.%s", column, relation["Column"]))
				} else {
					conditionBuilder.WriteString(fmt.Sprintf(" AND T1.%s = T2.%s", column, relation["Column"]))
				}
			}
		}
		// form the SELECT query for SELF-JOIN
		builder.WriteString(fmt.Sprintf("SELECT %s\n", selectBuilder.String()))
		builder.WriteString(fmt.Sprintf("FROM %s AS T1, %s AS T2\n", refTable, problemTable))
		builder.WriteString(fmt.Sprintf("WHERE %s;", conditionBuilder.String()))
	}

}

func appendForeignKey(foreignKeyBuilder *strings.Builder, relationships map[string]map[string]map[string]string, problemTable string, problemTableKeys []string, refTable string, refTablePks []string, newTablePks []string) {
	// builds a foreign key constraint
	var refBuilder strings.Builder
	allKeys := strings.Join(newTablePks[0:len(problemTableKeys)], ", ") // all the problem keys as a string
	if problemTable == refTable {                                       // check if it's a self reference
		for _, key := range problemTableKeys { // loop through keys to create the reference
			col := relationships[problemTable][key]["Column"] // the column that the key is referencing
			// format the referenced keys
			if refBuilder.String() == "" {
				refBuilder.WriteString(col)
			} else {
				refBuilder.WriteString(fmt.Sprintf(", %s", col))
			}
		}
	} else {
		// if it isn't a self reference then just join the keys
		refBuilder.WriteString(strings.Join(problemTableKeys, ", ")) // append to ref builder
	}
	// format foreign key for the problem table
	foreignKeyBuilder.WriteString(fmt.Sprintf("\n\tFOREIGN KEY (%s) REFERENCES %s(%s),", allKeys, problemTable, refBuilder.String()))
	allKeys = strings.Join(newTablePks[len(problemTableKeys):], ", ") // format the referenced tables primary keys as a string
	refPks := strings.Join(refTablePks, ", ")                         // string version of the referenced primary keys
	// format foreign key for ref table
	foreignKeyBuilder.WriteString(fmt.Sprintf("\n\tFOREIGN KEY (%s) REFERENCES %s(%s),", allKeys, refTable, refPks))
}

func appendPrimaryKeys(primaryKeyBuilder *strings.Builder, pks []string) {
	// builds the primary key constraint
	allKeys := strings.Join(pks, ", ")                          // convert array to string and add it to builder
	primaryKeyBuilder.WriteString(fmt.Sprintf("(%s)", allKeys)) // format string and add it to builder
}
