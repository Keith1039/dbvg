package parameters

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
)

type queryBatch struct {
	size       int
	slider     int
	queryList  []string
	paramsList [][]any
}

func (q *queryBatch) init(size int) {
	q.queryList = make([]string, size)
	q.paramsList = make([][]any, size)
	q.size = size
}

func (q *queryBatch) append(query string, params []any) {
	q.queryList[q.slider] = query
	q.paramsList[q.slider] = params
	q.slider++
}

func (q *queryBatch) reverseAppend(query string, params []any) {
	q.queryList[q.size-q.slider-1] = query
	q.paramsList[q.size-q.slider-1] = params
	q.slider++
}

func (q *queryBatch) executeBatch(ctx context.Context, db *sql.DB, verbose bool) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback() // don't care about error
	err = q.executeBatchAsTransaction(ctx, tx, verbose)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (q *queryBatch) executeBatchAsTransaction(ctx context.Context, tx *sql.Tx, verbose bool) error {
	for i, query := range q.queryList {
		params := q.paramsList[i]
		if verbose {
			fmt.Println(fmt.Sprintf("executing query %d: '%s' with parameters: %v", i+1, query, arrayAsString(params)))
		}
		_, err := tx.ExecContext(ctx, query, params...)
		if err != nil {
			return err
		}
	}
	return nil
}

func (q *queryBatch) Exec(db *sql.DB, verbose bool) error {
	return q.executeBatch(context.Background(), db, verbose)
}

func (q *queryBatch) ExecContext(ctx context.Context, db *sql.DB, verbose bool) error {
	return q.executeBatch(ctx, db, verbose)
}

func (q *queryBatch) ExecTransact(tx *sql.Tx, verbose bool) error {
	return q.executeBatchAsTransaction(context.Background(), tx, verbose)
}

func (q *queryBatch) ExecTransactContext(ctx context.Context, tx *sql.Tx, verbose bool) error {
	return q.executeBatchAsTransaction(ctx, tx, verbose)
}

func (q *queryBatch) Size() int {
	return len(q.queryList)
}

type InsertBatch struct {
	queryBatch
}

type DeleteBatch struct {
	queryBatch
}

func arrayAsString(arr []any) string {
	var builder strings.Builder
	builder.WriteString("[")
	for i, val := range arr {
		if i == len(arr)-1 {
			builder.WriteString(fmt.Sprintf("'%v']", val))
		} else {
			builder.WriteString(fmt.Sprintf("'%v', ", val))
		}
	}
	return builder.String()
}
