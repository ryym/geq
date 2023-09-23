package tests

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/ryym/geq"
	"github.com/ryym/geq/internal/tests/d"
	"github.com/ryym/geq/internal/tests/mdl"
)

var benchDBSrc = "root:root@tcp(:3990)/geq?parseTime=true"

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
func clearTransactions(db *sql.DB, b *testing.B) {
	_, err := db.Exec("TRUNCATE TABLE transactions")
	if err != nil {
		b.Fatalf("failed to clear transactions: %v", err)
	}
}

func BenchmarkSql(b *testing.B) {
	db, err := sql.Open("mysql", benchDBSrc)
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	insertTransactions(db)
	defer clearTransactions(db, b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rows, err := db.Query("SELECT * FROM transactions ORDER BY id LIMIT 100")
		if err != nil {
			b.Fatal(err)
		}
		var ts []mdl.Transaction
		for rows.Next() {
			var t mdl.Transaction
			rows.Scan(&t.ID, &t.UserID, &t.Amount, &t.Description, &t.CreatedAt)
			ts = append(ts, t)
		}
		if len(ts) < 100 {
			b.Error("unexpected transaction records")
		}
	}
}

func BenchmarkGeq(b *testing.B) {
	db, err := sql.Open("mysql", benchDBSrc)
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	insertTransactions(db)
	defer clearTransactions(db, b)

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q := geq.SelectFrom(d.Transactions).OrderBy(d.Transactions.ID).Limit(100)
		ts, err := q.Load(ctx, db)
		if err != nil {
			b.Fatal(err)
		}
		if len(ts) < 100 {
			b.Error("unexpected transaction records")
		}
	}
}
