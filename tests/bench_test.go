package tests

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/ryym/geq/tests/b"
	"github.com/ryym/geq/tests/mdl"
)

func insertTransactions(db *sql.DB) (err error) {
	sb := new(strings.Builder)
	sb.WriteString("INSERT INTO transactions (user_id, amount, description) VALUES ")
	for i := 0; i < 1000; i++ {
		if i > 0 {
			sb.WriteRune(',')
		}
		fmt.Fprintf(sb, `(1, %d, "%s-%d")`, i, "desc", i)
	}
	_, err = db.Exec(sb.String())
	return err
}
func clearTransactions(db *sql.DB, tb *testing.B) {
	_, err := db.Exec("TRUNCATE TABLE transactions")
	if err != nil {
		tb.Fatalf("failed to clear transactions: %v", err)
	}
}

func BenchmarkSql(tb *testing.B) {
	db, err := sql.Open("mysql", "root:root@tcp(:3990)/geq")
	if err != nil {
		tb.Fatal(err)
	}
	defer db.Close()

	insertTransactions(db)
	defer clearTransactions(db, tb)

	tb.ResetTimer()
	for i := 0; i < tb.N; i++ {
		rows, err := db.Query("SELECT * FROM transactions ORDER BY id LIMIT 100")
		if err != nil {
			tb.Fatal(err)
		}
		var ts []mdl.Transaction
		for rows.Next() {
			var t mdl.Transaction
			rows.Scan(&t.ID, &t.UserID, &t.Amount, &t.Description)
			ts = append(ts, t)
		}
		if len(ts) < 100 {
			tb.Error("unexpected transaction records")
		}
	}
}

func BenchmarkGeq(tb *testing.B) {
	db, err := sql.Open("mysql", "root:root@tcp(:3990)/geq")
	if err != nil {
		tb.Fatal(err)
	}
	defer db.Close()

	insertTransactions(db)
	defer clearTransactions(db, tb)

	ctx := context.Background()

	tb.ResetTimer()
	for i := 0; i < tb.N; i++ {
		q := b.SelectFrom(b.Transactions).OrderBy(b.Transactions.ID).Limit(100)
		ts, err := q.Load(ctx, db)
		if err != nil {
			tb.Fatal(err)
		}
		if len(ts) < 100 {
			tb.Error("unexpected transaction records")
		}
	}
}
