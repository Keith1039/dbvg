package parameters

import (
	"container/list"
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
)

type queryBatch struct {
	queryList  *list.List
	paramsList *list.List
}

func (q *queryBatch) init() {
	q.queryList = list.New()
	q.paramsList = list.New()
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
	i := 1
	node := q.queryList.Front()
	paramNode := q.paramsList.Front()
	for node != nil {
		query := node.Value.(string)
		params := paramNode.Value.([]any)
		if verbose {
			fmt.Println(fmt.Sprintf("executing query %d: '%s' with parameters: %v", i, query, arrayAsString(params)))
		}
		_, err := tx.ExecContext(ctx, query, params...)
		if err != nil {
			return err
		}
		i++
		node = node.Next()
		paramNode = paramNode.Next()
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
	return q.queryList.Len()
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
